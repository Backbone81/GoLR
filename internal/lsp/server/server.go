package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/backbone81/golr/internal/lsp/base"
	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/protocol"
)

// errEndOfSession ends Run after the exit notification.
var errEndOfSession = errors.New("end of session")

// Server is one session of the language server. The lifecycle, the position encoding and the documents are only
// accessed by the reader, a request in its own goroutine gets the documents of the moment it was read.
type Server struct {
	logger          *slog.Logger
	version         string
	parallel        bool
	clientProcessID int
	watcher         *clientProcessWatcher

	lifecycle Lifecycle
	encoding  protocol.PositionEncodingKind
	documents *Documents

	writeMutex sync.Mutex
	writer     io.Writer

	pendingMutex sync.Mutex
	pending      map[string]context.CancelFunc
	requests     sync.WaitGroup
}

// New returns a server for one session. Parallel runs feature requests in their own goroutine. ClientProcessID names
// the client process to watch, 0 watches the one of initialize.
func New(logger *slog.Logger, version string, parallel bool, clientProcessID int) *Server {
	return &Server{
		logger:          logger,
		version:         version,
		parallel:        parallel,
		clientProcessID: clientProcessID,
		encoding:        protocol.PositionEncodingKindUTF16,
		documents:       &Documents{snapshots: map[protocol.DocumentUri]*Snapshot{}},
		pending:         make(map[string]context.CancelFunc),
	}
}

// Run runs the session and watches the client process. It returns the result of the session, or
// ErrClientProcessGone as soon as the client process is gone. The session then goes on in the background until it
// waits for its next message, which a read cannot be interrupted for, and its running requests are cancelled.
func (s *Server) Run(ctx context.Context, reader io.Reader, writer io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.watcher = newClientProcessWatcher(ctx, s.logger)
	if s.clientProcessID != 0 {
		s.watcher.Watch(s.clientProcessID)
	}

	done := make(chan error, 1)
	go func() {
		done <- s.run(ctx, reader, writer)
	}()
	select {
	case err := <-done:
		return err
	case <-s.watcher.Gone():
		return ErrClientProcessGone
	}
}

// run reads and handles messages until the client exits or the input ends, and waits for the running requests before
// it returns. An error means a message could not be handled completely, which ends the session: a skipped message
// could leave the documents out of sync with the client. Ending the session without shutdown is an error, too.
func (s *Server) run(ctx context.Context, reader io.Reader, writer io.Writer) error {
	s.writer = writer
	defer s.requests.Wait()

	protocolParser := base.NewProtocolParser(reader)
	for protocolParser.Next() {
		if err := s.validateHeader(protocolParser.Header()); err != nil {
			return err
		}
		if err := s.dispatchRequestObject(ctx, protocolParser.Content()); err != nil {
			if errors.Is(err, errEndOfSession) {
				return s.ensureShutdown()
			}
			return err
		}
	}
	if err := protocolParser.Err(); err != nil {
		return err
	}
	return s.ensureShutdown()
}

// dispatchRequestObject unmarshals a message and dispatches it to the handler of its method. A message without method
// is a response. The method decides the handler, the id whether the message is answered, as JSON-RPC defines it. The
// rules for every message apply here, the handler checks what only concerns its method. An error ends the session.
//
//nolint:cyclop
func (s *Server) dispatchRequestObject(ctx context.Context, content []byte) error {
	var requestObject jsonrpc.RequestObject
	if err := s.unmarshalRequestObject(content, &requestObject); err != nil {
		return err
	}

	if requestObject.Method == nil {
		if requestObject.ID == nil {
			return errors.New("message has neither method nor id")
		}
		// The server sends no requests yet, so there is no response to expect. A response changes no state.
		s.logger.Error("Response to an unknown request.", "id", requestObject.ID.String())
		return nil
	}

	admitted, err := s.admitRequestObject(&requestObject)
	if !admitted {
		return err
	}
	if requestObject.ID != nil && strings.HasPrefix(*requestObject.Method, "$/") {
		// The specification demands MethodNotFound for a $/ request, the $/ methods are notifications.
		return s.writeMethodNotFound(&requestObject)
	}

	switch *requestObject.Method {
	case protocol.MethodInitialize:
		return runImmediate(ctx, s, &requestObject, protocol.DecodeInitializeParams, s.handleInitialize)
	case protocol.MethodInitialized:
		return runImmediate(ctx, s, &requestObject, protocol.DecodeInitializedParams, s.handleInitialized)
	case protocol.MethodShutdown:
		return runImmediate(ctx, s, &requestObject, s.decodeNoParams, s.handleShutdown)
	case protocol.MethodExit:
		return runImmediate(ctx, s, &requestObject, s.decodeNoParams, s.handleExit)
	case protocol.MethodCancelRequest:
		return runImmediate(ctx, s, &requestObject, protocol.DecodeCancelParams, s.handleCancel)
	case protocol.MethodTextDocumentDidOpen:
		return runImmediate(ctx, s, &requestObject, protocol.DecodeDidOpenTextDocumentParams, s.handleDidOpenTextDocument)
	case protocol.MethodTextDocumentDidChange:
		return runImmediate(ctx, s, &requestObject, protocol.DecodeDidChangeTextDocumentParams, s.handleDidChangeTextDocument)
	case protocol.MethodTextDocumentDidClose:
		return runImmediate(ctx, s, &requestObject, protocol.DecodeDidCloseTextDocumentParams, s.handleDidCloseTextDocument)
	case protocol.MethodTextDocumentFormatting:
		return runAsync(ctx, s, &requestObject, protocol.DecodeDocumentFormattingParams, s.handleDocumentFormatting)
	default:
		if requestObject.ID == nil {
			// The server keeps no state for a notification it does not handle.
			return nil
		}
		return s.writeMethodNotFound(&requestObject)
	}
}

// decodeNoParams decodes the params of a method without params, which are ignored.
func (s *Server) decodeNoParams(any) (struct{}, error) {
	return struct{}{}, nil
}

// writeMethodNotFound answers a request whose method the server does not handle.
func (s *Server) writeMethodNotFound(requestObject *jsonrpc.RequestObject) error {
	return s.writeResponseObject(jsonrpc.ResponseObject{
		JSONRPC: jsonrpc.Version,
		Error: &jsonrpc.ErrorObject{
			Code:    jsonrpc.ErrorCodeMethodNotFound,
			Message: "method not found: " + *requestObject.Method,
		},
		ID: requestObject.ID,
	})
}

// admitRequestObject reports whether a message goes on to its handler, see
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#initialize
// Initialize and exit pass, as they change the Lifecycle. Before initialize and after shutdown, a message with id is
// answered with its error, and one without id is dropped: the specification drops notifications before initialize,
// and after shutdown no request reads the state.
func (s *Server) admitRequestObject(requestObject *jsonrpc.RequestObject) (bool, error) {
	method := *requestObject.Method
	if method == protocol.MethodInitialize || method == protocol.MethodExit || s.lifecycle == LifecycleRunning {
		return true, nil
	}
	if requestObject.ID == nil {
		return false, nil
	}

	errorObject := &jsonrpc.ErrorObject{
		Code:    int(protocol.ErrorCodesServerNotInitialized),
		Message: "server not initialized",
	}
	if s.lifecycle == LifecycleShutDown {
		errorObject = &jsonrpc.ErrorObject{
			Code:    jsonrpc.ErrorCodeInvalidRequest,
			Message: "server shut down",
		}
	}
	return false, s.writeResponseObject(jsonrpc.ResponseObject{
		JSONRPC: jsonrpc.Version,
		Error:   errorObject,
		ID:      requestObject.ID,
	})
}

// ensureShutdown returns an error if the session ended without shutdown.
func (s *Server) ensureShutdown() error {
	if s.lifecycle != LifecycleShutDown {
		return ErrSessionEndedWithoutShutdown
	}
	return nil
}

// unmarshalRequestObject unmarshals the content into a request, a notification or a response, which all fit
// RequestObject.
func (s *Server) unmarshalRequestObject(content []byte, requestObject *jsonrpc.RequestObject) error {
	if err := json.Unmarshal(content, requestObject); err != nil {
		return fmt.Errorf("unmarshaling request object: %w", err)
	}
	if requestObject.JSONRPC != jsonrpc.Version {
		return fmt.Errorf("unsupported JSON-RPC version %q", requestObject.JSONRPC)
	}
	return nil
}

// writeResponseObject writes a complete response object. Writes are serialized, so responses do not interleave. An
// error ends the session: the output is broken, or the response lacks its version or id, which is a bug.
func (s *Server) writeResponseObject(response jsonrpc.ResponseObject) error {
	if response.JSONRPC != jsonrpc.Version {
		return fmt.Errorf("response has JSON-RPC version %q wanted %q", response.JSONRPC, jsonrpc.Version)
	}
	if response.ID == nil {
		return errors.New("response has no id")
	}

	content, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshaling response: %w", err)
	}

	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	if err := base.WriteMessage(s.writer, content); err != nil {
		return fmt.Errorf("writing response: %w", err)
	}
	return nil
}

// validateHeader validates the header fields of a message. It accepts a message without Content-Type or with the
// charset utf-8, and utf8, which the specification accepts for backwards compatibility. Any other charset ends the
// session, as the message could change the documents.
func (s *Server) validateHeader(header map[string]string) error {
	value, ok := header[base.HeaderContentType]
	if !ok {
		return nil
	}

	for parameter := range strings.SplitSeq(value, ";") {
		name, charset, ok := strings.Cut(strings.TrimSpace(parameter), "=")
		if !ok || !strings.EqualFold(strings.TrimSpace(name), "charset") {
			continue
		}
		charset = strings.Trim(strings.TrimSpace(charset), `"`)
		if strings.EqualFold(charset, "utf-8") || strings.EqualFold(charset, "utf8") {
			return nil
		}
		return fmt.Errorf("%w %q in Content-Type", ErrUnsupportedCharset, charset)
	}
	return nil
}

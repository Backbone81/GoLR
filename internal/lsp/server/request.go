package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/protocol"
)

// overloadThreshold is the number of pending requests above which the server warns that it falls behind.
const overloadThreshold = 100

// pendingRequest is a request registered by registerPendingRequest.
type pendingRequest struct {
	message *jsonrpc.RequestObject
	cancel  context.CancelFunc
}

// runImmediate runs the handler of a message before the next message is read, so the methods which change server
// state apply in order. A message with id is answered. One without id is a notification, which gets no response, and
// whose failure ends the session: a lost notification could leave the documents out of sync with the client, which a
// notification has no response to learn about. An error ends the session.
func runImmediate[P, R any](
	ctx context.Context,
	s *Server,
	requestObject *jsonrpc.RequestObject,
	handler func(ctx context.Context, documents *Documents, params *P) (R, error),
) error {
	if requestObject.ID == nil {
		if _, err := callHandler(ctx, s, requestObject, s.documents, handler); err != nil {
			return fmt.Errorf("handling %s: %w", *requestObject.Method, err)
		}
		return nil
	}
	requestCtx, request := s.registerPendingRequest(ctx, requestObject)
	defer s.unregisterPendingRequest(request)
	return processRequest(requestCtx, s, request, s.documents, handler)
}

// runAsync answers a feature request in its own goroutine, so the next message is read while it runs. Without
// parallelism, and for a message without id, which is a notification whose failure can end the session, it runs
// immediately. The goroutine only logs a response it cannot write, the session then ends with the next response the
// reader cannot write or with the end of the input.
func runAsync[P, R any](
	ctx context.Context,
	s *Server,
	requestObject *jsonrpc.RequestObject,
	handler func(ctx context.Context, documents *Documents, params *P) (R, error),
) error {
	if !s.parallel || requestObject.ID == nil {
		return runImmediate(ctx, s, requestObject, handler)
	}

	requestCtx, request := s.registerPendingRequest(ctx, requestObject)
	documents := s.documents
	s.requests.Go(func() {
		defer s.unregisterPendingRequest(request)
		if err := processRequest(requestCtx, s, request, documents, handler); err != nil {
			s.logger.Error("Answering request failed.", "method", *requestObject.Method, "error", err)
		}
	})
	return nil
}

// processRequest runs the handler of a request and writes the response. An error means the response could not be
// written.
func processRequest[P, R any](
	ctx context.Context,
	s *Server,
	request pendingRequest,
	documents *Documents,
	handler func(ctx context.Context, documents *Documents, params *P) (R, error),
) error {
	if ctx.Err() != nil {
		return s.writeResponse(request, nil, ctx.Err())
	}

	result, err := callHandler(ctx, s, request.message, documents, handler)
	if errors.Is(err, errEndOfSession) {
		// Exit ends the session even with an id.
		return err
	}
	return s.writeResponse(request, result, err)
}

// callHandler unmarshals the params and runs the handler. Params which do not unmarshal give an InvalidParams
// Error, a panic gives an error.
//
//nolint:nonamedreturns,ireturn // Named returns are required to report a recovered panic.
func callHandler[P, R any](
	ctx context.Context,
	s *Server,
	message *jsonrpc.RequestObject,
	documents *Documents,
	handler func(ctx context.Context, documents *Documents, params *P) (R, error),
) (result R, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			s.logger.Error("Handler panicked.",
				"method", *message.Method,
				"panic", recovered,
				"stack", string(debug.Stack()),
			)
			err = errors.New("handler panicked")
		}
	}()

	params, err := unmarshalParams[P](message.Params)
	if err != nil {
		return result, NewError(jsonrpc.ErrorCodeInvalidParams, err.Error())
	}
	return handler(ctx, documents, params)
}

// writeResponse writes the response with the result, or with the error: the code of an Error, RequestCancelled
// for a cancellation and InternalError otherwise.
func (s *Server) writeResponse(request pendingRequest, result any, err error) error {
	if err == nil {
		content, marshalErr := json.Marshal(result)
		if marshalErr == nil {
			return s.writeResponseObject(jsonrpc.ResponseObject{
				JSONRPC: jsonrpc.Version,
				Result:  content,
				ID:      request.message.ID,
			})
		}
		err = marshalErr
	}

	var coded *Error
	var errorObject *jsonrpc.ErrorObject
	switch {
	case errors.As(err, &coded):
		errorObject = &jsonrpc.ErrorObject{
			Code:    coded.code,
			Message: coded.message,
		}
	case errors.Is(err, context.Canceled):
		errorObject = &jsonrpc.ErrorObject{
			Code:    int(protocol.LSPErrorCodesRequestCancelled),
			Message: "request cancelled",
		}
	default:
		s.logger.Error("Request failed.", "method", *request.message.Method, "error", err)
		errorObject = &jsonrpc.ErrorObject{
			Code:    jsonrpc.ErrorCodeInternalError,
			Message: err.Error(),
		}
	}
	return s.writeResponseObject(jsonrpc.ResponseObject{
		JSONRPC: jsonrpc.Version,
		Error:   errorObject,
		ID:      request.message.ID,
	})
}

// registerPendingRequest registers a request for $/cancelRequest and returns its context. It runs on the reader before
// the request runs, so a $/cancelRequest read later finds the request. It warns each time the number of pending
// requests rises above overloadThreshold.
func (s *Server) registerPendingRequest(
	ctx context.Context,
	message *jsonrpc.RequestObject,
) (context.Context, pendingRequest) {
	s.pendingMutex.Lock()
	defer s.pendingMutex.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	s.pending[message.ID.String()] = cancel

	if len(s.pending) == overloadThreshold+1 {
		s.logger.Warn("Server falls behind the client.", "pending", len(s.pending))
	}
	return ctx, pendingRequest{
		message: message,
		cancel:  cancel,
	}
}

// unregisterPendingRequest removes a request registered by registerPendingRequest and cancels its context, which
// releases the context's resources.
func (s *Server) unregisterPendingRequest(request pendingRequest) {
	s.pendingMutex.Lock()
	defer s.pendingMutex.Unlock()

	delete(s.pending, request.message.ID.String())
	request.cancel()
}

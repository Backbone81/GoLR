package protocol

// This file holds the subset of the protocol which formatting needs, written by hand. The code generated from the meta
// model replaces it in the future.

import "encoding/json"

const (
	// MethodCancelRequest is defined in the LSP specification at
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#cancelRequest
	MethodCancelRequest = "$/cancelRequest"
	// MethodInitialize is defined in the LSP specification at
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#initialize
	MethodInitialize = "initialize"
	// MethodInitialized is defined in the LSP specification at
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#initialized
	MethodInitialized = "initialized"
	// MethodShutdown is defined in the LSP specification at
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#shutdown
	MethodShutdown = "shutdown"
	// MethodExit is defined in the LSP specification at
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#exit
	MethodExit = "exit"
	// MethodTextDocumentDidOpen is defined in the LSP specification at
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_didOpen
	MethodTextDocumentDidOpen = "textDocument/didOpen"
	// MethodTextDocumentDidChange is defined in the LSP specification at
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_didChange
	MethodTextDocumentDidChange = "textDocument/didChange"
	// MethodTextDocumentDidClose is defined in the LSP specification at
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_didClose
	MethodTextDocumentDidClose = "textDocument/didClose"
	// MethodTextDocumentFormatting is defined in the LSP specification at
	// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocument_formatting
	MethodTextDocumentFormatting = "textDocument/formatting"
)

// ErrorCodes is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#errorCodes
type ErrorCodes int

const ErrorCodesServerNotInitialized ErrorCodes = -32002

// LSPErrorCodes is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#errorCodes
type LSPErrorCodes int

const LSPErrorCodesRequestCancelled LSPErrorCodes = -32800

// CancelParams is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#cancelRequest
type CancelParams struct {
	// Id is an integer or a string, kept raw until the generated union type replaces it.
	Id json.RawMessage `json:"id"`
}

// DocumentUri is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#documentUri
type DocumentUri string

// PositionEncodingKind is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#positionEncodingKind
type PositionEncodingKind string

const (
	PositionEncodingKindUTF8  PositionEncodingKind = "utf-8"
	PositionEncodingKindUTF16 PositionEncodingKind = "utf-16"
)

// TextDocumentSyncKind is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocumentSyncKind
type TextDocumentSyncKind uint32

const TextDocumentSyncKindFull TextDocumentSyncKind = 1

// Position is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#position
type Position struct {
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}

// Range is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#range
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// TextEdit is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textEdit
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// TextDocumentIdentifier is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocumentIdentifier
type TextDocumentIdentifier struct {
	Uri DocumentUri `json:"uri"`
}

// VersionedTextDocumentIdentifier is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#versionedTextDocumentIdentifier
type VersionedTextDocumentIdentifier struct {
	Uri     DocumentUri `json:"uri"`
	Version int32       `json:"version"`
}

// TextDocumentItem is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocumentItem
type TextDocumentItem struct {
	Uri     DocumentUri `json:"uri"`
	Version int32       `json:"version"`
	Text    string      `json:"text"`
}

// GeneralClientCapabilities is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#generalClientCapabilities
type GeneralClientCapabilities struct {
	PositionEncodings []PositionEncodingKind `json:"positionEncodings"`
}

// ClientCapabilities is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#clientCapabilities
type ClientCapabilities struct {
	General GeneralClientCapabilities `json:"general"`
}

// InitializeParams is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#initializeParams
type InitializeParams struct {
	Capabilities ClientCapabilities `json:"capabilities"`
}

// ServerCapabilities is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#serverCapabilities
type ServerCapabilities struct {
	PositionEncoding           PositionEncodingKind `json:"positionEncoding,omitempty"`
	TextDocumentSync           TextDocumentSyncKind `json:"textDocumentSync,omitempty"`
	DocumentFormattingProvider bool                 `json:"documentFormattingProvider,omitempty"`
}

// ServerInfo is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#serverInfo
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// InitializeResult is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#initializeResult
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

// InitializedParams is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#initialized
type InitializedParams struct{}

// DidOpenTextDocumentParams is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#didOpenTextDocumentParams
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentContentChangeEvent covers only the arm of full text synchronization. It is defined in the LSP
// specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#textDocumentContentChangeEvent
type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// DidChangeTextDocumentParams is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#didChangeTextDocumentParams
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// DidCloseTextDocumentParams is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#didCloseTextDocumentParams
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentFormattingParams is defined in the LSP specification at
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification#documentFormattingParams
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

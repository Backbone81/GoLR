package jsonrpc

import "encoding/json"

// RequestObject is a request to the Server.
// See https://www.jsonrpc.org/specification#request_object for details.
type RequestObject struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  *string         `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`

	// Deviating from JSON-RPC, a null id decodes like a missing one, so such a request is treated as a notification.
	// This does not matter for LSP, which allows only integers and strings as request ids.
	ID *ID `json:"id,omitempty"`
}

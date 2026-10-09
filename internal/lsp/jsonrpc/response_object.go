package jsonrpc

import (
	"encoding/json"
)

// Error codes for pre-defined errors.
const (
	ErrorCodeParseError     int = -32700
	ErrorCodeInvalidRequest int = -32600
	ErrorCodeMethodNotFound int = -32601
	ErrorCodeInvalidParams  int = -32602
	ErrorCodeInternalError  int = -32603
)

// ResponseObject is a response of the server.
// See https://www.jsonrpc.org/specification#response_object for details.
type ResponseObject struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *ErrorObject    `json:"error,omitempty"`
	ID      *ID             `json:"id"`
}

// responseObjectJSON is the wire form of ResponseObject, where a nil Result omits the member.
type responseObjectJSON struct {
	JSONRPC string           `json:"jsonrpc"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *ErrorObject     `json:"error,omitempty"`
	ID      *ID              `json:"id"`
}

// MarshalJSON writes result exactly when there is no error. An empty Result is written as null.
func (r ResponseObject) MarshalJSON() ([]byte, error) {
	wire := responseObjectJSON{
		JSONRPC: r.JSONRPC,
		Error:   r.Error,
		ID:      r.ID,
	}
	if r.Error == nil {
		result := r.Result
		if len(result) == 0 {
			result = json.RawMessage("null")
		}
		wire.Result = &result
	}
	return json.Marshal(wire)
}

// ErrorObject is an error reported by the server.
// See https://www.jsonrpc.org/specification#error_object for details.
type ErrorObject struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

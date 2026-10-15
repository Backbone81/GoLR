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

// UnmarshalJSON decodes a request object with case-sensitive member names, as JSON-RPC demands. encoding/json alone
// would also take "Method" for "method". A member with another spelling is ignored like any unknown member.
func (r *RequestObject) UnmarshalJSON(data []byte) error {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(data, &members); err != nil {
		return err
	}

	*r = RequestObject{}
	if value, ok := members["jsonrpc"]; ok {
		if err := json.Unmarshal(value, &r.JSONRPC); err != nil {
			return err
		}
	}
	if value, ok := members["method"]; ok {
		if err := json.Unmarshal(value, &r.Method); err != nil {
			return err
		}
	}
	if value, ok := members["params"]; ok {
		r.Params = value
	}
	if value, ok := members["id"]; ok {
		if err := json.Unmarshal(value, &r.ID); err != nil {
			return err
		}
	}
	return nil
}

package jsonrpc

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// ID is the id of a request, an integer, a string or null.
type ID struct {
	// number is an integer, as LSP restricts numeric ids to integers where JSON-RPC allows any number.
	number *int
	text   *string
}

func (id ID) MarshalJSON() ([]byte, error) {
	if id.number != nil {
		return json.Marshal(id.number)
	}
	if id.text != nil {
		return json.Marshal(*id.text)
	}
	return json.Marshal(nil)
}

func (id *ID) UnmarshalJSON(data []byte) error {
	*id = ID{}
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	if len(data) > 0 && data[0] == '"' {
		return json.Unmarshal(data, &id.text)
	}
	return json.Unmarshal(data, &id.number)
}

func (id ID) String() string {
	if id.number != nil {
		return strconv.Itoa(*id.number)
	}
	if id.text != nil {
		return strconv.Quote(*id.text)
	}
	return "null"
}

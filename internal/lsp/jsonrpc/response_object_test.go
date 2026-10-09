package jsonrpc_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
)

var _ = Describe("ResponseObject", func() {
	var id jsonrpc.ID

	BeforeEach(func() {
		Expect(json.Unmarshal([]byte(`7`), &id)).To(Succeed())
	})

	It("encodes a result", func() {
		response := jsonrpc.ResponseObject{JSONRPC: "2.0", Result: json.RawMessage(`{"a":1}`), ID: &id}
		Expect(json.Marshal(response)).To(MatchJSON(`{"jsonrpc":"2.0","result":{"a":1},"id":7}`))
	})

	It("encodes an empty result as null", func() {
		response := jsonrpc.ResponseObject{JSONRPC: "2.0", ID: &id}
		Expect(json.Marshal(response)).To(MatchJSON(`{"jsonrpc":"2.0","result":null,"id":7}`))
	})

	It("encodes an error without result", func() {
		response := jsonrpc.ResponseObject{
			JSONRPC: "2.0",
			Result:  json.RawMessage(`{"a":1}`),
			Error:   &jsonrpc.ErrorObject{Code: jsonrpc.ErrorCodeMethodNotFound, Message: "method not found"},
			ID:      &id,
		}
		Expect(json.Marshal(response)).
			To(MatchJSON(`{"jsonrpc":"2.0","error":{"code":-32601,"message":"method not found"},"id":7}`))
	})

	It("encodes a missing id as null", func() {
		response := jsonrpc.ResponseObject{
			JSONRPC: "2.0",
			Error:   &jsonrpc.ErrorObject{Code: jsonrpc.ErrorCodeParseError, Message: "parse error"},
		}
		Expect(json.Marshal(response)).
			To(MatchJSON(`{"jsonrpc":"2.0","error":{"code":-32700,"message":"parse error"},"id":null}`))
	})

	It("encodes the same through a pointer", func() {
		response := &jsonrpc.ResponseObject{JSONRPC: "2.0", ID: &id}
		Expect(json.Marshal(response)).To(MatchJSON(`{"jsonrpc":"2.0","result":null,"id":7}`))
	})

	It("encodes error data", func() {
		response := jsonrpc.ResponseObject{
			JSONRPC: "2.0",
			Error: &jsonrpc.ErrorObject{
				Code:    jsonrpc.ErrorCodeInvalidParams,
				Message: "invalid params",
				Data:    json.RawMessage(`{"path":"a.b"}`),
			},
			ID: &id,
		}
		Expect(json.Marshal(response)).To(MatchJSON(
			`{"jsonrpc":"2.0","error":{"code":-32602,"message":"invalid params","data":{"path":"a.b"}},"id":7}`,
		))
	})

	It("tells a null result apart from a missing one when decoding", func() {
		var withNull, withError jsonrpc.ResponseObject
		Expect(json.Unmarshal([]byte(`{"jsonrpc":"2.0","result":null,"id":7}`), &withNull)).To(Succeed())
		Expect(json.Unmarshal(
			[]byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"internal error"},"id":7}`),
			&withError,
		)).To(Succeed())
		Expect(withNull.Result).To(Equal(json.RawMessage("null")))
		Expect(withNull.Error).To(BeNil())
		Expect(withError.Result).To(BeNil())
		Expect(withError.Error).
			To(Equal(&jsonrpc.ErrorObject{Code: jsonrpc.ErrorCodeInternalError, Message: "internal error"}))
		Expect(withError.ID.String()).To(Equal("7"))
	})
})

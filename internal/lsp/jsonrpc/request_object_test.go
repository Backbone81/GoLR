package jsonrpc_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
)

var _ = Describe("RequestObject", func() {
	It("decodes a request", func() {
		var request jsonrpc.RequestObject
		Expect(json.Unmarshal([]byte(`{"jsonrpc":"2.0","id":1,"method":"m","params":{"a":1}}`), &request)).To(Succeed())
		Expect(request.JSONRPC).To(Equal("2.0"))
		Expect(request.Method).To(HaveValue(Equal("m")))
		Expect(request.Params).To(MatchJSON(`{"a":1}`))
		Expect(request.ID).NotTo(BeNil())
		Expect(request.ID.String()).To(Equal("1"))
	})

	It("decodes a notification without id and params", func() {
		var request jsonrpc.RequestObject
		Expect(json.Unmarshal([]byte(`{"jsonrpc":"2.0","method":"m"}`), &request)).To(Succeed())
		Expect(request.ID).To(BeNil())
		Expect(request.Params).To(BeNil())
	})

	It("decodes a null id like a missing one", func() {
		var request jsonrpc.RequestObject
		Expect(json.Unmarshal([]byte(`{"jsonrpc":"2.0","method":"m","id":null}`), &request)).To(Succeed())
		Expect(request.ID).To(BeNil())
	})

	It("tells null params apart from missing params", func() {
		var request jsonrpc.RequestObject
		Expect(json.Unmarshal([]byte(`{"jsonrpc":"2.0","method":"m","params":null}`), &request)).To(Succeed())
		Expect(request.Params).To(Equal(json.RawMessage("null")))
	})

	It("tells a missing method apart from an empty one", func() {
		var missing, empty jsonrpc.RequestObject
		Expect(json.Unmarshal([]byte(`{"jsonrpc":"2.0","id":1}`), &missing)).To(Succeed())
		Expect(json.Unmarshal([]byte(`{"jsonrpc":"2.0","id":1,"method":""}`), &empty)).To(Succeed())
		Expect(missing.Method).To(BeNil())
		Expect(empty.Method).To(HaveValue(BeEmpty()))
	})

	It("encodes a notification without id and params", func() {
		method := "m"
		Expect(json.Marshal(jsonrpc.RequestObject{JSONRPC: "2.0", Method: &method})).
			To(MatchJSON(`{"jsonrpc":"2.0","method":"m"}`))
	})

	It("encodes a request with id and params", func() {
		var id jsonrpc.ID
		Expect(json.Unmarshal([]byte(`"x"`), &id)).To(Succeed())
		method := "m"
		request := jsonrpc.RequestObject{JSONRPC: "2.0", Method: &method, Params: json.RawMessage(`[1]`), ID: &id}
		Expect(json.Marshal(request)).To(MatchJSON(`{"jsonrpc":"2.0","method":"m","params":[1],"id":"x"}`))
	})
})

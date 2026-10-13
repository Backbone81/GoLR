package lsptest

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/protocol"
)

// resultDecoders are the decode functions of the results of the methods the server answers with a result.
var resultDecoders = map[string]func(any) error{
	protocol.MethodInitialize:             decodeResult(protocol.DecodeInitializeResult),
	protocol.MethodTextDocumentFormatting: decodeTextEdits,
}

// ExpectDecodableResults expects the result of every response to decode with the decode function of the method of its
// request, which checks the JSON kinds and literal values. Requests and responses are message contents, contents
// which are no request are skipped.
func ExpectDecodableResults(requests []string, responses []string) {
	methods := map[string]string{}
	for _, content := range requests {
		var request struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if json.Unmarshal([]byte(content), &request) != nil || request.ID == nil || request.Method == "" {
			continue
		}
		methods[string(request.ID)] = request.Method
	}

	for _, content := range responses {
		var response struct {
			ID     json.RawMessage `json:"id"`
			Result json.RawMessage `json:"result"`
		}
		gomega.Expect(json.Unmarshal([]byte(content), &response)).To(gomega.Succeed())
		if len(response.Result) == 0 || string(response.Result) == "null" {
			continue
		}
		method, ok := methods[string(response.ID)]
		gomega.Expect(ok).To(gomega.BeTrue(), "response %s answers no request", content)
		decode, ok := resultDecoders[method]
		gomega.Expect(ok).To(gomega.BeTrue(), "no decode function for the result of %s", method)

		decoder := json.NewDecoder(bytes.NewReader(response.Result))
		decoder.UseNumber()
		var tree any
		gomega.Expect(decoder.Decode(&tree)).To(gomega.Succeed())
		gomega.Expect(decode(tree)).To(gomega.Succeed(), "result of %s: %s", method, content)
	}
}

// decodeResult returns a decode function giving only the error.
func decodeResult[T any](decode func(any) (T, error)) func(any) error {
	return func(value any) error {
		_, err := decode(value)
		return err
	}
}

// decodeTextEdits decodes an array of text edits.
func decodeTextEdits(value any) error {
	edits, ok := value.([]any)
	if !ok {
		return fmt.Errorf("expected array, got %T", value)
	}
	for _, edit := range edits {
		if _, err := protocol.DecodeTextEdit(edit); err != nil {
			return err
		}
	}
	return nil
}

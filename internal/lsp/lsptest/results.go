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
	protocol.MethodTextDocumentDefinition: decodeResult(protocol.DecodeDefinition),
	protocol.MethodTextDocumentFormatting: decodeTextEdits,
}

// ExpectDecodableResults expects the result of every response to decode with the decode function of the method of its
// request, see DecodableResults.
func ExpectDecodableResults(requests []string, responses []string) {
	gomega.Expect(DecodableResults(requests, responses)).To(gomega.Succeed())
}

// DecodableResults checks that the result of every response decodes with the decode function of the method of its
// request, which checks the JSON kinds and literal values. Requests and responses are message contents, contents
// which are no request are skipped.
func DecodableResults(requests []string, responses []string) error {
	// Members are looked up by their exact names, as the server does; encoding/json alone would also take "ID" for
	// "id".
	methods := map[string]string{}
	for _, content := range requests {
		var request map[string]json.RawMessage
		var method string
		if json.Unmarshal([]byte(content), &request) != nil || request["id"] == nil ||
			json.Unmarshal(request["method"], &method) != nil || method == "" {
			continue
		}
		methods[string(request["id"])] = method
	}

	for _, content := range responses {
		var response map[string]json.RawMessage
		if err := json.Unmarshal([]byte(content), &response); err != nil {
			return fmt.Errorf("response %s: %w", content, err)
		}
		result := response["result"]
		if len(result) == 0 || string(result) == "null" {
			continue
		}
		method, ok := methods[string(response["id"])]
		if !ok {
			return fmt.Errorf("response %s answers no request", content)
		}
		decode, ok := resultDecoders[method]
		if !ok {
			return fmt.Errorf("no decode function for the result of %s", method)
		}

		decoder := json.NewDecoder(bytes.NewReader(result))
		decoder.UseNumber()
		var tree any
		if err := decoder.Decode(&tree); err != nil {
			return fmt.Errorf("result of %s: %s: %w", method, content, err)
		}
		if err := decode(tree); err != nil {
			return fmt.Errorf("result of %s: %s: %w", method, content, err)
		}
	}
	return nil
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

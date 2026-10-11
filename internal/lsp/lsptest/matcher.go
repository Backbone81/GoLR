// Package lsptest holds test helpers shared by the tests of the language server packages.
package lsptest

import (
	"encoding/json"
	"errors"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// BeErrorResponse succeeds for the content of an error response with the id and the error code. The message is not
// compared, as it is free text for humans and no part of the contract with the client.
//
//nolint:ireturn
func BeErrorResponse(id int, code int) types.GomegaMatcher {
	return gomega.WithTransform(errorResponseOf, gomega.Equal(errorResponse{ID: id, Code: code}))
}

// errorResponse holds what BeErrorResponse compares.
type errorResponse struct {
	ID   int
	Code int
}

// errorResponseOf returns the id and the error code of an error response.
func errorResponseOf(content string) (errorResponse, error) {
	var response struct {
		ID    int `json:"id"`
		Error *struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return errorResponse{}, err
	}
	if response.Error == nil {
		return errorResponse{}, errors.New("response has no error")
	}
	return errorResponse{ID: response.ID, Code: response.Error.Code}, nil
}

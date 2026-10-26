package provider

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/epilayer-public/epilayer-go"
)

type ErrorResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	Error        *epilayer.Error
}

var ErrResourceInErrorState = errors.New("the resource is in error state")

func generateErrorMessage(verb string, err error) string {
	return fmt.Sprintf("Error during %s: %s", verb, err)
}

func generateClientErrorMessage(verb string, resp ErrorResponse) string {
	if resp.Error != nil {
		return fmt.Sprintf("Error during %s: [%s] %s %s", verb, resp.HTTPResponse.Status, resp.Error.Code, resp.Error.Message)
	} else {
		body := resp.Body
		if len(body) > 200 {
			body = append(body[:200], []byte("... (truncated)")...)
		}
		return fmt.Sprintf("Error during %s: [%s] %s", verb, resp.HTTPResponse.Status, body)
	}
}

func sliceStringify[T ~string](arr []T) []string {
	ret := make([]string, len(arr))
	for i, value := range arr {
		ret[i] = string(value)
	}

	return ret
}

func pointer[T any](v T) *T {
	return &v
}

const (
	regionDescription        = "The legacy region identifier, e.g. `NORD-NO-KRS-1`. The region follows from the provider's `endpoint`; setting it has no effect."
	regionDeprecationMessage = "The region follows from the provider's endpoint. Remove `region` from your configuration; it is ignored and will be removed in the next major version."
)

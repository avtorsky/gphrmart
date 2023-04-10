package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func ReadPayload(
	r *http.Request,
	allowedContentType string,
	payloadReader func(bodyBytes []byte) (any, error),
) (any, error) {
	headerContentType := r.Header.Get("Content-Type")
	if headerContentType != allowedContentType {
		return nil, fmt.Errorf("%w: invalid request content-type", ErrInvalidContentType)
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		return nil, fmt.Errorf("%w: invalid request payload", ErrInvalidPayload)
	}
	body, err := payloadReader(bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid request body %v", ErrInvalidRequest, err.Error())
	}
	validate = validator.New()
	if err := validate.Struct(body); err != nil {
		return nil, fmt.Errorf("%w: invalid request body validation", ErrInvalidPayload)
	}
	return body, nil
}

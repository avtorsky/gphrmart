package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/avtorsky/gphrmart/internal/storage"
)

const sessionContentType = "application/json"

type sessionRequestBody struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Session struct {
	db             storage.Storager
	signingKey     []byte
	expireDuration time.Duration
}

func (h *Session) ReadBody(r *http.Request) (*sessionRequestBody, int, error) {
	payloadReader := func(bodyBytes []byte) (any, error) {
		body := sessionRequestBody{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			return nil, fmt.Errorf("%w: invalid request body deserialization", ErrInvalidRequest)
		}
		return &body, nil
	}
	body, err := ReadPayload(r, sessionContentType, payloadReader)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if bodySerialized, ok := body.(*sessionRequestBody); ok {
		return bodySerialized, http.StatusOK, nil
	} else {
		return nil, http.StatusInternalServerError, nil
	}
}

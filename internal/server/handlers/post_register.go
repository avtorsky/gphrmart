package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/avtorsky/gphrmart/internal/auth"
	"github.com/avtorsky/gphrmart/internal/storage"
)

type Register struct {
	Session
}

func (h *Register) Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, status, err := h.ReadBody(r)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	user, err := h.db.AddUser(ctx, body.Login, body.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := auth.GenerateJWTToken(user, h.signingKey, h.expireDuration)
	tokenSign, err := token.SignedString(h.signingKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer: %v", tokenSign))
	w.WriteHeader(http.StatusOK)
}

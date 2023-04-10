package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/avtorsky/gphrmart/internal/auth"
	"github.com/avtorsky/gphrmart/internal/storage"
)

type Login struct {
	Session
}

func (h *Login) Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, status, err := h.ReadBody(r)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	user, err := h.db.FindUser(ctx, body.Login, body.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if hash := auth.GenerateHash(body.Password, user.Salt); hash != user.HashedPassword {
		err := fmt.Errorf("%v: %v %v", ErrInvalidCredentials, body.Login, body.Password)
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

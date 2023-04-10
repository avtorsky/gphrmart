package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/avtorsky/gphrmart/internal/storage"
)

type Balance struct {
	db storage.Storager
}

func (h *Balance) Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := GetUID(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	balance, err := h.db.GetBalance(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	balanceSerialized, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(balanceSerialized)
}

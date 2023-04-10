package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/avtorsky/gphrmart/internal/storage"
)

type GetOrder struct {
	db storage.Storager
}

func (h *GetOrder) Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := GetUID(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orders, err := h.db.FindOrdersByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrEmptyResult) {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ordersSerialized, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(ordersSerialized)
}

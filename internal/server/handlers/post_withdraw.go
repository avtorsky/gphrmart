package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/avtorsky/gphrmart/internal/storage"
)

const withdrawContentType = "application/json"

type Withdraw struct {
	db storage.Storager
}

type withdrawRequestBody struct {
	OrderID string  `json:"order" validation:"required,luhn_checksum"`
	Sum     float64 `json:"sum" validation:"required"`
}

func (h *Withdraw) ReadBody(r *http.Request) (*withdrawRequestBody, int, error) {
	payloadReader := func(bodyBytes []byte) (any, error) {
		body := withdrawRequestBody{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			return nil, fmt.Errorf("%w: invalid request body deserialization", ErrInvalidRequest)
		}
		return &body, nil
	}
	body, err := ReadPayload(r, withdrawContentType, payloadReader)
	if err != nil {
		if errors.Is(err, ErrInvalidPayload) {
			return nil, http.StatusUnprocessableEntity, err
		}
		return nil, http.StatusBadRequest, err
	}
	if bodySerialized, ok := body.(*withdrawRequestBody); ok {
		return bodySerialized, http.StatusOK, nil
	} else {
		return nil, http.StatusInternalServerError, nil
	}
}

func (h *Withdraw) Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, status, err := h.ReadBody(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), status)
		return
	}
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
	if balance.Current.Float64 < body.Sum {
		http.Error(
			w,
			fmt.Errorf("%w: userID=%v request=%+v", ErrInvalidBalance, userID, body).Error(),
			http.StatusPaymentRequired,
		)
		return
	}
	err = h.db.AddWithdrawalRecord(ctx, body.OrderID, body.Sum, userID)
	if err != nil {
		if errors.Is(err, storage.ErrOrderNotFound) {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

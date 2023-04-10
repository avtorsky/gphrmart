package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/avtorsky/gphrmart/internal/accrual"
	"github.com/avtorsky/gphrmart/internal/storage"
	"github.com/avtorsky/gphrmart/internal/storage/queue"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

const postOrderContentType = "text/plain"

type PostOrder struct {
	ctx           context.Context
	db            storage.Storager
	orderQueue    queue.OrderStatusNotifier
	accrualSystem accrual.Accrualer
	wg            *sync.WaitGroup
}

type postOrderRequestBody struct {
	OrderID string `validate:"required,luhn_checksum"`
}

type accrualSystemResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func NewPostOrder(serverCtx context.Context, db storage.Storager, nWorkers int, accrualSystem accrual.Accrualer) *PostOrder {
	order := PostOrder{
		db:            db,
		orderQueue:    db.Queue(),
		accrualSystem: accrualSystem,
		wg:            new(sync.WaitGroup),
	}
	g, _ := errgroup.WithContext(serverCtx)
	for i := 0; i < nWorkers; i++ {
		order.wg.Add(1)
		g.Go(order.IterateQueue)
	}
	order.ctx = serverCtx
	return &order
}

func (h *PostOrder) IterateQueue() error {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case <-h.ctx.Done():
			log.Println("Waiting for IterateQueue goroutine done...")
			h.wg.Done()
			return ErrServerShutdown
		case <-ticker.C:
			err := h.NotifyQueue()
			if err != nil {
				return err
			}
		}
	}
}

func (h *PostOrder) NotifyQueue() error {
	task, err := h.orderQueue.Acquire(h.ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	res, err := h.accrualSystem.GetOrderInfo(task.OrderID)
	if err != nil {
		if errors.Is(err, accrual.ErrTooManyRequests) {
			err = h.orderQueue.UpdateStatusAndRelease(h.ctx, task.OrderID, task.StatusID)
			if err != nil {
				return err
			}
			time.Sleep(5 * time.Second)
			return nil
		}
		return err
	}
	resp := accrualSystemResponse{}
	json.Unmarshal(res, &resp)
	err = h.db.UpdateOrderStatus(context.Background(), task.OrderID, resp.Status)
	if err != nil {
		return err
	}
	switch {
	case resp.Status == "PROCESSED":
		err = h.db.AddAccrualRecord(context.Background(), task.OrderID, resp.Accrual)
		if err != nil {
			return err
		}
		err = h.orderQueue.Delete(h.ctx, task.OrderID)
		if err != nil {
			return nil
		}
	case resp.Status == "REGISTERED" || resp.Status == "PROCESSING":
		err = h.orderQueue.UpdateStatusAndRelease(h.ctx, task.OrderID, storage.OrderStatusMap[resp.Status])
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *PostOrder) WaitDone() {
	h.wg.Wait()
}

func (h *PostOrder) ReadBody(r *http.Request) (*postOrderRequestBody, int, error) {
	payloadReader := func(bodyBytes []byte) (any, error) {
		body := postOrderRequestBody{OrderID: string(bodyBytes)}
		return &body, nil
	}
	body, err := ReadPayload(r, postOrderContentType, payloadReader)
	if err != nil {
		if errors.Is(err, ErrInvalidPayload) {
			return nil, http.StatusUnprocessableEntity, err
		}
		return nil, http.StatusBadRequest, err
	}
	if bodySerialized, ok := body.(*postOrderRequestBody); ok {
		return bodySerialized, http.StatusOK, nil
	} else {
		return nil, http.StatusInternalServerError, nil
	}
}

func (h *PostOrder) Handler(w http.ResponseWriter, r *http.Request) {
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
	err = h.db.AddOrder(ctx, body.OrderID, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrOrderAlreadyAddedByThisUser):
			http.Error(w, err.Error(), http.StatusOK)
		case errors.Is(err, storage.ErrOrderAlreadyAddedByOtherUser):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	h.orderQueue.Add(ctx, body.OrderID)
	w.WriteHeader(http.StatusAccepted)
}

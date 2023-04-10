package handlers

import (
	"context"
	"log"

	"github.com/avtorsky/gphrmart/internal/accrual"
	"github.com/avtorsky/gphrmart/internal/config"
	"github.com/avtorsky/gphrmart/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
)

type Router struct {
	*chi.Mux
	register    *Register
	login       *Login
	postOrder   *PostOrder
	getOrder    *GetOrder
	balance     *Balance
	withdraw    *Withdraw
	withdrawals *Withdrawals
}

func (r *Router) Shutdown() {
	log.Println("Waiting for router goroutines done...")
	r.postOrder.WaitDone()
}

func NewRouter(serverCtx context.Context, db storage.Storager, cfg *config.Config) Router {
	router := Router{
		Mux: chi.NewRouter(),
	}
	router.register = &Register{
		Session: Session{
			db:             db,
			signingKey:     []byte(cfg.JWTSigningKey),
			expireDuration: cfg.JWTExpireDuration,
		},
	}
	router.login = &Login{
		Session: Session{
			db:             db,
			signingKey:     []byte(cfg.JWTSigningKey),
			expireDuration: cfg.JWTExpireDuration,
		},
	}
	tokenAuth := jwtauth.New("HS256", []byte(cfg.JWTSigningKey), nil)
	accrualService := accrual.NewAccrualService(cfg.AccrualSystemAddress)
	router.postOrder = NewPostOrder(serverCtx, db, 2, accrualService)
	router.getOrder = &GetOrder{db: db}
	router.balance = &Balance{db: db}
	router.withdraw = &Withdraw{db: db}
	router.withdrawals = &Withdrawals{db: db}

	router.Use(middleware.Logger)
	router.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/register", router.register.Handler)
			r.Post("/login", router.login.Handler)
		})
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator)
			r.Post("/orders", router.postOrder.Handler)
			r.Get("/orders", router.getOrder.Handler)
			r.Get("/balance", router.balance.Handler)
			r.Post("/balance/withdraw", router.withdraw.Handler)
			r.Get("/withdrawals", router.withdrawals.Handler)
		})
	})
	return router
}

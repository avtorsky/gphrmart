package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/avtorsky/gphrmart/internal/config"
	"github.com/avtorsky/gphrmart/internal/server/handlers"
	"github.com/avtorsky/gphrmart/internal/storage"
)

func RunServer(cfg *config.Config) {
	shutdownCtx, _ := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	db, err := storage.NewStorageService(shutdownCtx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	r := handlers.NewRouter(shutdownCtx, db, cfg)

	go func() {
		<-shutdownCtx.Done()
		log.Printf("Gracefull shutdown...")
		r.Shutdown()
		db.Close()
		log.Printf("Done...")
		os.Exit(0)
	}()

	log.Printf("Launching server with config %+v\n", cfg)
	log.Fatal(http.ListenAndServe(cfg.RunAddress, r))
}

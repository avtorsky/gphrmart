package main

import (
	"log"
	"os"

	"github.com/avtorsky/gphrmart/internal/config"
	"github.com/avtorsky/gphrmart/internal/server"
	"github.com/joho/godotenv"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	godotenv.Load("../../.env")
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	server.RunServer(cfg)
}

package main

import (
	"os"

	"agg-data-per-shift/internal/core"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	// loads values from .env into the system
	if err := godotenv.Load(".env"); err != nil {
		log.Error("No .env file found")
	}
}

func main() {
	core.InitCore()
}

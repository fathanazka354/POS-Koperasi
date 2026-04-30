package main

import (
	"log"

	"github.com/yourname/pos-koperasi/internal/config"
	"github.com/yourname/pos-koperasi/internal/seed"
)

func main() {
	cfg := config.Load()
	db := config.NewDB(cfg)
	defer db.Close()

	if err := seed.Run(db); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
	log.Println("Seed selesai.")
}

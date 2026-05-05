package main

import (
	"log"

	"github.com/fathanazka354/pos-koperasi/internal/config"
	infraDB "github.com/fathanazka354/pos-koperasi/internal/infra/db"
	"github.com/fathanazka354/pos-koperasi/internal/seed"
)

func main() {
	cfg := config.Load()
	gdb, err := infraDB.OpenGormPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer func() { _ = infraDB.CloseGorm(gdb) }()

	if err := seed.Run(gdb); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
	log.Println("Seed selesai.")
}

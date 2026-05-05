package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/fathanazka354/pos-koperasi/internal/config"
	"github.com/fathanazka354/pos-koperasi/internal/health"
	"github.com/fathanazka354/pos-koperasi/internal/middleware"

	productroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/product"

	addrroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/address"
	authroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/auth"
	chatroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/chat"
	notifroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/notification"
	shoproute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/shop"
	txroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/transaction"
)

//go:embed all:demo
var demoAssets embed.FS

// newHTTPRouter mendaftarkan seluruh routing HTTP (static demo + API v1).
//
// Diletakkan terpisah dari inject agar inject.go fokus pada dependency injection.
func newHTTPRouter(
	cfg *config.Config,
	authR *authroute.Routes,
	txR *txroute.Routes,
	chatR *chatroute.Routes,
	productR *productroute.Routes,
	notifR *notifroute.Routes,
	addrR *addrroute.Routes,
	shopR *shoproute.Routes,
) http.Handler {
	demoFS, err := fs.Sub(demoAssets, "demo")
	if err != nil {
		log.Fatalf("demo static: %v", err)
	}
	demoFileServer := http.FileServer(http.FS(demoFS))

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	r.Get("/healthz", health.Liveness)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/demo/", http.StatusTemporaryRedirect)
	})
	r.Get("/demo", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/demo/", http.StatusTemporaryRedirect)
	})
	r.Handle("/demo/*", http.StripPrefix("/demo", demoFileServer))
	// Backward-compat aliases
	r.Get("/demo/chat", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/demo/", http.StatusTemporaryRedirect)
	})
	r.Get("/demo/shop", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/demo/", http.StatusTemporaryRedirect)
	})

	r.Route("/api/v1", func(r chi.Router) {
		authR.Register(r)
		txR.RegisterPublic(r)
		chatR.Register(r, cfg.JWTSecret)

		// Public shop endpoints (tanpa auth)
		productR.RegisterPublic(r)

		// Notification WS + member routes
		notifR.Register(r, cfg.JWTSecret)

		// Address member routes
		addrR.Register(r, cfg.JWTSecret)

		// Shop checkout + orders
		shopR.Register(r, cfg.JWTSecret)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTSecret))

			txR.RegisterProtected(r)
			productR.RegisterProtected(r)
			productR.RegisterSeller(r)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("supervisor", "admin"))
				productR.RegisterSupervisor(r)
			})
		})
	})

	return r
}


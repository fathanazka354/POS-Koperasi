package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	"github.com/yourname/pos-koperasi/internal/config"
	"github.com/yourname/pos-koperasi/internal/middleware"
	midtransClient "github.com/yourname/pos-koperasi/internal/midtrans"
	authcontract "github.com/yourname/pos-koperasi/internal/modules/auth/contract"
	authcontroller "github.com/yourname/pos-koperasi/internal/modules/auth/controller"
	authrouter "github.com/yourname/pos-koperasi/internal/modules/auth/router"
	authsvc "github.com/yourname/pos-koperasi/internal/modules/auth/service/impl"
	chatcontract "github.com/yourname/pos-koperasi/internal/modules/chat/contract"
	chatcontroller "github.com/yourname/pos-koperasi/internal/modules/chat/controller"
	chatrepo "github.com/yourname/pos-koperasi/internal/modules/chat/repository/impl"
	chatrouter "github.com/yourname/pos-koperasi/internal/modules/chat/router"
	chatsvc "github.com/yourname/pos-koperasi/internal/modules/chat/service/impl"
	chatutil "github.com/yourname/pos-koperasi/internal/modules/chat/utility"
	productcontract "github.com/yourname/pos-koperasi/internal/modules/product/contract"
	productcontroller "github.com/yourname/pos-koperasi/internal/modules/product/controller"
	productrepo "github.com/yourname/pos-koperasi/internal/modules/product/repository/impl"
	productrouter "github.com/yourname/pos-koperasi/internal/modules/product/router"
	productsvc "github.com/yourname/pos-koperasi/internal/modules/product/service/impl"
	txcontract "github.com/yourname/pos-koperasi/internal/modules/transaction/contract"
	txcontroller "github.com/yourname/pos-koperasi/internal/modules/transaction/controller"
	txrepo "github.com/yourname/pos-koperasi/internal/modules/transaction/repository/impl"
	txrouter "github.com/yourname/pos-koperasi/internal/modules/transaction/router"
	txsvc "github.com/yourname/pos-koperasi/internal/modules/transaction/service/impl"
)

func provideDB(lc fx.Lifecycle, cfg *config.Config) (*sqlx.DB, error) {
	db, err := config.OpenDB(cfg)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return db.Close()
		},
	})
	return db, nil
}

func provideTransactionRepository(db *sqlx.DB) txcontract.TransactionRepository {
	return txrepo.New(db)
}

func provideChatRepository(db *sqlx.DB) chatcontract.ChatRepository {
	return chatrepo.New(db)
}

func provideProductRepository(db *sqlx.DB) productcontract.ProductRepository {
	return productrepo.New(db)
}

func provideAuthService(db *sqlx.DB, cfg *config.Config) authcontract.AuthService {
	return authsvc.New(db, cfg.JWTSecret, cfg.JWTExpiry)
}

func provideTransactionService(
	repo txcontract.TransactionRepository,
	mt *midtransClient.Client,
) txcontract.TransactionService {
	return txsvc.New(repo, mt)
}

func provideProductService(repo productcontract.ProductRepository) productcontract.ProductService {
	return productsvc.New(repo)
}

func provideChatService(
	db *sqlx.DB,
	repo chatcontract.ChatRepository,
	hub *chatutil.Hub,
) chatcontract.ChatService {
	return chatsvc.New(db, repo, hub)
}

func newTxController(svc txcontract.TransactionService, cfg *config.Config) *txcontroller.Controller {
	return txcontroller.New(svc, cfg.MidtransServerKey)
}

func newChatController(svc chatcontract.ChatService, cfg *config.Config) *chatcontroller.Controller {
	return chatcontroller.New(svc, cfg.JWTSecret)
}

func newHTTPRouter(
	cfg *config.Config,
	authR *authrouter.Router,
	txR *txrouter.Router,
	chatR *chatrouter.Router,
	productR *productrouter.Router,
	productCtl *productcontroller.Controller,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/demo/chat", http.StatusTemporaryRedirect)
	})
	r.Get("/demo/chat", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(chatDemoHTML)
	})

	r.Route("/api/v1", func(r chi.Router) {
		authR.Register(r)
		txR.RegisterPublic(r)
		chatR.Register(r, cfg.JWTSecret)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTSecret))

			txR.RegisterProtected(r)
			productR.RegisterProtected(r)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("supervisor", "admin"))
				r.Get("/products/low-stock/detail", productCtl.GetLowStock)
			})
		})
	})

	return r
}

func registerHTTPServer(lc fx.Lifecycle, cfg *config.Config, handler http.Handler) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.AppPort),
		Handler: handler,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("Server running on %s [%s]", srv.Addr, cfg.AppEnv)
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTP server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}

func fxModule() fx.Option {
	return fx.Options(
		fx.Provide(
			config.Load,
			provideDB,
			midtransClient.NewClient,
			chatutil.NewHub,
			provideTransactionRepository,
			provideChatRepository,
			provideProductRepository,
			provideAuthService,
			provideTransactionService,
			provideProductService,
			provideChatService,
			authcontroller.New,
			authrouter.New,
			newTxController,
			txrouter.New,
			productcontroller.New,
			productrouter.New,
			newChatController,
			chatrouter.New,
			newHTTPRouter,
		),
		fx.Invoke(registerHTTPServer),
	)
}

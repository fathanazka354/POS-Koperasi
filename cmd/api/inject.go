package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"

	"github.com/yourname/pos-koperasi/internal/config"
	"github.com/yourname/pos-koperasi/internal/middleware"
	midtransClient "github.com/yourname/pos-koperasi/internal/midtrans"

	// Auth
	authcontract "github.com/yourname/pos-koperasi/internal/modules/auth/contract"
	authcontroller "github.com/yourname/pos-koperasi/internal/modules/auth/controller"
	authrouter "github.com/yourname/pos-koperasi/internal/modules/auth/router"
	authsvc "github.com/yourname/pos-koperasi/internal/modules/auth/service/impl"

	// Chat
	chatcontract "github.com/yourname/pos-koperasi/internal/modules/chat/contract"
	chatcontroller "github.com/yourname/pos-koperasi/internal/modules/chat/controller"
	chatrepo "github.com/yourname/pos-koperasi/internal/modules/chat/repository/impl"
	chatrouter "github.com/yourname/pos-koperasi/internal/modules/chat/router"
	chatsvc "github.com/yourname/pos-koperasi/internal/modules/chat/service/impl"
	chatutil "github.com/yourname/pos-koperasi/internal/modules/chat/utility"

	// Product
	productcontract "github.com/yourname/pos-koperasi/internal/modules/product/contract"
	productcontroller "github.com/yourname/pos-koperasi/internal/modules/product/controller"
	productrepo "github.com/yourname/pos-koperasi/internal/modules/product/repository/impl"
	productrouter "github.com/yourname/pos-koperasi/internal/modules/product/router"
	productsvc "github.com/yourname/pos-koperasi/internal/modules/product/service/impl"

	// Transaction
	txcontract "github.com/yourname/pos-koperasi/internal/modules/transaction/contract"
	txcontroller "github.com/yourname/pos-koperasi/internal/modules/transaction/controller"
	txrepo "github.com/yourname/pos-koperasi/internal/modules/transaction/repository/impl"
	txrouter "github.com/yourname/pos-koperasi/internal/modules/transaction/router"
	txsvc "github.com/yourname/pos-koperasi/internal/modules/transaction/service/impl"

	// Notification
	notifcontract "github.com/yourname/pos-koperasi/internal/modules/notification/contract"
	notifcontroller "github.com/yourname/pos-koperasi/internal/modules/notification/controller"
	notifrepo "github.com/yourname/pos-koperasi/internal/modules/notification/repository/impl"
	notifmem "github.com/yourname/pos-koperasi/internal/modules/notification/repository/memory"
	notifrouter "github.com/yourname/pos-koperasi/internal/modules/notification/router"
	notifsvc "github.com/yourname/pos-koperasi/internal/modules/notification/service/impl"
	notifws "github.com/yourname/pos-koperasi/internal/modules/notification/ws"

	// Address
	addrcontract "github.com/yourname/pos-koperasi/internal/modules/address/contract"
	addrcontroller "github.com/yourname/pos-koperasi/internal/modules/address/controller"
	addrrepo "github.com/yourname/pos-koperasi/internal/modules/address/repository/impl"
	addrrouter "github.com/yourname/pos-koperasi/internal/modules/address/router"
	addrsvc "github.com/yourname/pos-koperasi/internal/modules/address/service/impl"

	// Voucher
	vouchercontract "github.com/yourname/pos-koperasi/internal/modules/voucher/contract"
	voucherrepo "github.com/yourname/pos-koperasi/internal/modules/voucher/repository/impl"
	vouchersvc "github.com/yourname/pos-koperasi/internal/modules/voucher/service/impl"

	// Shop
	shopcontract "github.com/yourname/pos-koperasi/internal/modules/shop/contract"
	shopcontroller "github.com/yourname/pos-koperasi/internal/modules/shop/controller"
	shoprouter "github.com/yourname/pos-koperasi/internal/modules/shop/router"
	shopsvc "github.com/yourname/pos-koperasi/internal/modules/shop/service/impl"
)

//go:embed all:demo
var demoAssets embed.FS

func provideDB(lc fx.Lifecycle, cfg *config.Config) (*sqlx.DB, error) {
	db, err := config.OpenDB(cfg)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error { return db.Close() },
	})
	return db, nil
}

func provideMongoDB(lc fx.Lifecycle, cfg *config.Config) *mongo.Database {
	db, err := config.OpenMongo(cfg)
	if err != nil {
		log.Printf("MongoDB tidak tersedia, gunakan in-memory notifications: %v", err)
		return nil
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return db.Client().Disconnect(ctx)
		},
	})
	return db
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

func provideAddressRepository(db *sqlx.DB) addrcontract.AddressRepository {
	return addrrepo.New(db)
}

func provideVoucherRepository(db *sqlx.DB) vouchercontract.VoucherRepository {
	return voucherrepo.New(db)
}

func provideNotificationRepository(mongoDB *mongo.Database) notifcontract.NotificationRepository {
	if mongoDB == nil {
		log.Println("MongoDB nil — menggunakan in-memory notification repository")
		return notifmem.New()
	}
	return notifrepo.New(mongoDB)
}

func provideAuthService(db *sqlx.DB, cfg *config.Config) authcontract.AuthService {
	return authsvc.New(db, cfg.JWTSecret, cfg.JWTExpiry)
}

func provideTransactionService(
	repo txcontract.TransactionRepository,
	mt *midtransClient.Client,
) *txsvc.Service {
	return txsvc.New(repo, mt)
}

func provideTransactionServiceAsContract(svc *txsvc.Service) txcontract.TransactionService {
	return svc
}

func provideProductService(repo productcontract.ProductRepository) productcontract.ProductService {
	return productsvc.New(repo)
}

func provideChatService(
	db *sqlx.DB,
	repo chatcontract.ChatRepository,
	hub *chatutil.Hub,
	presence *chatutil.Presence,
) chatcontract.ChatService {
	return chatsvc.New(db, repo, hub, presence)
}

func provideNotificationService(
	repo notifcontract.NotificationRepository,
	hub *notifws.NotifyHub,
) notifcontract.NotificationService {
	return notifsvc.New(repo, hub)
}

func provideAddressService(repo addrcontract.AddressRepository) addrcontract.AddressService {
	return addrsvc.New(repo)
}

func provideVoucherService(repo vouchercontract.VoucherRepository) vouchercontract.VoucherService {
	return vouchersvc.New(repo)
}

func provideShopService(
	txRepo txcontract.TransactionRepository,
	addrRepo addrcontract.AddressRepository,
	voucherSvc vouchercontract.VoucherService,
	mt *midtransClient.Client,
) *shopsvc.Service {
	return shopsvc.New(txRepo, addrRepo, voucherSvc, mt)
}

// wireNotification menyuntikkan notifyFn ke transaction service dan shop service.
func wireNotification(
	txService *txsvc.Service,
	shopService *shopsvc.Service,
	notifSvc notifcontract.NotificationService,
) {
	txService.WithNotify(notifSvc.Notify)
	shopService.WithNotify(notifSvc.Notify)
}

func newTxController(svc txcontract.TransactionService, cfg *config.Config) *txcontroller.Controller {
	return txcontroller.New(svc, cfg.MidtransServerKey)
}

func newChatController(svc chatcontract.ChatService, cfg *config.Config) *chatcontroller.Controller {
	return chatcontroller.New(svc, cfg.JWTSecret)
}

func newNotifController(svc notifcontract.NotificationService, hub *notifws.NotifyHub, cfg *config.Config) *notifcontroller.Controller {
	return notifcontroller.New(svc, hub, cfg.JWTSecret)
}

func newHTTPRouter(
	cfg *config.Config,
	authR *authrouter.Router,
	txR *txrouter.Router,
	chatR *chatrouter.Router,
	productR *productrouter.Router,
	productCtl *productcontroller.Controller,
	notifR *notifrouter.Router,
	addrR *addrrouter.Router,
	shopR *shoprouter.Router,
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
			log.Printf("Demo App   : http://localhost:%s/demo", cfg.AppPort)
			log.Printf("Webhook    : POST http://localhost:%s/api/v1/midtrans/webhook", cfg.AppPort)
			log.Printf("==> Untuk terima webhook dari Midtrans, jalankan: ngrok start api8080")
			log.Printf("    Lalu set Notification URL di dashboard Midtrans ke:")
			log.Printf("    https://<ngrok-url>/api/v1/midtrans/webhook")
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTP server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error { return srv.Shutdown(ctx) },
	})
}

func fxModule() fx.Option {
	return fx.Options(
		fx.Provide(
			config.Load,
			provideDB,
			provideMongoDB,
			midtransClient.NewClient,
			chatutil.NewHub,
			chatutil.NewPresence,
			notifws.NewNotifyHub,

			// Repositories
			provideTransactionRepository,
			provideChatRepository,
			provideProductRepository,
			provideAddressRepository,
			provideVoucherRepository,
			provideNotificationRepository,

			// Services
			provideAuthService,
			provideTransactionService,
			provideTransactionServiceAsContract,
			provideProductService,
			provideChatService,
			provideNotificationService,
			provideAddressService,
			provideVoucherService,
			provideShopService,
			func(s *shopsvc.Service) shopcontract.ShopService { return s },

			// Controllers
			authcontroller.New,
			newTxController,
			productcontroller.New,
			newChatController,
			newNotifController,
			addrcontroller.New,
			func(svc shopcontract.ShopService) *shopcontroller.Controller { return shopcontroller.New(svc) },

			// Routers
			authrouter.New,
			txrouter.New,
			productrouter.New,
			chatrouter.New,
			notifrouter.New,
			addrrouter.New,
			func(ctl *shopcontroller.Controller) *shoprouter.Router { return shoprouter.New(ctl) },

			newHTTPRouter,
		),
		fx.Invoke(
			registerHTTPServer,
			wireNotification, // hubungkan notifyFn ke transaction service
		),
	)
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/fathanazka354/pos-koperasi/internal/config"
	authhandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/auth"
	chathandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/chat"
	notifhandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/notification"
	producthandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/product"
	authroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/auth"
	addrhandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/address"
	addrroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/address"
	chatroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/chat"
	notifroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/notification"
	productroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/product"
	shophandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/shop"
	shoproute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/shop"
	txhandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/transaction"
	txroute "github.com/fathanazka354/pos-koperasi/internal/delivery/http/route/transaction"
	"github.com/fathanazka354/pos-koperasi/internal/gateway/outbox"
	"github.com/fathanazka354/pos-koperasi/internal/gateway/queue"
	infraDB "github.com/fathanazka354/pos-koperasi/internal/infra/db"
	midtransClient "github.com/fathanazka354/pos-koperasi/internal/midtrans"

	// Auth
	authuc "github.com/fathanazka354/pos-koperasi/internal/usecase/auth"
	authrepo "github.com/fathanazka354/pos-koperasi/internal/repository/auth"

	// Chat
	chatuc "github.com/fathanazka354/pos-koperasi/internal/usecase/chat"
	chatutil "github.com/fathanazka354/pos-koperasi/internal/gateway/chatws/utility"
	chatrepo "github.com/fathanazka354/pos-koperasi/internal/repository/chat"

	// Product
	productuc "github.com/fathanazka354/pos-koperasi/internal/usecase/product"
	productrepo "github.com/fathanazka354/pos-koperasi/internal/repository/product"

	// Transaction
	txuc "github.com/fathanazka354/pos-koperasi/internal/usecase/transaction"
	txrepo "github.com/fathanazka354/pos-koperasi/internal/repository/transaction"

	// Notification
	notifuc "github.com/fathanazka354/pos-koperasi/internal/usecase/notification"
	notifrepo "github.com/fathanazka354/pos-koperasi/internal/repository/notification"
	notifws "github.com/fathanazka354/pos-koperasi/internal/gateway/notifyws"

	// Address
	addruc "github.com/fathanazka354/pos-koperasi/internal/usecase/address"
	addrrepo "github.com/fathanazka354/pos-koperasi/internal/repository/address"

	// Voucher
	voucheruc "github.com/fathanazka354/pos-koperasi/internal/usecase/voucher"
	voucherrepo "github.com/fathanazka354/pos-koperasi/internal/repository/voucher"

	// Shop
	shopuc "github.com/fathanazka354/pos-koperasi/internal/usecase/shop"
)

func provideGormDB(lc fx.Lifecycle, cfg *config.Config) (*gorm.DB, error) {
	g, err := infraDB.OpenGormPostgres(cfg)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			_ = ctx
			return infraDB.CloseGorm(g)
		},
	})
	return g, nil
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

func provideTransactionRepository(db *gorm.DB) txuc.Repository {
	return txrepo.New(db)
}

func provideChatRepository(db *gorm.DB) chatuc.Repository {
	return chatrepo.New(db)
}

func provideProductRepository(db *gorm.DB) productuc.Repository {
	return productrepo.New(db)
}

func provideAddressRepository(db *gorm.DB) addruc.Repository {
	return addrrepo.New(db)
}

func provideVoucherRepository(db *gorm.DB) voucheruc.Repository {
	return voucherrepo.New(db)
}

func provideNotificationRepository(mongoDB *mongo.Database) notifuc.Repository {
	if mongoDB == nil {
		log.Println("MongoDB nil — menggunakan in-memory notification repository")
		return notifrepo.NewMemory()
	}
	return notifrepo.NewMongo(mongoDB)
}

func provideAuthRepository(db *gorm.DB) authuc.Repository {
	return authrepo.New(db)
}

func provideAuthService(repo authuc.Repository, cfg *config.Config) authuc.Usecase {
	return authuc.New(repo, cfg.JWTSecret, cfg.JWTExpiry)
}

func provideTransactionService(
	repo txuc.Repository,
	mt *midtransClient.Client,
	ob outbox.PaymentOutboxStore,
) txuc.Usecase {
	return txuc.New(repo, mt, ob)
}

func providePaymentOutboxStore(
	mongoDB *mongo.Database,
	nq *queue.RedisMemberNotifyQueue,
	notifSvc notifuc.Usecase,
) outbox.PaymentOutboxStore {
	if mongoDB == nil {
		return outbox.NewDirectPaymentOutboxStore(nq, notifSvc)
	}
	return outbox.NewMongoPaymentOutboxStore(mongoDB)
}

func provideProductService(repo productuc.Repository) productuc.Usecase {
	return productuc.New(repo)
}

func provideChatService(
	repo chatuc.Repository,
	hub *chatutil.Hub,
	presence *chatutil.Presence,
) chatuc.Usecase {
	return chatuc.New(repo, hub, presence)
}

func provideNotificationService(
	repo notifuc.Repository,
	hub *notifws.NotifyHub,
) notifuc.Usecase {
	return notifuc.New(repo, hub)
}

func provideMidtransServerKey(cfg *config.Config) string {
	return cfg.MidtransServerKey
}

func provideAddressService(repo addruc.Repository) addruc.Usecase {
	return addruc.New(repo)
}

func provideVoucherService(repo voucheruc.Repository) voucheruc.Usecase {
	return voucheruc.New(repo)
}

func provideShopService(
	txRepo txuc.Repository,
	addrRepo addruc.Repository,
	voucherSvc voucheruc.Usecase,
	mt *midtransClient.Client,
	ob outbox.PaymentOutboxStore,
) shopuc.Usecase {
	return shopuc.New(txRepo, addrRepo, voucherSvc, mt, ob)
}

// registerOutboxPaymentNotifyWorker memublikasikan event outbox MongoDB → notifikasi (Redis atau sinkron).
func registerOutboxPaymentNotifyWorker(
	lc fx.Lifecycle,
	store outbox.PaymentOutboxStore,
	nq *queue.RedisMemberNotifyQueue,
	notifSvc notifuc.Usecase,
) {
	wctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go outbox.RunPaymentNotifyWorker(wctx, store, nq, notifSvc)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
}

// provideRedisClient koneksi opsional untuk antrian webhook; gagal ping → nil (webhook sinkron).
func provideRedisClient(lc fx.Lifecycle, cfg *config.Config) *redis.Client {
	if cfg.RedisAddr == "" {
		return nil
	}
	c := queue.NewRedisFromConfig(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(pingCtx).Err(); err != nil {
		log.Printf("Redis tidak tersedia (%v): webhook Midtrans diproses sinkron", err)
		_ = c.Close()
		return nil
	}
	log.Printf("Redis terhubung — stream webhook %q, notifikasi member %q",
		queue.StreamMidtransWebhooks, queue.StreamMemberNotifications)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error { return c.Close() },
	})
	return c
}

func provideMidtransWebhookQueue(rdb *redis.Client) *queue.RedisMidtransQueue {
	if rdb == nil {
		return nil
	}
	return queue.NewRedisMidtransQueue(rdb)
}

func provideMemberNotifyQueue(rdb *redis.Client) *queue.RedisMemberNotifyQueue {
	if rdb == nil {
		return nil
	}
	return queue.NewRedisMemberNotifyQueue(rdb)
}

// registerMemberNotifyConsumer memproses antrian notifikasi member (Notify → DB + WS).
func registerMemberNotifyConsumer(lc fx.Lifecycle, q *queue.RedisMemberNotifyQueue, notifSvc notifuc.Usecase) {
	if q == nil {
		return
	}
	wctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go q.RunConsumer(wctx, notifSvc)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
}

// registerMidtransWebhookConsumer menjalankan worker yang memanggil HandleMidtransNotification.
func registerMidtransWebhookConsumer(lc fx.Lifecycle, q *queue.RedisMidtransQueue, svc txuc.Usecase) {
	if q == nil {
		return
	}
	wctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go q.RunConsumer(wctx, svc)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
}

func registerHTTPServer(lc fx.Lifecycle, cfg *config.Config, handler http.Handler) {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.AppPort),
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("Server running on %s [%s]", srv.Addr, cfg.AppEnv)
			log.Printf("HTTP timeouts: read_header=%s read=%s write=%s idle=%s shutdown=%s",
				cfg.HTTPReadHeaderTimeout, cfg.HTTPReadTimeout, cfg.HTTPWriteTimeout, cfg.HTTPIdleTimeout, cfg.HTTPShutdownTimeout)
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
		OnStop: func(_ context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("http shutdown: %w", err)
			}
			return nil
		},
	})
}

func fxModule() fx.Option {
	return fx.Options(
		fx.Provide(
			config.Load,
			provideGormDB,
			provideMongoDB,
			midtransClient.NewClient,
			chatutil.NewHub,
			chatutil.NewPresence,
			notifws.NewNotifyHub,

			// Repositories
			provideAuthRepository,
			provideTransactionRepository,
			provideChatRepository,
			provideProductRepository,
			provideAddressRepository,
			provideVoucherRepository,
			provideNotificationRepository,
			providePaymentOutboxStore,

			// Services
			provideAuthService,
			provideTransactionService,
			provideProductService,
			provideChatService,
			provideNotificationService,
			provideMidtransServerKey,
			provideAddressService,
			provideVoucherService,
			provideShopService,

			// Delivery (Auth)
			authhandler.New,
			authroute.New,
			// Delivery (Product)
			producthandler.New,
			productroute.New,
			// Delivery (Address)
			addrhandler.New,
			addrroute.New,
			// Delivery (Shop)
			shophandler.New,
			shoproute.New,
			// Delivery (Transaction)
			txhandler.New,
			txroute.New,
			// Delivery (Chat)
			chathandler.New,
			chatroute.New,
			// Delivery (Notification)
			notifhandler.New,
			notifroute.New,

			// Controllers (legacy modules)
			provideRedisClient,
			provideMidtransWebhookQueue,
			provideMemberNotifyQueue,
			// Routers

			newHTTPRouter,
		),
		fx.Invoke(
			registerHTTPServer,
			registerOutboxPaymentNotifyWorker,
			registerMidtransWebhookConsumer,
			registerMemberNotifyConsumer,
		),
	)
}

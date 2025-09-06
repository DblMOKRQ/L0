package main

import (
	"L0/internal/application"
	"L0/internal/config"
	"L0/internal/messagebroker"
	"L0/internal/redis_client"
	"L0/internal/repository"
	"L0/internal/router"
	"L0/internal/router/handlers"
	"L0/internal/service"
	"L0/pkg/logger"
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
)

func main() {
	// TODO: Тесты написать
	// TODO: Сделать readme файл
	// TODO: Создать пользователя и выдать ему права
	// TODO: Подкорректировать миграции (убрать update)
	// TODO: Желательно сделать чтобы если не нашли order то возвращать 404 вместо 500
	// TODO: убрать debug mode в gin
	// TODO: сделать заполнение кэша при запуске
	// TODO: вынести ттл в конфиг
	ctx := context.Background()
	cfg := config.MustLoad()
	log, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}
	defer log.Sync()
	log.Info("config", zap.Any("cfg", cfg))
	//producer := messagebroker.NewProducer(cfg.Brokers, cfg.Topic, log)
	//if err := producer.SendMessage(context.Background(), "123", models.Order{
	//	OrderUID:    "563feb7b2b84b6test",
	//	TrackNumber: "WBILMTESTTRACK",
	//	Entry:       "WBIL",
	//	Delivery: models.Delivery{
	//		Name:    "Test Testov",
	//		Phone:   "+9720000000",
	//		Zip:     "2639809",
	//		City:    "Kiryat Mozkin",
	//		Address: "Ploshad Mira 15",
	//		Region:  "Kraiot",
	//		Email:   "test@gmail.com",
	//	},
	//	Payment: models.Payment{
	//		Transaction:  "563feb7b2b84b6test",
	//		Currency:     "USD",
	//		Provider:     "wbpay",
	//		Amount:       1817,
	//		PaymentDt:    1637907727,
	//		Bank:         "alpha",
	//		DeliveryCost: 1500,
	//		GoodsTotal:   317,
	//		CustomFee:    0,
	//	},
	//	Items: []models.Item{
	//		{
	//			ChrtID:      9934930,
	//			TrackNumber: "WBILMTESTTRACK",
	//			Price:       453,
	//			Rid:         "ab4219087a764ae0btest",
	//			Name:        "Mascaras",
	//			Sale:        30,
	//			Size:        "0",
	//			TotalPrice:  317,
	//			NmID:        2389212,
	//			Brand:       "Vivienne Sabo",
	//			Status:      202,
	//		},
	//	},
	//	Locale:            "en",
	//	InternalSignature: "",
	//	CustomerID:        "test",
	//	DeliveryService:   "meest",
	//	Shardkey:          "9",
	//	SmID:              99,
	//	DateCreated:       time.Date(2021, 11, 26, 6, 22, 19, 0, time.UTC),
	//	OofShard:          "1",
	//}); err != nil {
	//	log.Fatal("failed to send message", zap.Error(err))
	//}
	//log.Info("send message successfully")

	storage, err := repository.NewStorage(ctx, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode, log)
	if err != nil {
		log.Fatal("failed to initialize storage")
	}
	repo := storage.NewRepository()
	redisClient, err := redisClient.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.DB, log)
	if err != nil {
		log.Fatal("failed to initialize redis_client client")
	}

	consumer := messagebroker.NewConsumer(cfg.Brokers, cfg.Topic, log)

	orderService := service.NewOrderService(consumer, repo, redisClient, log)
	handler := handlers.NewOrderHandlers(orderService)
	rout := router.NewRouter(handler, log)

	//go func() {
	//	for {
	//		time.Sleep(10 * time.Second)
	//		producer.SendMessage(ctx, "123", "12313")
	//	}
	//}()
	//
	//for {
	//	mes, err := orderService.ReadMessage(ctx)
	//	if err != nil {
	//		log.Fatal("failed to read message", zap.Error(err))
	//	}
	//	log.Info("Message from kafka", zap.Any("message", mes))
	//	time.Sleep(time.Second)
	//}
	//producer := messagebroker.NewProducer(cfg.Brokers, cfg.Topic, log)
	//go func() {
	//	for i := 0; i < 10; i++ {
	//		time.Sleep(5 * time.Second)
	//		if err := producer.SendMessage(ctx, "123", i); err != nil {
	//			log.Fatal("failed to send message", zap.Error(err))
	//		}
	//		log.Info("send message successfully")
	//	}
	//
	//}()
	app := application.NewApp(orderService, rout, cfg.Addr, log)
	if err := app.Run(time.Hour); err != nil {
		log.Fatal("failed to initialize application", zap.Error(err))
	}

}

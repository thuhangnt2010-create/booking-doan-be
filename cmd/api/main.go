package main

import (
	"context"
	"log"
	"net/http"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/config"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/db"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/handlers"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/realtime"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/service"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	pgPool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres connect failed: %v", err)
	}
	defer pgPool.Close()

	redisClient, err := db.NewRedisClient(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("redis connect failed: %v", err)
	}
	defer redisClient.Close()

	qrSessionService := &service.QRSessionService{
		QR:      &repository.QRRepository{DB: pgPool},
		Session: &repository.SessionRepository{DB: pgPool},
		Table:   &repository.TableRepository{DB: pgPool},
	}

	menuRepo := &repository.MenuRepository{DB: pgPool}
	sessionRepo := &repository.SessionRepository{DB: pgPool}
	tableRepo := &repository.TableRepository{DB: pgPool}
	orderRepo := &repository.OrderRepository{DB: pgPool}
	hub := realtime.NewHub()
	menuHandler := &handlers.MenuHandler{Repo: menuRepo}

	orderService := &service.OrderService{
		Session: sessionRepo,
		Table:   tableRepo,
		Menu:    menuRepo,
		Order:   orderRepo,
		Hub:     hub,
	}
	orderHandler := &handlers.OrderHandler{Service: orderService, OrderRepo: orderRepo}

	mux := http.NewServeMux()
	mux.Handle("/health", &handlers.HealthHandler{DB: pgPool, Redis: redisClient})
	mux.Handle("/qr/", &handlers.QRHandler{Service: qrSessionService})
	mux.HandleFunc("/menu", menuHandler.List)
	mux.HandleFunc("/menu-items/", menuHandler.Detail)
	mux.Handle("/ws/menu/", &handlers.MenuWSHandler{Hub: hub})
	mux.HandleFunc("/admin/menu-items/", (&handlers.AdminMenuHandler{Repo: menuRepo, Hub: hub}).Update)
	mux.HandleFunc("/orders", orderHandler.Root)
	mux.HandleFunc("/orders/", orderHandler.SubRoute)
	mux.HandleFunc("/order-items/", orderHandler.ItemSubRoute)
	mux.Handle("/ws/orders/branch/", &handlers.OrderBranchWSHandler{Hub: hub})
	mux.Handle("/ws/orders/session/", &handlers.OrderSessionWSHandler{Hub: hub})

	log.Printf("booking-doan-be listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

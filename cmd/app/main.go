package main

import (
	"context"
	"go-chat/internal/config"
	"go-chat/internal/db"
	"go-chat/internal/handlers"
	"go-chat/internal/middleware"
	"go-chat/internal/repositories"
	"go-chat/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	postgresPool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	defer postgresPool.Close()

	userRepositories := repositories.NewUserRepository(postgresPool)
	chatRepository := repositories.NewChatReposytory(postgresPool)
	messageRepository := repositories.NewMessageRepository(postgresPool)

	authService := service.NewAuthService(userRepositories, cfg.JWTSecret, time.Duration(cfg.JWTHours)*time.Hour)
	chatService := service.NewChatService(chatRepository, userRepositories)
	messageService := service.NewMessageService(messageRepository, chatRepository)

	healthHandler := handlers.NewHealthHandler(postgresPool)
	authHandler := handlers.NewAuthHandler(authService)
	chatHandler := handlers.NewChatHandler(chatService)
	messageHandler := handlers.NewMessageHandler(messageService)

	authMiddlevare := middleware.AuthMiddleware(cfg.JWTSecret)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Real-time chat API is running"))
	})

	mux.HandleFunc("/health", healthHandler.HealthCheck)
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	mux.Handle("/api/me", authMiddlevare(http.HandlerFunc(authHandler.GetUser)))
	mux.Handle("/api/chats", authMiddlevare(http.HandlerFunc(chatHandler.GetMyChats)))
	mux.Handle("/api/chats/private", authMiddlevare(http.HandlerFunc(chatHandler.CreatePrivateChat)))
	mux.Handle("/api/chats/group", authMiddlevare(http.HandlerFunc(chatHandler.CreateGroupChat)))
	mux.Handle("POST /api/chats/{chatID}/messages", authMiddlevare(http.HandlerFunc(messageHandler.SendMessage)))
	mux.Handle("GET /api/chats/{chatID}/messages", authMiddlevare(http.HandlerFunc(messageHandler.GetMessanges)))

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Server started on port %s", cfg.HTTPPort)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}

	}()

	shutdownServer(server)

}

func shutdownServer(server *http.Server) {
	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-quit

	log.Println("server shuttdown")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("failed to shutdown server: %v", err)

		return
	}

	log.Printf("server stopped gracefully")
}

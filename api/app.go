package main

import (
	"api/internal/message"
	"api/internal/message/repository"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const defaultRabbitmqUrl = "amqp://guest:guest@localhost:5672/"
const defaultPort = "8080"

func main() {
	log.Println("Starting the server")

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("could not load .env: %v", err)
	}

	rabbitmqUrl := getEnv("RABBITMQ_URL", defaultRabbitmqUrl)
	webPort := ":" + getEnv("PORT", defaultPort)

	messageRepo, err := connectWithRetry(rabbitmqUrl, 10, 3*time.Second)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer messageRepo.Close()

	var router = mux.NewRouter()
	messageService := message.NewDefaultMessageService(messageRepo)
	message.AddMessageHandlers(router, messageService)

	server := &http.Server{
		Addr:              webPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Println("Starting Web Server on port", webPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("web server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func connectWithRetry(url string, attempts int, delay time.Duration) (*repository.RabbitMQMessageRepository, error) {
	var err error
	for i := 1; i <= attempts; i++ {
		var repo *repository.RabbitMQMessageRepository
		repo, err = repository.NewRabbitMQMessageRepository(url)
		if err == nil {
			return repo, nil
		}
		log.Printf("RabbitMQ not ready (attempt %d/%d): %v", i, attempts, err)
		time.Sleep(delay)
	}
	return nil, err
}

package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"taxigo/services/rating-service/internal/events"
	grpcHandler "taxigo/services/rating-service/internal/infrastructure/grpc"
	"taxigo/services/rating-service/internal/infrastructure/repository"
	"taxigo/services/rating-service/internal/service"
	"taxigo/shared/env"
	"taxigo/shared/messaging"
	"taxigo/shared/tracing"

	pb "taxigo/shared/proto/rating"

	"google.golang.org/grpc"
)

var grpcAddr = env.GetString("GRPC_ADDR", ":9005")

func main() {
	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:    "rating-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}

	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("rating-service: failed to initialize tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	repo := repository.NewInMemRatingRepository()
	svc := service.NewRatingService(repo)

	// RabbitMQ
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatalf("rating-service: failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	paymentConsumer := events.NewPaymentConsumer(rabbitmq, svc)
	go paymentConsumer.Listen()

	// gRPC server
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("rating-service: failed to listen on %s: %v", grpcAddr, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRatingServiceServer(grpcServer, grpcHandler.NewGrpcHandler(svc))

	log.Printf("rating-service: gRPC server listening on %s", grpcAddr)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("rating-service: gRPC server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("rating-service: shutting down...")
	grpcServer.GracefulStop()
}

package main

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"time"
)

func main() {
	ctx := context.Background()

	inMemRepository := repository.NewInMemRepository()
	fare := &domain.RideFareModel{
		UserID: "42",
	}

	service := service.NewService(inMemRepository)
	t, err := service.CreateTrip(ctx, fare)
	if err != nil {
		log.Println(err)
	}
	log.Println(t)

	// keep listening (for tilt)
	for {
		time.Sleep(time.Second)
	}
}

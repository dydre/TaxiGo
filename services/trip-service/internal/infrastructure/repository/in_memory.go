package repository

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
)

type inMemRepository struct {
	trips map[string]*domain.TripModel
	fares map[string]*domain.RideFareModel
}

func NewInMemRepository() *inMemRepository {
	return &inMemRepository{
		trips: make(map[string]*domain.TripModel),
		fares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inMemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (r *inMemRepository) CreateFare(ctx context.Context, fare *domain.RideFareModel) (*domain.RideFareModel, error) {
	r.fares[fare.ID.Hex()] = fare
	return fare, nil
}

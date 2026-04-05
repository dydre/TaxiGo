package repository

import (
	"context"
	"fmt"
	"sync"

	"taxigo/services/rating-service/internal/domain"
)

type inMemRatingRepository struct {
	mu      sync.RWMutex
	ratings map[string]*domain.Rating // key: tripID
}

func NewInMemRatingRepository() domain.RatingRepository {
	return &inMemRatingRepository{
		ratings: make(map[string]*domain.Rating),
	}
}

func (r *inMemRatingRepository) SaveRating(ctx context.Context, rating *domain.Rating) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ratings[rating.TripID] = rating
	return nil
}

func (r *inMemRatingRepository) GetRatingByTripID(ctx context.Context, tripID string) (*domain.Rating, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rating, ok := r.ratings[tripID]
	if !ok {
		return nil, fmt.Errorf("rating not found for trip %s", tripID)
	}
	return rating, nil
}

func (r *inMemRatingRepository) GetDriverStats(ctx context.Context, driverID string) (*domain.DriverStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var total int32
	var count int
	for _, rating := range r.ratings {
		if rating.DriverID == driverID {
			total += rating.Score
			count++
		}
	}

	if count == 0 {
		return &domain.DriverStats{
			DriverID:      driverID,
			AverageRating: 0,
			TotalRatings:  0,
		}, nil
	}

	return &domain.DriverStats{
		DriverID:      driverID,
		AverageRating: float64(total) / float64(count),
		TotalRatings:  count,
	}, nil
}

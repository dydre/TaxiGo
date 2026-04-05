package service

import (
	"context"
	"fmt"
	"time"

	"taxigo/services/rating-service/internal/domain"

	"github.com/google/uuid"
)

// eligibleTrips holds trips that completed payment and are eligible for rating.
// In production this would be a DB table.
type eligibleTrip struct {
	UserID   string
	DriverID string
}

type ratingService struct {
	repo          domain.RatingRepository
	eligibleTrips map[string]*eligibleTrip // key: tripID
}

func NewRatingService(repo domain.RatingRepository) domain.RatingService {
	return &ratingService{
		repo:          repo,
		eligibleTrips: make(map[string]*eligibleTrip),
	}
}

// MarkTripEligible is called when a payment succeeds and the trip can be rated.
func (s *ratingService) MarkTripEligible(ctx context.Context, tripID, userID, driverID string) error {
	s.eligibleTrips[tripID] = &eligibleTrip{
		UserID:   userID,
		DriverID: driverID,
	}
	return nil
}

// SubmitRating validates and persists a rating for a completed trip.
func (s *ratingService) SubmitRating(ctx context.Context, tripID, userID, driverID string, score int32, comment string) (*domain.Rating, error) {
	if score < 1 || score > 5 {
		return nil, fmt.Errorf("score must be between 1 and 5, got %d", score)
	}

	eligible, ok := s.eligibleTrips[tripID]
	if !ok {
		return nil, fmt.Errorf("trip %s is not eligible for rating", tripID)
	}
	if eligible.UserID != userID {
		return nil, fmt.Errorf("user %s is not authorized to rate trip %s", userID, tripID)
	}

	// Check duplicate rating
	if _, err := s.repo.GetRatingByTripID(ctx, tripID); err == nil {
		return nil, fmt.Errorf("trip %s has already been rated", tripID)
	}

	rating := &domain.Rating{
		ID:        uuid.New().String(),
		TripID:    tripID,
		UserID:    userID,
		DriverID:  driverID,
		Score:     score,
		Comment:   comment,
		CreatedAt: time.Now(),
	}

	if err := s.repo.SaveRating(ctx, rating); err != nil {
		return nil, fmt.Errorf("failed to save rating: %w", err)
	}

	delete(s.eligibleTrips, tripID)
	return rating, nil
}

// GetDriverStats returns aggregate rating stats for a driver.
func (s *ratingService) GetDriverStats(ctx context.Context, driverID string) (*domain.DriverStats, error) {
	return s.repo.GetDriverStats(ctx, driverID)
}

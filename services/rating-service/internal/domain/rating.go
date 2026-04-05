package domain

import (
	"context"
	"time"
)

type Rating struct {
	ID        string    `json:"id"`
	TripID    string    `json:"tripID"`
	UserID    string    `json:"userID"`
	DriverID  string    `json:"driverID"`
	Score     int32     `json:"score"` // 1-5
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

type DriverStats struct {
	DriverID      string  `json:"driverID"`
	AverageRating float64 `json:"averageRating"`
	TotalRatings  int     `json:"totalRatings"`
}

type RatingRepository interface {
	SaveRating(ctx context.Context, r *Rating) error
	GetRatingByTripID(ctx context.Context, tripID string) (*Rating, error)
	GetDriverStats(ctx context.Context, driverID string) (*DriverStats, error)
}

type RatingService interface {
	SubmitRating(ctx context.Context, tripID, userID, driverID string, score int32, comment string) (*Rating, error)
	GetDriverStats(ctx context.Context, driverID string) (*DriverStats, error)
	MarkTripEligible(ctx context.Context, tripID, userID, driverID string) error
}

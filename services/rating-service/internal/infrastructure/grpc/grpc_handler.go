package grpc

import (
	"context"
	"log"

	"taxigo/services/rating-service/internal/domain"
	pb "taxigo/shared/proto/rating"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcHandler struct {
	pb.UnimplementedRatingServiceServer
	service domain.RatingService
}

func NewGrpcHandler(service domain.RatingService) *GrpcHandler {
	return &GrpcHandler{service: service}
}

func (h *GrpcHandler) SubmitRating(ctx context.Context, req *pb.SubmitRatingRequest) (*pb.SubmitRatingResponse, error) {
	log.Printf("rating-service: SubmitRating tripID=%s score=%d", req.TripID, req.Score)

	rating, err := h.service.SubmitRating(ctx, req.TripID, req.UserID, req.DriverID, req.Score, req.Comment)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to submit rating: %v", err)
	}

	return &pb.SubmitRatingResponse{
		RatingID:  rating.ID,
		TripID:    rating.TripID,
		Score:     rating.Score,
		CreatedAt: rating.CreatedAt.String(),
	}, nil
}

func (h *GrpcHandler) GetDriverStats(ctx context.Context, req *pb.GetDriverStatsRequest) (*pb.GetDriverStatsResponse, error) {
	log.Printf("rating-service: GetDriverStats driverID=%s", req.DriverID)

	stats, err := h.service.GetDriverStats(ctx, req.DriverID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get driver stats: %v", err)
	}

	return &pb.GetDriverStatsResponse{
		DriverID:      stats.DriverID,
		AverageRating: stats.AverageRating,
		TotalRatings:  int32(stats.TotalRatings),
	}, nil
}

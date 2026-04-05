package events

import (
	"context"
	"encoding/json"
	"log"

	"taxigo/services/rating-service/internal/domain"
	"taxigo/shared/contracts"
	"taxigo/shared/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PaymentConsumer listens for payment success events so the rating service
// can mark trips as eligible for rating.
type PaymentConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.RatingService
}

func NewPaymentConsumer(rabbitmq *messaging.RabbitMQ, service domain.RatingService) *PaymentConsumer {
	return &PaymentConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *PaymentConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.RatingTripCompletedQueue, func(ctx context.Context, msg amqp.Delivery) error {
		var message contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			log.Printf("rating-service: failed to unmarshal message: %v", err)
			return err
		}

		switch msg.RoutingKey {
		case contracts.PaymentEventSuccess:
			var payload messaging.PaymentStatusUpdateData
			if err := json.Unmarshal(message.Data, &payload); err != nil {
				log.Printf("rating-service: failed to unmarshal payload: %v", err)
				return err
			}
			return c.handlePaymentSuccess(ctx, payload)
		}

		return nil
	})
}

func (c *PaymentConsumer) handlePaymentSuccess(ctx context.Context, payload messaging.PaymentStatusUpdateData) error {
	log.Printf("rating-service: marking trip %s eligible for rating", payload.TripID)
	return c.service.MarkTripEligible(ctx, payload.TripID, payload.UserID, payload.DriverID)
}

package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type DeliveryConfirmation struct {
	Delivery *amqp.Delivery
	Timer    *prometheus.Timer

	Logger zerolog.Logger
}

func (dc DeliveryConfirmation) Confirm() error {
	err := dc.Delivery.Ack(false)
	if err != nil {
		dc.Logger.Error().
			Err(err).
			Str("Method", "Confirm").
			Str("routing", dc.Delivery.RoutingKey).
			Str("exchange", dc.Delivery.Exchange).
			Msg("Error during Acknowledgment")
		return err
	}
	dc.Logger.Debug().
		Str("Method", "Confirm").
		Str("routing", dc.Delivery.RoutingKey).
		Str("exchange", dc.Delivery.Exchange).
		Msg("Acknowledged")
	dc.Timer.ObserveDuration()
	return nil
}
func (dc DeliveryConfirmation) Reject() error {
	err := dc.Delivery.Reject(false)
	if err != nil {
		dc.Logger.Error().
			Err(err).
			Str("Method", "Reject").
			Str("routing", dc.Delivery.RoutingKey).
			Str("exchange", dc.Delivery.Exchange).
			Msg("Error during Rejection")
		return err
	}
	dc.Logger.Warn().
		Str("Method", "Reject").
		Str("routing", dc.Delivery.RoutingKey).
		Str("exchange", dc.Delivery.Exchange).
		Msg("Rejected")
	dc.Timer.ObserveDuration()
	return nil
}

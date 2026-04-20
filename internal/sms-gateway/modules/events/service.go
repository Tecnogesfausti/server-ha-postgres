package events

import (
	"context"
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/devices"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/sse"
	"github.com/android-sms-gateway/server/internal/sms-gateway/pubsub"
	"go.uber.org/zap"
)

const (
	pubsubTopic   = "events"
	pubsubTimeout = 5 * time.Second
)

type Service struct {
	deviceSvc *devices.Service

	sseSvc  *sse.Service
	pushSvc *push.Service

	pubsub pubsub.PubSub

	metrics *metrics

	logger *zap.Logger
}

func NewService(
	devicesSvc *devices.Service,
	sseSvc *sse.Service,
	pushSvc *push.Service,
	pubsub pubsub.PubSub,
	metrics *metrics,
	logger *zap.Logger,
) *Service {
	return &Service{
		deviceSvc: devicesSvc,
		sseSvc:    sseSvc,
		pushSvc:   pushSvc,

		metrics: metrics,

		pubsub: pubsub,

		logger: logger,
	}
}

func (s *Service) Notify(userID string, deviceID *string, event Event) error {
	notifyStartedAt := time.Now().UTC()

	if event.EventType == "" {
		return fmt.Errorf("%w: event type is empty", ErrValidationFailed)
	}

	subCtx, cancel := context.WithTimeout(context.Background(), pubsubTimeout)
	defer cancel()

	wrapper := eventWrapper{
		UserID:   userID,
		DeviceID: deviceID,
		Event:    event,
	}

	wrapperBytes, err := wrapper.serialize()
	if err != nil {
		s.metrics.IncrementFailed(string(event.EventType), DeliveryTypeUnknown, FailureReasonSerializationError)
		return fmt.Errorf("failed to serialize event wrapper: %w", err)
	}

	if pubErr := s.pubsub.Publish(subCtx, pubsubTopic, wrapperBytes); pubErr != nil {
		s.metrics.IncrementFailed(string(event.EventType), DeliveryTypeUnknown, FailureReasonPublishError)
		return fmt.Errorf("failed to publish event: %w", pubErr)
	}

	s.metrics.IncrementEnqueued(string(event.EventType))

	s.logger.Info(
		"event published for device delivery",
		zap.String("user_id", userID),
		zap.Stringp("device_id", deviceID),
		zap.String("event_type", string(event.EventType)),
		zap.Time("notify_started_at", notifyStartedAt),
		zap.Duration("publish_elapsed", time.Since(notifyStartedAt)),
	)

	return nil
}

func (s *Service) Run(ctx context.Context) error {
	sub, err := s.pubsub.Subscribe(ctx, pubsubTopic)
	if err != nil {
		return fmt.Errorf("failed to subscribe to pubsub: %w", err)
	}
	defer sub.Close()

	ch := sub.Receive()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Event service stopped")
			return nil
		case msg, ok := <-ch:
			if !ok {
				s.logger.Info("Subscription closed")
				return nil
			}
			wrapper := new(eventWrapper)
			if jsonErr := wrapper.deserialize(msg.Data); jsonErr != nil {
				s.metrics.IncrementFailed(EventTypeUnknown, DeliveryTypeUnknown, FailureReasonSerializationError)
				s.logger.Error("failed to deserialize event wrapper", zap.Error(jsonErr))
				continue
			}
			s.processEvent(wrapper)
		}
	}
}

func (s *Service) processEvent(wrapper *eventWrapper) {
	processStartedAt := time.Now().UTC()

	// Load devices from database
	filters := []devices.SelectFilter{}
	if wrapper.DeviceID != nil {
		filters = append(filters, devices.WithID(*wrapper.DeviceID))
	}

	devices, err := s.deviceSvc.Select(wrapper.UserID, filters...)
	if err != nil {
		s.logger.Error("failed to select devices", zap.String("user_id", wrapper.UserID), zap.Error(err))
		return
	}

	if len(devices) == 0 {
		s.logger.Info("no devices found for user", zap.String("user_id", wrapper.UserID))
		return
	}

	// Process each device
	for _, device := range devices {
		if device.PushToken != nil && *device.PushToken != "" {
			pushStartedAt := time.Now().UTC()

			// Device has push token, use push service
			if enqErr := s.pushSvc.Enqueue(*device.PushToken, push.Event{
				Type: wrapper.Event.EventType,
				Data: wrapper.Event.Data,
			}); enqErr != nil {
				s.logger.Error(
					"failed to enqueue push notification",
					zap.String("user_id", wrapper.UserID),
					zap.String("device_id", device.ID),
					zap.String("event_type", string(wrapper.Event.EventType)),
					zap.Time("process_started_at", processStartedAt),
					zap.Time("push_started_at", pushStartedAt),
					zap.Duration("push_elapsed", time.Since(pushStartedAt)),
					zap.Error(enqErr),
				)
				s.metrics.IncrementFailed(
					string(wrapper.Event.EventType),
					DeliveryTypePush,
					FailureReasonProviderFailed,
				)
			} else {
				s.logger.Info(
					"push notification dispatched to provider",
					zap.String("user_id", wrapper.UserID),
					zap.String("device_id", device.ID),
					zap.String("event_type", string(wrapper.Event.EventType)),
					zap.Time("process_started_at", processStartedAt),
					zap.Time("push_started_at", pushStartedAt),
					zap.Duration("process_to_push_elapsed", time.Since(processStartedAt)),
					zap.Duration("push_elapsed", time.Since(pushStartedAt)),
				)
				s.metrics.IncrementSent(string(wrapper.Event.EventType), DeliveryTypePush)
			}
			continue
		}

		// No push token, use SSE service
		sseStartedAt := time.Now().UTC()
		if sseErr := s.sseSvc.Send(device.ID, sse.Event{
			Type: wrapper.Event.EventType,
			Data: wrapper.Event.Data,
		}); sseErr != nil {
			s.logger.Error(
				"failed to send SSE notification",
				zap.String("user_id", wrapper.UserID),
				zap.String("device_id", device.ID),
				zap.String("event_type", string(wrapper.Event.EventType)),
				zap.Time("process_started_at", processStartedAt),
				zap.Time("sse_started_at", sseStartedAt),
				zap.Duration("sse_elapsed", time.Since(sseStartedAt)),
				zap.Error(sseErr),
			)
			s.metrics.IncrementFailed(string(wrapper.Event.EventType), DeliveryTypeSSE, FailureReasonProviderFailed)
		} else {
			s.logger.Info(
				"sse notification delivered",
				zap.String("user_id", wrapper.UserID),
				zap.String("device_id", device.ID),
				zap.String("event_type", string(wrapper.Event.EventType)),
				zap.Time("process_started_at", processStartedAt),
				zap.Time("sse_started_at", sseStartedAt),
				zap.Duration("process_to_sse_elapsed", time.Since(processStartedAt)),
				zap.Duration("sse_elapsed", time.Since(sseStartedAt)),
			)
			s.metrics.IncrementSent(string(wrapper.Event.EventType), DeliveryTypeSSE)
		}
	}
}

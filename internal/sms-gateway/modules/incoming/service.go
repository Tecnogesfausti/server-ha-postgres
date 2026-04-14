package incoming

import (
	"errors"
	"fmt"
	"time"
)

type CreateMessageInput struct {
	ID             string
	DeviceID       string
	Type           MessageType
	Sender         string
	Recipient      *string
	SimNumber      *int
	SubscriptionID *int
	ContentPreview string
	ReceivedAt     time.Time
}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Insert(input CreateMessageInput) error {
	message := newMessage(
		input.ID,
		input.DeviceID,
		input.Type,
		input.Sender,
		input.Recipient,
		input.SimNumber,
		input.SubscriptionID,
		input.ContentPreview,
		input.ReceivedAt,
	)

	if err := s.repository.Insert(message); err != nil {
		if errors.Is(err, ErrMessageAlreadyExists) {
			return nil
		}

		return fmt.Errorf("failed to insert incoming message: %w", err)
	}

	return nil
}

func (s *Service) Select(userID string, filter SelectFilter, options SelectOptions) ([]Message, int64, error) {
	filter.UserID = userID
	return s.repository.Select(filter, options)
}

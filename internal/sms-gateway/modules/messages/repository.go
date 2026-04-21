package messages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/capcom6/go-infra-fx/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxPendingBatch = 100

var ErrMessageNotFound = errors.New("message not found")
var ErrMessageAlreadyExists = errors.New("duplicate id")
var ErrMultipleMessagesFound = errors.New("multiple messages found")

type Repository struct {
	db      *gorm.DB
	dialect db.Dialect
}

func NewRepository(dbConn *gorm.DB, cfg db.Config) *Repository {
	return &Repository{
		db:      dbConn,
		dialect: cfg.Dialect,
	}
}

func (r *Repository) list(filter SelectFilter, options SelectOptions) ([]messageModel, int64, error) {
	query := r.db.Model((*messageModel)(nil))

	// Apply date range filter
	if !filter.StartDate.IsZero() {
		query = query.Where("messages.created_at >= ?", filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		query = query.Where("messages.created_at < ?", filter.EndDate)
	}

	// Apply ID filter
	if filter.ExtID != "" {
		query = query.Where("messages.ext_id = ?", filter.ExtID)
	}

	// Apply user filter
	if filter.UserID != "" {
		query = query.
			Joins("JOIN devices ON messages.device_id = devices.id").
			Where("devices.user_id = ?", filter.UserID)
	}

	// Apply state filter
	if filter.State != "" {
		query = query.Where("messages.state = ?", filter.State)
	}

	// Apply device filter
	if filter.DeviceID != "" {
		query = query.Where("messages.device_id = ?", filter.DeviceID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset > 0 {
		query = query.Offset(options.Offset)
	}

	// Apply ordering
	if options.OrderBy == MessagesOrderFIFO {
		query = query.Order("messages.priority DESC, messages.id ASC")
	} else {
		query = query.Order("messages.priority DESC, messages.id DESC")
	}

	// Preload related data
	if options.WithRecipients {
		query = query.Preload("Recipients")
	}
	if filter.UserID == "" && options.WithDevice {
		query = query.Joins("Device")
	}
	if options.WithStates {
		query = query.Preload("States")
	}

	// Apply content filter
	if !options.WithContent {
		query = query.Omit("Content")
	}

	messages := make([]messageModel, 0, min(options.Limit, int(total)))
	if err := query.Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to select messages: %w", err)
	}

	return messages, total, nil
}

func (r *Repository) listPending(deviceID string, order Order) ([]messageModel, error) {
	messages, _, err := r.list(
		*new(SelectFilter).WithDeviceID(deviceID).WithState(ProcessingStatePending),
		*new(SelectOptions).IncludeContent().IncludeRecipients().WithLimit(maxPendingBatch).WithOrderBy(order),
	)

	return messages, err
}

func (r *Repository) get(filter SelectFilter, options SelectOptions) (messageModel, error) {
	messages, _, err := r.list(filter, options)
	if err != nil {
		return messageModel{}, fmt.Errorf("failed to get message: %w", err)
	}

	if len(messages) == 0 {
		return messageModel{}, ErrMessageNotFound
	}

	if len(messages) > 1 {
		return messageModel{}, ErrMultipleMessagesFound
	}

	return messages[0], nil
}

func (r *Repository) Insert(message *messageModel) error {
	err := r.db.Omit("Device").Create(message).Error
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrMessageAlreadyExists
	}

	return fmt.Errorf("failed to insert message: %w", err)
}

func (r *Repository) UpdateState(message *messageModel) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(message).Select("State").Updates(message).Error; err != nil {
			return err
		}

		for _, v := range message.States {
			v.MessageID = message.ID
			if err := tx.Model(&v).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&v).Error; err != nil {
				return err
			}
		}

		for _, v := range message.Recipients {
			if err := tx.Model((*messageRecipientModel)(nil)).
				Where("message_id = ? AND phone_number = ?", message.ID, v.PhoneNumber).
				Select("state", "error").
				Updates(map[string]any{"state": v.State, "error": v.Error}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update message state: %w", err)
	}

	return nil
}

func (r *Repository) HashProcessed(ctx context.Context, ids []uint64) (int64, error) {
	query := r.db.WithContext(ctx).
		Preload("Recipients").
		Where("is_hashed = ?", false).
		Where("is_encrypted = ?", false).
		Where("state <> ?", ProcessingStatePending)
	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}

	var messages []Message
	if err := query.Find(&messages).Error; err != nil {
		return 0, fmt.Errorf("failed to select messages for hashing: %w", err)
	}

	var rowsAffected int64
	for i := range messages {
		message := &messages[i]
		hashedContent, err := hashMessageContent(*message)
		if err != nil {
			return rowsAffected, fmt.Errorf("failed to hash message %d: %w", message.ID, err)
		}

		err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&Message{}).
				Where("id = ?", message.ID).
				Updates(map[string]any{
					"is_hashed": true,
					"content":   hashedContent,
				}).Error; err != nil {
				return err
			}

			for _, recipient := range message.Recipients {
				hashedPhone := hashString(recipient.PhoneNumber)
				if len(hashedPhone) > 16 {
					hashedPhone = hashedPhone[:16]
				}

				if err := tx.Model(&MessageRecipient{}).
					Where("id = ?", recipient.ID).
					Update("phone_number", hashedPhone).Error; err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return rowsAffected, fmt.Errorf("failed to persist hashes for message %d: %w", message.ID, err)
		}

		rowsAffected++
	}

	return rowsAffected, nil
}

func (r *Repository) Cleanup(ctx context.Context, until time.Time) (int64, error) {
	res := r.db.
		WithContext(ctx).
		Where("state <> ?", ProcessingStatePending).
		Where("created_at < ?", until).
		Delete(new(messageModel))
	return res.RowsAffected, res.Error
}

func hashMessageContent(message Message) (string, error) {
	if content, err := message.GetTextContent(); err != nil {
		return "", err
	} else if content != nil {
		return hashString(content.Text), nil
	}

	if content, err := message.GetDataContent(); err != nil {
		return "", err
	} else if content != nil {
		return hashString(content.Data), nil
	}

	return hashString(message.Content), nil
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

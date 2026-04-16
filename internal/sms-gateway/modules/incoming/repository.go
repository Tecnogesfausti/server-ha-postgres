package incoming

import (
	"errors"
	"fmt"

	"github.com/android-sms-gateway/server/pkg/mysql"
	"gorm.io/gorm"
)

var ErrMessageAlreadyExists = errors.New("duplicate id")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Insert(message *Message) error {
	err := r.db.Omit("Device").Create(message).Error
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) || mysql.IsDuplicateKeyViolation(err) {
		return ErrMessageAlreadyExists
	}

	return fmt.Errorf("failed to insert incoming message: %w", err)
}

func (r *Repository) Select(filter SelectFilter, options SelectOptions) ([]Message, int64, error) {
	query := r.db.Model((*Message)(nil))

	if filter.UserID != "" {
		query = query.
			Joins("JOIN devices ON incoming_messages.device_id = devices.id").
			Where("devices.user_id = ?", filter.UserID)
	}
	if filter.DeviceID != "" {
		query = query.Where("incoming_messages.device_id = ?", filter.DeviceID)
	}
	if filter.Sender != "" {
		query = query.Where("incoming_messages.sender = ?", filter.Sender)
	}
	if filter.ExtID != "" {
		query = query.Where("incoming_messages.ext_id = ?", filter.ExtID)
	}
	if filter.Type != "" {
		query = query.Where("incoming_messages.type = ?", filter.Type)
	}
	if !filter.StartDate.IsZero() {
		query = query.Where("incoming_messages.received_at >= ?", filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		query = query.Where("incoming_messages.received_at < ?", filter.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count incoming messages: %w", err)
	}

	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset > 0 {
		query = query.Offset(options.Offset)
	}

	query = query.Order("incoming_messages.received_at DESC, incoming_messages.id DESC")

	items := make([]Message, 0, min(options.Limit, int(total)))
	if err := query.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to select incoming messages: %w", err)
	}

	return items, total, nil
}

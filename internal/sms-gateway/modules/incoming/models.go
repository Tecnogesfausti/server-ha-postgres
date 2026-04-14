package incoming

import (
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	"gorm.io/gorm"
)

type MessageType string

const (
	MessageTypeSMS           MessageType = "SMS"
	MessageTypeDataSMS       MessageType = "DATA_SMS"
	MessageTypeMMS           MessageType = "MMS"
	MessageTypeMMSDownloaded MessageType = "MMS_DOWNLOADED"
)

type Message struct {
	models.SoftDeletableModel

	ID uint64 `json:"-" gorm:"->;primaryKey;type:BIGINT UNSIGNED;autoIncrement"`

	ExtID string `json:"id" gorm:"not null;type:varchar(64);uniqueIndex:unq_incoming_messages_device_extid,priority:2"`

	DeviceID string `json:"deviceId" gorm:"not null;type:char(21);uniqueIndex:unq_incoming_messages_device_extid,priority:1;index:idx_incoming_messages_device_id"`

	Type MessageType `json:"type" gorm:"not null;type:varchar(32);index:idx_incoming_messages_type"`

	Sender         string  `json:"sender"         gorm:"not null;type:varchar(64)"`
	Recipient      *string `json:"recipient"      gorm:"type:varchar(64)"`
	SimNumber      *int    `json:"simNumber"      gorm:"type:int"`
	SubscriptionID *int    `json:"subscriptionId" gorm:"type:int"`

	ContentPreview string    `json:"contentPreview" gorm:"not null;type:text"`
	ReceivedAt     time.Time `json:"receivedAt"     gorm:"not null;index:idx_incoming_messages_received_at"`

	Device models.Device `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE"`
}

func (Message) TableName() string {
	return "incoming_messages"
}

func newMessage(
	extID string,
	deviceID string,
	messageType MessageType,
	sender string,
	recipient *string,
	simNumber *int,
	subscriptionID *int,
	contentPreview string,
	receivedAt time.Time,
) *Message {
	return &Message{
		ExtID:          extID,
		DeviceID:       deviceID,
		Type:           messageType,
		Sender:         sender,
		Recipient:      recipient,
		SimNumber:      simNumber,
		SubscriptionID: subscriptionID,
		ContentPreview: contentPreview,
		ReceivedAt:     receivedAt,
	}
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(new(Message)); err != nil {
		return fmt.Errorf("incoming messages migration failed: %w", err)
	}

	return nil
}

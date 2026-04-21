package messages

import (
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
)

type TextMessageContent = smsgateway.TextMessage
type DataMessageContent = smsgateway.DataMessage
type HashedMessageContent = smsgateway.HashedMessage

type MessageContent struct {
	TextContent *TextMessageContent `json:"textContent,omitempty"`
	DataContent *DataMessageContent `json:"dataContent,omitempty"`
}

type MessageStateContent struct {
	MessageContent

	HashedContent *HashedMessageContent `json:"hashedContent,omitempty"`
}

type MessageInput struct {
	MessageContent

	ID string

	PhoneNumbers []string
	IsEncrypted  bool

	SimNumber          *uint8
	WithDeliveryReport *bool
	TTL                *uint64
	ValidUntil         *time.Time
	Priority           smsgateway.MessagePriority
}

type Message struct {
	MessageInput

	CreatedAt time.Time
}

type MessageStateInput struct {
	ID         string                      `json:"id"`         // Message ID
	State      ProcessingState             `json:"state"`      // State
	Recipients []smsgateway.RecipientState `json:"recipients"` // Recipients states
	States     map[string]time.Time        `json:"states"`     // History of states

	Message      string              `json:"message"`       // Plain text message when available
	TextMessage  *TextMessageContent `json:"text_message"`  // Plain text payload when available
	DataMessage  *DataMessageContent `json:"data_message"`  // Plain binary payload when available
	PhoneNumbers []string            `json:"phone_numbers"` // Plain recipients when available
}

type MessageState struct {
	MessageStateInput
	MessageStateContent

	DeviceID    string `json:"deviceId"`    // Device ID
	IsHashed    bool   `json:"isHashed"`    // Hashed
	IsEncrypted bool   `json:"isEncrypted"` // Encrypted
}

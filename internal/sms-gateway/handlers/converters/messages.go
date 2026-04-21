package converters

import (
	"time"

	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
)

func MessageToMobileDTO(m messages.Message) smsgateway.MobileMessage {
	var message string
	var textMessage *smsgateway.TextMessage
	var dataMessage *smsgateway.DataMessage

	if m.TextContent != nil {
		message = m.TextContent.Text
		textMessage = &smsgateway.TextMessage{
			Text: m.TextContent.Text,
		}
	} else if m.DataContent != nil {
		dataMessage = &smsgateway.DataMessage{
			Data: m.DataContent.Data,
			Port: m.DataContent.Port,
		}
	}

	return smsgateway.MobileMessage{
		Message: smsgateway.Message{
			ID:       m.ID,
			DeviceID: "",

			Message:     message,
			TextMessage: textMessage,
			DataMessage: dataMessage,

			SimNumber:          m.SimNumber,
			WithDeliveryReport: m.WithDeliveryReport,
			IsEncrypted:        m.IsEncrypted,
			PhoneNumbers:       m.PhoneNumbers,
			TTL:                m.TTL,
			ValidUntil:         m.ValidUntil,
			Priority:           m.Priority,
		},
		CreatedAt: m.CreatedAt,
	}
}

type MessageStateDTO struct {
	ID             string                      `json:"id"`
	DeviceID       string                      `json:"deviceId"`
	State          smsgateway.ProcessingState  `json:"state"`
	IsHashed       bool                        `json:"isHashed"`
	IsEncrypted    bool                        `json:"isEncrypted"`
	Recipients     []smsgateway.RecipientState `json:"recipients"`
	States         map[string]time.Time        `json:"states"`
	ContentPreview string                      `json:"contentPreview,omitempty"`

	Message       string                    `json:"message,omitempty"`
	TextMessage   *smsgateway.TextMessage   `json:"textMessage,omitempty"`
	DataMessage   *smsgateway.DataMessage   `json:"dataMessage,omitempty"`
	HashedMessage *smsgateway.HashedMessage `json:"hashedMessage,omitempty"`
	PhoneNumbers  []string                  `json:"phoneNumbers,omitempty"`
}

func MessageStateToDTO(state messages.MessageState) MessageStateDTO {
	dto := MessageStateDTO{
		ID:          state.ID,
		DeviceID:    state.DeviceID,
		State:       smsgateway.ProcessingState(state.State),
		IsHashed:    state.IsHashed,
		IsEncrypted: state.IsEncrypted,
		Recipients:  state.Recipients,
		States:      state.States,

		TextMessage:   state.TextContent,
		DataMessage:   state.DataContent,
		HashedMessage: state.HashedContent,
	}

	if state.TextContent != nil {
		dto.TextMessage = &smsgateway.TextMessage{Text: state.TextContent.Text}
		dto.Message = state.TextContent.Text
	} else if state.Message != "" {
		dto.Message = state.Message
	}
	if state.DataContent != nil {
		dto.DataMessage = &smsgateway.DataMessage{
			Data: state.DataContent.Data,
			Port: state.DataContent.Port,
		}
	}
	if len(state.PhoneNumbers) > 0 {
		dto.PhoneNumbers = state.PhoneNumbers
	}
	if dto.Message != "" {
		dto.ContentPreview = dto.Message
	} else if dto.DataMessage != nil {
		dto.ContentPreview = "[DATA message]"
	} else if state.IsHashed {
		dto.ContentPreview = "[HASHED]"
	}

	return dto
}

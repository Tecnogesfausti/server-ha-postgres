package incoming

import (
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/deviceauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/models"
	incomingmodule "github.com/android-sms-gateway/server/internal/sms-gateway/modules/incoming"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type mobileControllerParams struct {
	fx.In

	IncomingSvc *incomingmodule.Service

	Validator *validator.Validate
	Logger    *zap.Logger
}

type MobileController struct {
	base.Handler

	incomingSvc *incomingmodule.Service
}

type mobilePostRequest struct {
	ID             string                     `json:"id" validate:"required"`
	Type           incomingmodule.MessageType `json:"type" validate:"required"`
	Sender         string                     `json:"sender" validate:"required"`
	Recipient      *string                    `json:"recipient"`
	SimNumber      *int                       `json:"simNumber"`
	SubscriptionID *int                       `json:"subscriptionId"`
	ContentPreview string                     `json:"contentPreview" validate:"required"`
	ReceivedAtMs   int64                      `json:"receivedAtMs" validate:"required"`
}

func NewMobileController(params mobileControllerParams) *MobileController {
	return &MobileController{
		Handler: base.Handler{
			Logger:    params.Logger,
			Validator: params.Validator,
		},
		incomingSvc: params.IncomingSvc,
	}
}

func (h *MobileController) post(device models.Device, c *fiber.Ctx) error {
	req := new(mobilePostRequest)
	if err := h.BodyParserValidator(c, req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err := h.incomingSvc.Insert(incomingmodule.CreateMessageInput{
		ID:             req.ID,
		DeviceID:       device.ID,
		Type:           req.Type,
		Sender:         req.Sender,
		Recipient:      req.Recipient,
		SimNumber:      req.SimNumber,
		SubscriptionID: req.SubscriptionID,
		ContentPreview: req.ContentPreview,
		ReceivedAt:     time.UnixMilli(req.ReceivedAtMs).UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to persist incoming message: %w", err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *MobileController) Register(router fiber.Router) {
	router.Post("", deviceauth.WithDevice(h.post))
}

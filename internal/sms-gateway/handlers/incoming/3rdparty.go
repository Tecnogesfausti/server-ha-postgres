package incoming

import (
	"fmt"
	"strconv"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/permissions"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	incomingmodule "github.com/android-sms-gateway/server/internal/sms-gateway/modules/incoming"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type thirdPartyGetQueryParams struct {
	From     *time.Time `query:"from"`
	To       *time.Time `query:"to"`
	DeviceID string     `query:"deviceId" validate:"omitempty,min=21,max=21"`
	Sender   string     `query:"sender" validate:"omitempty,max=64"`
	Limit    int        `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset   int        `query:"offset" validate:"omitempty,min=0"`
}

type ThirdPartyController struct {
	base.Handler

	incomingSvc *incomingmodule.Service
}

type messageDTO struct {
	ID             string                      `json:"id"`
	DeviceID       string                      `json:"deviceId"`
	Type           incomingmodule.MessageType  `json:"type"`
	Sender         string                      `json:"sender"`
	Recipient      *string                     `json:"recipient"`
	SimNumber      *int                        `json:"simNumber"`
	SubscriptionID *int                        `json:"subscriptionId"`
	ContentPreview string                      `json:"contentPreview"`
	ReceivedAt     time.Time                   `json:"receivedAt"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
}

func NewThirdPartyController(
	incomingSvc *incomingmodule.Service,
	logger *zap.Logger,
	validator *validator.Validate,
) *ThirdPartyController {
	return &ThirdPartyController{
		Handler: base.Handler{
			Logger:    logger,
			Validator: validator,
		},
		incomingSvc: incomingSvc,
	}
}

func (h *ThirdPartyController) get(userID string, c *fiber.Ctx) error {
	params := new(thirdPartyGetQueryParams)
	if err := h.QueryParserValidator(c, params); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	filter := incomingmodule.SelectFilter{
		DeviceID: params.DeviceID,
		Sender:   params.Sender,
	}
	if params.From != nil || params.To != nil {
		var from time.Time
		var to time.Time
		if params.From != nil {
			from = *params.From
		}
		if params.To != nil {
			to = *params.To
		}
		filter.WithDateRange(from, to)
	}

	options := incomingmodule.SelectOptions{
		Limit:  50,
		Offset: 0,
	}
	if params.Limit > 0 {
		options.Limit = params.Limit
	}
	if params.Offset > 0 {
		options.Offset = params.Offset
	}

	items, total, err := h.incomingSvc.Select(userID, filter, options)
	if err != nil {
		return fmt.Errorf("failed to select incoming messages: %w", err)
	}

	c.Set("X-Total-Count", strconv.Itoa(int(total)))
	response := make([]messageDTO, 0, len(items))
	for _, item := range items {
		response = append(response, messageDTO{
			ID:             item.ExtID,
			DeviceID:       item.DeviceID,
			Type:           item.Type,
			Sender:         item.Sender,
			Recipient:      item.Recipient,
			SimNumber:      item.SimNumber,
			SubscriptionID: item.SubscriptionID,
			ContentPreview: item.ContentPreview,
			ReceivedAt:     item.ReceivedAt,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}

	return c.JSON(response)
}

func (h *ThirdPartyController) Register(router fiber.Router) {
	router.Get("", permissions.RequireScope(ScopeList), userauth.WithUserID(h.get))
}

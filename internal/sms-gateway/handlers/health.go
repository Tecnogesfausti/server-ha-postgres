package handlers

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/base"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/android-sms-gateway/server/internal/version"
	"github.com/android-sms-gateway/server/pkg/health"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type HealthHandler struct {
	base.Handler

	healthSvc               *health.Service
	messagesHashingInterval int
}

func NewHealthHandler(
	healthSvc *health.Service,
	messagesCfg messages.Config,
	logger *zap.Logger,
) *HealthHandler {
	return &HealthHandler{
		Handler: base.Handler{
			Logger:    logger,
			Validator: nil,
		},
		healthSvc:               healthSvc,
		messagesHashingInterval: int(messagesCfg.HashingInterval.Seconds()),
	}
}

//	@Summary		Liveness probe
//	@Description	Checks if service is running (liveness probe)
//	@Tags			System
//	@Produce		json
//	@Success		200	{object}	smsgateway.HealthResponse	"Service is alive"
//	@Failure		503	{object}	smsgateway.HealthResponse	"Service is not alive"
//	@Router			/health/live [get]
//
// Liveness probe.
func (h *HealthHandler) getLiveness(c *fiber.Ctx) error {
	return h.writeProbe(c, h.healthSvc.CheckLiveness(c.Context()))
}

//	@Summary		Readiness probe
//	@Description	Checks if service is ready to serve traffic (readiness probe)
//	@Tags			System
//	@Produce		json
//	@Success		200	{object}	smsgateway.HealthResponse	"Service is ready"
//	@Failure		503	{object}	smsgateway.HealthResponse	"Service is not ready"
//	@Router			/health/ready [get]
//	@Router			/3rdparty/v1/health [get]
//
// Readiness probe.
func (h *HealthHandler) getReadiness(c *fiber.Ctx) error {
	return h.writeProbe(c, h.healthSvc.CheckReadiness(c.Context()))
}

//	@Summary		Startup probe
//	@Description	Checks if service has completed initialization (startup probe)
//	@Tags			System
//	@Produce		json
//	@Success		200	{object}	smsgateway.HealthResponse	"Service has completed initialization"
//	@Failure		503	{object}	smsgateway.HealthResponse	"Service has not completed initialization"
//	@Router			/health/startup [get]
//
// Startup probe.
func (h *HealthHandler) getStartup(c *fiber.Ctx) error {
	return h.writeProbe(c, h.healthSvc.CheckStartup(c.Context()))
}

func (h *HealthHandler) writeProbe(c *fiber.Ctx, r health.CheckResult) error {
	status := fiber.StatusOK
	if r.Status() == health.StatusFail {
		status = fiber.StatusServiceUnavailable
	}
	return c.Status(status).JSON(h.makeResponse(r))
}

func (h *HealthHandler) makeResponse(result health.CheckResult) smsgateway.HealthResponse {
	checks := lo.MapValues(
		result.Checks,
		func(value health.CheckDetail, _ string) smsgateway.HealthCheck {
			return smsgateway.HealthCheck{
				Description:   value.Description,
				ObservedUnit:  value.ObservedUnit,
				ObservedValue: value.ObservedValue,
				Status:        smsgateway.HealthStatus(value.Status),
			}
		},
	)
	checks["messages:hashing_interval_seconds"] = smsgateway.HealthCheck{
		Description:   "Outgoing message hashing interval",
		ObservedUnit:  "seconds",
		ObservedValue: h.messagesHashingInterval,
		Status:        smsgateway.HealthStatusPass,
	}

	return smsgateway.HealthResponse{
		Status:    smsgateway.HealthStatus(result.Status()),
		Version:   version.AppVersion,
		ReleaseID: version.AppReleaseID(),
		Checks:    checks,
	}
}

func (h *HealthHandler) Register(router fiber.Router) {
	router.Get("/health", h.getReadiness)
	router.Get("/health/live", h.getLiveness)
	router.Get("/health/ready", h.getReadiness)
	router.Get("/health/startup", h.getStartup)
}

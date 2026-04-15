package handlers

import (
	"fmt"
	"path"
	"strings"

	"github.com/android-sms-gateway/server/internal/sms-gateway/handlers/middlewares/userauth"
	"github.com/android-sms-gateway/server/internal/sms-gateway/openapi"
	"github.com/android-sms-gateway/server/internal/sms-gateway/users"
	"github.com/gofiber/fiber/v2"
)

type rootHandler struct {
	config Config

	healthHandler  *HealthHandler
	openapiHandler *openapi.Handler
	usersSvc       *users.Service
}

func (h *rootHandler) Register(app *fiber.App) {
	if h.config.PublicPath != "/api" {
		app.Use(func(c *fiber.Ctx) error {
			err := c.Next()

			location := c.GetRespHeader(fiber.HeaderLocation)
			if after, ok := strings.CutPrefix(location, "/api"); ok {
				c.Set(fiber.HeaderLocation, path.Join(h.config.PublicPath, after))
			}

			return err //nolint:wrapcheck // passed through to fiber's error handler
		})
	}

	h.healthHandler.Register(app)

	h.registerOpenAPI(app)
	h.registerUI(app)
}

func (h *rootHandler) registerOpenAPI(router fiber.Router) {
	if !h.config.OpenAPIEnabled {
		return
	}

	router.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/api" || c.Path() == "/api/" {
			return c.Redirect("/api/docs", fiber.StatusMovedPermanently)
		}

		return c.Next()
	})
	h.openapiHandler.Register(router.Group("/api/docs"), h.config.PublicHost, h.config.PublicPath)
}

func (h *rootHandler) registerUI(app *fiber.App) {
	group := app.Group("/ui",
		userauth.NewBasic(h.usersSvc),
		userauth.UserRequired(),
	)

	group.Get("", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(uiHTML)
	})

	group.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(uiHTML)
	})
}

func newRootHandler(
	cfg Config,
	healthHandler *HealthHandler,
	openapiHandler *openapi.Handler,
	usersSvc *users.Service,
) *rootHandler {
	return &rootHandler{
		config: cfg,

		healthHandler:  healthHandler,
		openapiHandler: openapiHandler,
		usersSvc:       usersSvc,
	}
}

var uiHTML = fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>SMSGate UI</title>
  <style>
    :root { color-scheme: light; }
    body { font-family: -apple-system, Segoe UI, Roboto, sans-serif; margin: 20px; background: #f7f8fb; color: #1f2937; }
    h1 { margin: 0 0 12px; }
    h2 { margin: 18px 0 10px; font-size: 18px; }
    .row { display: flex; gap: 12px; flex-wrap: wrap; }
    .card { background: white; border: 1px solid #e5e7eb; border-radius: 10px; padding: 14px; flex: 1 1 360px; min-width: 320px; }
    label { display: block; font-size: 12px; margin: 8px 0 4px; color: #4b5563; }
    input, textarea, button { width: 100%%; box-sizing: border-box; border-radius: 8px; border: 1px solid #d1d5db; padding: 8px; font-size: 14px; }
    textarea { min-height: 80px; resize: vertical; }
    button { background: #111827; color: #fff; border: none; cursor: pointer; margin-top: 10px; }
    button.secondary { background: #374151; }
    pre { background: #0b1020; color: #e5e7eb; border-radius: 8px; padding: 10px; overflow: auto; max-height: 360px; }
    .muted { color: #6b7280; font-size: 12px; }
  </style>
</head>
<body>
  <h1>SMSGate UI (MVP)</h1>
  <p class="muted">Same backend, browser authenticated with Basic auth. This page calls /api/3rdparty/v1/* directly.</p>

  <div class="row">
    <section class="card">
      <h2>Send SMS</h2>
      <label for="phone">Phone Number</label>
      <input id="phone" placeholder="+34600111222" />
      <label for="text">Message</label>
      <textarea id="text" placeholder="Message text"></textarea>
      <button id="sendBtn">Send</button>
      <pre id="sendResult"></pre>
    </section>

    <section class="card">
      <h2>Incoming</h2>
      <button class="secondary" id="refreshIncoming">Refresh Incoming</button>
      <pre id="incomingResult"></pre>
    </section>
  </div>

  <section class="card" style="margin-top: 12px;">
    <h2>Outgoing</h2>
    <button class="secondary" id="refreshOutgoing">Refresh Outgoing</button>
    <pre id="outgoingResult"></pre>
  </section>

  <script>
    const apiBase = "/api/3rdparty/v1";

    function pretty(value) {
      return JSON.stringify(value, null, 2);
    }

    async function request(path, options = {}) {
      const res = await fetch(apiBase + path, {
        ...options,
        headers: {
          "Content-Type": "application/json",
          ...(options.headers || {}),
        },
        credentials: "same-origin",
      });
      const contentType = res.headers.get("content-type") || "";
      const body = contentType.includes("application/json")
        ? await res.json()
        : await res.text();
      if (!res.ok) {
        throw new Error("HTTP " + res.status + " " + res.statusText + "\n" + pretty(body));
      }
      return body;
    }

    async function refreshIncoming() {
      const target = document.getElementById("incomingResult");
      target.textContent = "Loading...";
      try {
        const data = await request("/incoming?limit=25");
        target.textContent = pretty(data);
      } catch (err) {
        target.textContent = String(err);
      }
    }

    async function refreshOutgoing() {
      const target = document.getElementById("outgoingResult");
      target.textContent = "Loading...";
      try {
        const data = await request("/messages?limit=25");
        target.textContent = pretty(data);
      } catch (err) {
        target.textContent = String(err);
      }
    }

    async function sendSMS() {
      const phone = document.getElementById("phone").value.trim();
      const text = document.getElementById("text").value;
      const target = document.getElementById("sendResult");
      if (!phone || !text.trim()) {
        target.textContent = "Phone and message are required.";
        return;
      }

      target.textContent = "Sending...";
      try {
        const payload = {
          phoneNumbers: [phone],
          textMessage: { text: text },
          withDeliveryReport: true
        };
        const data = await request("/messages", {
          method: "POST",
          body: JSON.stringify(payload)
        });
        target.textContent = pretty(data);
        await refreshOutgoing();
      } catch (err) {
        target.textContent = String(err);
      }
    }

    document.getElementById("refreshIncoming").addEventListener("click", refreshIncoming);
    document.getElementById("refreshOutgoing").addEventListener("click", refreshOutgoing);
    document.getElementById("sendBtn").addEventListener("click", sendSMS);

    refreshIncoming();
    refreshOutgoing();
  </script>
</body>
</html>`)

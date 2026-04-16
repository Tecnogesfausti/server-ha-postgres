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
		func(c *fiber.Ctx) error {
			if !userauth.HasUser(c) {
				c.Set(fiber.HeaderWWWAuthenticate, `Basic realm="SMSGate UI"`)
				return fiber.ErrUnauthorized
			}
			return c.Next()
		},
		userauth.UserRequired(),
	)

	group.Get("", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(uiHTML)
	})

	group.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(uiHTML)
	})

	group.Get("/st-904l", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(uiST904LHTML)
	})

	group.Get("/st-904la", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(uiST904LHTML)
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
    .list { display: grid; gap: 8px; max-height: 420px; overflow: auto; margin-top: 10px; }
    .item { border: 1px solid #e5e7eb; border-radius: 8px; padding: 10px; background: #fafafa; }
    .item-top { display: flex; justify-content: space-between; gap: 8px; margin-bottom: 6px; font-size: 12px; color: #374151; }
    .item-title { font-size: 14px; color: #111827; margin: 0 0 4px; font-weight: 600; }
    .item-preview { font-size: 13px; color: #1f2937; white-space: pre-wrap; word-break: break-word; }
    .tag { padding: 2px 6px; border-radius: 999px; background: #e5e7eb; color: #111827; font-size: 11px; }
    .error { color: #991b1b; background: #fee2e2; border: 1px solid #fecaca; border-radius: 8px; padding: 8px; }
  </style>
</head>
<body>
  <h1>SMSGate UI (MVP)</h1>
  <p class="muted">Same backend, browser authenticated with Basic auth. This page calls /api/3rdparty/v1/* directly.</p>

  <div class="row">
    <section class="card">
      <h2>API Auth</h2>
      <label for="authUser">Username</label>
      <input id="authUser" placeholder="KN_UH0" />
      <label for="authPass">Password</label>
      <input id="authPass" type="password" placeholder="password" />
      <p class="muted">Used for /api/3rdparty/v1 calls from this UI.</p>
    </section>

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
      <div id="incomingResult" class="list"></div>
    </section>
  </div>

  <section class="card" style="margin-top: 12px;">
    <h2>Outgoing</h2>
    <button class="secondary" id="refreshOutgoing">Refresh Outgoing</button>
    <div id="outgoingResult" class="list"></div>
  </section>

  <script>
    const apiBase = window.location.protocol + "//" + window.location.host + "/api/3rdparty/v1";

    function pretty(value) {
      return JSON.stringify(value, null, 2);
    }

    function escapeHTML(value) {
      return String(value ?? "")
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll("\"", "&quot;")
        .replaceAll("'", "&#39;");
    }

    function toLocalDate(value) {
      if (!value) return "";
      const d = new Date(value);
      if (Number.isNaN(d.getTime())) return String(value);
      return d.toLocaleString();
    }

    function renderEmpty(target, text) {
      target.innerHTML = "<div class=\"muted\">" + escapeHTML(text) + "</div>";
    }

    function renderError(target, err) {
      target.innerHTML = "<div class=\"error\">" + escapeHTML(String(err)) + "</div>";
    }

    function renderIncomingList(target, items) {
      if (!Array.isArray(items) || items.length === 0) {
        renderEmpty(target, "No incoming messages.");
        return;
      }
      target.innerHTML = items.map((item) => {
        const sender = item.sender || "-";
        const preview = item.contentPreview || "";
        const receivedAt = toLocalDate(item.receivedAt);
        const id = item.id || "";
        return "<article class=\"item\">"
          + "<div class=\"item-top\"><span class=\"tag\">Incoming</span><span>" + escapeHTML(receivedAt) + "</span></div>"
          + "<p class=\"item-title\">From: " + escapeHTML(sender) + "</p>"
          + "<div class=\"item-preview\">" + escapeHTML(preview) + "</div>"
          + "<div class=\"muted\">id: " + escapeHTML(id) + "</div>"
          + "</article>";
      }).join("");
    }

    function renderOutgoingList(target, items) {
      if (!Array.isArray(items) || items.length === 0) {
        renderEmpty(target, "No outgoing messages.");
        return;
      }
      target.innerHTML = items.map((item) => {
        const state = item.state || "-";
        const phone = Array.isArray(item.phoneNumbers) && item.phoneNumbers.length > 0
          ? item.phoneNumbers.join(", ")
          : ((Array.isArray(item.recipients) && item.recipients[0] && item.recipients[0].phoneNumber) || "-");
        const preview = item.contentPreview || item.message || "";
        const sentAt = item.states && (item.states.Delivered || item.states.Sent || item.states.Processed || item.states.Pending);
        return "<article class=\"item\">"
          + "<div class=\"item-top\"><span class=\"tag\">" + escapeHTML(state) + "</span><span>" + escapeHTML(toLocalDate(sentAt)) + "</span></div>"
          + "<p class=\"item-title\">To: " + escapeHTML(phone) + "</p>"
          + "<div class=\"item-preview\">" + escapeHTML(preview || "[no content]") + "</div>"
          + "<div class=\"muted\">id: " + escapeHTML(item.id || "") + (item.isHashed ? " | hashed" : "") + "</div>"
          + "</article>";
      }).join("");
    }

    async function request(path, options = {}) {
      const authUser = document.getElementById("authUser").value.trim();
      const authPass = document.getElementById("authPass").value;
      const headers = {
        "Content-Type": "application/json",
        ...(options.headers || {}),
      };
      if (authUser) {
        headers["Authorization"] = "Basic " + btoa(authUser + ":" + authPass);
      }

      const res = await fetch(apiBase + path, {
        ...options,
        headers: headers,
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
      target.innerHTML = "<div class=\"muted\">Loading...</div>";
      try {
        const data = await request("/incoming?limit=25");
        renderIncomingList(target, data);
      } catch (err) {
        renderError(target, err);
      }
    }

    async function refreshOutgoing() {
      const target = document.getElementById("outgoingResult");
      target.innerHTML = "<div class=\"muted\">Loading...</div>";
      try {
        const data = await request("/messages?limit=25");
        renderOutgoingList(target, data);
      } catch (err) {
        renderError(target, err);
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

    renderEmpty(document.getElementById("incomingResult"), "Fill API Auth and click Refresh Incoming.");
    renderEmpty(document.getElementById("outgoingResult"), "Fill API Auth and click Refresh Outgoing.");
  </script>
</body>
</html>`)

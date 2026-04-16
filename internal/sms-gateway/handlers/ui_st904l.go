package handlers

var uiST904LHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>SMSGate ST-904L Console</title>
  <style>
    body { font-family: -apple-system, Segoe UI, Roboto, sans-serif; margin: 20px; background: #f3f4f6; color: #111827; }
    h1, h2 { margin: 0 0 10px; }
    .muted { color: #6b7280; font-size: 12px; }
    .grid { display: grid; gap: 12px; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); }
    .card { background: #fff; border: 1px solid #e5e7eb; border-radius: 10px; padding: 12px; }
    label { display: block; font-size: 12px; margin: 8px 0 4px; color: #4b5563; }
    input, textarea, button, select { width: 100%; box-sizing: border-box; border: 1px solid #d1d5db; border-radius: 8px; padding: 8px; font-size: 14px; }
    textarea { min-height: 70px; resize: vertical; }
    .row { display: flex; gap: 8px; flex-wrap: wrap; }
    .row > button { flex: 1 1 130px; }
    button { border: none; background: #111827; color: #fff; cursor: pointer; }
    button.secondary { background: #374151; }
    .list { margin-top: 10px; display: grid; gap: 8px; max-height: 420px; overflow: auto; }
    .item { border: 1px solid #e5e7eb; border-radius: 8px; background: #fafafa; padding: 8px; }
    .item-top { display: flex; justify-content: space-between; font-size: 12px; color: #4b5563; }
    .item-main { margin-top: 4px; white-space: pre-wrap; word-break: break-word; }
    .kv { margin-top: 6px; display: grid; gap: 4px; }
    .kv-row { display: grid; grid-template-columns: 120px 1fr; gap: 8px; font-size: 12px; align-items: start; }
    .kv-key { color: #6b7280; }
    .kv-val { color: #111827; word-break: break-word; }
    .tag { background: #e5e7eb; border-radius: 999px; padding: 2px 6px; font-size: 11px; color: #111827; }
    pre { background: #0b1020; color: #e5e7eb; border-radius: 8px; padding: 10px; overflow: auto; max-height: 180px; }
    .doc { border: 1px dashed #d1d5db; border-radius: 8px; padding: 8px; background: #f9fafb; margin-top: 8px; }
    .doc p { margin: 6px 0; font-size: 13px; }
    .doc code { background: #e5e7eb; padding: 1px 4px; border-radius: 4px; }
  </style>
</head>
<body>
  <h1>ST-904L Command Console</h1>
  <p class="muted">Route: /ui/st-904l. Catálogo precargado de comandos ST-904L/ST-904LA y envío por SMS gateway.</p>

  <div class="grid">
    <section class="card">
      <h2>Session</h2>
      <label for="authUser">API User</label>
      <input id="authUser" placeholder="M6KFDA" />
      <label for="authPass">API Password</label>
      <input id="authPass" type="password" placeholder="password" />
      <label for="trackerPhone">Tracker Phone (SIM in tracker)</label>
      <input id="trackerPhone" placeholder="+346XXXXXXXX" />
      <div class="row" style="margin-top:10px;">
        <button id="saveSession" class="secondary">Save Session</button>
        <button id="loadSession" class="secondary">Load Session</button>
      </div>
    </section>

    <section class="card">
      <h2>Send Command</h2>
      <label for="commandPreset">Command Catalog</label>
      <select id="commandPreset"></select>
      <div class="doc">
        <p><strong id="docTitle">-</strong></p>
        <p id="docDescription" class="muted">Select a command to see documentation.</p>
        <p><strong>Template:</strong> <code id="docTemplate">-</code></p>
        <p><strong>Example:</strong> <code id="docExample">-</code></p>
        <p class="muted" id="docNotes"></p>
      </div>
      <label for="commandText">SMS Command</label>
      <textarea id="commandText" placeholder="RCONF"></textarea>
      <button id="sendCommand">Send To Tracker</button>
      <pre id="sendResult"></pre>
    </section>
  </div>

  <section class="card" style="margin-top:12px;">
    <h2>Tracker Timeline</h2>
    <div class="row">
      <button id="refreshAll" class="secondary">Refresh</button>
      <button id="toggleAuto" class="secondary">Auto Refresh: Off</button>
    </div>
    <div class="grid" style="margin-top:10px;">
      <div>
        <h2 style="font-size:16px;">Outgoing Commands</h2>
        <div id="outgoingList" class="list"></div>
      </div>
      <div>
        <h2 style="font-size:16px;">Incoming Replies</h2>
        <div id="incomingList" class="list"></div>
      </div>
    </div>
  </section>

  <script>
    const apiBase = window.location.protocol + "//" + window.location.host + "/api/3rdparty/v1";
    let autoTimer = null;
    const commandCatalog = [
      { key: "POS_SMS", label: "6690000 - SMS tracking", template: "6690000", example: "6690000", description: "Solicita posición por SMS (link Google Maps).", notes: "Comando directo al SIM del tracker." },
      { key: "ADMIN_SET_1", label: "Admin set #1", template: "{ADMIN_PHONE}0000 1", example: "132657901800000 1", description: "Define número administrador principal.", notes: "Formato exacto: phone + password + espacio + índice." },
      { key: "VOICE_MONITOR_66", label: "66 - Voice monitor callback", template: "66", example: "66", description: "Tracker devuelve llamada al número admin.", notes: "Solo funciona si ya existe admin number configurado." },
      { key: "ADMIN_CANCEL_D101", label: "D101# - admin cancel", template: "D101#", example: "D101#", description: "Cancela número admin (según manual).", notes: "La sintaxis puede variar por firmware; confirmar respuesta SET OK!." },
      { key: "SPEED_SET", label: "1220000 070 - overspeed set", template: "1220000 070", example: "1220000 070", description: "Alarma exceso velocidad a 70 km/h.", notes: "Usa 3 dígitos para velocidad (070, 090, etc.)." },
      { key: "SPEED_CANCEL", label: "1220000 0 - overspeed off", template: "1220000 0", example: "1220000 0", description: "Desactiva alarma de velocidad.", notes: "Debe responder SET OK!." },
      { key: "SHAKE_SET", label: "1810000T10 - shake alarm set", template: "1810000T10", example: "1810000T10", description: "Activa alarma de vibración/golpe.", notes: "Tras activar, dejar el tracker quieto 5 minutos." },
      { key: "SHAKE_CANCEL", label: "1800000 - shake alarm off", template: "1800000", example: "1800000", description: "Desactiva alarma de vibración.", notes: "Debe responder SET OK!." },
      { key: "MODE_WORK", label: "WORK0000 - keep working mode", template: "WORK0000", example: "WORK0000", description: "Modo trabajo continuo.", notes: "Mayor consumo batería (aprox 18-24h)." },
      { key: "MODE_MOVE", label: "MOVE0000 - move mode", template: "MOVE0000", example: "MOVE0000", description: "Modo trabajo al detectar movimiento.", notes: "Modo fábrica habitual; mejor autonomía." },
      { key: "MODE_STANDBY", label: "STANDBY0000 - standby mode", template: "STANDBY0000", example: "STANDBY0000", description: "Modo standby; despierta por SMS/llamada.", notes: "Autonomía superior según manual." },
      { key: "RCONF", label: "RCONF - read config", template: "RCONF", example: "RCONF", description: "Solicita configuración actual del equipo.", notes: "Útil para validar APN, servidor y modo." },
      { key: "MODE_GPRS", label: "7100000 - set GPRS mode", template: "7100000", example: "7100000", description: "Cambia a modo GPRS.", notes: "Puede requerir comando adicional de reporte." },
      { key: "MODE_SMS", label: "7000000 - set SMS mode", template: "7000000", example: "7000000", description: "Cambia a modo SMS.", notes: "Adecuado para uso solo por comandos SMS." },
      { key: "RESET", label: "RESET - restart tracker", template: "RESET", example: "RESET", description: "Reinicio del tracker.", notes: "Puede quedar offline unos segundos." },
      { key: "APN_SET", label: "8030000 - set APN", template: "8030000 {APN}", example: "8030000 internet", description: "Configura APN.", notes: "Depende de operador SIM." },
      { key: "APN_USER_PASS", label: "APN user/pass", template: "8030000 {APN} {APN_USER} {APN_PASS}", example: "8030000 iot.movistar.es movistar movistar", description: "APN con usuario/clave.", notes: "Solo si el operador exige autenticación." },
      { key: "SERVER_SET", label: "8040000 - server host/port", template: "8040000 {HOST} {PORT}", example: "8040000 47.254.77.28 8090", description: "Configura host/puerto servidor.", notes: "Úsalo para modo plataforma TCP." },
      { key: "TIMEZONE", label: "8960000 - timezone", template: "8960000 {TZ}", example: "8960000 E00", description: "Configura zona horaria.", notes: "Ejemplos: E00, E01, W03." },
      { key: "CHECK_SIM", label: "CHECK - SIM/network info", template: "CHECK", example: "CHECK", description: "Consulta estado SIM/red.", notes: "Útil para diagnosticar sin datos." },
      { key: "FACTORY", label: "FACTORY - reset defaults", template: "FACTORY", example: "FACTORY", description: "Reset a valores de fábrica.", notes: "Impacto alto: hay que reconfigurar todo." }
    ];

    function $(id) { return document.getElementById(id); }
    function pretty(v) { return JSON.stringify(v, null, 2); }
    function esc(v) {
      return String(v ?? "")
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll("\"", "&quot;")
        .replaceAll("'", "&#39;");
    }
    function toDate(v) {
      if (!v) return "";
      const d = new Date(v);
      if (Number.isNaN(d.getTime())) return String(v);
      return d.toLocaleString();
    }
    function normalizePhone(v) {
      return String(v || "").replace(/\s+/g, "").replace(/^\+/, "");
    }

    function initCommandCatalog() {
      const select = $("commandPreset");
      select.innerHTML = commandCatalog.map((item) =>
        "<option value=\"" + esc(item.key) + "\">" + esc(item.label) + "</option>"
      ).join("");
      select.addEventListener("change", () => applyPreset(select.value));
      if (commandCatalog.length > 0) {
        applyPreset(commandCatalog[0].key);
      }
    }

    function applyPreset(key) {
      const item = commandCatalog.find((v) => v.key === key);
      if (!item) return;
      $("docTitle").textContent = item.label;
      $("docDescription").textContent = item.description;
      $("docTemplate").textContent = item.template;
      $("docExample").textContent = item.example;
      $("docNotes").textContent = "Note: " + item.notes;
      $("commandText").value = item.template;
    }

    async function request(path, options = {}) {
      const authUser = $("authUser").value.trim();
      const authPass = $("authPass").value;
      const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
      if (authUser) {
        headers["Authorization"] = "Basic " + btoa(authUser + ":" + authPass);
      }
      const res = await fetch(apiBase + path, { ...options, headers, credentials: "same-origin" });
      const ct = res.headers.get("content-type") || "";
      const body = ct.includes("application/json") ? await res.json() : await res.text();
      if (!res.ok) throw new Error("HTTP " + res.status + " " + res.statusText + "\n" + pretty(body));
      return body;
    }

    function renderList(target, items, emptyText, renderItem) {
      if (!Array.isArray(items) || items.length === 0) {
        target.innerHTML = "<div class=\"muted\">" + esc(emptyText) + "</div>";
        return;
      }
      target.innerHTML = items.map(renderItem).join("");
    }

    function pickStateTime(states) {
      if (!states) return "";
      return states.Delivered || states.Sent || states.Processed || states.Pending || "";
    }

    function looksLikeRconf(text) {
      const t = String(text || "");
      return t.includes("ID:") && t.includes("MODE:") && t.includes("APN:");
    }

    function parseRconf(text) {
      const out = {};
      const chunks = String(text || "").split(",");
      chunks.forEach((chunk) => {
        const c = chunk.trim();
        const i = c.indexOf(":");
        if (i <= 0) return;
        const key = c.slice(0, i).trim();
        const val = c.slice(i + 1).trim();
        if (!key || !val) return;
        out[key.toUpperCase()] = val;
      });
      const flags = [];
      ["DAILY", "GEO FENCE", "OVER SPEED", "VOICE", "SHAKE ALARM", "SLEEP"].forEach((k) => {
        if (out[k]) flags.push(k + ":" + out[k]);
      });
      return {
        model: out["ST-904/ST-904L"] || "",
        id: out["ID"] || "",
        password: out["UP"] || "",
        u1: out["U1"] || "",
        u2: out["U2"] || "",
        u3: out["U3"] || "",
        mode: out["MODE"] || "",
        apn: out["APN"] || "",
        ip: out["IP"] || "",
        upload: out["GPRS UPLOAD TIME"] || "",
        timezone: out["TIME ZONE"] || "",
        flags: flags.join(" | ")
      };
    }

    function renderRconfBody(text) {
      const p = parseRconf(text);
      const row = (k, v) => "<div class=\"kv-row\"><div class=\"kv-key\">" + esc(k) + "</div><div class=\"kv-val\">" + esc(v || "-") + "</div></div>";
      return "<div class=\"item-main\"><strong>RCONF decoded</strong></div>"
        + "<div class=\"kv\">"
        + row("Model", p.model)
        + row("ID", p.id)
        + row("Password", p.password)
        + row("Admin U1", p.u1)
        + row("Admin U2", p.u2)
        + row("Admin U3", p.u3)
        + row("Mode", p.mode)
        + row("APN", p.apn)
        + row("IP:Port", p.ip)
        + row("Upload(s)", p.upload)
        + row("Timezone", p.timezone)
        + row("Flags", p.flags)
        + "</div>"
        + "<div class=\"item-main muted\">Raw: " + esc(text) + "</div>";
    }

    function renderIncomingBody(text) {
      if (looksLikeRconf(text)) return renderRconfBody(text);
      return "<div class=\"item-main\">" + esc(text || "") + "</div>";
    }

    async function refreshTrackerTimeline() {
      const trackerPhone = $("trackerPhone").value.trim();
      const normTracker = normalizePhone(trackerPhone);
      if (!normTracker) {
        $("incomingList").innerHTML = "<div class=\"muted\">Set tracker phone first.</div>";
        $("outgoingList").innerHTML = "<div class=\"muted\">Set tracker phone first.</div>";
        return;
      }

      try {
        const senderVariants = [trackerPhone, "+" + normTracker, normTracker];
        const incomingBatches = await Promise.all(
          senderVariants.map((sender) => request("/incoming?limit=30&sender=" + encodeURIComponent(sender)).catch(() => []))
        );
        const incoming = [].concat(...incomingBatches)
          .filter((v, i, arr) => arr.findIndex((x) => x.id === v.id) === i)
          .sort((a, b) => String(b.receivedAt).localeCompare(String(a.receivedAt)));

        const outgoingAll = await request("/messages?limit=80");
        const outgoing = outgoingAll.filter((item) => {
          const phones = Array.isArray(item.phoneNumbers) ? item.phoneNumbers : [];
          const recipientPhones = Array.isArray(item.recipients) ? item.recipients.map((r) => r.phoneNumber) : [];
          return phones.concat(recipientPhones).some((p) => normalizePhone(p) === normTracker);
        });

        renderList(
          $("incomingList"),
          incoming,
          "No incoming replies for tracker.",
          (item) => "<article class=\"item\">"
            + "<div class=\"item-top\"><span class=\"tag\">IN</span><span>" + esc(toDate(item.receivedAt)) + "</span></div>"
            + renderIncomingBody(item.contentPreview || "")
            + "<div class=\"muted\">from " + esc(item.sender || "-") + " | id " + esc(item.id || "") + "</div>"
            + "</article>"
        );

        renderList(
          $("outgoingList"),
          outgoing,
          "No outgoing commands for tracker.",
          (item) => "<article class=\"item\">"
            + "<div class=\"item-top\"><span class=\"tag\">" + esc(item.state || "-") + "</span><span>" + esc(toDate(pickStateTime(item.states))) + "</span></div>"
            + "<div class=\"item-main\">" + esc(item.contentPreview || item.message || "") + "</div>"
            + "<div class=\"muted\">id " + esc(item.id || "") + (item.isHashed ? " | hashed" : "") + "</div>"
            + "</article>"
        );
      } catch (err) {
        const msg = "<div class=\"muted\" style=\"color:#991b1b;\">" + esc(String(err)) + "</div>";
        $("incomingList").innerHTML = msg;
        $("outgoingList").innerHTML = msg;
      }
    }

    async function sendCommand() {
      const phone = $("trackerPhone").value.trim();
      const text = $("commandText").value.trim();
      if (!phone || !text) {
        $("sendResult").textContent = "Tracker phone and command are required.";
        return;
      }

      $("sendResult").textContent = "Sending...";
      try {
        const payload = { phoneNumbers: [phone], textMessage: { text: text }, withDeliveryReport: true };
        const data = await request("/messages", { method: "POST", body: JSON.stringify(payload) });
        $("sendResult").textContent = pretty(data);
        await refreshTrackerTimeline();
      } catch (err) {
        $("sendResult").textContent = String(err);
      }
    }

    function saveSession() {
      localStorage.setItem("st904l_auth_user", $("authUser").value.trim());
      localStorage.setItem("st904l_auth_pass", $("authPass").value);
      localStorage.setItem("st904l_phone", $("trackerPhone").value.trim());
    }

    function loadSession() {
      $("authUser").value = localStorage.getItem("st904l_auth_user") || "";
      $("authPass").value = localStorage.getItem("st904l_auth_pass") || "";
      $("trackerPhone").value = localStorage.getItem("st904l_phone") || "";
    }

    function toggleAuto() {
      if (autoTimer) {
        clearInterval(autoTimer);
        autoTimer = null;
        $("toggleAuto").textContent = "Auto Refresh: Off";
        return;
      }
      autoTimer = setInterval(refreshTrackerTimeline, 15000);
      $("toggleAuto").textContent = "Auto Refresh: On (15s)";
      refreshTrackerTimeline();
    }

    $("sendCommand").addEventListener("click", sendCommand);
    $("refreshAll").addEventListener("click", refreshTrackerTimeline);
    $("toggleAuto").addEventListener("click", toggleAuto);
    $("saveSession").addEventListener("click", saveSession);
    $("loadSession").addEventListener("click", loadSession);

    initCommandCatalog();
    loadSession();
    $("incomingList").innerHTML = "<div class=\"muted\">Set tracker phone and click Refresh.</div>";
    $("outgoingList").innerHTML = "<div class=\"muted\">Set tracker phone and click Refresh.</div>";
    $("sendResult").textContent = "Ready.";
  </script>
</body>
</html>`

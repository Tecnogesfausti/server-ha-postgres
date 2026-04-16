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
    input, textarea, button { width: 100%; box-sizing: border-box; border: 1px solid #d1d5db; border-radius: 8px; padding: 8px; font-size: 14px; }
    textarea { min-height: 70px; resize: vertical; }
    .row { display: flex; gap: 8px; flex-wrap: wrap; }
    .row > button { flex: 1 1 130px; }
    button { border: none; background: #111827; color: #fff; cursor: pointer; }
    button.secondary { background: #374151; }
    .list { margin-top: 10px; display: grid; gap: 8px; max-height: 420px; overflow: auto; }
    .item { border: 1px solid #e5e7eb; border-radius: 8px; background: #fafafa; padding: 8px; }
    .item-top { display: flex; justify-content: space-between; font-size: 12px; color: #4b5563; }
    .item-main { margin-top: 4px; white-space: pre-wrap; word-break: break-word; }
    .tag { background: #e5e7eb; border-radius: 999px; padding: 2px 6px; font-size: 11px; color: #111827; }
    pre { background: #0b1020; color: #e5e7eb; border-radius: 8px; padding: 10px; overflow: auto; max-height: 180px; }
  </style>
</head>
<body>
  <h1>ST-904L Command Console</h1>
  <p class="muted">Route: /ui/st-904l. Usa la API 3rdparty existente para enviar comandos SMS y leer respuestas.</p>

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
      <label for="commandText">SMS Command</label>
      <textarea id="commandText" placeholder="RCONF"></textarea>
      <div class="row">
        <button class="secondary quick" data-cmd="RCONF">RCONF</button>
        <button class="secondary quick" data-cmd="6690000">6690000</button>
        <button class="secondary quick" data-cmd="7100000">7100000</button>
      </div>
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
      return String(v || "").replaceAll(/\s+/g, "").replaceAll(/^\+/, "");
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
            + "<div class=\"item-main\">" + esc(item.contentPreview || "") + "</div>"
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
        const payload = { phoneNumbers: [phone], textMessage: { text }, withDeliveryReport: true };
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
    document.querySelectorAll(".quick").forEach((btn) => {
      btn.addEventListener("click", () => { $("commandText").value = btn.dataset.cmd || ""; });
    });

    loadSession();
    $("incomingList").innerHTML = "<div class=\"muted\">Set tracker phone and click Refresh.</div>";
    $("outgoingList").innerHTML = "<div class=\"muted\">Set tracker phone and click Refresh.</div>";
    $("sendResult").textContent = "Ready.";
  </script>
</body>
</html>`

package api

const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>OCPP Learning Test Lab</title>
  <style>
    :root {
      --bg: #eef2f5;
      --ink: #14202b;
      --muted: #526170;
      --panel: #ffffff;
      --line: #d4dee8;
      --accent: #065f8c;
      --accent-soft: #d8edf8;
      --good: #0f7d49;
      --warn: #a66a00;
      --bad: #a32222;
      --terminal: #0f1a24;
      --terminal-ink: #cee7ff;
      --hero-grad-a: #f8fcff;
      --hero-grad-b: #d6eaf7;
    }

    * { box-sizing: border-box; }

    body {
      margin: 0;
      color: var(--ink);
      font-family: "Sora", "Avenir Next", "Trebuchet MS", sans-serif;
      background:
        radial-gradient(1200px 480px at 110% -10%, #cce8f9 0%, transparent 60%),
        radial-gradient(900px 520px at -10% 120%, #d8f2ea 0%, transparent 55%),
        var(--bg);
    }

    .shell {
      max-width: 1280px;
      margin: 0 auto;
      padding: 18px;
      display: grid;
      gap: 14px;
    }

    .card {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 14px;
      box-shadow: 0 10px 28px rgba(16, 41, 64, 0.08);
      padding: 14px;
    }

    .hero {
      padding: 16px;
      background: linear-gradient(135deg, var(--hero-grad-a), var(--hero-grad-b));
      display: grid;
      gap: 12px;
    }

    .hero h1 {
      margin: 0;
      font-size: 26px;
      letter-spacing: 0.1px;
    }

    .hero .sub {
      margin-top: 4px;
      color: var(--muted);
      font-size: 14px;
      line-height: 1.45;
      max-width: 900px;
    }

    .chip-row {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }

    .chip {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      padding: 6px 10px;
      border-radius: 999px;
      border: 1px solid #b9cfe0;
      background: #fff;
      font-size: 12px;
      color: #1f3648;
    }

    .dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: #6f8292;
      display: inline-block;
    }

    .dot.good { background: var(--good); }
    .dot.warn { background: var(--warn); }
    .dot.bad { background: var(--bad); }

    .grid-2 {
      display: grid;
      gap: 14px;
      grid-template-columns: 1.15fr 0.85fr;
    }

    .grid-3 {
      display: grid;
      gap: 10px;
      grid-template-columns: repeat(6, minmax(0, 1fr));
    }

    .stat {
      border: 1px solid var(--line);
      border-radius: 12px;
      background: linear-gradient(180deg, #ffffff 0%, #f7fbff 100%);
      padding: 12px;
      min-height: 78px;
    }

    .stat .k {
      font-size: 12px;
      color: var(--muted);
    }

    .stat .v {
      margin-top: 5px;
      font-size: 24px;
      font-weight: 700;
    }

    h2 {
      margin: 0 0 10px;
      font-size: 18px;
      letter-spacing: 0.2px;
    }

    .muted {
      color: var(--muted);
      font-size: 12px;
    }

    .row {
      display: grid;
      gap: 10px;
      grid-template-columns: repeat(12, minmax(0, 1fr));
    }

    .c3 { grid-column: span 3; }
    .c4 { grid-column: span 4; }
    .c6 { grid-column: span 6; }
    .c8 { grid-column: span 8; }
    .c12 { grid-column: span 12; }

    label {
      display: block;
      margin-bottom: 4px;
      font-size: 12px;
      color: var(--muted);
    }

    input,
    select,
    textarea {
      width: 100%;
      border: 1px solid var(--line);
      border-radius: 9px;
      padding: 8px 10px;
      font-size: 14px;
      color: var(--ink);
      background: #fff;
    }

    input:focus,
    select:focus,
    textarea:focus {
      outline: 2px solid #b6ddf2;
      border-color: #7ebadb;
    }

    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-top: 8px;
    }

    button {
      border: 1px solid #0f6c9b;
      background: var(--accent);
      color: #fff;
      border-radius: 9px;
      padding: 8px 12px;
      font-size: 13px;
      cursor: pointer;
      transition: transform 0.15s ease, box-shadow 0.15s ease;
    }

    button:hover {
      transform: translateY(-1px);
      box-shadow: 0 6px 14px rgba(4, 66, 99, 0.2);
    }

    button.secondary {
      background: #fff;
      color: #0f5f89;
      border-color: #8ec3de;
    }

    button.warn {
      background: #b97400;
      border-color: #9b6200;
    }

    button.bad {
      background: #9b2525;
      border-color: #822121;
    }

    .scenario-box {
      border: 1px solid var(--line);
      border-radius: 10px;
      background: #f8fcff;
      padding: 10px;
      margin-top: 10px;
    }

    .hint-box {
      border: 1px dashed #8fc3de;
      border-radius: 10px;
      background: #f6fbff;
      padding: 10px;
      font-size: 13px;
      color: #234054;
      line-height: 1.4;
    }

    .hint-list {
      margin: 0;
      padding-left: 18px;
      line-height: 1.45;
      font-size: 13px;
    }

    .hint-list li {
      margin-bottom: 6px;
    }

    .kv {
      margin-bottom: 10px;
      border: 1px solid var(--line);
      border-radius: 10px;
      background: #f9fcff;
      padding: 10px;
    }

    .kv strong {
      display: block;
      margin-bottom: 4px;
      font-size: 12px;
      color: #2f4c63;
    }

    code {
      font-family: "IBM Plex Mono", "Consolas", monospace;
      font-size: 12px;
      color: #173a52;
      background: #eaf4fb;
      border: 1px solid #c8deed;
      border-radius: 6px;
      padding: 2px 6px;
    }

    .table-wrap {
      overflow-x: auto;
      border: 1px solid var(--line);
      border-radius: 10px;
    }

    table {
      width: 100%;
      border-collapse: collapse;
      font-size: 12px;
    }

    th,
    td {
      padding: 8px 7px;
      border-bottom: 1px solid #e3ebf2;
      text-align: left;
      vertical-align: top;
    }

    th {
      background: #f3f9fd;
      color: #3b566b;
      position: sticky;
      top: 0;
      z-index: 1;
    }

    tr:last-child td {
      border-bottom: none;
    }

    .tag {
      display: inline-block;
      padding: 2px 8px;
      border-radius: 999px;
      font-size: 11px;
      border: 1px solid #c8d7e3;
      background: #fff;
      margin: 0 5px 4px 0;
      white-space: nowrap;
    }

    .tag.good {
      border-color: #b7e3cc;
      background: #ecfbf3;
      color: #0f6b3f;
    }

    .tag.warn {
      border-color: #f0d9af;
      background: #fff8e8;
      color: #8d5b00;
    }

    .tag.bad {
      border-color: #eec1c1;
      background: #fff0f0;
      color: #8f1f1f;
    }

    .timeline-controls {
      display: grid;
      gap: 8px;
      grid-template-columns: repeat(12, minmax(0, 1fr));
      margin-bottom: 10px;
    }

    .timeline-table {
      max-height: 360px;
      overflow: auto;
      border: 1px solid var(--line);
      border-radius: 10px;
    }

    .level-info td { background: #ffffff; }
    .level-warn td { background: #fffaf0; }
    .level-error td { background: #fff2f2; }

    .mono {
      font-family: "IBM Plex Mono", "Consolas", monospace;
    }

    .terminal {
      margin: 0;
      border: 1px solid #243544;
      border-radius: 10px;
      background: var(--terminal);
      color: var(--terminal-ink);
      font-family: "IBM Plex Mono", "Consolas", monospace;
      font-size: 12px;
      line-height: 1.45;
      padding: 10px;
      max-height: 220px;
      overflow: auto;
      white-space: pre-wrap;
    }

    .pm-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 10px;
      margin-top: 8px;
    }

    .pm-card {
      border: 1px solid var(--line);
      border-radius: 10px;
      background: #f8fcff;
      padding: 10px;
    }

    .pm-card h3 {
      margin: 0 0 8px;
      font-size: 14px;
    }

    .pm-card ol,
    .pm-card ul {
      margin: 0;
      padding-left: 17px;
      font-size: 13px;
      line-height: 1.45;
    }

    .pm-card li { margin-bottom: 5px; }

    .alert {
      border: 1px solid #e8c798;
      background: #fff8ed;
      color: #734600;
      border-radius: 10px;
      padding: 8px 10px;
      font-size: 12px;
      margin-top: 8px;
    }

    .footer-note {
      color: #466177;
      font-size: 12px;
      line-height: 1.45;
    }

    @media (max-width: 1080px) {
      .grid-2 { grid-template-columns: 1fr; }
      .grid-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); }
      .pm-grid { grid-template-columns: 1fr; }
    }

    @media (max-width: 720px) {
      .c3, .c4, .c6, .c8 { grid-column: span 12; }
      .grid-3 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
      .hero h1 { font-size: 22px; }
    }
  </style>
</head>
<body>
  <div class="shell">
    <section class="hero card">
      <div>
        <h1>OCPP Learning Test Lab</h1>
        <div class="sub">
          Beginner-friendly tool for validating station behavior step-by-step. Use scenarios, connector-level actions,
          and event timeline to confirm your backend really interacts with the charge point.
        </div>
      </div>
      <div class="chip-row">
        <div class="chip"><span id="dot-api" class="dot"></span> API <strong id="api-status">unknown</strong></div>
        <div class="chip"><span id="dot-ws" class="dot"></span> OCPP WS <strong id="ws-status">unknown</strong></div>
        <div class="chip">Active CP <strong id="active-cp-chip">my-test</strong></div>
        <div class="chip">Active Connector <strong id="active-connector-chip">1</strong></div>
      </div>
    </section>

    <section class="grid-2">
      <div class="card">
        <h2>Control Center</h2>
        <div class="kv">
          <strong>Station connection URL</strong>
          <code id="ocpp-url">ws://&lt;backend-ip&gt;:9005/my-test</code>
        </div>

        <div class="row">
          <div class="c6">
            <label>API base URL</label>
            <input id="api-base" />
          </div>
          <div class="c3">
            <label>Charge point</label>
            <select id="cp-select"></select>
          </div>
          <div class="c3">
            <label>Connector</label>
            <select id="connector-select"></select>
          </div>

          <div class="c4">
            <label>User ID</label>
            <input id="user-id" value="user-001" />
          </div>
          <div class="c4">
            <label>Amount paid</label>
            <input id="amount-paid" type="number" step="0.01" value="20" />
          </div>
          <div class="c4">
            <label>Transaction ID</label>
            <input id="tx-id" type="number" min="1" />
          </div>
        </div>

        <div class="actions">
          <button onclick="startSession()">Start Session</button>
          <button class="secondary" onclick="syncConnectors()">Sync Connectors From Charger</button>
          <button class="secondary" onclick="getEnergy()">Get Energy</button>
          <button class="secondary" onclick="stopSession()">Stop Session</button>
          <button class="secondary" onclick="refreshAll(true)">Refresh Now</button>
        </div>

        <div class="scenario-box">
          <label>Scenario</label>
          <select id="scenario-select" onchange="renderScenarioDescription()">
            <option value="happy_cycle">Happy path: start -> energy -> stop</option>
            <option value="start_poll_stop">Observe meter updates (multi-poll)</option>
            <option value="already_active">Error case: start twice on same connector</option>
            <option value="wrong_connector">Error case: wrong connector ID</option>
            <option value="unknown_cp">Error case: unknown charge point</option>
          </select>
          <div id="scenario-description" class="hint-box" style="margin-top:8px;"></div>
          <div class="actions">
            <button class="warn" onclick="runScenario()">Run Scenario</button>
            <button class="bad" onclick="clearLocalLog()">Clear Local Log</button>
          </div>
        </div>
      </div>

      <div class="card">
        <h2>Beginner Hints</h2>
        <div class="hint-box">
          If you are new: connect station first, wait for <strong>BootNotification</strong> and <strong>StatusNotification</strong>
          in timeline, then run the happy-path scenario.
        </div>
        <ul id="hints" class="hint-list" style="margin-top:10px;"></ul>

        <div class="alert">
          Safety: EVSE testing involves mains voltage/current. Use PEAKMETER PM701 according to its manual,
          local electrical regulations, and your site safety procedures.
        </div>
      </div>
    </section>

    <section class="card">
      <h2>System Stats</h2>
      <div class="grid-3">
        <div class="stat"><div class="k">Total Charge Points</div><div class="v" id="st-total-cp">0</div></div>
        <div class="stat"><div class="k">Online Charge Points</div><div class="v" id="st-online-cp">0</div></div>
        <div class="stat"><div class="k">Total Connectors</div><div class="v" id="st-total-connectors">0</div></div>
        <div class="stat"><div class="k">Active Sessions</div><div class="v" id="st-active-sessions">0</div></div>
        <div class="stat"><div class="k">Total Sessions</div><div class="v" id="st-total-sessions">0</div></div>
        <div class="stat"><div class="k">Timeline Events</div><div class="v" id="st-total-events">0</div></div>
      </div>
    </section>

    <section class="grid-2">
      <div class="card">
        <h2>Charge Points and Connectors</h2>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Charge Point</th>
                <th>Connection</th>
                <th>Registration</th>
                <th>Connector State</th>
                <th>Next Allowed Statuses</th>
              </tr>
            </thead>
            <tbody id="cp-table"></tbody>
          </table>
        </div>
      </div>

      <div class="card">
        <h2>Active Sessions</h2>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Tx</th>
                <th>CP</th>
                <th>Conn</th>
                <th>User</th>
                <th>Energy Wh</th>
                <th>Duration s</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody id="session-table"></tbody>
          </table>
        </div>
      </div>
    </section>

    <section class="card">
      <h2>Detailed History Timeline</h2>
      <div class="timeline-controls">
        <div class="c3">
          <label>Source filter</label>
          <select id="filter-source" onchange="renderTimeline()">
            <option value="all">All</option>
            <option value="api">API</option>
            <option value="ocpp">OCPP</option>
            <option value="system">System</option>
          </select>
        </div>
        <div class="c3">
          <label>Level filter</label>
          <select id="filter-level" onchange="renderTimeline()">
            <option value="all">All</option>
            <option value="info">Info</option>
            <option value="warn">Warn</option>
            <option value="error">Error</option>
          </select>
        </div>
        <div class="c3">
          <label>Fetch limit</label>
          <select id="event-limit" onchange="refreshEvents()">
            <option>100</option>
            <option selected>200</option>
            <option>400</option>
            <option>800</option>
          </select>
        </div>
        <div class="c3">
          <label>Actions</label>
          <div class="actions" style="margin-top:0;">
            <button class="secondary" onclick="refreshEvents()">Refresh Events</button>
          </div>
        </div>
      </div>

      <div class="timeline-table">
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Source</th>
              <th>Level</th>
              <th>Action</th>
              <th>CP/Conn/Tx</th>
              <th>Message</th>
              <th>Details</th>
            </tr>
          </thead>
          <tbody id="timeline-table"></tbody>
        </table>
      </div>
    </section>

    <section class="card">
      <h2>PEAKMETER PM701 EVSE Tester Playbook</h2>
      <div class="pm-grid">
        <div class="pm-card">
          <h3>Per-Connector Test Flow</h3>
          <ol>
            <li>Select charge point and connector in Control Center.</li>
            <li>Connect PM701 to the target connector and set expected test mode on PM701.</li>
            <li>Confirm timeline shows connector status from station (Available/Preparing/etc.).</li>
            <li>Run <strong>Happy path</strong> scenario to send RemoteStart, read energy, then RemoteStop.</li>
            <li>Verify timeline received <span class="mono">StartTransaction</span>, meter updates, then <span class="mono">StopTransaction</span>.</li>
            <li>Repeat the same flow for every connector and compare behavior.</li>
          </ol>
        </div>
        <div class="pm-card">
          <h3>What to Validate with PM701 + OCPP</h3>
          <ul>
            <li>Pilot/control-state changes align with OCPP <span class="mono">StatusNotification</span>.</li>
            <li>Session start command triggers actual charging state at tester side.</li>
            <li>Meter values trend upward during charging period (energy Wh increases).</li>
            <li>Remote stop ends session and connector returns to available/idle state.</li>
            <li>Error scenarios (wrong connector / unknown CP) fail safely and are clear in timeline.</li>
          </ul>
        </div>
      </div>
      <p class="footer-note" style="margin-top:10px;">
        Treat this dashboard as a practical test framework: choose scenario, execute, inspect event trail, and
        keep notes per connector for repeatability.
      </p>
    </section>

    <section class="card">
      <h2>Local Action Log</h2>
      <pre id="local-log" class="terminal"></pre>
    </section>
  </div>

  <script>
    var app = {
      state: null,
      events: [],
      refreshTimer: null
    };

    function nowISO() {
      return new Date().toISOString();
    }

    function escapeHtml(value) {
      return String(value)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/\"/g, "&quot;")
        .replace(/'/g, "&#39;");
    }

    function localLog(text, payload) {
      var node = document.getElementById("local-log");
      var line = "[" + nowISO() + "] " + text;
      if (payload !== undefined) {
        try {
          line += "\n" + JSON.stringify(payload, null, 2);
        } catch (e) {
          line += "\n" + String(payload);
        }
      }
      node.textContent += line + "\n\n";
      node.scrollTop = node.scrollHeight;
    }

    function clearLocalLog() {
      document.getElementById("local-log").textContent = "";
    }

    function baseUrl() {
      var val = document.getElementById("api-base").value.trim();
      if (val.endsWith("/")) {
        val = val.slice(0, -1);
      }
      return val;
    }

    function activeCpId() {
      return document.getElementById("cp-select").value;
    }

    function activeConnectorId() {
      var raw = Number(document.getElementById("connector-select").value);
      if (!raw || raw < 1) {
        return 0;
      }
      return raw;
    }

    function activeUserId() {
      return document.getElementById("user-id").value.trim();
    }

    function activeAmount() {
      return Number(document.getElementById("amount-paid").value);
    }

    function activeTxId() {
      return Number(document.getElementById("tx-id").value);
    }

    function setTxId(v) {
      document.getElementById("tx-id").value = v ? String(v) : "";
    }

    function setStatusDot(id, kind) {
      var node = document.getElementById(id);
      node.className = "dot" + (kind ? " " + kind : "");
    }

    async function fetchJSON(method, path, body) {
      var opts = {
        method: method,
        headers: { "Content-Type": "application/json" }
      };
      if (body !== undefined) {
        opts.body = JSON.stringify(body);
      }

      var response = await fetch(baseUrl() + path, opts);
      var data = null;
      try {
        data = await response.json();
      } catch (e) {
        data = { success: false, error: "invalid-json", userMessage: "invalid JSON" };
      }
      return { status: response.status, data: data };
    }

    function normalizeStats(raw) {
      raw = raw || {};
      return {
        totalChargePoints: raw.TotalChargePoints || raw.totalChargePoints || 0,
        onlineChargePoints: raw.OnlineChargePoints || raw.onlineChargePoints || 0,
        totalConnectors: raw.TotalConnectors || raw.totalConnectors || 0,
        activeSessions: raw.ActiveSessions || raw.activeSessions || 0,
        totalSessions: raw.TotalSessions || raw.totalSessions || 0,
        totalEvents: raw.TotalEvents || raw.totalEvents || 0
      };
    }

    function buildConnectorOptions(connectors, selected) {
      var opts = [];
      for (var i = 0; i < connectors.length; i++) {
        var c = connectors[i];
        var sel = Number(c.connectorId) === Number(selected) ? " selected" : "";
        opts.push("<option value='" + c.connectorId + "'" + sel + ">#" + c.connectorId + " (" + escapeHtml(c.status) + ")</option>");
      }
      return opts.join("");
    }

    function applySelectorsFromState(state) {
      var cpSelect = document.getElementById("cp-select");
      var connectorSelect = document.getElementById("connector-select");
      var prevCP = cpSelect.value;
      var prevConnector = Number(connectorSelect.value || "1");

      var cps = state.chargePoints || [];
      if (cps.length === 0) {
        cpSelect.innerHTML = "<option value='my-test'>my-test</option>";
        connectorSelect.innerHTML = "<option value='0'>Not discovered yet</option>";
        return;
      }

      var cpOpts = [];
      for (var i = 0; i < cps.length; i++) {
        var cp = cps[i];
        var selected = cp.id === prevCP ? " selected" : "";
        cpOpts.push("<option value='" + escapeHtml(cp.id) + "'" + selected + ">" + escapeHtml(cp.id) + "</option>");
      }
      cpSelect.innerHTML = cpOpts.join("");

      if (!cpSelect.value && cps.length > 0) {
        cpSelect.value = cps[0].id;
      }

      var selectedCP = null;
      for (var j = 0; j < cps.length; j++) {
        if (cps[j].id === cpSelect.value) {
          selectedCP = cps[j];
          break;
        }
      }

      if (!selectedCP) {
        selectedCP = cps[0];
      }

      var connectors = selectedCP && selectedCP.connectors ? selectedCP.connectors : [];
      if (connectors.length === 0) {
        connectorSelect.innerHTML = "<option value='0'>Not discovered yet</option>";
      } else {
        connectorSelect.innerHTML = buildConnectorOptions(connectors, prevConnector);
      }

      if (!connectorSelect.value && connectors.length > 0) {
        connectorSelect.value = String(connectors[0].connectorId);
      }

      document.getElementById("active-cp-chip").textContent = cpSelect.value || "-";
      document.getElementById("active-connector-chip").textContent = connectorSelect.value || "-";

      var host = window.location.hostname || "127.0.0.1";
      var cpPart = cpSelect.value || "my-test";
      document.getElementById("ocpp-url").textContent = "ws://" + host + ":9005/" + cpPart;
    }

    function renderChargePoints(state) {
      var tbody = document.getElementById("cp-table");
      var cps = state.chargePoints || [];

      if (cps.length === 0) {
        tbody.innerHTML = "<tr><td colspan='5'>No charge points available</td></tr>";
        return;
      }

      var rows = [];
      for (var i = 0; i < cps.length; i++) {
        var cp = cps[i];
        var connTag = cp.connected ? "<span class='tag good'>online</span>" : "<span class='tag bad'>offline</span>";
        var regClass = cp.registrationStatus === "accepted" ? "good" : "warn";
        var regTag = "<span class='tag " + regClass + "'>" + escapeHtml(cp.registrationStatus) + "</span>";

        var connectorChips = [];
        var nextStatusText = [];
        for (var j = 0; j < (cp.connectors || []).length; j++) {
          var c = cp.connectors[j];
          var stateClass = "";
          if (c.status === "Charging") {
            stateClass = " good";
          } else if (c.status === "Faulted" || c.status === "Unavailable") {
            stateClass = " bad";
          }
          connectorChips.push("<span class='tag" + stateClass + "'>#" + c.connectorId + " " + escapeHtml(c.status) + "</span>");
          nextStatusText.push("<div><strong>#" + c.connectorId + "</strong>: " + escapeHtml((c.allowedNextStatuses || []).join(", ")) + "</div>");
        }

        rows.push(
          "<tr>" +
            "<td><strong>" + escapeHtml(cp.id) + "</strong><div class='muted'>" + escapeHtml(cp.vendor || "-") + " " + escapeHtml(cp.model || "") + "</div></td>" +
            "<td>" + connTag + "</td>" +
            "<td>" + regTag + "</td>" +
            "<td>" + (connectorChips.join("") || "-") + "</td>" +
            "<td>" + (nextStatusText.join("") || "-") + "</td>" +
          "</tr>"
        );
      }

      tbody.innerHTML = rows.join("");
    }

    function renderSessions(state) {
      var tbody = document.getElementById("session-table");
      var sessions = state.activeSessions || [];
      if (sessions.length === 0) {
        tbody.innerHTML = "<tr><td colspan='7'>No active sessions</td></tr>";
        return;
      }

      var rows = [];
      for (var i = 0; i < sessions.length; i++) {
        var s = sessions[i];
        rows.push(
          "<tr>" +
            "<td>" + s.transactionId + "</td>" +
            "<td>" + escapeHtml(s.chargePointId) + "</td>" +
            "<td>#" + s.connectorId + "</td>" +
            "<td>" + escapeHtml(s.userId) + "</td>" +
            "<td>" + Number(s.currentEnergyWh || 0).toFixed(0) + "</td>" +
            "<td>" + Number(s.durationSeconds || 0) + "</td>" +
            "<td><span class='tag good'>" + escapeHtml(s.status || "active") + "</span></td>" +
          "</tr>"
        );
      }
      tbody.innerHTML = rows.join("");
    }

    function renderHints(state) {
      var hints = [];
      var cps = state.chargePoints || [];
      var connected = state.connectedChargePoints || [];
      var activeCP = activeCpId();
      var selectedCP = null;

      for (var i = 0; i < cps.length; i++) {
        if (cps[i].id === activeCP) {
          selectedCP = cps[i];
          break;
        }
      }

      if (connected.length === 0) {
        hints.push("No station connected right now. Configure charger URL exactly as shown and wait for BootNotification in timeline.");
      }

      if (selectedCP && !selectedCP.connected) {
        hints.push("Selected charge point is offline. You can still run error scenarios, but start/stop will fail until it connects.");
      }

      if (selectedCP && (!selectedCP.connectors || selectedCP.connectors.length === 0)) {
        hints.push("No connectors discovered yet. Click 'Sync Connectors From Charger' after the station is connected.");
      }

      if (selectedCP && selectedCP.registrationStatus !== "accepted") {
        hints.push("Charge point registration is not accepted. Remote operations may fail until registration status is accepted.");
      }

      if ((state.activeSessions || []).length === 0) {
        hints.push("Start with the 'Happy path' scenario to verify full command-response cycle.");
      }

      var hasMeterValues = false;
      for (var j = 0; j < app.events.length; j++) {
        if (app.events[j].action === "MeterValues") {
          hasMeterValues = true;
          break;
        }
      }
      if (!hasMeterValues) {
        hints.push("No MeterValues seen yet. After starting a session, check station configuration to send periodic MeterValues.");
      }

      hints.push("For PM701 testing: repeat the same scenario per connector and compare status transitions and energy behavior.");

      var items = [];
      for (var k = 0; k < hints.length; k++) {
        items.push("<li>" + escapeHtml(hints[k]) + "</li>");
      }
      document.getElementById("hints").innerHTML = items.join("");
    }

    function normalizeEvents(raw) {
      if (!raw || !Array.isArray(raw)) {
        return [];
      }
      return raw;
    }

    function renderTimeline() {
      var sourceFilter = document.getElementById("filter-source").value;
      var levelFilter = document.getElementById("filter-level").value;
      var tbody = document.getElementById("timeline-table");

      var rows = [];
      for (var i = 0; i < app.events.length; i++) {
        var ev = app.events[i];
        if (sourceFilter !== "all" && String(ev.source || "") !== sourceFilter) {
          continue;
        }
        if (levelFilter !== "all" && String(ev.level || "") !== levelFilter) {
          continue;
        }

        var cls = "level-" + (ev.level || "info");
        var cpConnTx = (ev.chargePointId || "-") +
          " / " + (ev.connectorId !== null && ev.connectorId !== undefined ? ev.connectorId : "-") +
          " / " + (ev.transactionId !== null && ev.transactionId !== undefined ? ev.transactionId : "-");

        var detailsText = "-";
        if (ev.details) {
          try {
            detailsText = JSON.stringify(ev.details);
          } catch (e) {
            detailsText = String(ev.details);
          }
        }

        rows.push(
          "<tr class='" + cls + "'>" +
            "<td class='mono'>" + escapeHtml(ev.timestamp || "") + "</td>" +
            "<td>" + escapeHtml(ev.source || "") + "</td>" +
            "<td>" + escapeHtml(ev.level || "") + "</td>" +
            "<td class='mono'>" + escapeHtml(ev.action || "") + "</td>" +
            "<td class='mono'>" + escapeHtml(cpConnTx) + "</td>" +
            "<td>" + escapeHtml(ev.message || "") + "</td>" +
            "<td class='mono'>" + escapeHtml(detailsText) + "</td>" +
          "</tr>"
        );
      }

      if (rows.length === 0) {
        rows.push("<tr><td colspan='7'>No timeline items for current filter</td></tr>");
      }

      tbody.innerHTML = rows.join("");
    }

    function renderScenarioDescription() {
      var id = document.getElementById("scenario-select").value;
      var msg = "";

      if (id === "happy_cycle") {
        msg = "Starts one session on selected connector, reads energy once, then stops. Use first when validating full backend loop.";
      } else if (id === "start_poll_stop") {
        msg = "Starts session and polls energy multiple times to observe meter updates before stopping.";
      } else if (id === "already_active") {
        msg = "Attempts second start on same connector while first session is active. Useful to validate safety/error handling.";
      } else if (id === "wrong_connector") {
        msg = "Attempts start on invalid connector number to confirm connector validation and clear API errors.";
      } else if (id === "unknown_cp") {
        msg = "Attempts start on unknown charge point ID to verify DB/config mismatch behavior.";
      }

      document.getElementById("scenario-description").textContent = msg;
    }

    async function refreshState(silent) {
      var result = await fetchJSON("GET", "/state");
      if (!result.data || !result.data.success) {
        setStatusDot("dot-api", "bad");
        document.getElementById("api-status").textContent = "error";
        if (!silent) {
          localLog("State refresh failed", result.data);
        }
        return;
      }

      app.state = result.data;
      setStatusDot("dot-api", "good");
      document.getElementById("api-status").textContent = "ok";

      var connectedCount = (result.data.connectedChargePoints || []).length;
      setStatusDot("dot-ws", connectedCount > 0 ? "good" : "warn");
      document.getElementById("ws-status").textContent = connectedCount > 0 ? "connected" : "no station";

      var stats = normalizeStats(result.data.stats);
      document.getElementById("st-total-cp").textContent = String(stats.totalChargePoints);
      document.getElementById("st-online-cp").textContent = String(stats.onlineChargePoints);
      document.getElementById("st-total-connectors").textContent = String(stats.totalConnectors);
      document.getElementById("st-active-sessions").textContent = String(stats.activeSessions);
      document.getElementById("st-total-sessions").textContent = String(stats.totalSessions);
      document.getElementById("st-total-events").textContent = String(stats.totalEvents);

      applySelectorsFromState(result.data);
      renderChargePoints(result.data);
      renderSessions(result.data);
      renderHints(result.data);

      if (Array.isArray(result.data.recentEvents) && result.data.recentEvents.length > 0) {
        app.events = result.data.recentEvents;
        renderTimeline();
      }
    }

    async function refreshEvents() {
      var limit = Number(document.getElementById("event-limit").value || "200");
      var result = await fetchJSON("GET", "/events?limit=" + encodeURIComponent(limit));
      if (!result.data || !result.data.success) {
        localLog("Event refresh failed", result.data);
        return;
      }
      app.events = normalizeEvents(result.data.events);
      renderTimeline();
    }

    async function startSession(custom) {
      var payload = custom || {
        chargePointId: activeCpId(),
        connectorId: activeConnectorId(),
        userId: activeUserId(),
        amountPaid: activeAmount()
      };
      if (!payload.connectorId || payload.connectorId < 1) {
        var msg = "Connector ID is unknown (0). Click 'Sync Connectors From Charger' and select a real connector ID.";
        localLog("Start skipped: " + msg);
        window.alert(msg);
        return {
          status: 0,
          data: {
            success: false,
            userMessage: msg
          }
        };
      }

      localLog("POST /sessions/start", payload);
      var result = await fetchJSON("POST", "/sessions/start", payload);
      localLog("Response " + result.status, result.data);

      if (result.data && result.data.transactionId) {
        setTxId(result.data.transactionId);
      }

      await refreshAll(true);
      return result;
    }

    async function syncConnectors() {
      var cp = activeCpId();
      if (!cp) {
        localLog("Sync skipped: no charge point selected");
        return;
      }
      var path = "/charge-points/" + encodeURIComponent(cp) + "/sync-connectors";
      localLog("POST " + path);
      var result = await fetchJSON("POST", path);
      localLog("Response " + result.status, result.data);
      await refreshAll(true);
      return result;
    }

    async function getEnergy(customTx, customCP) {
      var tx = customTx || activeTxId();
      var cp = customCP || activeCpId();
      if (!tx) {
        localLog("Get energy skipped: transaction ID is empty");
        return null;
      }

      var path = "/sessions/" + tx + "/energy?chargePointId=" + encodeURIComponent(cp);
      localLog("GET " + path);
      var result = await fetchJSON("GET", path);
      localLog("Response " + result.status, result.data);
      await refreshAll(true);
      return result;
    }

    async function stopSession(customTx, customCP) {
      var tx = customTx || activeTxId();
      var cp = customCP || activeCpId();
      if (!tx) {
        localLog("Stop skipped: transaction ID is empty");
        return null;
      }

      var path = "/sessions/" + tx + "/stop?chargePointId=" + encodeURIComponent(cp);
      localLog("POST " + path);
      var result = await fetchJSON("POST", path);
      localLog("Response " + result.status, result.data);
      await refreshAll(true);
      return result;
    }

    function sleep(ms) {
      return new Promise(function(resolve) {
        setTimeout(resolve, ms);
      });
    }

    async function runScenario() {
      var scenario = document.getElementById("scenario-select").value;
      localLog("Scenario started: " + scenario);

      if (scenario === "happy_cycle") {
        var started = await startSession();
        if (!started || !started.data || !started.data.success) {
          localLog("Scenario aborted: start failed");
          return;
        }
        await sleep(3000);
        await getEnergy();
        await sleep(1200);
        await stopSession();
      } else if (scenario === "start_poll_stop") {
        var started2 = await startSession();
        if (!started2 || !started2.data || !started2.data.success) {
          localLog("Scenario aborted: start failed");
          return;
        }
        var tx = started2.data.transactionId;
        var cp = started2.data.chargePointId || activeCpId();

        await sleep(2500);
        await getEnergy(tx, cp);
        await sleep(2500);
        await getEnergy(tx, cp);
        await sleep(2500);
        await getEnergy(tx, cp);
        await stopSession(tx, cp);
      } else if (scenario === "already_active") {
        var first = await startSession();
        if (!first || !first.data || !first.data.success) {
          localLog("Scenario aborted: first start failed");
          return;
        }
        await sleep(1200);
        await startSession({
          chargePointId: activeCpId(),
          connectorId: activeConnectorId(),
          userId: activeUserId(),
          amountPaid: activeAmount()
        });
        await sleep(1000);
        await stopSession(first.data.transactionId, first.data.chargePointId || activeCpId());
      } else if (scenario === "wrong_connector") {
        await startSession({
          chargePointId: activeCpId(),
          connectorId: 999,
          userId: activeUserId(),
          amountPaid: activeAmount()
        });
      } else if (scenario === "unknown_cp") {
        await startSession({
          chargePointId: "unknown-cp-lab",
          connectorId: 1,
          userId: activeUserId(),
          amountPaid: activeAmount()
        });
      }

      localLog("Scenario finished: " + scenario);
      await refreshAll(true);
    }

    async function refreshAll(silent) {
      await refreshState(!!silent);
      await refreshEvents();
    }

    function startAutoRefresh() {
      if (app.refreshTimer) {
        clearInterval(app.refreshTimer);
      }
      app.refreshTimer = setInterval(function() {
        refreshAll(true);
      }, 4000);
    }

    function bindUI() {
      document.getElementById("cp-select").addEventListener("change", function() {
        if (app.state) {
          applySelectorsFromState(app.state);
          renderHints(app.state);
        }
      });

      document.getElementById("connector-select").addEventListener("change", function() {
        document.getElementById("active-connector-chip").textContent = document.getElementById("connector-select").value || "-";
      });
    }

    (function init() {
      document.getElementById("api-base").value = window.location.origin;
      document.getElementById("cp-select").innerHTML = "<option value='my-test'>my-test</option>";
      document.getElementById("connector-select").innerHTML = "<option value='0'>Not discovered yet</option>";

      renderScenarioDescription();
      bindUI();
      refreshAll(false);
      startAutoRefresh();
      localLog("Dashboard initialized. Start by connecting charger to ws://<backend-ip>:9005/my-test");
    })();
  </script>
</body>
</html>`

package main

import (
	"html/template"
	"sort"
	"strings"
	"time"
)

type dashboardView struct {
	Attempts []Attempt
}

var dashboardTmpl = template.Must(template.New("dashboard").Funcs(template.FuncMap{
	"formatTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	},
	"joinHeaders": func(headers map[string]string) string {
		keys := make([]string, 0, len(headers))
		for key := range headers {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, key+": "+headers[key])
		}
		return strings.Join(parts, "\n")
	},
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Bypass Playground Dashboard</title>
  <style>
    :root {
      --bg: #111111;
      --panel: #181818;
      --panel-2: #1f1f1f;
      --panel-3: #252526;
      --ink: #d4d4d4;
      --muted: #9da1a6;
      --line: #2d2d30;
      --accent: #569cd6;
      --accent-2: #4ec9b0;
      --ok: #73c991;
      --warn: #d7ba7d;
      --bad: #f48771;
      --shadow: rgba(0, 0, 0, 0.35);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      padding: 20px;
      background:
        radial-gradient(circle at top left, rgba(86, 156, 214, 0.08), transparent 24%),
        radial-gradient(circle at bottom right, rgba(78, 201, 176, 0.06), transparent 22%),
        var(--bg);
      color: var(--ink);
      font-family: "Segoe UI", "Helvetica Neue", Arial, sans-serif;
    }
    .shell {
      max-width: 1480px;
      margin: 0 auto;
    }
    .hero {
      display: flex;
      justify-content: space-between;
      align-items: end;
      gap: 16px;
      margin-bottom: 16px;
      padding: 20px 22px;
      border: 1px solid var(--line);
      background: var(--panel);
      box-shadow: 0 16px 40px var(--shadow);
      border-radius: 12px;
    }
    h1 {
      margin: 0;
      font-size: 1.45rem;
      font-weight: 600;
      letter-spacing: 0.02em;
    }
    p {
      margin: 6px 0 0;
      color: var(--muted);
      line-height: 1.45;
      max-width: 780px;
    }
    .grid {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 12px;
      margin-bottom: 16px;
    }
    .stat {
      padding: 16px 18px;
      border: 1px solid var(--line);
      background: linear-gradient(180deg, var(--panel-2), var(--panel));
      border-radius: 12px;
      box-shadow: 0 10px 24px var(--shadow);
    }
    .stat strong {
      display: block;
      font-size: 1.35rem;
      margin-bottom: 4px;
    }
    .filters {
      margin-bottom: 16px;
      padding: 16px 18px;
      border: 1px solid var(--line);
      background: var(--panel);
      border-radius: 12px;
      box-shadow: 0 10px 24px var(--shadow);
    }
    .filters-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      margin-bottom: 14px;
    }
    .filters-actions {
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .filters-grid {
      display: grid;
      grid-template-columns: 2fr repeat(4, minmax(0, 1fr)) auto;
      gap: 10px;
    }
    .field {
      display: grid;
      gap: 6px;
    }
    .field label {
      font-size: 0.78rem;
      color: var(--muted);
      text-transform: uppercase;
      letter-spacing: 0.06em;
    }
    input, select, button {
      width: 100%;
      min-height: 40px;
      border-radius: 8px;
      border: 1px solid #3c3c3c;
      background: var(--panel-3);
      color: var(--ink);
      padding: 0 12px;
      font-size: 0.95rem;
      outline: none;
    }
    input:focus, select:focus {
      border-color: var(--accent);
      box-shadow: 0 0 0 1px var(--accent);
    }
    button {
      cursor: pointer;
      background: #0e639c;
      border-color: #1177bb;
      font-weight: 600;
      align-self: end;
    }
    button:hover {
      background: #1177bb;
    }
    .button-secondary {
      background: #2a2d2e;
      border-color: #3c3c3c;
    }
    .button-secondary:hover {
      background: #33373a;
    }
    .table-meta {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 12px;
      margin-bottom: 10px;
      color: var(--muted);
      font-size: 0.9rem;
    }
    .table-wrap {
      overflow: auto;
      border: 1px solid var(--line);
      background: var(--panel);
      box-shadow: 0 16px 40px var(--shadow);
      border-radius: 12px;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      min-width: 980px;
    }
    th, td {
      padding: 12px;
      border-bottom: 1px solid var(--line);
      vertical-align: top;
      text-align: left;
      font-size: 0.92rem;
    }
    th {
      position: sticky;
      top: 0;
      background: #202024;
      z-index: 1;
      color: #c5c5c5;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      font-size: 0.76rem;
    }
    .code {
      font-family: Consolas, "Courier New", monospace;
      white-space: pre-wrap;
      word-break: break-word;
      font-size: 0.85rem;
    }
    .request-cell {
      min-width: 300px;
    }
    .ua-cell {
      min-width: 210px;
      max-width: 320px;
    }
    .signal-cell {
      min-width: 260px;
      max-width: 380px;
    }
    .pill {
      display: inline-block;
      padding: 4px 9px;
      border-radius: 999px;
      font-size: 0.8rem;
      font-weight: 700;
      color: #111;
    }
    .s2xx { background: var(--ok); }
    .s3xx { background: var(--accent); }
    .s4xx { background: var(--warn); }
    .s5xx { background: var(--bad); }
    .muted { color: var(--muted); }
    .hero-badge {
      padding: 7px 10px;
      border-radius: 999px;
      background: rgba(86, 156, 214, 0.12);
      color: #9cdcfe;
      border: 1px solid rgba(86, 156, 214, 0.28);
      white-space: nowrap;
      font-size: 0.86rem;
    }
    tbody tr:hover {
      background: rgba(255, 255, 255, 0.025);
    }
    .empty {
      padding: 28px;
      text-align: center;
      color: var(--muted);
    }
    @media (max-width: 900px) {
      body { padding: 12px; }
      .hero {
        display: grid;
        align-items: start;
      }
      .grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
      .filters-grid { grid-template-columns: 1fr; }
      .table-meta { display: grid; }
    }
  </style>
</head>
<body>
  <div class="shell">
    <section class="hero">
      <div>
        <h1>Bypass Playground Dashboard</h1>
        <p>Live request log for the local lab. Filter by status, route, method, path, user agent, or bypass trigger to quickly isolate the traffic pattern you care about.</p>
      </div>
      <div class="hero-badge">Live Lab Traffic</div>
    </section>

    <section class="grid">
      <div class="stat"><strong id="storedCount">{{len .Attempts}}</strong><span class="muted">Attempts currently stored</span></div>
      <div class="stat"><strong>/dashboard</strong><span class="muted">Human-readable dashboard</span></div>
      <div class="stat"><strong>/api/attempts</strong><span class="muted">JSON feed for automation</span></div>
      <div class="stat"><strong>Instant filters</strong><span class="muted">Client-side, no page reload needed</span></div>
    </section>

    <section class="filters">
      <div class="filters-head">
        <strong>Filters</strong>
        <div class="filters-actions">
          <span class="muted" id="resultCount">Showing {{len .Attempts}} of {{len .Attempts}} attempts</span>
          <button id="clearLogs" class="button-secondary" type="button">Clear Logs</button>
        </div>
      </div>
      <div class="filters-grid">
        <div class="field">
          <label for="searchFilter">Search</label>
          <input id="searchFilter" type="text" placeholder="Path, header, user-agent, host, IP">
        </div>
        <div class="field">
          <label for="methodFilter">Method</label>
          <select id="methodFilter">
            <option value="">All methods</option>
          </select>
        </div>
        <div class="field">
          <label for="statusFilter">Status</label>
          <select id="statusFilter">
            <option value="">All statuses</option>
            <option value="2">2xx</option>
            <option value="3">3xx</option>
            <option value="4">4xx</option>
            <option value="5">5xx</option>
          </select>
        </div>
        <div class="field">
          <label for="routeFilter">Route</label>
          <select id="routeFilter">
            <option value="">All route families</option>
          </select>
        </div>
        <div class="field">
          <label for="outcomeFilter">Outcome</label>
          <select id="outcomeFilter">
            <option value="">Any outcome</option>
            <option value="blocked">blocked</option>
          </select>
        </div>
        <button id="resetFilters" type="button">Reset</button>
      </div>
    </section>

    <div class="table-meta">
      <span>Newest attempts first. The table stays focused on the path tested, response status, user-agent, and the request signals most likely to affect access control.</span>
      <span class="muted">Use search for specific headers, paths, or user-agents when you need to drill in.</span>
    </div>

    <section class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Time</th>
            <th>Method</th>
            <th>Request</th>
            <th>Status</th>
            <th>Outcome</th>
            <th>User-Agent</th>
            <th>Signals</th>
          </tr>
        </thead>
        <tbody id="attemptTableBody">
          {{range .Attempts}}
          <tr
            data-id="{{.ID}}"
            data-method="{{.Method}}"
            data-status="{{.StatusCode}}"
            data-route="{{.Route}}"
            data-outcome="{{.Outcome}}"
            data-search="{{.Host}} {{.RequestURI}} {{.RemoteAddr}} {{.UserAgent}} {{joinHeaders .Headers}} {{.Route}} {{.Outcome}}">
            <td>{{formatTime .Time}}</td>
            <td>{{.Method}}</td>
            <td class="code request-cell">{{.Host}}{{.RequestURI}}&#10;<span class="muted">{{.RemoteAddr}}</span></td>
            <td>
              <span class="pill {{if and (ge .StatusCode 200) (lt .StatusCode 300)}}s2xx{{else if and (ge .StatusCode 300) (lt .StatusCode 400)}}s3xx{{else if and (ge .StatusCode 400) (lt .StatusCode 500)}}s4xx{{else}}s5xx{{end}}">
                {{.StatusCode}}
              </span>
            </td>
            <td>{{.Outcome}}</td>
            <td class="code ua-cell">{{.UserAgent}}</td>
            <td class="code signal-cell">{{joinHeaders .Headers}}</td>
          </tr>
          {{else}}
          <tr>
            <td colspan="7" class="empty">No traffic yet. Send a request to <span class="code">/admin</span> and the dashboard will populate automatically.</td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </section>
  </div>
  <script>
    (function () {
      const searchFilter = document.getElementById("searchFilter");
      const methodFilter = document.getElementById("methodFilter");
      const statusFilter = document.getElementById("statusFilter");
      const routeFilter = document.getElementById("routeFilter");
      const outcomeFilter = document.getElementById("outcomeFilter");
      const resetFilters = document.getElementById("resetFilters");
      const clearLogs = document.getElementById("clearLogs");
      const resultCount = document.getElementById("resultCount");
      const storedCount = document.getElementById("storedCount");
      const tableBody = document.getElementById("attemptTableBody");
      let rows = Array.from(document.querySelectorAll("#attemptTableBody tr[data-method]"));
      let lastSeenID = rows.length ? Math.max(...rows.map((row) => Number(row.dataset.id) || 0)) : 0;

      function escapeHtml(value) {
        return String(value || "")
          .replace(/&/g, "&amp;")
          .replace(/</g, "&lt;")
          .replace(/>/g, "&gt;")
          .replace(/"/g, "&quot;")
          .replace(/'/g, "&#39;");
      }

      function joinHeaders(headers) {
        return Object.keys(headers || {})
          .sort()
          .map((key) => key + ": " + headers[key])
          .join("\n");
      }

      function statusClass(status) {
        if (status >= 200 && status < 300) return "s2xx";
        if (status >= 300 && status < 400) return "s3xx";
        if (status >= 400 && status < 500) return "s4xx";
        return "s5xx";
      }

      function formatTime(value) {
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return value;
        return date.getFullYear()
          + "-" + String(date.getMonth() + 1).padStart(2, "0")
          + "-" + String(date.getDate()).padStart(2, "0")
          + " " + String(date.getHours()).padStart(2, "0")
          + ":" + String(date.getMinutes()).padStart(2, "0")
          + ":" + String(date.getSeconds()).padStart(2, "0");
      }

      function populateOptions(select, values) {
        const currentValue = select.value;
        const baseOptions = Array.from(select.querySelectorAll("option")).slice(0, select.dataset.baseCount || 1);
        select.innerHTML = "";
        baseOptions.forEach((option) => select.appendChild(option.cloneNode(true)));
        values.forEach((value) => {
          const option = document.createElement("option");
          option.value = value;
          option.textContent = value;
          select.appendChild(option);
        });
        if ([...select.options].some((option) => option.value === currentValue)) {
          select.value = currentValue;
        }
      }

      [methodFilter, routeFilter, outcomeFilter].forEach((select) => {
        select.dataset.baseCount = String(select.options.length);
      });

      function refreshSelectOptions() {
        populateOptions(methodFilter, [...new Set(rows.map((row) => row.dataset.method))].sort());
        populateOptions(routeFilter, [...new Set(rows.map((row) => row.dataset.route))].sort());
        populateOptions(
          outcomeFilter,
          [...new Set(rows.map((row) => row.dataset.outcome))].sort().filter((value) => value !== "blocked")
        );
      }

      function buildRowMarkup(attempt) {
        const outcome = attempt.Outcome || "blocked";
        const headers = joinHeaders(attempt.Headers);
        const search = [
          attempt.Host,
          attempt.RequestURI,
          attempt.RemoteAddr,
          attempt.UserAgent,
          headers,
          attempt.Route,
          outcome
        ].join(" ");

        return '<tr'
          + ' data-id="' + escapeHtml(attempt.ID) + '"'
          + ' data-method="' + escapeHtml(attempt.Method) + '"'
          + ' data-status="' + escapeHtml(attempt.StatusCode) + '"'
          + ' data-route="' + escapeHtml(attempt.Route) + '"'
          + ' data-outcome="' + escapeHtml(outcome) + '"'
          + ' data-search="' + escapeHtml(search) + '">'
          + '<td>' + escapeHtml(formatTime(attempt.Time)) + '</td>'
          + '<td>' + escapeHtml(attempt.Method) + '</td>'
          + '<td class="code request-cell">' + escapeHtml((attempt.Host || "") + (attempt.RequestURI || "")) + '\n<span class="muted">' + escapeHtml(attempt.RemoteAddr) + '</span></td>'
          + '<td><span class="pill ' + statusClass(attempt.StatusCode) + '">' + escapeHtml(attempt.StatusCode) + '</span></td>'
          + '<td>' + escapeHtml(outcome) + '</td>'
          + '<td class="code ua-cell">' + escapeHtml(attempt.UserAgent) + '</td>'
          + '<td class="code signal-cell">' + escapeHtml(headers) + '</td>'
          + '</tr>';
      }

      function refreshRows() {
        rows = Array.from(document.querySelectorAll("#attemptTableBody tr[data-method]"));
        lastSeenID = rows.length ? Math.max(...rows.map((row) => Number(row.dataset.id) || 0)) : 0;
        storedCount.textContent = String(rows.length);
      }

      function renderRows(attempts) {
        if (!attempts.length) {
          tableBody.innerHTML = '<tr><td colspan="7" class="empty">No traffic yet. Send a request to <span class="code">/admin</span> and the dashboard will populate automatically.</td></tr>';
          rows = [];
          lastSeenID = 0;
          refreshSelectOptions();
          applyFilters();
          return;
        }

        tableBody.innerHTML = attempts.map(buildRowMarkup).join("");
        refreshRows();
        refreshSelectOptions();
        applyFilters();
      }

      function prependRows(attempts) {
        if (!attempts.length) {
          return;
        }

        const emptyState = tableBody.querySelector(".empty");
        if (emptyState) {
          renderRows(attempts.slice().reverse());
          return;
        }

        const fragment = document.createDocumentFragment();
        attempts.slice().reverse().forEach((attempt) => {
          const template = document.createElement("template");
          template.innerHTML = buildRowMarkup(attempt).trim();
          fragment.appendChild(template.content.firstChild);
        });
        tableBody.prepend(fragment);
        refreshRows();
        refreshSelectOptions();
        applyFilters();
      }

      function applyFilters() {
        const search = searchFilter.value.trim().toLowerCase();
        const method = methodFilter.value;
        const statusPrefix = statusFilter.value;
        const route = routeFilter.value;
        const outcome = outcomeFilter.value;

        let visible = 0;
        rows.forEach((row) => {
          const matchesSearch = !search || row.dataset.search.toLowerCase().includes(search);
          const matchesMethod = !method || row.dataset.method === method;
          const matchesStatus = !statusPrefix || row.dataset.status.startsWith(statusPrefix);
          const matchesRoute = !route || row.dataset.route === route;
          const matchesOutcome = !outcome || row.dataset.outcome === outcome;
          const show = matchesSearch && matchesMethod && matchesStatus && matchesRoute && matchesOutcome;
          row.style.display = show ? "" : "none";
          if (show) visible += 1;
        });

        resultCount.textContent = "Showing " + visible + " of " + rows.length + " attempts";
      }

      [searchFilter, methodFilter, statusFilter, routeFilter, outcomeFilter].forEach((element) => {
        element.addEventListener("input", applyFilters);
        element.addEventListener("change", applyFilters);
      });

      resetFilters.addEventListener("click", () => {
        searchFilter.value = "";
        methodFilter.value = "";
        statusFilter.value = "";
        routeFilter.value = "";
        outcomeFilter.value = "";
        applyFilters();
      });

      clearLogs.addEventListener("click", async () => {
        const confirmed = window.confirm("Clear all recorded attempts from the dashboard?");
        if (!confirmed) return;

        clearLogs.disabled = true;
        try {
          const response = await fetch("/api/attempts/clear", { method: "POST" });
          if (!response.ok) {
            throw new Error("Failed to clear logs");
          }
          renderRows([]);
          clearLogs.disabled = false;
        } catch (error) {
          window.alert("Unable to clear logs right now.");
          clearLogs.disabled = false;
        }
      });

      async function pollAttempts() {
        try {
          const response = await fetch("/api/attempts?since_id=" + encodeURIComponent(String(lastSeenID)), { cache: "no-store" });
          if (!response.ok) {
            throw new Error("Failed to fetch attempts");
          }
          const attempts = await response.json();
          prependRows(attempts);
        } catch (error) {
          console.error(error);
        }
      }

      refreshSelectOptions();
      applyFilters();
      window.setInterval(pollAttempts, 3000);
    })();
  </script>
</body>
</html>`))

document.addEventListener('DOMContentLoaded', () => {
  // DOM Elements
  const authView = document.getElementById('auth-view');
  const dashboardView = document.getElementById('dashboard-view');
  const authForm = document.getElementById('auth-form');
  const authEmailInput = document.getElementById('auth-email');
  const authPasswordInput = document.getElementById('auth-password');
  const authSubmitBtn = document.getElementById('auth-submit-btn');
  const authError = document.getElementById('auth-error');
  const tabLogin = document.getElementById('tab-login');
  const tabSignup = document.getElementById('tab-signup');
  const userEmailDisplay = document.getElementById('user-email-display');
  const authBtn = document.getElementById('auth-btn');
  const logoutBtn = document.getElementById('logout-btn');

  // Stats Elements
  const statTotal = document.getElementById('stat-total');
  const statUp = document.getElementById('stat-up');
  const statDown = document.getElementById('stat-down');
  const statSla = document.getElementById('stat-sla');

  // Monitor Elements
  const monitorGrid = document.getElementById('monitor-grid');
  const addModal = document.getElementById('add-modal');
  const openAddModalBtn = document.getElementById('open-add-modal-btn');
  const closeAddModalBtn = document.getElementById('close-add-modal-btn');
  const cancelAddBtn = document.getElementById('cancel-add-btn');
  const addMonitorForm = document.getElementById('add-monitor-form');
  const addError = document.getElementById('add-error');

  // Telemetry Modal Elements
  const telemetryModal = document.getElementById('telemetry-modal');
  const closeTelemetryModalBtn = document.getElementById('close-telemetry-modal-btn');
  const telemetryModalTitle = document.getElementById('telemetry-modal-title');
  const telemetryTimeline = document.getElementById('telemetry-timeline');
  const telemetryLogsTbody = document.getElementById('telemetry-logs-tbody');

  // App State
  let isSignupMode = false;
  let pollTimer = null;

  // Initialize App Session
  initSession();

  function getToken() {
    return localStorage.getItem('pinggopher_token');
  }

  function setSession(token, email) {
    localStorage.setItem('pinggopher_token', token);
    localStorage.setItem('pinggopher_email', email);
    initSession();
  }

  function clearSession() {
    localStorage.removeItem('pinggopher_token');
    localStorage.removeItem('pinggopher_email');
    if (pollTimer) clearInterval(pollTimer);
    initSession();
  }

  function initSession() {
    const token = getToken();
    const email = localStorage.getItem('pinggopher_email');

    if (token) {
      authView.classList.add('hidden');
      dashboardView.classList.remove('hidden');
      authBtn.classList.add('hidden');
      logoutBtn.classList.remove('hidden');
      userEmailDisplay.classList.remove('hidden');
      userEmailDisplay.textContent = email || 'Authenticated Tenant';
      
      fetchMonitors();
      if (!pollTimer) {
        pollTimer = setInterval(fetchMonitors, 10000);
      }
    } else {
      authView.classList.remove('hidden');
      dashboardView.classList.add('hidden');
      authBtn.classList.remove('hidden');
      logoutBtn.classList.add('hidden');
      userEmailDisplay.classList.add('hidden');
    }
  }

  // API Fetch Wrapper
  async function fetchAPI(endpoint, method = 'GET', body = null) {
    const token = getToken();
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = `Bearer ${token}`;

    const config = { method, headers };
    if (body) config.body = JSON.stringify(body);

    const res = await fetch(endpoint, config);
    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || 'API Request Failed');
    }
    return data;
  }

  // Auth Tab Handlers
  tabLogin.addEventListener('click', () => {
    isSignupMode = false;
    authSubmitBtn.textContent = 'Sign In';
    tabLogin.style.background = 'rgba(255, 255, 255, 0.15)';
    tabSignup.style.background = 'transparent';
    authError.classList.add('hidden');
  });

  tabSignup.addEventListener('click', () => {
    isSignupMode = true;
    authSubmitBtn.textContent = 'Create Account';
    tabSignup.style.background = 'rgba(255, 255, 255, 0.15)';
    tabLogin.style.background = 'transparent';
    authError.classList.add('hidden');
  });

  // Auth Form Submit Handler
  authForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    authError.classList.add('hidden');

    const email = authEmailInput.value.trim();
    const password = authPasswordInput.value.trim();
    const endpoint = isSignupMode ? '/v1/auth/signup' : '/v1/auth/login';

    try {
      const res = await fetchAPI(endpoint, 'POST', { email, password });
      setSession(res.data.token, res.data.user.email);
    } catch (err) {
      authError.textContent = err.message;
      authError.classList.remove('hidden');
    }
  });

  logoutBtn.addEventListener('click', clearSession);

  // Fetch Monitors & Update Dashboard
  async function fetchMonitors() {
    try {
      const res = await fetchAPI('/v1/monitors');
      const monitors = res.data || [];
      renderMonitors(monitors);
    } catch (err) {
      if (err.message.includes('Unauthorized')) {
        clearSession();
      }
    }
  }

  function renderMonitors(monitors) {
    statTotal.textContent = monitors.length;
    
    let upCount = 0;
    let downCount = 0;

    monitors.forEach(m => {
      if (m.status === 'UP') upCount++;
      if (m.status === 'DOWN') downCount++;
    });

    statUp.textContent = upCount;
    statDown.textContent = downCount;

    const totalActive = upCount + downCount;
    const sla = totalActive > 0 ? ((upCount / totalActive) * 100).toFixed(1) + '%' : '99.9%';
    statSla.textContent = sla;

    if (monitors.length === 0) {
      monitorGrid.innerHTML = `
        <div class="glass-panel" style="grid-column: 1 / -1; padding: 3rem; text-align: center; color: var(--text-muted);">
          <h3>No Monitors Configured</h3>
          <p style="margin-top: 0.5rem;">Click "+ Add New Target" above to configure synthetic uptime monitoring.</p>
        </div>
      `;
      return;
    }

    monitorGrid.innerHTML = monitors.map(m => {
      const statusClass = m.status === 'UP' ? 'badge-up' : (m.status === 'DOWN' ? 'badge-down' : 'badge-paused');
      const isPaused = m.status === 'PAUSED';

      return `
        <div class="glass-panel monitor-card">
          <div class="monitor-card-header">
            <div>
              <div class="monitor-name">${escapeHTML(m.name)}</div>
              <div class="monitor-url">${escapeHTML(m.url)}</div>
            </div>
            <span class="badge ${statusClass}">
              <span class="pulse-dot"></span>
              ${m.status}
            </span>
          </div>

          <div style="font-size: 0.85rem; color: var(--text-muted); margin-bottom: 1rem;">
            Check Interval: <strong>${m.check_interval_seconds}s</strong>
          </div>

          <div style="display: flex; gap: 0.5rem; justify-content: flex-end;">
            <button class="btn btn-secondary" onclick="viewLogs('${m.id}', '${escapeHTML(m.name)}')">Logs & SLA</button>
            <button class="btn btn-secondary" onclick="togglePause('${m.id}', '${m.status}')">${isPaused ? 'Resume' : 'Pause'}</button>
            <button class="btn btn-danger" onclick="deleteMonitorTarget('${m.id}')">Delete</button>
          </div>
        </div>
      `;
    }).join('');
  }

  // Add Monitor Modal Handlers
  openAddModalBtn.addEventListener('click', () => {
    addError.classList.add('hidden');
    addMonitorForm.reset();
    addModal.classList.add('active');
  });

  closeAddModalBtn.addEventListener('click', () => addModal.classList.remove('active'));
  cancelAddBtn.addEventListener('click', () => addModal.classList.remove('active'));

  addMonitorForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    addError.classList.add('hidden');

    const name = document.getElementById('monitor-name').value.trim();
    const url = document.getElementById('monitor-url').value.trim();
    const check_interval_seconds = parseInt(document.getElementById('monitor-interval').value);

    try {
      await fetchAPI('/v1/monitors', 'POST', { name, url, check_interval_seconds });
      addModal.classList.remove('active');
      fetchMonitors();
    } catch (err) {
      addError.textContent = err.message;
      addError.classList.remove('hidden');
    }
  });

  // Global Scope Handlers for Dynamic Cards
  window.togglePause = async (id, currentStatus) => {
    const nextStatus = currentStatus === 'PAUSED' ? 'UP' : 'PAUSED';
    try {
      await fetchAPI(`/v1/monitors/${id}`, 'PUT', { status: nextStatus });
      fetchMonitors();
    } catch (err) {
      alert('Failed to toggle monitor status: ' + err.message);
    }
  };

  window.deleteMonitorTarget = async (id) => {
    if (!confirm('Are you sure you want to delete this target monitor?')) return;
    try {
      await fetchAPI(`/v1/monitors/${id}`, 'DELETE');
      fetchMonitors();
    } catch (err) {
      alert('Failed to delete monitor: ' + err.message);
    }
  };

  window.viewLogs = async (id, name) => {
    telemetryModalTitle.textContent = `${name} — Execution Telemetry`;
    telemetryTimeline.innerHTML = '';
    telemetryLogsTbody.innerHTML = '<tr><td colspan="3" style="padding:1rem; text-align:center;">Loading logs...</td></tr>';
    telemetryModal.classList.add('active');

    try {
      const res = await fetchAPI(`/v1/monitors/${id}/logs?limit=30`);
      const logs = res.data || [];

      if (logs.length === 0) {
        telemetryTimeline.innerHTML = '<div style="color:var(--text-muted); font-size:0.85rem;">No probe execution logs recorded yet.</div>';
        telemetryLogsTbody.innerHTML = '<tr><td colspan="3" style="padding:1rem; text-align:center; color:var(--text-muted);">No records found.</td></tr>';
        return;
      }

      // Render Latency Bars
      telemetryTimeline.innerHTML = logs.slice().reverse().map(l => {
        const isUp = l.status_code >= 200 && l.status_code < 400;
        const barClass = isUp ? 'up' : 'down';
        const height = Math.min(Math.max((l.response_time_ms / 300) * 40, 6), 40);
        return `<div class="log-bar ${barClass}" style="height: ${height}px;" title="${l.status_code} - ${l.response_time_ms}ms"></div>`;
      }).join('');

      // Render Logs Table
      telemetryLogsTbody.innerHTML = logs.map(l => {
        const isUp = l.status_code >= 200 && l.status_code < 400;
        const statusColor = isUp ? 'var(--emerald)' : 'var(--rose)';
        const dateStr = new Date(l.created_at).toLocaleTimeString();
        return `
          <tr style="border-bottom: 1px solid rgba(255,255,255,0.05);">
            <td style="padding: 0.5rem; color: ${statusColor}; font-weight:600;">${l.status_code || 'ERR'}</td>
            <td style="padding: 0.5rem;">${l.response_time_ms} ms</td>
            <td style="padding: 0.5rem; color: var(--text-muted);">${dateStr}</td>
          </tr>
        `;
      }).join('');

    } catch (err) {
      telemetryLogsTbody.innerHTML = `<tr><td colspan="3" style="color:var(--rose); padding:1rem;">Failed to load logs: ${err.message}</td></tr>`;
    }
  };

  closeTelemetryModalBtn.addEventListener('click', () => telemetryModal.classList.remove('active'));

  function escapeHTML(str) {
    return str.replace(/[&<>'"]/g, 
      tag => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[tag] || tag)
    );
  }
});

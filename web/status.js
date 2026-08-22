document.addEventListener('DOMContentLoaded', () => {
  const statusBanner = document.getElementById('status-banner');
  const bannerPulse = document.getElementById('banner-pulse');
  const bannerTitle = document.getElementById('banner-title');
  const bannerSubtitle = document.getElementById('banner-subtitle');
  const componentsList = document.getElementById('status-components-list');
  const incidentsList = document.getElementById('status-incidents-list');

  fetchPublicStatus();
  setInterval(fetchPublicStatus, 15000);

  async function fetchPublicStatus() {
    try {
      const res = await fetch('/v1/status/public');
      const data = await res.json();

      if (!res.ok || !data.success) {
        throw new Error(data.error || 'Failed to fetch status');
      }

      renderStatus(data.data);
    } catch (err) {
      bannerTitle.textContent = 'Service Status Unavailable';
      bannerSubtitle.textContent = err.message;
      bannerPulse.style.backgroundColor = 'var(--amber)';
    }
  }

  function renderStatus(statusData) {
    const isOperational = statusData.down_monitors === 0;

    bannerTitle.textContent = statusData.system_status;
    bannerSubtitle.textContent = `Monitoring ${statusData.total_monitors} total service components. Last updated: ${new Date().toLocaleTimeString()}`;

    if (isOperational) {
      statusBanner.style.borderColor = 'rgba(16, 185, 129, 0.4)';
      bannerPulse.style.backgroundColor = 'var(--emerald)';
      bannerPulse.style.color = 'var(--emerald)';
      bannerTitle.style.color = 'var(--emerald)';
    } else {
      statusBanner.style.borderColor = 'rgba(244, 63, 94, 0.4)';
      bannerPulse.style.backgroundColor = 'var(--rose)';
      bannerPulse.style.color = 'var(--rose)';
      bannerTitle.style.color = 'var(--rose)';
    }

    // Render Component Cards
    const monitors = statusData.monitors || [];
    if (monitors.length === 0) {
      componentsList.innerHTML = `
        <div class="glass-panel" style="padding: 1.5rem; text-align: center; color: var(--text-muted);">
          No public components configured.
        </div>
      `;
    } else {
      componentsList.innerHTML = monitors.map(m => {
        const isUp = m.status === 'UP';
        const badgeClass = isUp ? 'badge-up' : 'badge-down';
        return `
          <div class="glass-panel" style="padding: 1.25rem 1.5rem; display: flex; justify-content: space-between; align-items: center;">
            <div style="font-weight: 600; font-size: 1.05rem;">${escapeHTML(m.name)}</div>
            <span class="badge ${badgeClass}">
              <span class="pulse-dot"></span>
              ${isUp ? 'Operational' : 'Service Outage'}
            </span>
          </div>
        `;
      }).join('');
    }

    // Render Active Incidents
    const incidents = statusData.active_incidents || [];
    if (incidents.length === 0) {
      incidentsList.innerHTML = `
        <div class="glass-panel" style="padding: 1.5rem; text-align: center; color: var(--text-muted);">
          ✨ No active incidents reported in the last 90 days.
        </div>
      `;
    } else {
      incidentsList.innerHTML = incidents.map(inc => `
        <div class="glass-panel" style="padding: 1.5rem; margin-bottom: 1rem; border-left: 4px solid var(--rose);">
          <div style="display: flex; justify-content: space-between; margin-bottom: 0.5rem;">
            <div style="font-weight: 600; color: var(--rose);">${escapeHTML(inc.monitor_name)} — Outage Reported</div>
            <div style="font-size: 0.85rem; color: var(--text-muted);">${new Date(inc.started_at).toLocaleString()}</div>
          </div>
          <div style="font-size: 0.9rem; color: var(--text-main);">${escapeHTML(inc.cause || 'Target target failed HTTP health probe')}</div>
        </div>
      `).join('');
    }
  }

  function escapeHTML(str) {
    return str.replace(/[&<>'"]/g, 
      tag => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[tag] || tag)
    );
  }
});

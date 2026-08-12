let state = {};
let inputsInit = false;
let currentPage = 'dashboard';
let sseConnected = false;
let importPreviewContent = '';
let logTab = 'session';
let textTab = 'personal';
let pollTimer = null;
let pagesReady = false;
let bootAt = 0;
const BOOT_GRACE_MS = 8000;
const UI_ASSET_VER = '2.4';
const activityLog = [];
const PAGE_ORDER = ['dashboard', 'devices', 'accounts', 'skrip', 'text', 'settings', 'kontrol', 'log'];

/** Lucide CDN — markup placeholder; call refreshLucideIcons() after dynamic HTML. */
function lucideIcon(name, className) {
  return '<i data-lucide="' + name + '"' + (className ? ' class="' + className + '"' : '') + ' aria-hidden="true"></i>';
}

function refreshLucideIcons(root) {
  if (typeof lucide === 'undefined' || !lucide.createIcons) return;
  try {
    const opts = { attrs: { 'stroke-width': '1.75' } };
    if (root) opts.root = root;
    lucide.createIcons(opts);
  } catch (e) {
    console.warn('lucide icons:', e);
  }
}

const STATUS_LABEL = {
  idle: 'Siap',
  running: 'Berjalan',
  paused: 'Dijeda',
  done: 'Selesai',
  error: 'Error'
};

function statusLabel(s) {
  const k = (s || 'idle').toLowerCase();
  return STATUS_LABEL[k] || k;
}

function setSSEStatus(connected, reconnecting) {
  sseConnected = connected;
  const dot = document.getElementById('sseDot');
  const lbl = document.getElementById('sseLabel');
  if (!dot || !lbl) return;
  if (reconnecting) {
    dot.className = 'dot-sse dot-sse-reconnect';
    lbl.textContent = 'menyambung…';
  } else if (connected) {
    dot.className = 'dot-sse dot-sse-ok';
    lbl.textContent = 'live';
  } else {
    dot.className = 'dot-sse dot-sse-off';
    lbl.textContent = 'offline';
  }
}

function showPage(page, navEl) {
  if (!PAGE_ORDER.includes(page)) return;
  currentPage = page;
  document.querySelectorAll('.page').forEach(p => {
    p.classList.toggle('active', p.dataset.page === page);
  });
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  const btn = navEl || document.querySelector('.nav-item[data-page="' + page + '"]');
  if (btn) btn.classList.add('active');
  if (page === 'log' && logTab === 'server') fetchServerLog();
  if (page === 'text') loadPostTexts(textTab);
}

function logHtml(entries) {
  if (!entries.length) {
    return '<div class="empty" style="padding:12px">Belum ada aktivitas.</div>';
  }
  return entries.map(e =>
    '<div class="act-row"><span class="act-time">' + esc(e.t) + '</span><span class="act-msg">' + esc(e.msg) + '</span></div>'
  ).join('');
}

function dashActivityBadge(msg) {
  const m = (msg || '').toLowerCase();
  if (/jalankan|run|automation|start/.test(m)) return { cls: 'run', text: 'RUN' };
  if (/aktif|worker|mirror|connect|terhubung|assign|import|refresh/.test(m)) return { cls: 'aktif', text: 'AKTIF' };
  return { cls: 'info', text: 'INFO' };
}

function dashLogHtml(entries) {
  if (!entries.length) {
    return '<div class="dash-activity-empty">Belum ada aktivitas.</div>';
  }
  return entries.map(e => {
    const badge = dashActivityBadge(e.msg);
    return '<div class="dash-activity-item">' +
      '<span class="dash-activity-dot" aria-hidden="true"></span>' +
      '<div class="dash-activity-body">' +
        '<div class="dash-activity-meta">' +
          '<span class="dash-activity-time">' + esc(e.t) + '</span>' +
          '<span class="dash-activity-badge dash-activity-badge--' + badge.cls + '">' + badge.text + '</span>' +
        '</div>' +
        '<div class="dash-activity-msg">' + esc(e.msg) + '</div>' +
      '</div>' +
      lucideIcon('chevron-right', 'dash-activity-chevron') +
    '</div>';
  }).join('');
}

function log(msg) {
  const t = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  activityLog.unshift({ t, msg });
  if (activityLog.length > 50) activityLog.pop();
  renderLog();
}

function renderLog() {
  const full = document.getElementById('logPanel');
  const dash = document.getElementById('dashLog');
  const html = logHtml(activityLog);
  const dashHtml = dashLogHtml(activityLog.slice(0, 8));
  if (full) full.innerHTML = html;
  if (dash) dash.innerHTML = dashHtml;
  if (full) refreshLucideIcons(full);
  if (dash) refreshLucideIcons(dash);
}

function esc(s) {
  return String(s).replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/'/g,'&#39;');
}

function escAttr(s) {
  return esc(s).replace(/\u0060/g, '&#96;');
}

function applyState(data) {
  if (!data) return false;
  if (data.ui_rev != null && data.ui_rev !== uiRev) {
    const pastBoot = Date.now() - bootAt > BOOT_GRACE_MS;
    if (uiRev !== 0 && pagesReady && pastBoot) {
      location.reload();
      return false;
    }
    uiRev = data.ui_rev;
  }
  if (data.state_rev != null) {
    if (data.state_rev < stateRev) return false;
    stateRev = data.state_rev;
  }
  if (!Array.isArray(data.devices) && data.run_status === undefined) return false;

  // Jangan timpa switch yang sedang di-toggle (race dengan polling/SSE).
  if (Array.isArray(data.devices) && pendingToggle.size) {
    data = Object.assign({}, data, {
      devices: data.devices.map(d => {
        if (pendingToggle.has(d.serial) && pendingToggleWant.has(d.serial)) {
          return Object.assign({}, d, { enabled: pendingToggleWant.get(d.serial) });
        }
        return d;
      })
    });
    if (data.enabled_count == null) {
      data.enabled_count = data.devices.filter(d => d.enabled && d.connected !== false).length;
    }
  }

  state = data;
  render();
  renderDevBadge();
  return true;
}

function renderDevBadge() {
  const el = document.getElementById('devModeBadge');
  if (!el) return;
  el.style.display = state.panel_dev ? 'inline-flex' : 'none';
}

function showApiError(msg) {
  const meta = document.getElementById('deviceListMeta');
  if (meta) meta.innerHTML = '<span style="color:#b91c1c">' + esc(msg) + '</span>';
  const list = document.getElementById('deviceList');
  if (list) list.innerHTML = '<div class="dev-empty">' + esc(msg) + '</div>';
}

async function fetchStateAndRender() {
  const r = await fetch('/api/state?v=' + UI_ASSET_VER, { cache: 'no-store' });
  if (!r.ok) {
    showApiError('API /state → HTTP ' + r.status);
    log('state gagal — ' + r.status);
    return false;
  }
  const data = await r.json();
  return applyState(data);
}

async function refreshState() {
  try {
    await fetchStateAndRender();
  } catch (e) {
    if (pagesReady) showApiError('API refresh gagal');
  }
}

function setButtons() {
  const s = (state.run_status || 'idle').toLowerCase();
  const statusText = document.getElementById('statusText');
  if (statusText) statusText.textContent = statusLabel(s);
  const dot = document.getElementById('statusDot');
  if (dot) dot.className = 'status-dot dot-' + (['idle','running','paused','done','error'].includes(s) ? s : 'idle');
  const pf = state.preflight || {};
  const canRun = !!pf.can_run;
  const runIds = ['btnRun', 'btnRunPage', 'btnDashRun'];
  const pauseIds = ['btnPause', 'btnPausePage'];
  const resumeIds = ['btnResume', 'btnResumePage'];
  const stopIds = ['btnStop', 'btnStopPage'];
  const ackIds = ['btnAck', 'btnAckPage'];
  runIds.forEach(id => { const el = document.getElementById(id); if (el) el.disabled = !canRun; });
  pauseIds.forEach(id => { const el = document.getElementById(id); if (el) el.disabled = s !== 'running'; });
  resumeIds.forEach(id => { const el = document.getElementById(id); if (el) el.disabled = s !== 'paused'; });
  stopIds.forEach(id => { const el = document.getElementById(id); if (el) el.disabled = s !== 'running' && s !== 'paused'; });
  const showAck = s === 'done';
  ackIds.forEach(id => { const el = document.getElementById(id); if (el) el.style.display = showAck ? '' : 'none'; });
  renderPreflight();
  renderLastResults();
}

function renderPreflight() {
  const el = document.getElementById('preflightBanner');
  if (!el) return;
  const pf = state.preflight || {};
  const reasons = pf.reasons || [];
  if (reasons.length === 0 && pf.can_run) {
    el.style.display = 'none';
    return;
  }
  el.style.display = '';
  if (reasons.length === 0) {
    el.className = 'preflight-banner ok';
    el.textContent = 'Siap dijalankan';
    return;
  }
  el.className = 'preflight-banner';
  el.innerHTML = '<strong>Belum siap run:</strong><ul>' +
    reasons.map(r => '<li>' + esc(r) + '</li>').join('') + '</ul>';
}

function renderLastResults() {
  const box = document.getElementById('lastResultsBox');
  const list = document.getElementById('lastResultsList');
  if (!box || !list) return;
  const results = state.last_results || [];
  if (!results.length) {
    box.style.display = 'none';
    return;
  }
  box.style.display = '';
  list.innerHTML = results.map(r => {
    const tail = (r.serial || '').slice(-8);
    const ok = !(r.errors > 0);
    return '<div class="result-row"><span>…' + esc(tail) + '</span>' +
      '<span class="' + (ok ? 'result-ok' : 'result-err') + '">' +
      (ok ? r.tasks + ' task OK' : r.errors + ' error') + '</span></div>';
  }).join('');
}

function dashStatCard(iconClass, iconName, value, label, solo) {
  return '<div class="dash-stat-card' + (solo ? ' dash-stat-card--solo' : '') + '">' +
    '<div class="dash-stat-icon ' + iconClass + '">' + lucideIcon(iconName) + '</div>' +
    '<span class="dash-stat-value">' + value + '</span>' +
    '<span class="dash-stat-label">' + label + '</span></div>';
}

const DASH_STAT_ICONS = {
  phone: 'smartphone',
  toggle: 'toggle-right',
  monitor: 'monitor',
  db: 'database',
  users: 'users',
  worker: 'user-cog',
  standby: 'circle-dot',
  total: 'layers'
};

function renderDashboard() {
  const ec = state.enabled_count || 0;
  const mo = state.mirror_open_count || 0;
  const ac = state.account_count || 0;
  const asg = state.assigned_count || 0;
  const devN = (state.devices || []).length;
  const wk = state.workers || document.getElementById('workers')?.value || 2;
  const dash = document.getElementById('dashStats');
  if (dash) {
    dash.innerHTML =
      dashStatCard('dash-stat-icon--blue', DASH_STAT_ICONS.phone, devN, 'HP terhubung') +
      dashStatCard('dash-stat-icon--green', DASH_STAT_ICONS.toggle, ec, 'Switch ON') +
      dashStatCard('dash-stat-icon--purple', DASH_STAT_ICONS.monitor, mo, 'Mirror terbuka') +
      dashStatCard('dash-stat-icon--orange', DASH_STAT_ICONS.db, ac, 'Akun DB') +
      dashStatCard('dash-stat-icon--teal', DASH_STAT_ICONS.users, asg, 'Assigned') +
      dashStatCard('dash-stat-icon--pink', DASH_STAT_ICONS.worker, wk, 'Workers', true);
  }
  const health = state.health || {};
  const hr = document.getElementById('healthRow');
  if (hr) {
    hr.innerHTML = 'ADB: <strong>' + (health.adb ? 'OK' : '✗') + '</strong> · DB: <strong>' + (health.db ? 'OK' : '✗') + '</strong> · Scrcpy: <strong>' + (health.scrcpy_count || 0) + '</strong>';
  }
  const hb = document.getElementById('healthBadge');
  if (hb) {
    const issues = (health.issues || []).length;
    const allOk = health.adb && health.db && issues === 0;
    hb.innerHTML = allOk
      ? '<span class="dash-health-ok">✓ Semua sistem berjalan normal</span>'
      : '<span class="dash-health-warn">⚠ ' + (issues ? issues + ' isu perlu diperbaiki' : 'Perlu perhatian') + '</span>';
  }
  const cl = document.getElementById('dashChecklist');
  if (cl) {
    const items = state.checklist || [];
    if (!items.length) {
      cl.innerHTML = '';
    } else {
      cl.innerHTML = items.map(it => {
        const ok = !!it.ok;
        return '<div class="dash-check-item ' + (ok ? '' : 'fail') + '" onclick="showPage(\'' + esc(it.page || 'dashboard') + '\')">' +
          '<span class="dash-check-mark" aria-hidden="true">' + (ok ? '✓' : '!') + '</span>' +
          '<span>' + esc(it.label || '') + '</span></div>';
      }).join('');
    }
  }
  const dashEl = document.getElementById('dashStats');
  if (dashEl) refreshLucideIcons(dashEl);
}

const PIPELINE_STEPS = [
  { id: 'pm_clear', label: 'PM Clear' },
  { id: 'login', label: 'Login' },
  { id: 'auto_post', label: 'Post Beranda' },
  { id: 'fanpage_post', label: 'Post Fanpage' },
  { id: 'logout', label: 'Logout' },
];

function legacyPipelineSteps(flow) {
  const map = {
    facebook_login_logout: ['login', 'logout'],
    facebook_login_auto_post: ['login', 'auto_post'],
    facebook_login_auto_post_logout: ['login', 'auto_post', 'logout'],
    facebook_login_fanpage_post: ['login', 'fanpage_post'],
    facebook_login_fanpage_post_logout: ['login', 'fanpage_post', 'logout'],
  };
  return map[flow] || ['login', 'logout'];
}

function pipelineChecksHtml(account, boxId) {
  const id = boxId || ('accPipe' + account.id);
  const active = new Set(
    (account.pipeline_steps && account.pipeline_steps.length)
      ? account.pipeline_steps
      : legacyPipelineSteps(account.automation_flow || 'facebook_login_logout')
  );
  return '<div class="acc-pipeline" id="' + id + '">' +
    PIPELINE_STEPS.map(s =>
      '<label class="acc-step"><input type="checkbox" data-step="' + escAttr(s.id) + '"' +
      (active.has(s.id) ? ' checked' : '') +
      ' onchange="saveAccountPipeline(' + account.id + ', \'' + id + '\')"> ' + esc(s.label) + '</label>'
    ).join('') + '</div>';
}

function pipelineSummary(account) {
  const steps = account.automation_steps || [];
  if (steps.length) return steps.join(' → ');
  const active = (account.pipeline_steps && account.pipeline_steps.length)
    ? account.pipeline_steps
    : legacyPipelineSteps(account.automation_flow || 'facebook_login_logout');
  return PIPELINE_STEPS.filter(s => active.includes(s.id)).map(s => s.label).join(' → ');
}

let accDrawerAccountId = null;
let accDrawerMode = 'settings';

const ACC_ICON_GEAR = lucideIcon('settings', 'acc-lucide');
const ACC_ICON_DETAIL = lucideIcon('info', 'acc-lucide');

const DEV_INFO_ICON = lucideIcon('info', 'dev-lucide');
const DEV_PHONE_ICON = lucideIcon('smartphone', 'dev-lucide');
const DEV_MIRROR_ICON = lucideIcon('monitor', 'dev-lucide');
const DEV_DETAIL_ICON = lucideIcon('file-text', 'dev-lucide');

function closeAccDrawer() {
  const drawer = document.getElementById('accDrawer');
  if (drawer) {
    drawer.classList.remove('open');
    drawer.setAttribute('aria-hidden', 'true');
  }
  accDrawerAccountId = null;
  accDrawerMode = 'settings';
}

function accDetailRows(account) {
  const assigned = account.assigned_serial || '';
  const fps = account.fanpages || [];
  const fpList = fps.length
    ? '<ul class="acc-detail-fplist">' + fps.map(f =>
        '<li><strong>' + esc(f.name || '—') + '</strong><span>' + esc(f.fb_page_id || '') + '</span></li>'
      ).join('') + '</ul>'
    : '<span class="acc-drawer-muted">Belum ada fanpage</span>';
  return [
    ['ID akun', '#' + account.id],
    ['Nama', account.name || '—'],
    ['Login', account.login_id || '—'],
    ['Slot', account.slot_no != null ? String(account.slot_no) : '—'],
    ['HP', assigned ? esc(assigned) : '—'],
    ['Automation', account.automation_enabled === false ? 'Nonaktif' : 'Aktif'],
    ['Flow', account.automation_flow || '—'],
    ['Pipeline', pipelineSummary(account) || '—'],
    ['Fanpage (' + (account.fanpage_count || 0) + ')', fpList],
  ];
}

function renderAccDetailBody(account) {
  return accDetailRows(account).map(([label, val]) =>
    '<div class="acc-detail-row">' +
      '<div class="acc-detail-k">' + esc(label) + '</div>' +
      '<div class="acc-detail-v">' + (typeof val === 'string' && val.indexOf('<') >= 0 ? val : esc(String(val))) + '</div>' +
    '</div>'
  ).join('');
}

function renderAccSettingsBody(account, accountId) {
  const devices = state.devices || [];
  const assigned = account.assigned_serial || '';
  const devOpts = devices.map((d, i) => {
    const tail = (d.serial || '').slice(-8);
    const sel = d.serial === assigned ? ' selected' : '';
    return '<option value="' + escAttr(d.serial) + '"' + sel + '>HP' + (i + 1) + ' …' + esc(tail) + '</option>';
  }).join('');

  let assignBlock = '';
  if (assigned) {
    assignBlock =
      '<div class="acc-drawer-section">' +
        '<div class="acc-drawer-label">HP terassign</div>' +
        '<div class="acc-drawer-hp">…' + esc(assigned.slice(-8)) + '</div>' +
        '<button type="button" class="btn btn-sm btn-danger acc-drawer-lepas" onclick="unassignAccount(' + accountId + ')">Lepas dari HP</button>' +
      '</div>';
  } else if (devices.length) {
    assignBlock =
      '<div class="acc-drawer-section">' +
        '<div class="acc-drawer-label">Assign ke HP</div>' +
        '<div class="acc-assign acc-drawer-assign">' +
          '<select id="accDrawerDev">' + devOpts + '</select>' +
          '<button type="button" class="btn-dev on" onclick="assignAccount(' + accountId + ', true)">Assign</button>' +
        '</div>' +
      '</div>';
  } else {
    assignBlock = '<div class="acc-drawer-section acc-drawer-muted">Colok HP di tab Device untuk assign.</div>';
  }

  return assignBlock +
    '<div class="acc-drawer-section">' +
      '<div class="acc-drawer-label">Pipeline (per akun)</div>' +
      pipelineChecksHtml(account, 'accDrawerPipe' + accountId) +
      '<div class="acc-drawer-steps" id="accDrawerSteps">' + esc(pipelineSummary(account)) + '</div>' +
    '</div>';
}

function openAccDrawer(accountId, mode) {
  const account = (state.accounts || []).find(a => String(a.id) === String(accountId));
  if (!account) return;
  accDrawerAccountId = accountId;
  accDrawerMode = mode === 'detail' ? 'detail' : 'settings';
  const drawer = document.getElementById('accDrawer');
  const title = document.getElementById('accDrawerTitle');
  const sub = document.getElementById('accDrawerSub');
  const body = document.getElementById('accDrawerBody');
  if (!drawer || !body) return;

  if (title) title.textContent = accDrawerMode === 'detail' ? 'Detail akun' : (account.name || ('Akun #' + accountId));
  if (sub) sub.textContent = accDrawerMode === 'detail' ? (account.name || '') : (account.login_id || '');

  body.innerHTML = accDrawerMode === 'detail'
    ? renderAccDetailBody(account)
    : renderAccSettingsBody(account, accountId);

  drawer.classList.add('open');
  drawer.setAttribute('aria-hidden', 'false');
  refreshLucideIcons(body);
}

async function saveAccountPipeline(accountId, boxId) {
  const id = boxId || ('accPipe' + accountId);
  const box = document.getElementById(id);
  if (!box) return;
  const steps = Array.from(box.querySelectorAll('input[data-step]:checked')).map(el => el.dataset.step);
  if (!steps.length) {
    log('Minimal 1 langkah aktif');
    return;
  }
  await api('accounts/automation', { account_id: accountId, steps: steps, enabled: true });
  const summary = document.getElementById('accDrawerSteps');
  if (summary) {
    summary.textContent = PIPELINE_STEPS.filter(s => steps.includes(s.id)).map(s => s.label).join(' → ');
  }
  log('Pipeline akun #' + accountId + ': ' + steps.join(' → '));
}

function fanpageCell(a) {
  const n = a.fanpage_count || 0;
  if (!n) return '<span class="fp-count zero">0</span>';
  const names = (a.fanpages || []).map(f => (f.name && f.name !== f.fb_page_id) ? f.name : (f.name || f.fb_page_id)).join(' · ');
  const title = names || (n + ' fanpage');
  return '<span class="fp-count" title="' + escAttr(title) + '">' + n + '</span>' +
    (names ? '<div class="fp-names" title="' + escAttr(title) + '">' + esc(names) + '</div>' : '');
}

function renderAccountsList() {
  const el = document.getElementById('accountsList');
  if (!el) return;
  const accounts = state.accounts || [];
  if (!accounts.length) {
    el.innerHTML = '<div class="empty" style="padding:12px">Belum ada akun — klik Import.</div>';
    closeAccDrawer();
    return;
  }
  el.innerHTML = '<table class="acc-table"><thead><tr><th>#</th><th>Nama</th><th>Login</th><th>Fanpage</th><th>HP</th><th>Aksi</th></tr></thead><tbody>' +
    accounts.map((a, i) => {
      const assigned = a.assigned_serial || '';
      const tail = assigned ? assigned.slice(-8) : '—';
      const pipeHint = pipelineSummary(a);
      return '<tr><td>' + (i+1) + '</td><td>' + esc(a.name || '—') +
        (pipeHint ? '<div class="acc-pipe-hint" title="' + escAttr(pipeHint) + '">' + esc(pipeHint) + '</div>' : '') +
        '</td><td>' + esc(a.login_id || '') +
        '</td><td class="fp-col">' + fanpageCell(a) + '</td><td>' +
        (assigned ? '…' + esc(tail) : '<span class="acc-unassigned">—</span>') +
        '</td><td class="acc-actions-col">' +
        '<div class="acc-action-btns">' +
        '<button type="button" class="btn-acc-icon" onclick="openAccDrawer(' + a.id + ', \'detail\')" title="Detail akun">' + ACC_ICON_DETAIL + '</button>' +
        '<button type="button" class="btn-acc-icon" onclick="openAccDrawer(' + a.id + ', \'settings\')" title="Pipeline &amp; assign">' + ACC_ICON_GEAR + '</button>' +
        '</div></td></tr>';
    }).join('') + '</tbody></table>';

  refreshLucideIcons(el);

  if (accDrawerAccountId != null) {
    const still = accounts.find(a => String(a.id) === String(accDrawerAccountId));
    if (still) openAccDrawer(accDrawerAccountId, accDrawerMode);
    else closeAccDrawer();
  }
}

async function setAccountFlow(accountId, flow) {
  await api('accounts/automation', { account_id: accountId, flow: flow, enabled: true });
  log('Skrip akun #' + accountId + ' → ' + flow);
}

async function assignAccount(accountId, fromDrawer) {
  const sel = document.getElementById(fromDrawer ? 'accDrawerDev' : ('accDev' + accountId));
  if (!sel || !sel.value) { log('Pilih HP dulu'); return; }
  await api('accounts/assign', { account_id: accountId, serial: sel.value, slot_no: 1 });
  log('Assign akun #' + accountId);
}

async function unassignAccount(accountId) {
  if (!confirm('Lepas assign akun #' + accountId + '?')) return;
  await api('accounts/unassign', { account_id: accountId });
  log('Unassign akun #' + accountId);
  closeAccDrawer();
}

function confirmEnableAll() {
  const n = (state.devices || []).length;
  if (!n) { log('Tidak ada HP terhubung'); return; }
  if (!confirm('Aktifkan mirror untuk semua ' + n + ' HP terhubung?')) return;
  api('devices/enable-all');
}

function confirmDisableAll() {
  const ec = state.enabled_count || 0;
  const mo = state.mirror_open_count || 0;
  if (!ec) { log('Tidak ada HP aktif'); return; }
  if (!confirm('Matikan mirror untuk ' + ec + ' HP aktif?')) return;
  api('devices/disable-all');
}

async function api(path, body) {
  const opts = { method: 'POST', headers: { 'Content-Type': 'application/json' } };
  if (body) opts.body = JSON.stringify(body);
  const r = await fetch('/api/' + path, opts);
  const text = await r.text();
  if (!r.ok) { log('✗ ' + path + ' — ' + text.slice(0, 120)); return false; }
  try {
    if (text) applyState(JSON.parse(text));
  } catch (e) {
    await refreshState();
  }
  log(path.replace(/\//g, ' · '));
  return true;
}

const pendingToggle = new Set();
const pendingToggleWant = new Map();

function patchDeviceToggleUI(serial, enabled) {
  const sel = '.dev-item[data-serial="' + (typeof CSS !== 'undefined' && CSS.escape ? CSS.escape(serial) : serial.replace(/"/g, '\\"')) + '"]';
  const row = document.querySelector(sel);
  if (!row) return;
  const cb = row.querySelector('input[type=checkbox]');
  if (cb && cb.checked !== enabled) cb.checked = enabled;
  row.classList.toggle('dev-item--active', !!enabled);
  const badge = row.querySelector('.dev-badge');
  if (badge) {
    if (!enabled) {
      badge.className = 'dev-badge dev-badge--standby';
      badge.textContent = 'Standby';
    } else {
      badge.className = 'dev-badge dev-badge--wait';
      badge.textContent = 'Membuka mirror…';
    }
  }
}

function devDotClass(st) {
  const s = (st || 'idle').toLowerCase();
  return ['idle','running','paused','done','error'].includes(s) ? s : 'idle';
}

async function toggleDev(serial, enabled, el) {
  if (pendingToggle.has(serial)) return;
  pendingToggle.add(serial);
  pendingToggleWant.set(serial, enabled);
  const dev = (state.devices || []).find(d => d.serial === serial);
  const prev = dev ? !!dev.enabled : false;
  if (dev) {
    dev.enabled = enabled;
    patchDeviceToggleUI(serial, enabled);
    const ec = (state.devices || []).filter(d => d.enabled && d.connected !== false).length;
    state.enabled_count = ec;
    const statsLine = document.getElementById('statsLine');
    if (statsLine) {
      const mo = state.mirror_open_count || 0;
      const wk = state.workers ?? document.getElementById('workers')?.value ?? 2;
      const ac = state.account_count || 0;
      statsLine.innerHTML =
        '<strong>' + ec + '</strong> switch ON · <strong>' + mo + '</strong> mirror terbuka · <strong>' + wk + '</strong> Worker · <strong>' + ac + '</strong> Akun';
    }
  }
  if (el) el.disabled = true;
  try {
    const r = await fetch('/api/devices/toggle', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ serial, enabled })
    });
    const text = await r.text();
    if (!r.ok) {
      if (dev) dev.enabled = prev;
      pendingToggleWant.set(serial, prev);
      patchDeviceToggleUI(serial, prev);
      renderDevicesPage();
      log('toggle gagal: ' + serial.slice(-8));
      return;
    }
    try {
      if (text) applyState(JSON.parse(text));
      else await refreshState();
    } catch (e) {
      await refreshState();
    }
    log((enabled ? '✓ Aktif' : '○ Nonaktif') + ' …' + serial.slice(-8));
  } catch (e) {
    if (dev) dev.enabled = prev;
    pendingToggleWant.set(serial, prev);
    patchDeviceToggleUI(serial, prev);
    renderDevicesPage();
    log('toggle error');
  } finally {
    pendingToggle.delete(serial);
    pendingToggleWant.delete(serial);
    if (el) el.disabled = false;
  }
}

async function doRun() {
  if (!(state.preflight || {}).can_run) {
    log('Run diblok — cek checklist di atas');
    return;
  }
  await saveSettings(true);
  await api('run', {
    max_devices: +document.getElementById('maxDev').value || 1,
    workers: +document.getElementById('workers').value || 1
  });
}

let settingsDirty = false;
let settingsFormBound = false;

const SETTINGS_INPUT_IDS = [
  'setMaxDev', 'setWorkers', 'setLoopCount', 'setDelayActions', 'setPostAction',
  'setDelayLoops', 'setRetryDelay', 'setRetryMax', 'setPollSec', 'setLaunchWait',
  'setDelayForceStop', 'setDriver', 'setAdbTimeout', 'setAdbRetries',
  'setWakeScreen', 'setForceStopBefore', 'setForceStopAfter',
  'setScreenshotErr', 'setMirrorOnRun'
];

function markSettingsDirty() {
  settingsDirty = true;
  updateSettingsSaveUI();
}

function clearSettingsDirty() {
  settingsDirty = false;
  updateSettingsSaveUI();
}

function isRunActive() {
  const s = (state.run_status || 'idle').toLowerCase();
  return s === 'running' || s === 'paused';
}

function updateSettingsSaveUI() {
  const btnTop = document.getElementById('btnSaveSettings');
  const btnBottom = document.getElementById('btnSaveSettingsBottom');
  const dirtyHint = document.getElementById('settingsDirtyHint');
  const runBanner = document.getElementById('settingsRunBanner');
  const running = isRunActive();

  [btnTop, btnBottom].forEach(btn => {
    if (btn) btn.disabled = !settingsDirty;
  });
  if (dirtyHint) dirtyHint.style.display = settingsDirty ? '' : 'none';

  if (runBanner) {
    if (running) {
      runBanner.style.display = '';
      runBanner.innerHTML = '<strong>Run sedang berjalan.</strong> Anda tetap bisa simpan setup — perubahan disimpan ke config.json dan berlaku untuk <strong>run berikutnya</strong>, bukan run yang sedang jalan.';
    } else {
      runBanner.style.display = 'none';
      runBanner.innerHTML = '';
    }
  }
}

function setupSettingsForm() {
  if (settingsFormBound) return;
  SETTINGS_INPUT_IDS.forEach(id => {
    const el = document.getElementById(id);
    if (!el) return;
    el.addEventListener('input', markSettingsDirty);
    el.addEventListener('change', markSettingsDirty);
  });
  settingsFormBound = true;
  updateSettingsSaveUI();
}

function collectSettingsForm() {
  return {
    max_devices: +document.getElementById('setMaxDev')?.value || 1,
    parallel_workers: +document.getElementById('setWorkers')?.value || 1,
    loop_count: +document.getElementById('setLoopCount')?.value || 1,
    delay_between_actions_sec: +document.getElementById('setDelayActions')?.value || 0,
    post_action_delay_sec: +document.getElementById('setPostAction')?.value || 0,
    delay_between_loops_sec: +document.getElementById('setDelayLoops')?.value || 0,
    retry_delay_sec: +document.getElementById('setRetryDelay')?.value || 0,
    retry_max_attempts: +document.getElementById('setRetryMax')?.value || 2,
    poll_sec: +document.getElementById('setPollSec')?.value || 0,
    app_launch_wait_sec: +document.getElementById('setLaunchWait')?.value || 0,
    delay_after_force_stop_sec: +document.getElementById('setDelayForceStop')?.value || 0,
    automation_driver: document.getElementById('setDriver')?.value || 'adb',
    adb_timeout_sec: +document.getElementById('setAdbTimeout')?.value || 30,
    adb_retries: +document.getElementById('setAdbRetries')?.value || 2,
    wake_screen_before_task: !!document.getElementById('setWakeScreen')?.checked,
    clear_app_before_open: false,
    force_stop_before_open: !!document.getElementById('setForceStopBefore')?.checked,
    force_stop_after_task: !!document.getElementById('setForceStopAfter')?.checked,
    screenshot_on_error: !!document.getElementById('setScreenshotErr')?.checked,
    mirror_on_run: !!document.getElementById('setMirrorOnRun')?.checked
  };
}

function applySettingsForm(s) {
  if (!s) return;
  const set = (id, v) => { const el = document.getElementById(id); if (el && v != null) el.value = v; };
  const setChk = (id, v) => { const el = document.getElementById(id); if (el) el.checked = !!v; };
  set('setMaxDev', s.max_devices);
  set('setWorkers', s.parallel_workers);
  set('maxDev', s.max_devices);
  set('workers', s.parallel_workers);
  set('setLoopCount', s.loop_count);
  set('setDelayActions', s.delay_between_actions_sec);
  set('setPostAction', s.post_action_delay_sec);
  set('setDelayLoops', s.delay_between_loops_sec);
  set('setRetryDelay', s.retry_delay_sec);
  set('setRetryMax', s.retry_max_attempts);
  set('setPollSec', s.poll_sec);
  set('setLaunchWait', s.app_launch_wait_sec);
  set('setDelayForceStop', s.delay_after_force_stop_sec);
  set('setAdbTimeout', s.adb_timeout_sec);
  set('setAdbRetries', s.adb_retries);
  if (s.automation_driver) {
    const d = document.getElementById('setDriver');
    if (d) d.value = s.automation_driver;
  }
  setChk('setWakeScreen', s.wake_screen_before_task);
  setChk('setForceStopBefore', s.force_stop_before_open);
  setChk('setForceStopAfter', s.force_stop_after_task);
  setChk('setScreenshotErr', s.screenshot_on_error);
  setChk('setMirrorOnRun', s.mirror_on_run);
}

async function saveSettings(silent) {
  const body = collectSettingsForm();
  const maxEl = document.getElementById('maxDev');
  const wkEl = document.getElementById('workers');
  if (maxEl) maxEl.value = body.max_devices;
  if (wkEl) wkEl.value = body.parallel_workers;
  try {
    const r = await fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
    if (!r.ok) {
      const t = await r.text();
      const hint = document.getElementById('settingsSaveHint');
      if (hint && !silent) {
        hint.className = 'set-save-hint err';
        hint.textContent = 'Gagal simpan: ' + t.slice(0, 80);
      }
      if (!silent) log('Simpan gagal: ' + t);
      return false;
    }
    let data = {};
    try { data = await r.json(); } catch (e) {}
    clearSettingsDirty();
    const hint = document.getElementById('settingsSaveHint');
    if (hint && !silent) {
      hint.className = 'set-save-hint';
      hint.textContent = data.run_active
        ? '✓ Tersimpan — berlaku untuk run berikutnya (run sedang jalan)'
        : '✓ Pengaturan tersimpan ke config.json';
      setTimeout(() => { hint.textContent = ''; hint.className = 'set-save-hint'; }, 4500);
    }
    if (!silent) {
      log(data.run_active ? 'Pengaturan disimpan (run berikutnya)' : 'Pengaturan disimpan');
    }
    await refreshState();
    return true;
  } catch (e) {
    if (!silent) log('Simpan error');
    return false;
  }
}

async function ackRun() {
  await api('run/ack');
  log('Run di-reset — siap lagi');
}

async function retryMirror(serial) {
  await api('devices/mirror-retry', { serial });
  log('Retry mirror …' + serial.slice(-8));
}

async function toggleTask(name) { await api('tasks/toggle', { name }); }
async function pauseDev(serial) { await api('devices/pause', { serial }); }
async function resumeDev(serial) { await api('devices/resume', { serial }); }

async function registerDevice() {
  const serial = (document.getElementById('devSerial')?.value || '').trim();
  const label = (document.getElementById('devLabel')?.value || '').trim();
  if (!serial) { log('Serial HP wajib diisi'); return; }
  await api('devices/register', { serial, label });
  document.getElementById('devSerial').value = '';
  log('HP terdaftar …' + serial.slice(-8));
}

async function toggleWindowMaximize() {
  try {
    const r = await fetch('/api/window/maximize', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' });
    if (!r.ok) return;
    await syncWindowChrome();
  } catch (e) {}
}

async function toggleWindowFullscreen() {
  try {
    const r = await fetch('/api/window/fullscreen', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' });
    if (!r.ok) return;
    await syncWindowChrome();
    log('Layar penuh toggled (F11)');
  } catch (e) {}
}

async function syncWindowChrome() {
  try {
    const r = await fetch('/api/window/state');
    if (!r.ok) return;
    const data = await r.json();
    const maxBtn = document.getElementById('btnMaximize');
    const fsBtn = document.getElementById('btnFullscreen');
    if (maxBtn) maxBtn.classList.toggle('icon-btn-active', !!data.maximized);
    if (fsBtn) fsBtn.classList.toggle('icon-btn-active', !!data.fullscreen);
  } catch (e) {}
}

async function launchLDPlayer(index) {
  const body = index != null ? { index: index } : { all: true };
  const ok = await api('emulators/launch', body);
  if (!ok) {
    log('Gagal launch LDPlayer — cek Log atau jalankan panel sebagai Administrator jika masih gagal');
    return;
  }
  log(index != null ? ('LDPlayer #' + index + ' diluncurkan — tunggu boot…') : 'Semua LDPlayer diluncurkan — tunggu boot…');
}

async function launchAllLDPlayer() {
  await launchLDPlayer(null);
}

async function addLDPlayerInstance() {
  await api('emulators/add', { clone_from: 0 });
  log('Instance LDPlayer baru ditambahkan');
}

async function quitLDPlayer(index) {
  await api('emulators/quit', { index: index });
  log('LDPlayer index ' + index + ' dihentikan');
}

function renderLDPlayerPanel() {
  const lp = state.ldplayer || {};
  const meta = document.getElementById('ldplayerMeta');
  const list = document.getElementById('ldplayerList');
  if (!meta || !list) return;

  if (!lp.available) {
    meta.textContent = lp.error ? ('LDPlayer: ' + lp.error) : 'LDPlayer tidak ditemukan';
    list.innerHTML = '<div class="dev-empty">Set install_path di config.json atau install LDPlayer 14.</div>';
    return;
  }

  const instances = lp.instances || [];
  const runningN = instances.filter(i => i.running).length;
  const connN = instances.filter(i => i.connected).length;
  meta.innerHTML = '<strong>' + instances.length + '</strong> instance · <strong>' + runningN + '</strong> running · <strong>' + connN + '</strong> ADB';

  if (!instances.length) {
    list.innerHTML = '<div class="dev-empty">Belum ada instance — klik Tambah instance.</div>';
    return;
  }

  list.innerHTML = instances.map(inst => {
    const status = inst.connected ? 'ADB OK' : (inst.running ? 'Booting…' : 'Stopped');
    const cls = inst.connected ? 'dev-emulator-row--on' : (inst.running ? 'dev-emulator-row--boot' : '');
    return '<div class="dev-emulator-row ' + cls + '">' +
      '<div class="dev-emulator-row-main">' +
        '<span class="dev-emulator-name">LDPlayer #' + inst.index + '</span>' +
        '<span class="dev-tag">' + esc(inst.adb_serial || '') + '</span>' +
        '<span class="dev-tag">' + esc(status) + '</span>' +
      '</div>' +
      '<div class="dev-emulator-row-actions">' +
        (inst.running
          ? '<button type="button" class="dev-btn-ghost" onclick="quitLDPlayer(' + inst.index + ')">Stop</button>'
          : '<button type="button" class="dev-btn-ghost" onclick="launchLDPlayer(' + inst.index + ')">Launch</button>') +
      '</div></div>';
  }).join('');
}

async function createAccount() {
  const name = (document.getElementById('accName')?.value || '').trim();
  const login = (document.getElementById('accLogin')?.value || '').trim();
  const password = (document.getElementById('accPass')?.value || '').trim();
  if (!login || !password) { log('Login & password wajib'); return; }
  const body = { name, password };
  if (login.includes('@')) body.email = login;
  else body.profile_id = login;
  await api('accounts/create', body);
  document.getElementById('accName').value = '';
  document.getElementById('accLogin').value = '';
  document.getElementById('accPass').value = '';
  log('Akun ditambah: ' + login.slice(0, 12));
}

function startStatePolling() {
  if (pollTimer) clearInterval(pollTimer);
  pollTimer = setInterval(async () => {
    if (pendingToggle.size) return;
    await refreshState();
  }, sseConnected ? 5000 : 2000);
}

function startUIRevPolling() {
  setInterval(async () => {
    if (!state.panel_dev) return;
    try {
      const r = await fetch('/api/state');
      if (!r.ok) return;
      const data = await r.json();
      if (data.ui_rev != null && data.ui_rev !== uiRev) {
        const pastBoot = Date.now() - bootAt > BOOT_GRACE_MS;
        if (uiRev !== 0 && pagesReady && pastBoot) {
          location.reload();
          return;
        }
        uiRev = data.ui_rev;
      }
      if (data.panel_dev && !state.panel_dev) {
        state.panel_dev = true;
        renderDevBadge();
      }
    } catch (e) {}
  }, 1200);
}

function renderAdbBanner() {
  const el = document.getElementById('adbBanner');
  if (!el) return;
  const err = state.adb_error || '';
  const connected = (state.devices || []).filter(d => d.connected !== false).length;
  if (err) {
    el.style.display = '';
    el.className = 'dev-info-banner err';
    el.innerHTML = DEV_INFO_ICON + '<span>ADB error: ' + esc(err) + ' — cek USB driver / jalankan ulang panel</span>';
    refreshLucideIcons(el);
    return;
  }
  if (connected === 0) {
    el.style.display = '';
    el.className = 'dev-info-banner';
    el.innerHTML = DEV_INFO_ICON + '<span>Memindai USB… colok HP dan tunggu beberapa detik (otomatis).</span>';
    refreshLucideIcons(el);
    return;
  }
  el.style.display = '';
  el.className = 'dev-info-banner';
  el.innerHTML = DEV_INFO_ICON + '<span>' + connected + ' HP terhubung via USB — aktifkan switch untuk mirror.</span>';
  refreshLucideIcons(el);
}

function devStatCard(iconClass, iconName, value, label) {
  return '<div class="dev-stat-card">' +
    '<div class="dev-stat-icon ' + iconClass + '">' + lucideIcon(iconName) + '</div>' +
    '<span class="dev-stat-value">' + value + '</span>' +
    '<span class="dev-stat-label">' + label + '</span></div>';
}

function deviceBadge(d, isConn) {
  if (!isConn) return '<span class="dev-badge dev-badge--offline">Offline</span>';
  if (!d.enabled) return '<span class="dev-badge dev-badge--standby">Standby</span>';
  if (d.mirror_open) return '<span class="dev-badge dev-badge--mirror">Mirror aktif</span>';
  if (d.mirror_error) return '<span class="dev-badge dev-badge--fail">Gagal</span>';
  return '<span class="dev-badge dev-badge--wait">Membuka mirror…</span>';
}

function isEmulatorSerial(serial) {
  return /^emulator-\d+$/.test(serial || '') || /^127\.0\.0\.1:\d+$/.test(serial || '') || /^localhost:\d+$/.test(serial || '');
}

function deviceTags(d, isConn) {
  const tags = [];
  if (isEmulatorSerial(d.serial)) tags.push('LDPlayer');
  else if (isConn) tags.push('USB');
  else tags.push('Terdaftar');
  if (d.model) tags.push(esc(d.model));
  if (d.resolution && d.resolution !== '—') tags.push(esc(d.resolution));
  tags.push(esc(statusLabel(d.status)));
  return tags.map(t => '<span class="dev-tag">' + t + '</span>').join('');
}

function syncDevAutoOnToggle() {
  const el = document.getElementById('devAutoOn');
  if (!el) return;
  const connected = (state.devices || []).filter(d => d.connected !== false);
  suppressDevAutoOn = true;
  el.checked = connected.length > 0 && connected.every(d => d.enabled);
  suppressDevAutoOn = false;
}

let suppressDevAutoOn = false;

function onDevAutoOnToggle(el) {
  if (suppressDevAutoOn) return;
  if (el.checked) confirmEnableAll();
  else confirmDisableAll();
  setTimeout(syncDevAutoOnToggle, 600);
}

function showDeviceDetail(serial) {
  const d = (state.devices || []).find(x => x.serial === serial);
  if (!d) return;
  const lines = [
    'Serial: ' + d.serial,
    'Model: ' + (d.model || '—'),
    'Resolusi: ' + (d.resolution || '—'),
    'Status: ' + statusLabel(d.status),
    'Switch: ' + (d.enabled ? 'ON' : 'OFF'),
    'Mirror: ' + (d.mirror_open ? 'Terbuka' : 'Tutup')
  ];
  if (d.mirror_error) lines.push('Error mirror: ' + d.mirror_error);
  if (d.assigned) lines.push('Akun: ' + (d.account_name || d.account_login || '—'));
  else lines.push('Akun: belum assign');
  alert(lines.join('\n'));
}

async function openMirror(serial) {
  const d = (state.devices || []).find(x => x.serial === serial);
  if (!d || d.connected === false) return;
  if (!d.enabled) {
    await toggleDev(serial, true);
    return;
  }
  if (!d.mirror_open || d.mirror_error) await retryMirror(serial);
}

function renderDevicesPage() {
  const devices = state.devices || [];
  const connectedN = devices.filter(d => d.connected !== false).length;
  const ec = state.enabled_count || 0;
  const mo = state.mirror_open_count || 0;
  const standbyN = devices.filter(d => d.connected !== false && !d.enabled).length;
  const running = state.run_status === 'running' || state.run_status === 'paused';

  renderAdbBanner();
  renderLDPlayerPanel();

  const stats = document.getElementById('devStats');
  if (stats) {
    stats.innerHTML =
      devStatCard('dev-stat-icon--blue', DASH_STAT_ICONS.phone, connectedN, 'HP Terhubung') +
      devStatCard('dev-stat-icon--green', DASH_STAT_ICONS.toggle, ec, 'Switch ON') +
      devStatCard('dev-stat-icon--purple', DASH_STAT_ICONS.monitor, mo, 'Mirror Terbuka') +
      devStatCard('dev-stat-icon--orange', DASH_STAT_ICONS.standby, standbyN, 'Standby') +
      devStatCard('dev-stat-icon--indigo', DASH_STAT_ICONS.total, devices.length, 'Total Device');
  }

  const meta = document.getElementById('deviceListMeta');
  if (meta) {
    meta.innerHTML = '<strong>' + connectedN + '</strong> terhubung · <strong>' + ec + '</strong> switch ON · <strong>' + mo + '</strong> mirror terbuka';
  }

  const list = document.getElementById('deviceList');
  if (!list) return;

  if (!devices.length) {
    list.innerHTML = '<div class="dev-empty">Memindai USB… HP muncul otomatis saat terhubung.</div>';
    syncDevAutoOnToggle();
    const devPageEmpty = document.getElementById('page-devices');
    if (devPageEmpty) refreshLucideIcons(devPageEmpty);
    return;
  }

  list.innerHTML = devices.map((d, i) => {
    const serial = d.serial || '';
    const tail = serial.length > 8 ? serial.slice(-8) : serial;
    const dot = devDotClass(d.status);
    const isConn = d.connected !== false;
    const cls = 'dev-item' +
      (d.enabled && isConn ? ' dev-item--active' : '') +
      (!isConn ? ' dev-item--offline' : '');

    let sub = esc(statusLabel(d.status)) + ' · ' + esc(d.resolution || '—');
    if (d.assigned && d.account_login) {
      sub += ' · <span class="acc-tag">' + esc(d.account_name || d.account_login) + '</span>';
    } else if (!d.assigned) {
      sub += ' · <span class="warn-tag">belum assign</span>';
    }

    const mirrorDisabled = !isConn;
    const detailDisabled = !isConn;
    let runBtns = '';
    if (running && d.enabled && isConn) {
      runBtns = d.paused
        ? '<button type="button" class="dev-btn-ghost" onclick="resumeDev(\'' + escAttr(serial) + '\')">Resume</button>'
        : '<button type="button" class="dev-btn-ghost" onclick="pauseDev(\'' + escAttr(serial) + '\')">Pause</button>';
    }

    return '<div class="' + cls + '" data-serial="' + escAttr(serial) + '">' +
      '<div class="dev-item-icon">' + DEV_PHONE_ICON + '</div>' +
      '<div class="dev-item-main">' +
        '<div class="dev-item-title">' +
          '<span class="status-dot dot-' + dot + '"></span>' +
          '<span class="dev-item-name">HP' + (i + 1) + ' — …' + esc(tail) + '</span>' +
          deviceBadge(d, isConn) +
        '</div>' +
        '<div class="dev-item-sub">' + sub + '</div>' +
        '<div class="dev-item-tags">' + deviceTags(d, isConn) + '</div>' +
      '</div>' +
      '<div class="dev-item-actions">' +
        '<label class="switch dev-switch" title="' + (d.enabled ? 'Nonaktifkan mirror' : 'Aktifkan mirror') + '">' +
          '<input type="checkbox"' + (d.enabled ? ' checked' : '') + (isConn ? '' : ' disabled') +
          ' onchange="toggleDev(\'' + escAttr(serial) + '\', this.checked, this)">' +
          '<span class="slider"></span></label>' +
        '<button type="button" class="dev-btn-ghost"' + (mirrorDisabled ? ' disabled' : '') +
          ' onclick="openMirror(\'' + escAttr(serial) + '\')">' + DEV_MIRROR_ICON + ' Buka Mirror</button>' +
        '<button type="button" class="dev-btn-ghost"' + (detailDisabled ? ' disabled' : '') +
          ' onclick="showDeviceDetail(\'' + escAttr(serial) + '\')">' + DEV_DETAIL_ICON + ' Detail</button>' +
        runBtns +
      '</div></div>';
  }).join('');

  syncDevAutoOnToggle();
  const devPage = document.getElementById('page-devices');
  if (devPage) refreshLucideIcons(devPage);
}

async function relayoutMirrors() {
  const r = await fetch('/api/devices/mirror-relayout', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' });
  const text = await r.text();
  if (!r.ok) { log('Relayout gagal — ' + text.slice(0, 80)); return; }
  try {
    const data = JSON.parse(text);
    log('Mirror diatur: ' + (data.moved || 0) + ' dipindah, ' + (data.restarted || 0) + ' restart');
  } catch (e) { log('Mirror diatur ulang'); }
  await refreshState();
}

async function importDefaultFile() {
  await api('accounts/import', {});
  log('Import data/accounts.txt');
}

function onImportFileSelected(input) {
  const file = input.files && input.files[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = async () => {
    importPreviewContent = String(reader.result || '');
    await previewImportContent(importPreviewContent, file.name);
    input.value = '';
  };
  reader.readAsText(file);
}

async function previewImportContent(content, label) {
  const r = await fetch('/api/accounts/import-preview', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content })
  });
  const text = await r.text();
  if (!r.ok) { log('Preview gagal — ' + text.slice(0, 100)); return; }
  const data = JSON.parse(text);
  const box = document.getElementById('importPreview');
  const btn = document.getElementById('btnConfirmImport');
  if (!box) return;
  box.style.display = '';
  const rows = data.rows || [];
  box.innerHTML = '<div style="padding:6px 8px;background:#f8fafc;font-weight:600">' + esc(label || 'file') +
    ' — ' + (data.valid || 0) + '/' + (data.total || 0) + ' valid</div><table><thead><tr><th>#</th><th>Login</th><th>Nama</th><th>FP</th><th>Status</th></tr></thead><tbody>' +
    rows.map(row => '<tr><td>' + row.line_no + '</td><td>' + esc(row.login_id || '') + '</td><td>' + esc(row.name || '') +
    '</td><td>' + (row.fanpage_count || 0) + '</td><td>' + (row.valid ? '<span style="color:var(--ok-text)">OK</span>' : esc(row.error || 'invalid')) +
    '</td></tr>').join('') + '</tbody></table>';
  if (btn) btn.style.display = (data.valid > 0) ? '' : 'none';
  log('Preview ' + (data.valid || 0) + ' akun valid');
}

async function confirmImportPreview() {
  if (!importPreviewContent) return;
  await api('accounts/import', { content: importPreviewContent });
  importPreviewContent = '';
  const box = document.getElementById('importPreview');
  const btn = document.getElementById('btnConfirmImport');
  if (box) { box.style.display = 'none'; box.innerHTML = ''; }
  if (btn) btn.style.display = 'none';
  log('Import dari file selesai');
}

function setLogTab(tab) {
  logTab = tab;
  document.getElementById('logTabSession').classList.toggle('active', tab === 'session');
  document.getElementById('logTabServer').classList.toggle('active', tab === 'server');
  document.getElementById('logPanel').style.display = tab === 'session' ? '' : 'none';
  document.getElementById('serverLogPanel').style.display = tab === 'server' ? '' : 'none';
  if (tab === 'server') fetchServerLog();
}

const TEXT_TABS = [
  { id: 'personal', btn: 'textTabPersonal' },
  { id: 'fanpage', btn: 'textTabFanpage' },
  { id: 'group', btn: 'textTabGroup' },
];

function setTextTab(tab) {
  textTab = tab;
  TEXT_TABS.forEach(t => {
    const el = document.getElementById(t.btn);
    if (el) el.classList.toggle('active', t.id === tab);
  });
  loadPostTexts(tab);
}

async function loadPostTexts(category) {
  const list = document.getElementById('textList');
  const countEl = document.getElementById('textListCount');
  if (!list) return;
  try {
    const r = await fetch('/api/post-texts?category=' + encodeURIComponent(category) + '&_=' + Date.now(), { cache: 'no-store' });
    if (!r.ok) {
      const msg = await r.text();
      list.innerHTML = '<div class="empty" style="padding:12px;color:var(--warn)">' + esc(msg || 'Gagal memuat') + '</div>';
      return;
    }
    const data = await r.json();
    const items = data.items || [];
    if (countEl) countEl.textContent = items.length + ' teks';
    if (!items.length) {
      list.innerHTML = '<div class="empty" style="padding:12px">Belum ada teks — tambahkan di atas.</div>';
      return;
    }
    list.innerHTML = items.map(item => {
      const img = item.image_file ? ' <span class="script-tag">' + esc(item.image_file) + '</span>' : '';
      return '<div class="text-row">' +
        '<div class="text-row-body">' + esc(item.body) + img + '</div>' +
        '<button type="button" class="btn btn-sm btn-danger" onclick="deletePostText(' + item.id + ')">Hapus</button>' +
        '</div>';
    }).join('');
  } catch (e) {
    list.innerHTML = '<div class="empty" style="padding:12px">Error memuat teks</div>';
  }
}

async function addPostText() {
  const ta = document.getElementById('textInput');
  const img = document.getElementById('textImageFile');
  if (!ta) return;
  const body = (ta.value || '').trim();
  if (!body) { log('Teks kosong'); return; }
  await api('post-texts', {
    category: textTab,
    body: body,
    image_file: img ? (img.value || '').trim() : '',
  });
  ta.value = '';
  if (img) img.value = '';
  log('Teks ditambah (' + textTab + ')');
  loadPostTexts(textTab);
}

async function deletePostText(id) {
  if (!confirm('Hapus teks #' + id + '?')) return;
  await api('post-texts/delete', { id: id });
  log('Teks #' + id + ' dihapus');
  loadPostTexts(textTab);
}

async function fetchServerLog() {
  try {
    const r = await fetch('/api/logs/tail?lines=100');
    if (!r.ok) return;
    const data = await r.json();
    const el = document.getElementById('serverLogPanel');
    if (!el) return;
    const lines = data.lines || [];
    el.textContent = lines.length ? lines.join('\n') : '(log kosong — jalankan panel dari zautopanel.exe)';
  } catch (e) {}
}

function setupKeyboardShortcuts() {
  document.addEventListener('keydown', (e) => {
    if (e.target && (e.target.matches('input, textarea, select') || e.target.isContentEditable)) return;
    if (e.key === 'F5') { e.preventDefault(); api('devices/refresh'); log('Refresh (F5)'); }
    if (e.key === 'F11') { e.preventDefault(); toggleWindowFullscreen(); }
    if (e.ctrlKey && (e.key === 'r' || e.key === 'R')) { e.preventDefault(); doRun(); }
    if (e.key === 'Escape') {
      if (document.getElementById('accDrawer')?.classList.contains('open')) {
        closeAccDrawer();
        return;
      }
      const s = (state.run_status || '').toLowerCase();
      if (s === 'running' || s === 'paused') api('stop');
    }
  });
}

async function importAccounts() {
  await importDefaultFile();
}

async function autoAssign() {
  const r = await fetch('/api/accounts/auto-assign', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' });
  const text = await r.text();
  if (!r.ok) { log('Assign gagal — ' + text.slice(0, 80)); return; }
  try {
    const data = JSON.parse(text);
    log('Assign OK — ' + (data.assigned || 0) + ' HP');
  } catch (e) { log('Assign selesai'); }
  const st = await fetch('/api/state');
  if (st.ok) { state = await st.json(); render(); }
}

function scriptIcon(app) {
  const a = (app || '').toLowerCase();
  if (a.includes('facebook') || a.includes('fb')) {
    return '<div class="script-icon">' + lucideIcon('facebook') + '</div>';
  }
  return '<div class="script-icon script-icon--muted">' + lucideIcon('box') + '</div>';
}

function render() {
  try { setButtons(); } catch (e) { console.error('setButtons', e); }
  try { renderDashboard(); } catch (e) { console.error('renderDashboard', e); }
  try { updateSettingsSaveUI(); } catch (e) { console.error('updateSettingsSaveUI', e); }
  if (!inputsInit) {
    try {
      applySettingsForm(state.settings || {
        max_devices: state.max_devices,
        parallel_workers: state.workers
      });
      clearSettingsDirty();
      inputsInit = true;
    } catch (e) { console.error('applySettingsForm', e); }
  }

  const ec = state.enabled_count || 0;
  const mo = state.mirror_open_count || 0;
  const ac = state.account_count || 0;
  const asg = state.assigned_count || 0;
  const wk = state.workers ?? document.getElementById('workers')?.value ?? 2;
  const statsLine = document.getElementById('statsLine');
  if (statsLine) {
    statsLine.innerHTML =
      '<strong>' + ec + '</strong> switch ON · <strong>' + mo + '</strong> mirror terbuka · <strong>' + wk + '</strong> Worker · <strong>' + ac + '</strong> Akun';
  }
  const accEl = document.getElementById('accountsStats');
  if (accEl) {
    const fpTotal = (state.accounts || []).reduce((sum, a) => sum + (a.fanpage_count || 0), 0);
    accEl.innerHTML = '<strong>' + ac + '</strong> akun di database · <strong>' + asg + '</strong> HP sudah di-assign · <strong>' + fpTotal + '</strong> fanpage';
  }

  // Device page first — jangan tertahan error di halaman lain
  if (pagesReady) {
    try { renderDevicesPage(); } catch (e) { console.error('renderDevicesPage', e); }
  }

  try { renderAccountsList(); } catch (e) { console.error('renderAccountsList', e); }

  const taskList = document.getElementById('taskList');
  if (!taskList) return;
  try {
    taskList.innerHTML = (state.tasks || []).length
      ? (state.tasks || []).map(t => {
          const steps = (t.steps || []).map(s => '<span class="script-step">' + esc(s) + '</span>').join('<span class="script-step-arrow">→</span>');
          const desc = t.description ? '<div class="script-desc">' + esc(t.description) + '</div>' : '';
          return '<div class="script-row">' +
            scriptIcon(t.app) +
            '<input type="checkbox" class="script-check"' + (t.active ? ' checked' : '') +
            ' onchange="toggleTask(\'' + esc(t.name) + '\')">' +
            '<div class="script-body">' +
              '<div class="script-name">' + esc(t.name) +
              ' <span class="script-tag">' + esc(t.app || 'app') + '</span></div>' +
              (steps ? '<div class="script-pipeline">' + steps + '</div>' : '') +
              desc +
            '</div>' +
            '<button class="btn-edit" type="button" title="Edit di config/config.json">Edit Skrip</button></div>';
        }).join('')
      : '<div class="empty">Tidak ada skrip di config.</div>';
    refreshLucideIcons(taskList);
  } catch (e) { console.error('taskList', e); }
}

let evtSource = null;
let stateRev = 0;
let uiRev = 0;
function connectSSE() {
  if (evtSource) evtSource.close();
  setSSEStatus(false, true);
  evtSource = new EventSource('/api/events');
  evtSource.onopen = () => { setSSEStatus(true, false); };
  evtSource.onmessage = (e) => {
    try {
      const data = JSON.parse(e.data);
      applyState(data);
    }
    catch (err) { log('state parse gagal'); }
  };
  evtSource.onerror = () => {
    setSSEStatus(false, false);
    evtSource.close();
    evtSource = null;
    setTimeout(connectSSE, 3000);
  };
}

async function loadPages() {
  const container = document.getElementById('pages');
  if (!container) return;
  pagesReady = false;
  const parts = [];
  for (const name of PAGE_ORDER) {
    const r = await fetch('/assets/pages/' + name + '.html?v=' + UI_ASSET_VER, { cache: 'no-store' });
    if (!r.ok) throw new Error('page load failed: ' + name);
    parts.push(await r.text());
  }
  container.innerHTML = parts.join('\n');
  setupSettingsForm();
  renderLog();
  if (currentPage === 'text') loadPostTexts(textTab);
  refreshLucideIcons(container);
  pagesReady = true;
  if (state && (state.devices || state.run_status != null)) {
    try { renderDevicesPage(); } catch (e) { console.error('renderDevicesPage after loadPages', e); }
  }
}

async function startPanelServices() {
  connectSSE();
  startStatePolling();
  startUIRevPolling();
  setupPageSwipe();
  setupKeyboardShortcuts();
  syncWindowChrome();
  setTimeout(refreshState, 1500);
}

async function bootPanel() {
  bootAt = Date.now();
  try {
    await loadPages();
  } catch (e) {
    const c = document.getElementById('pages');
    if (c) c.innerHTML = '<div class="empty">Gagal memuat halaman — refresh panel (F5).</div>';
    console.error(e);
    log('Gagal memuat halaman UI');
    await startPanelServices();
    return;
  }

  try {
    await fetchStateAndRender();
  } catch (e) {
    showApiError('API tidak terjangkau — ' + (e.message || e));
    log('koneksi API gagal');
  }

  await startPanelServices();
  refreshLucideIcons(document.body);
}

function setupPageSwipe() {
  const pages = document.getElementById('pages');
  if (!pages) return;
  let startX = 0, startY = 0;
  pages.addEventListener('touchstart', (e) => {
    if (!e.touches.length) return;
    startX = e.touches[0].clientX;
    startY = e.touches[0].clientY;
  }, { passive: true });
  pages.addEventListener('touchend', (e) => {
    if (!e.changedTouches.length) return;
    const dx = e.changedTouches[0].clientX - startX;
    const dy = e.changedTouches[0].clientY - startY;
    if (Math.abs(dx) < 60 || Math.abs(dy) > Math.abs(dx)) return;
    const idx = PAGE_ORDER.indexOf(currentPage);
    if (idx < 0) return;
    if (dx < 0 && idx < PAGE_ORDER.length - 1) showPage(PAGE_ORDER[idx + 1]);
    else if (dx > 0 && idx > 0) showPage(PAGE_ORDER[idx - 1]);
  }, { passive: true });
}

bootPanel();

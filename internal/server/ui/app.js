let previewState = { title: '', format: '', nodes: [], selected: [] };
let upstreamEditorState = { index: -1 };
let bindingEditorState = { index: -1 };
let nodeEditorState = { index: -1 };
let subscriptionEditorState = { index: -1 };

async function api(url, options = {}) {
  const res = await fetch(url, {headers: {'Content-Type': 'application/json'}, credentials: 'same-origin', ...options});
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { data = text; }
  if (!res.ok) throw new Error((data && data.error) || text || ('HTTP ' + res.status));
  return data;
}

function metric(label, value) {
  const div = document.createElement('div');
  div.className = 'metric';
  div.innerHTML = `<div class="label">${label}</div><div class="value">${value ?? '-'}</div>`;
  return div;
}

function setLog(id, msg) { const el = document.getElementById(id); if (el) el.textContent = msg; }
function setField(name, value) { const el = document.querySelector(`[name="${name}"]`); if (el) el.value = value ?? ''; }
function readField(name) { const el = document.querySelector(`[name="${name}"]`); return el ? el.value : ''; }
function byId(id) { return document.getElementById(id); }
function valueOf(id) { return byId(id)?.value ?? ''; }
function setValue(id, value) { const el = byId(id); if (el) el.value = value ?? ''; }
function getEditorConfig() { return JSON.parse(byId('config-json').value); }
function setEditorConfig(cfg) { const editor = byId('config-json'); if (editor) editor.value = JSON.stringify(cfg, null, 2); }
function ensureArray(cfg, key) { if (!Array.isArray(cfg[key])) cfg[key] = []; }
function splitTags(s) { return String(s || '').split(',').map(v => v.trim()).filter(Boolean); }
function splitCSV(s) { return String(s || '').split(',').map(v => v.trim()).filter(Boolean); }
function escapeHTML(s) { return String(s ?? '').replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;').replaceAll('"', '&quot;'); }

function getPreviewNodes() { return previewState.nodes.filter((_, idx) => previewState.selected[idx]); }

function renderPreviewTable() {
  const wrap = byId('preview-table-wrap'); const tbody = byId('preview-tbody'); const summary = byId('preview-summary'); const checkAll = byId('preview-check-all');
  if (!wrap || !tbody || !summary) return;
  tbody.innerHTML = '';
  if (!previewState.nodes.length) { wrap.classList.add('hidden'); summary.textContent = '当前没有预览节点。'; if (checkAll) checkAll.checked = false; return; }
  wrap.classList.remove('hidden');
  const selectedCount = previewState.selected.filter(Boolean).length;
  summary.textContent = `预览格式：${previewState.format || '-'} ｜ 节点数：${previewState.nodes.length} ｜ 已选中：${selectedCount}`;
  if (checkAll) checkAll.checked = selectedCount === previewState.nodes.length;
  previewState.nodes.forEach((node, idx) => {
    const tr = document.createElement('tr'); const tags = Array.isArray(node.tags) ? node.tags : [];
    tr.innerHTML = `<td><input type="checkbox" data-preview-index="${idx}" ${previewState.selected[idx] ? 'checked' : ''} /></td><td>${escapeHTML(node.name || '')}<small>${escapeHTML(node.id || '')}</small></td><td>${escapeHTML(node.scheme || '')}</td><td>${escapeHTML(node.host || '')}</td><td>${node.port ?? ''}</td><td>${escapeHTML(node.subscription_id || '')}</td><td><div class="tag-list">${tags.map(tag => `<span class="tag">${escapeHTML(tag)}</span>`).join('')}</div></td>`;
    tbody.appendChild(tr);
  });
}

function renderPreview(nodes, title, format = '') {
  previewState = { title, format, nodes: Array.isArray(nodes) ? nodes : [], selected: (Array.isArray(nodes) ? nodes : []).map(() => true) };
  renderPreviewTable();
  setLog('import-preview', JSON.stringify({title, format, count: nodes.length, nodes}, null, 2));
}

function setPreviewSelection(checked) { previewState.selected = previewState.nodes.map(() => checked); renderPreviewTable(); }

function renderGenericList(containerId, items, activeIndex, titleFn, metaFn, attr) {
  const box = byId(containerId); if (!box) return; box.innerHTML = '';
  items.forEach((item, idx) => {
    const div = document.createElement('div');
    div.className = 'list-item' + (activeIndex === idx ? ' active' : '');
    div.setAttribute(attr, String(idx));
    div.innerHTML = `<div class="title">${escapeHTML(titleFn(item, idx))}</div><div class="meta">${escapeHTML(metaFn(item, idx))}</div>`;
    box.appendChild(div);
  });
}

function renderUpstreamList() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'upstreams');
  renderGenericList('upstream-list', cfg.upstreams, upstreamEditorState.index, (u, i) => u.name || u.id || `upstream-${i+1}`, u => `${u.base_url || ''} · ${u.binding_id || ''}`, 'data-upstream-index');
}
function renderBindingList() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'bindings');
  renderGenericList('binding-list', cfg.bindings, bindingEditorState.index, (b, i) => b.id || `binding-${i+1}`, b => `${b.upstream_id || ''} -> ${b.node_id || ''} · ${b.mode || ''}`, 'data-binding-index');
}
function renderNodeList() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'proxy_nodes');
  renderGenericList('node-list', cfg.proxy_nodes, nodeEditorState.index, (n, i) => n.name || n.id || `node-${i+1}`, n => `${n.scheme || ''} · ${n.host || ''}:${n.port ?? ''}`, 'data-node-index');
}
function renderSubscriptionList() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'subscriptions');
  renderGenericList('subscription-list', cfg.subscriptions, subscriptionEditorState.index, (s, i) => s.name || s.id || `sub-${i+1}`, s => s.url || '', 'data-subscription-index');
}

function renderFailoverSteps() {
  const cfg = getEditorConfig();
  if (!cfg.failover || typeof cfg.failover !== 'object') cfg.failover = {};
  ensureArray(cfg.failover, 'cooldown_steps');
  const tbody = byId('failover-steps-tbody');
  if (!tbody) return;
  tbody.innerHTML = '';
  cfg.failover.cooldown_steps.forEach((step, idx) => {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td><input data-failover-after="${idx}" type="number" min="1" value="${Number(step.after_failures ?? 1)}" /></td><td><input data-failover-duration="${idx}" type="number" min="0" value="${Number(step.duration_seconds ?? 0)}" /></td><td><button class="btn" type="button" data-action="failover-step-delete" data-failover-index="${idx}">删除</button></td>`;
    tbody.appendChild(tr);
  });
}

function syncFailoverStepsFromTable() {
  const cfg = getEditorConfig();
  if (!cfg.failover || typeof cfg.failover !== 'object') cfg.failover = {};
  const afterInputs = Array.from(document.querySelectorAll('[data-failover-after]'));
  const steps = afterInputs.map((input) => {
    const idx = input.getAttribute('data-failover-after');
    const durationInput = document.querySelector(`[data-failover-duration="${idx}"]`);
    return {
      after_failures: Number(input.value || 1),
      duration_seconds: Number(durationInput?.value || 0),
    };
  });
  cfg.failover.cooldown_steps = steps;
  setEditorConfig(cfg);
  renderFailoverSteps();
}

function loadUpstreamForm(index) {
  const cfg = getEditorConfig(); ensureArray(cfg, 'upstreams');
  const u = cfg.upstreams[index] || { id:'', name:'', base_url:'', api_key:'', models:[], binding_id:'', priority:100, enabled:true, timeout_seconds:120 };
  upstreamEditorState.index = index;
  setValue('upstream-id', u.id); setValue('upstream-name', u.name); setValue('upstream-base-url', u.base_url); setValue('upstream-api-key', u.api_key); setValue('upstream-binding-id', u.binding_id); setValue('upstream-priority', u.priority ?? 100); setValue('upstream-enabled', String(u.enabled !== false)); setValue('upstream-timeout-seconds', u.timeout_seconds); setValue('upstream-models', Array.isArray(u.models) ? u.models.join(',') : '');
  renderUpstreamList();
}
function saveUpstreamForm() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'upstreams');
  const u = { id:valueOf('upstream-id').trim(), name:valueOf('upstream-name').trim(), base_url:valueOf('upstream-base-url').trim(), api_key:valueOf('upstream-api-key'), models:splitCSV(valueOf('upstream-models')), binding_id:valueOf('upstream-binding-id').trim(), priority:Number(valueOf('upstream-priority') || 100), enabled:valueOf('upstream-enabled') === 'true', timeout_seconds:Number(valueOf('upstream-timeout-seconds') || 0) };
  if (upstreamEditorState.index >= 0 && upstreamEditorState.index < cfg.upstreams.length) cfg.upstreams[upstreamEditorState.index] = u; else { cfg.upstreams.push(u); upstreamEditorState.index = cfg.upstreams.length - 1; }
  setEditorConfig(cfg); renderUpstreamList(); setLog('config-log', '上游表单已同步到 JSON 编辑区，记得保存配置。');
}
function deleteUpstreamForm() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'upstreams'); if (upstreamEditorState.index < 0 || upstreamEditorState.index >= cfg.upstreams.length) return;
  cfg.upstreams.splice(upstreamEditorState.index, 1); setEditorConfig(cfg); upstreamEditorState.index = -1; loadUpstreamForm(-1); renderUpstreamList(); setLog('config-log', '已从 JSON 编辑区删除当前上游，记得保存配置。');
}

function renderSelectOptions(selectId, options, selected) {
  const sel = byId(selectId);
  if (!sel) return;
  sel.innerHTML = '';
  options.forEach(opt => {
    const option = document.createElement('option');
    option.value = opt;
    option.textContent = opt;
    if (opt === selected) option.selected = true;
    sel.appendChild(option);
  });
}

function loadBindingForm(index) {
  const cfg = getEditorConfig(); ensureArray(cfg, 'bindings');
  const b = cfg.bindings[index] || { id:'', upstream_id:'', node_id:'', mode:'fixed' };
  bindingEditorState.index = index;
  setValue('binding-id', b.id);
  // Populate upstream and node selects from current config
  const upstreamIds = (cfg.upstreams || []).map(u => u.id);
  const nodeIds = (cfg.proxy_nodes || []).map(n => n.id);
  renderSelectOptions('binding-upstream-id', upstreamIds, b.upstream_id);
  renderSelectOptions('binding-node-id', nodeIds, b.node_id);
  setValue('binding-mode', b.mode);
  renderBindingList();
}
function saveBindingForm() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'bindings');
  const upstreamId = valueOf('binding-upstream-id').trim();
  const nodeId = valueOf('binding-node-id').trim();
  if (!upstreamId || !nodeId) {
    setLog('config-log', '错误：必须选择 Upstream ID 和 Node ID');
    return;
  }
  const b = { id:valueOf('binding-id').trim(), upstream_id:upstreamId, node_id:nodeId, mode:valueOf('binding-mode').trim() };
  if (bindingEditorState.index >= 0 && bindingEditorState.index < cfg.bindings.length) cfg.bindings[bindingEditorState.index] = b; else { cfg.bindings.push(b); bindingEditorState.index = cfg.bindings.length - 1; }
  setEditorConfig(cfg); renderBindingList(); setLog('config-log', '绑定表单已同步到 JSON 编辑区，记得保存配置。');
}
function deleteBindingForm() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'bindings'); if (bindingEditorState.index < 0 || bindingEditorState.index >= cfg.bindings.length) return;
  cfg.bindings.splice(bindingEditorState.index, 1); setEditorConfig(cfg); bindingEditorState.index = -1; loadBindingForm(-1); renderBindingList(); setLog('config-log', '已从 JSON 编辑区删除当前绑定，记得保存配置。');
}

function loadNodeForm(index) {
  const cfg = getEditorConfig(); ensureArray(cfg, 'proxy_nodes');
  const node = cfg.proxy_nodes[index] || { id:'', name:'', scheme:'vless', host:'', port:443, subscription_id:'manual', tags:[], raw_uri:'' };
  nodeEditorState.index = index;
  setValue('node-id', node.id); setValue('node-name', node.name); setValue('node-scheme', node.scheme); setValue('node-host', node.host); setValue('node-port', node.port); setValue('node-subscription-id', node.subscription_id); setValue('node-tags', Array.isArray(node.tags) ? node.tags.join(',') : ''); setValue('node-raw-uri', node.raw_uri);
  renderNodeList();
}
function saveNodeForm() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'proxy_nodes');
  const node = { id:valueOf('node-id').trim(), name:valueOf('node-name').trim(), scheme:valueOf('node-scheme').trim(), host:valueOf('node-host').trim(), port:Number(valueOf('node-port') || 0), subscription_id:valueOf('node-subscription-id').trim(), tags:splitTags(valueOf('node-tags')), raw_uri:valueOf('node-raw-uri') };
  if (nodeEditorState.index >= 0 && nodeEditorState.index < cfg.proxy_nodes.length) cfg.proxy_nodes[nodeEditorState.index] = node; else { cfg.proxy_nodes.push(node); nodeEditorState.index = cfg.proxy_nodes.length - 1; }
  setEditorConfig(cfg); renderNodeList(); setLog('config-log', '节点表单已同步到 JSON 编辑区，记得保存配置。');
}
function deleteNodeForm() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'proxy_nodes'); if (nodeEditorState.index < 0 || nodeEditorState.index >= cfg.proxy_nodes.length) return;
  cfg.proxy_nodes.splice(nodeEditorState.index, 1); setEditorConfig(cfg); nodeEditorState.index = -1; loadNodeForm(-1); renderNodeList(); setLog('config-log', '已从 JSON 编辑区删除当前节点，记得保存配置。');
}

function loadSubscriptionForm(index) {
  const cfg = getEditorConfig(); ensureArray(cfg, 'subscriptions');
  const sub = cfg.subscriptions[index] || { id:'', name:'', url:'', refresh_interval_seconds:3600 };
  subscriptionEditorState.index = index;
  setValue('subscription-id', sub.id); setValue('subscription-name', sub.name); setValue('subscription-url', sub.url); setValue('subscription-refresh', sub.refresh_interval_seconds);
  renderSubscriptionList();
}
function saveSubscriptionForm() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'subscriptions');
  const sub = { id:valueOf('subscription-id').trim(), name:valueOf('subscription-name').trim(), url:valueOf('subscription-url').trim(), refresh_interval_seconds:Number(valueOf('subscription-refresh') || 0) };
  if (subscriptionEditorState.index >= 0 && subscriptionEditorState.index < cfg.subscriptions.length) cfg.subscriptions[subscriptionEditorState.index] = sub; else { cfg.subscriptions.push(sub); subscriptionEditorState.index = cfg.subscriptions.length - 1; }
  setEditorConfig(cfg); renderSubscriptionList(); setLog('config-log', '订阅表单已同步到 JSON 编辑区，记得保存配置。');
}
function deleteSubscriptionForm() {
  const cfg = getEditorConfig(); ensureArray(cfg, 'subscriptions'); if (subscriptionEditorState.index < 0 || subscriptionEditorState.index >= cfg.subscriptions.length) return;
  cfg.subscriptions.splice(subscriptionEditorState.index, 1); setEditorConfig(cfg); subscriptionEditorState.index = -1; loadSubscriptionForm(-1); renderSubscriptionList(); setLog('config-log', '已从 JSON 编辑区删除当前订阅，记得保存配置。');
}

function mergeNodesIntoConfig(nodes) {
  if (!nodes.length) throw new Error('当前没有选中的预览节点');
  const cfg = getEditorConfig(); ensureArray(cfg, 'proxy_nodes');
  const existing = new Set(cfg.proxy_nodes.map(n => n.id || n.raw_uri)); let added = 0;
  for (const node of nodes) { const key = node.id || node.raw_uri; if (existing.has(key)) continue; cfg.proxy_nodes.push(node); existing.add(key); added += 1; }
  setEditorConfig(cfg); renderNodeList(); setLog('config-log', `已把 ${added} 个选中预览节点加入当前配置编辑区。记得点击“保存配置”落盘。`);
}

function renderExitIPProbe(data) {
  const tbody = byId('exit-ip-tbody');
  if (!tbody) return;
  tbody.innerHTML = '';
  const tr = document.createElement('tr');
  tr.innerHTML = `<td>${escapeHTML(data.direct_ip || '')}</td><td>${escapeHTML(data.proxy_ip || '')}</td><td>${escapeHTML(data.proxy_address || '')}</td><td>${data.proxy_active ? 'yes' : 'no'}</td><td>${data.same_exit ? 'yes' : 'no'}</td>`;
  tbody.appendChild(tr);
  setLog('exit-ip-log', JSON.stringify(data || {}, null, 2));
}

async function loadExitIPProbe() {
  try {
    const data = await api('/api/diagnostics/exit-ip');
    renderExitIPProbe(data || {});
  } catch (err) {
    setLog('exit-ip-log', '出口 IP 自检失败: ' + err.message);
  }
}

function renderRouteHealth(routes) {
  const tbody = byId('route-health-tbody');
  if (!tbody) return;
  tbody.innerHTML = '';
  for (const r of (routes || [])) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td>${escapeHTML(r.upstream_id || '')}</td><td>${escapeHTML(r.binding_id || '')}</td><td>${escapeHTML(r.node_id || '')}</td><td>${r.consecutive_failures ?? 0}</td><td>${r.total_successes ?? 0}</td><td>${r.total_failures ?? 0}</td><td>${r.is_cooling_down ? 'yes' : 'no'}</td><td>${r.cooldown_seconds ?? 0}</td><td>${escapeHTML(r.last_error || '')}</td>`;
    tbody.appendChild(tr);
  }
  setLog('route-health-log', JSON.stringify(routes || [], null, 2));
}

async function loadRouteHealth() {
  try {
    const data = await api('/api/health/routes');
    renderRouteHealth(data.routes || []);
  } catch (err) {
    setLog('route-health-log', '加载线路状态失败: ' + err.message);
  }
}

async function loadStatus() {
  const grid = byId('status-grid'); if (!grid) return; grid.innerHTML = '';
  try { const health = await api('/healthz'); const cfg = await api('/api/config'); grid.appendChild(metric('健康状态', health.ok ? 'ok' : 'bad')); grid.appendChild(metric('监听地址', cfg.config.server.listen)); grid.appendChild(metric('上游数量', cfg.config.upstreams.length)); grid.appendChild(metric('绑定数量', cfg.config.bindings.length)); grid.appendChild(metric('节点数量', cfg.config.proxy_nodes.length)); grid.appendChild(metric('订阅数量', cfg.config.subscriptions.length)); } catch (err) { grid.appendChild(metric('状态', '加载失败')); grid.appendChild(metric('错误', err.message)); }
}

async function loadConfig() {
  const data = await api('/api/config'); const cfg = data.config;
  setField('server.listen', cfg.server.listen); setField('server.admin_token', ''); setField('runtime.dir', cfg.runtime.dir); setField('runtime.xray_binary', cfg.runtime.xray_binary); setField('runtime.base_port', cfg.runtime.base_port); setField('runtime.healthcheck_url', cfg.runtime.healthcheck_url); setField('runtime.subscription_cache_file', cfg.runtime.subscription_cache_file);
  setEditorConfig(cfg);
  renderUpstreamList(); renderBindingList(); renderNodeList(); renderSubscriptionList(); renderFailoverSteps();
  loadUpstreamForm(cfg.upstreams?.length ? 0 : -1); loadBindingForm(cfg.bindings?.length ? 0 : -1); loadNodeForm(cfg.proxy_nodes?.length ? 0 : -1); loadSubscriptionForm(cfg.subscriptions?.length ? 0 : -1);
  await loadRouteHealth();
  await loadExitIPProbe();
  setLog('config-log', `已加载配置: ${data.path}\n说明：控制台密钥不会回显；修改请使用“修改控制台密钥”按钮。`);
}

async function saveConfig() {
  const cfg = getEditorConfig(); cfg.server.listen = readField('server.listen'); cfg.runtime.dir = readField('runtime.dir'); cfg.runtime.xray_binary = readField('runtime.xray_binary'); cfg.runtime.base_port = Number(readField('runtime.base_port')); cfg.runtime.healthcheck_url = readField('runtime.healthcheck_url'); cfg.runtime.subscription_cache_file = readField('runtime.subscription_cache_file'); syncFailoverStepsFromTable();
  const data = await api('/api/config', {method: 'POST', body: JSON.stringify({config: cfg})});
  setEditorConfig(data.config); renderUpstreamList(); renderBindingList(); renderNodeList(); renderSubscriptionList(); renderFailoverSteps();
  await loadRouteHealth();
  setLog('config-log', '保存成功。已自动备份旧配置到 configs/.history/ 。');
}

async function applyConfig() {
  const cfg = getEditorConfig(); cfg.server.listen = readField('server.listen'); cfg.runtime.dir = readField('runtime.dir'); cfg.runtime.xray_binary = readField('runtime.xray_binary'); cfg.runtime.base_port = Number(readField('runtime.base_port')); cfg.runtime.healthcheck_url = readField('runtime.healthcheck_url'); cfg.runtime.subscription_cache_file = readField('runtime.subscription_cache_file'); syncFailoverStepsFromTable();
  const data = await api('/api/config/apply', {method: 'POST', body: JSON.stringify({config: cfg})});
  setEditorConfig(data.config); renderUpstreamList(); renderBindingList(); renderNodeList(); renderSubscriptionList(); renderFailoverSteps();
  await loadRouteHealth();
  const nodeCount = data.result?.summary?.node_count ?? data.result?.flat_result?.nodes?.length ?? 0;
  const generated = data.result?.summary?.generated_count ?? data.result?.flat_result?.generated_xray?.length ?? 0;
  setLog('config-log', `保存并应用成功。节点数=${nodeCount}，生成配置数=${generated}。已刷新路由与 Xray 产物。`);
}

async function previewURI() { const data = await api('/api/import/uri', {method: 'POST', body: JSON.stringify({raw_uri: valueOf('import-uri').trim()})}); renderPreview(data.nodes || [], 'URI 解析预览', data.format || 'uri'); }
async function previewSubscription() { const payload = { id: valueOf('import-sub-id').trim(), name: valueOf('import-sub-name').trim(), url: valueOf('import-sub-url').trim() }; const data = await api('/api/import/subscription', {method: 'POST', body: JSON.stringify(payload)}); renderPreview(data.nodes || [], '订阅抓取预览', data.format || 'remote'); }
async function previewCurrentSubscriptionForm() { const payload = { id: valueOf('subscription-id').trim(), name: valueOf('subscription-name').trim(), url: valueOf('subscription-url').trim() }; const data = await api('/api/import/subscription', {method: 'POST', body: JSON.stringify(payload)}); renderPreview(data.nodes || [], '当前订阅表单预览', data.format || 'remote'); }
async function previewRawImport() { const payload = { id: valueOf('import-sub-id').trim(), name: valueOf('import-sub-name').trim(), text: valueOf('import-raw-text') }; const data = await api('/api/import/subscription', {method: 'POST', body: JSON.stringify(payload)}); renderPreview(data.nodes || [], '文本导入预览', data.format || 'raw'); }

async function restartServer() {
  try {
    await api('/api/restart', {method: 'POST'});
    setLog('config-log', '已请求重启，页面将在 2 秒后刷新。');
    setTimeout(() => location.reload(), 2000);
  } catch (e) {
    setLog('config-log', '重启请求失败: ' + e.message);
  }
}
async function loadBootstrap(run) {
  const data = await api(run ? '/api/bootstrap/run' : '/api/bootstrap', {method: run ? 'POST' : 'GET'});
  const result = data.result || null; const summary = byId('bootstrap-summary'); const json = byId('bootstrap-json'); if (!summary || !json) return;
  summary.innerHTML = ''; if (!result) { summary.appendChild(metric('状态', '暂无结果')); json.textContent = JSON.stringify(data, null, 2); return; }
  summary.appendChild(metric('节点数', result.summary?.node_count ?? result.flat_result?.nodes?.length ?? 0)); summary.appendChild(metric('生成配置数', result.summary?.generated_count ?? result.flat_result?.generated_xray?.length ?? 0)); summary.appendChild(metric('运行阶段', result.runtime_stage ? 'available' : 'none')); summary.appendChild(metric('节点阶段', result.node_stage ? 'available' : 'none')); json.textContent = JSON.stringify(result, null, 2);
}

async function login() { await api('/api/login', {method: 'POST', body: JSON.stringify({token: byId('login-token').value})}); window.location.href = '/'; }
async function logout() { await api('/api/logout', {method: 'POST', body: JSON.stringify({})}); window.location.href = '/login'; }
async function changeToken() { await api('/api/admin/token', {method: 'POST', body: JSON.stringify({token: byId('new-admin-token').value})}); setLog('config-log', '控制台密钥已修改并更新当前登录态。'); byId('new-admin-token').value = ''; byId('token-panel').classList.add('hidden'); }

function addTemplate(kind) {
  const cfg = getEditorConfig();
  if (kind === 'upstream') { ensureArray(cfg, 'upstreams'); cfg.upstreams.push({id:'upstream-' + (cfg.upstreams.length + 1), name:'', base_url:'', api_key:'', models:[], binding_id:'', enabled:true, timeout_seconds:120}); upstreamEditorState.index = cfg.upstreams.length - 1; }
  if (kind === 'binding') { ensureArray(cfg, 'bindings'); cfg.bindings.push({id:'binding-' + (cfg.bindings.length + 1), upstream_id:'', node_id:'', mode:'fixed'}); bindingEditorState.index = cfg.bindings.length - 1; }
  if (kind === 'node') { ensureArray(cfg, 'proxy_nodes'); cfg.proxy_nodes.push({id:'node-' + (cfg.proxy_nodes.length + 1), name:'', scheme:'vless', host:'', port:443, subscription_id:'manual', tags:[], raw_uri:''}); nodeEditorState.index = cfg.proxy_nodes.length - 1; }
  if (kind === 'subscription') { ensureArray(cfg, 'subscriptions'); cfg.subscriptions.push({id:'sub-' + (cfg.subscriptions.length + 1), name:'', url:'', refresh_interval_seconds:3600}); subscriptionEditorState.index = cfg.subscriptions.length - 1; }
  if (kind === 'failover-step') { if (!cfg.failover || typeof cfg.failover !== 'object') cfg.failover = {}; ensureArray(cfg.failover, 'cooldown_steps'); cfg.failover.cooldown_steps.push({after_failures: 1, duration_seconds: 10}); }
  setEditorConfig(cfg); renderUpstreamList(); renderBindingList(); renderNodeList(); renderSubscriptionList(); renderFailoverSteps();
  if (kind === 'upstream') loadUpstreamForm(upstreamEditorState.index); if (kind === 'binding') loadBindingForm(bindingEditorState.index); if (kind === 'node') loadNodeForm(nodeEditorState.index); if (kind === 'subscription') loadSubscriptionForm(subscriptionEditorState.index);
  setLog('config-log', `已新增 ${kind} 模板。`);
}

document.addEventListener('change', (e) => {
  const cb = e.target.closest('[data-preview-index]');
  if (cb) { previewState.selected[Number(cb.getAttribute('data-preview-index'))] = cb.checked; renderPreviewTable(); return; }
  if (e.target && e.target.id === 'preview-check-all') setPreviewSelection(e.target.checked);
  if (e.target && (e.target.hasAttribute('data-failover-after') || e.target.hasAttribute('data-failover-duration'))) syncFailoverStepsFromTable();
});

document.addEventListener('click', async (e) => {
  const btn = e.target.closest('[data-action], [data-upstream-index], [data-binding-index], [data-node-index], [data-subscription-index]');
  if (!btn) return;
  try {
    if (btn.hasAttribute('data-upstream-index')) { loadUpstreamForm(Number(btn.getAttribute('data-upstream-index'))); return; }
    if (btn.hasAttribute('data-binding-index')) { loadBindingForm(Number(btn.getAttribute('data-binding-index'))); return; }
    if (btn.hasAttribute('data-node-index')) { loadNodeForm(Number(btn.getAttribute('data-node-index'))); return; }
    if (btn.hasAttribute('data-subscription-index')) { loadSubscriptionForm(Number(btn.getAttribute('data-subscription-index'))); return; }
    const action = btn.getAttribute('data-action');
    if (action === 'refresh-status') await loadStatus();
    if (action === 'reload-config') await loadConfig();
    if (action === 'refresh-route-health') await loadRouteHealth();
    if (action === 'refresh-exit-ip') await loadExitIPProbe();
    if (action === 'save-config') await saveConfig();
    if (action === 'apply-config') await applyConfig();
    if (action === 'restart') await restartServer();
    if (action === 'run-bootstrap') await loadBootstrap(true);
    if (action === 'login') await login();
    if (action === 'logout') await logout();
    if (action === 'change-token') byId('token-panel')?.classList.toggle('hidden');
    if (action === 'submit-token') await changeToken();
    if (action === 'add-upstream') addTemplate('upstream');
    if (action === 'add-binding') addTemplate('binding');
    if (action === 'add-node') addTemplate('node');
    if (action === 'add-subscription') addTemplate('subscription');
    if (action === 'failover-step-new') addTemplate('failover-step');
    if (action === 'failover-step-sync') { renderFailoverSteps(); }
    if (action === 'failover-step-delete') {
      const cfg = getEditorConfig();
      if (!cfg.failover || typeof cfg.failover !== 'object') cfg.failover = {};
      ensureArray(cfg.failover, 'cooldown_steps');
      const idx = Number(btn.getAttribute('data-failover-index'));
      if (idx >= 0 && idx < cfg.failover.cooldown_steps.length) {
        cfg.failover.cooldown_steps.splice(idx, 1);
        setEditorConfig(cfg);
        renderFailoverSteps();
        setLog('config-log', '已删除一个 failover 冷却档位。');
      }
    }
    if (action === 'preview-uri') await previewURI();
    if (action === 'preview-subscription') await previewSubscription();
    if (action === 'preview-raw-import') await previewRawImport();
    if (action === 'append-preview-nodes') mergeNodesIntoConfig(getPreviewNodes());
    if (action === 'preview-select-all') setPreviewSelection(true);
    if (action === 'preview-select-none') setPreviewSelection(false);
    if (action === 'upstream-form-new') loadUpstreamForm(-1);
    if (action === 'upstream-form-sync') { renderUpstreamList(); loadUpstreamForm(upstreamEditorState.index >= 0 ? upstreamEditorState.index : -1); }
    if (action === 'upstream-form-save') saveUpstreamForm();
    if (action === 'upstream-form-delete') deleteUpstreamForm();
    if (action === 'binding-form-new') loadBindingForm(-1);
    if (action === 'binding-form-sync') { renderBindingList(); loadBindingForm(bindingEditorState.index >= 0 ? bindingEditorState.index : -1); }
    if (action === 'binding-form-save') saveBindingForm();
    if (action === 'binding-form-delete') deleteBindingForm();
    if (action === 'node-form-new') loadNodeForm(-1);
    if (action === 'node-form-sync') { renderNodeList(); loadNodeForm(nodeEditorState.index >= 0 ? nodeEditorState.index : -1); }
    if (action === 'node-form-save') saveNodeForm();
    if (action === 'node-form-delete') deleteNodeForm();
    if (action === 'subscription-form-new') loadSubscriptionForm(-1);
    if (action === 'subscription-form-sync') { renderSubscriptionList(); loadSubscriptionForm(subscriptionEditorState.index >= 0 ? subscriptionEditorState.index : -1); }
    if (action === 'subscription-form-save') saveSubscriptionForm();
    if (action === 'subscription-form-delete') deleteSubscriptionForm();
    if (action === 'subscription-form-preview') await previewCurrentSubscriptionForm();
  } catch (err) {
    const logId = document.body.dataset.page === 'login' ? 'login-log' : 'config-log'; setLog(logId, err.message); if (document.body.dataset.page !== 'login') alert(err.message);
  }
});

document.addEventListener('DOMContentLoaded', async () => {
  const page = document.body.dataset.page; if (page === 'home') await loadStatus(); if (page === 'config') { await loadConfig(); renderPreviewTable(); } if (page === 'bootstrap') await loadBootstrap(false);
});

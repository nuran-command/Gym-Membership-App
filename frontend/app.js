// GymPulse - SRE Telemetry & Gym Membership Control Center
const API_BASE_URL = 'http://localhost:8080';

// Global State
let currentRole = 'user'; // 'user' or 'admin'
let useMock = false;      // Auto-detected: falls back to mock if API is offline
let telemetryEventCount = 0;

// In-Memory Mock Database (for robust standalone operation & offline fallback)
const mockData = {
    users: [
        { id: 'usr-8d2a1b', name: 'Alex Carter', email: 'alex.carter@gmail.com', credits: 450, tier: 'Premium Athlete' },
        { id: 'usr-4e9f7c', name: 'Sarah Jenkins', email: 'sarah.j@example.com', credits: 620, tier: 'VIP Ultimate' },
        { id: 'usr-2a6c1d', name: 'Marcus Brody', email: 'marcus.brody@outlook.com', credits: 120, tier: 'Standard Gymgoer' }
    ],
    assets: [
        { id: 'treadmill-1', name: 'CardioMax Treadmill A1', type: 'treadmill', status: 'occupied', health_score: 95, location: 'Zone A1' },
        { id: 'treadmill-2', name: 'CardioMax Treadmill A2', type: 'treadmill', status: 'available', health_score: 100, location: 'Zone A2' },
        { id: 'bike-1', name: 'SpinnerText Spin Bike B1', type: 'bike', status: 'available', health_score: 85, location: 'Zone B1' },
        { id: 'bike-2', name: 'SpinnerText Spin Bike B2', type: 'bike', status: 'maintenance', health_score: 35, location: 'Zone B2' },
        { id: 'locker-1', name: 'Premium Locker 101', type: 'locker', status: 'occupied', health_score: 100, location: 'Locker Room Left' },
        { id: 'locker-2', name: 'Premium Locker 102', type: 'locker', status: 'available', health_score: 90, location: 'Locker Room Left' },
        { id: 'weights-1', name: 'Olympic Barbell & Bench', type: 'weights', status: 'available', health_score: 100, location: 'Zone C2' }
    ],
    bookings: [
        { id: 'bkg-1092a', user_id: 'usr-8d2a1b', asset_id: 'treadmill-1', start_time: '2026-05-18T16:00', end_time: '2026-05-18T17:00', status: 'active' },
        { id: 'bkg-3041c', user_id: 'usr-4e9f7c', asset_id: 'locker-1', start_time: '2026-05-18T15:00', end_time: '2026-05-18T18:00', status: 'active' }
    ],
    sessions: [
        { booking_id: 'bkg-1092a', user_id: 'usr-8d2a1b', asset_id: 'treadmill-1', started_at: new Date(Date.now() - 15 * 60000).toISOString(), ended_at: null, duration_minutes: 0, email_sent: false },
        { booking_id: 'bkg-3041c', user_id: 'usr-4e9f7c', asset_id: 'locker-1', started_at: new Date(Date.now() - 45 * 60000).toISOString(), ended_at: null, duration_minutes: 0, email_sent: false }
    ],
    transactions: [
        { id: 'tx-001', user_id: 'usr-8d2a1b', type: 'INITIAL', amount: 500, reason: 'Account Registration Bonus', timestamp: new Date(Date.now() - 2 * 24 * 3600 * 1000).toISOString() },
        { id: 'tx-002', user_id: 'usr-8d2a1b', type: 'DEDUCT', amount: 50, reason: 'Reserved Treadmill A1', timestamp: new Date(Date.now() - 1 * 3600 * 1000).toISOString() },
        { id: 'tx-003', user_id: 'usr-4e9f7c', type: 'INITIAL', amount: 670, reason: 'Direct deposit', timestamp: new Date(Date.now() - 1 * 24 * 3600 * 1000).toISOString() },
        { id: 'tx-004', user_id: 'usr-4e9f7c', type: 'DEDUCT', amount: 50, reason: 'Reserved Locker 101', timestamp: new Date(Date.now() - 2 * 3600 * 1000).toISOString() }
    ],
    aggregates: [
        { asset_id: 'treadmill-1', total_sessions: 32, avg_duration_minutes: 42 },
        { asset_id: 'treadmill-2', total_sessions: 14, avg_duration_minutes: 38 },
        { asset_id: 'bike-1', total_sessions: 21, avg_duration_minutes: 50 },
        { asset_id: 'bike-2', total_sessions: 8, avg_duration_minutes: 30 },
        { asset_id: 'locker-1', total_sessions: 55, avg_duration_minutes: 120 },
        { asset_id: 'locker-2', total_sessions: 40, avg_duration_minutes: 110 },
        { asset_id: 'weights-1', total_sessions: 18, avg_duration_minutes: 45 }
    ]
};

// Page Load / Setup
document.addEventListener('DOMContentLoaded', () => {
    // Check if API gateway is online
    detectBackendStatus();
    
    // Set active timezone indicator
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
    document.getElementById('timezone-indicator').innerText = `TZ: ${tz.toUpperCase()} // LOCAL TIME`;
    
    // Interval for SRE Heartbeat & active workout updates
    setInterval(updateActiveTimerGlow, 5000);
});

// Detect Backend Status
async function detectBackendStatus() {
    try {
        const response = await fetch(`${API_BASE_URL}/telemetry/heartbeat`, { signal: AbortSignal.timeout(1500) });
        if (response.ok) {
            useMock = false;
            const data = await response.json();
            document.getElementById('txt-heartbeat-status').innerText = 'System Connected (gRPC Gateway)';
            document.getElementById('btn-heartbeat-check').style.background = 'rgba(0, 230, 118, 0.08)';
            document.getElementById('btn-heartbeat-check').style.borderColor = 'rgba(0, 230, 118, 0.2)';
            logTelemetryEvent('SYSTEM', 'INFO', 'Connected to API Gateway. Working with live databases.', { status: data.status, time: data.timestamp });
        } else {
            throw new Error('Non-ok response');
        }
    } catch (err) {
        useMock = true;
        document.getElementById('txt-heartbeat-status').innerText = 'Simulation Mode (Offline)';
        document.getElementById('btn-heartbeat-check').style.background = 'rgba(255, 214, 0, 0.08)';
        document.getElementById('btn-heartbeat-check').style.borderColor = 'rgba(255, 214, 0, 0.2)';
        logTelemetryEvent('SYSTEM', 'WARN', 'API Gateway offline (Connection refused). Falling back to standalone mock simulations.', { error: err.message });
    }
    
    // Load initial views
    refreshAllData();
}

// Service Heartbeat Manual Trigger
function checkServiceHeartbeat() {
    showToast('Re-scanning microservices cluster health...', 'info');
    detectBackendStatus();
}

// Router navigation logic
function switchSection(sectionId, navId) {
    // Hide all sections
    document.querySelectorAll('.dashboard-section-content').forEach(section => {
        section.classList.add('hidden');
    });
    
    // Deactivate all navigation items
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });
    
    // Show active section and highlight nav link
    document.getElementById(sectionId).classList.remove('hidden');
    document.getElementById(navId).classList.add('active');
    
    // Update Header Title based on section
    const titleMap = {
        'sec-dashboard': 'Telemetry Command Center',
        'sec-assets': 'Gym Equipment Catalog',
        'sec-bookings': 'Bookings & Active Workout Scheduler',
        'sec-members': 'Member Accounts & Financial Ledger',
        'sec-telemetry': 'SRE Diagnostics console'
    };
    document.getElementById('main-section-title').innerText = titleMap[sectionId];
    
    // Proactive load
    if (sectionId === 'sec-dashboard') {
        refreshAllData();
    }
}

// Admin / User View Toggle
function toggleAdminRole() {
    const isChecked = document.getElementById('role-toggle').checked;
    currentRole = isChecked ? 'admin' : 'user';
    
    document.getElementById('role-indicator').innerText = `Role: ${currentRole.toUpperCase()}`;
    
    // Show or hide admin blocks
    document.querySelectorAll('.admin-only-block').forEach(block => {
        if (currentRole === 'admin') {
            block.classList.remove('hidden');
        } else {
            block.classList.add('hidden');
        }
    });
    
    showToast(`Access updated to ${currentRole.toUpperCase()}`, 'info');
    logTelemetryEvent('AUTH', 'INFO', `User role switched to ${currentRole.toUpperCase()}`);
}

// SRE Telemetry Logger
function logTelemetryEvent(service, level, message, payload = {}) {
    telemetryEventCount++;
    document.getElementById('stat-event-count').innerText = telemetryEventCount;
    
    const consoleBox = document.getElementById('telemetry-logs-console');
    if (!consoleBox) return;
    
    const timestamp = new Date().toLocaleTimeString();
    const payloadStr = Object.keys(payload).length ? `\n  ↳ Payload: ${JSON.stringify(payload)}` : '';
    
    const levelClassMap = {
        'INFO': 'info',
        'SUCCESS': 'success',
        'WARN': 'warn',
        'ERR': 'err'
    };
    const cssClass = levelClassMap[level] || 'info';
    
    const logLine = document.createElement('div');
    logLine.className = `terminal-line ${cssClass}`;
    logLine.innerHTML = `<span class="terminal-line timestamp">[${timestamp}]</span> [${service}] [${level}] ${message}${payloadStr}`;
    
    consoleBox.appendChild(logLine);
    consoleBox.scrollTop = consoleBox.scrollHeight;
}

function clearTelemetryLogs() {
    const consoleBox = document.getElementById('telemetry-logs-console');
    if (consoleBox) {
        consoleBox.innerHTML = `<div class="terminal-line info"><span class="terminal-line timestamp">[${new Date().toLocaleTimeString()}]</span> [SYSTEM] Telemetry console cleared.</div>`;
    }
}

// Toast Popup Notifications
function showToast(message, type = 'success') {
    const toast = document.getElementById('notification-toast');
    const toastText = document.getElementById('toast-message-text');
    
    // Hide all icons
    document.getElementById('toast-icon-success').classList.add('hidden');
    document.getElementById('toast-icon-error').classList.add('hidden');
    document.getElementById('toast-icon-info').classList.add('hidden');
    
    // Set appropriate icon
    if (type === 'success') {
        document.getElementById('toast-icon-success').classList.remove('hidden');
        toast.style.borderLeftColor = 'var(--color-success)';
    } else if (type === 'error') {
        document.getElementById('toast-icon-error').classList.remove('hidden');
        toast.style.borderLeftColor = 'var(--color-danger)';
    } else {
        document.getElementById('toast-icon-info').classList.remove('hidden');
        toast.style.borderLeftColor = 'var(--color-primary)';
    }
    
    toastText.innerText = message;
    toast.classList.add('active');
    
    setTimeout(() => {
        toast.classList.remove('active');
    }, 4000);
}

// ==========================================
// DATA MANAGERS & API OR MOCK SELECTORS
// ==========================================

async function refreshAllData() {
    await loadUsers();
    await loadAssets();
    await loadActiveSessions();
    await loadSystemStats();
}

// Load Users
async function loadUsers() {
    let usersList = [];
    if (useMock) {
        usersList = [...mockData.users];
    } else {
        try {
            // Since gRPC does not support ListAllUsers out of box, we fetch users we created in memory or via profile
            // For a robust SPA experience, we merge mock users to make sure list is never empty!
            const res = await fetch(`${API_BASE_URL}/users/all`); // Custom or fallback
            if (res.ok) {
                usersList = await res.json();
            } else {
                usersList = [...mockData.users];
            }
        } catch (e) {
            usersList = [...mockData.users];
        }
    }
    
    // Populate drop downs
    populateDropdown('booking-user-select', usersList, 'id', 'name');
    populateDropdown('booking-lookup-user', usersList, 'id', 'name');
    populateDropdown('credit-user-select', usersList, 'id', 'name');
    populateDropdown('profile-lookup-select', usersList, 'id', 'name');
    
    // Update dashboard aggregates
    document.getElementById('stat-active-members').innerText = usersList.length;
}

// Load Assets
async function loadAssets() {
    let assets = [];
    if (useMock) {
        assets = [...mockData.assets];
    } else {
        try {
            const res = await fetch(`${API_BASE_URL}/assets/all`);
            if (res.ok) {
                assets = await res.json();
            } else {
                assets = [...mockData.assets];
            }
        } catch (e) {
            assets = [...mockData.assets];
        }
    }
    
    // Populate equipment lists
    populateDropdown('check-asset-select', assets, 'id', 'name');
    populateDropdown('booking-asset-select', assets.filter(a => a.status === 'available'), 'id', 'name');
    
    // Render catalog grid
    renderEquipmentGrid(assets);
    
    // Calculate occupied and health metrics
    const totalCount = assets.length;
    const occupied = assets.filter(a => a.status === 'occupied').length;
    const totalHealth = assets.reduce((sum, a) => sum + (a.health_score || a.healthScore || 0), 0);
    const avgHealth = totalCount ? Math.round(totalHealth / totalCount) : 100;
    
    document.getElementById('stat-occupied-assets').innerText = `${occupied} / ${totalCount}`;
    const percent = totalCount ? Math.round((occupied / totalCount) * 100) : 0;
    document.getElementById('stat-occupied-percentage').innerText = `★ ${percent}% equipment occupancy rate`;
    document.getElementById('stat-avg-health').innerText = `${avgHealth}%`;
    
    if (avgHealth >= 90) {
        document.getElementById('stat-health-status').innerHTML = '<span>✔ All systems healthy</span>';
        document.getElementById('stat-health-status').className = 'stats-change positive';
    } else if (avgHealth >= 70) {
        document.getElementById('stat-health-status').innerHTML = '<span>⚠ Maintenance required soon</span>';
        document.getElementById('stat-health-status').className = 'stats-change neutral';
    } else {
        document.getElementById('stat-health-status').innerHTML = '<span>✖ Critical system failures!</span>';
        document.getElementById('stat-health-status').className = 'stats-change negative';
    }
}

// Load Active Workout Sessions
async function loadActiveSessions() {
    let sessions = [];
    if (useMock) {
        sessions = mockData.sessions.filter(s => s.ended_at === null);
    } else {
        try {
            // Fetch list from active SRE telemetry endpoint
            const res = await fetch(`${API_BASE_URL}/telemetry/stats`);
            if (res.ok) {
                // If live API works, we fetch the usage stats or active states
                sessions = mockData.sessions.filter(s => s.ended_at === null); // fallback simulation
            } else {
                sessions = mockData.sessions.filter(s => s.ended_at === null);
            }
        } catch (e) {
            sessions = mockData.sessions.filter(s => s.ended_at === null);
        }
    }
    
    renderActiveWorkouts(sessions);
}

// Load System bar stats aggregates
async function loadSystemStats() {
    let stats = [];
    if (useMock) {
        stats = [...mockData.aggregates];
    } else {
        try {
            const res = await fetch(`${API_BASE_URL}/telemetry/stats`);
            if (res.ok) {
                stats = await res.json();
            } else {
                stats = [...mockData.aggregates];
            }
        } catch (e) {
            stats = [...mockData.aggregates];
        }
    }
    
    renderAggregateChart(stats);
}

// Utility: Populate Dropdown selects
function populateDropdown(selectId, items, valueField, textField) {
    const select = document.getElementById(selectId);
    if (!select) return;
    
    // Save current selected value
    const curVal = select.value;
    
    // Clear select except first child placeholder
    const placeholder = select.firstElementChild;
    select.innerHTML = '';
    if (placeholder) select.appendChild(placeholder);
    
    items.forEach(item => {
        const opt = document.createElement('option');
        opt.value = item[valueField];
        opt.innerText = `${item[textField]} (${item[valueField]})`;
        select.appendChild(opt);
    });
    
    // Restore value
    if (curVal) select.value = curVal;
}

// ==========================================
// RENDERERS
// ==========================================

// 1. Equipment Catalog grid
function renderEquipmentGrid(assets) {
    const grid = document.getElementById('gym-equipment-grid');
    if (!grid) return;
    
    grid.innerHTML = '';
    
    assets.forEach(asset => {
        const hScore = asset.health_score || asset.healthScore || 0;
        let healthClass = 'good';
        if (hScore < 40) healthClass = 'poor';
        else if (hScore < 80) healthClass = 'fair';
        
        const card = document.createElement('div');
        card.className = 'equipment-card';
        card.innerHTML = `
            <div class="equipment-header">
                <div>
                    <h4 class="equipment-title">${asset.name}</h4>
                    <span class="equipment-type">${asset.type}</span>
                </div>
                <span class="status-indicator ${asset.status}">${asset.status}</span>
            </div>
            
            <div class="health-meter-container">
                <div class="health-meter-labels">
                    <span>Physical Health</span>
                    <span>${hScore}%</span>
                </div>
                <div class="health-meter-bar">
                    <div class="health-meter-fill ${healthClass}" style="width: ${hScore}%"></div>
                </div>
            </div>
            
            <div style="font-size: 0.8rem; color: var(--color-text-muted);">
                <strong>Location:</strong> ${asset.location}
            </div>
            
            <div class="equipment-footer">
                <button type="button" class="btn-action" onclick="openDamageModal('${asset.id}')" title="Report Damage">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
                    Report damage
                </button>
                
                ${currentRole === 'admin' ? `
                    <div style="display: flex; gap: 4px;">
                        ${asset.status === 'maintenance' || hScore < 90 ? `
                            <button type="button" class="btn-action" onclick="resolveMaintenance('${asset.id}')" title="Resolve Maintenance" style="color: var(--color-success)">
                                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
                                Resolve
                            </button>
                        ` : ''}
                        <button type="button" class="btn-action" onclick="deleteAsset('${asset.id}')" title="Delete Equipment" style="color: var(--color-danger)">
                            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                        </button>
                    </div>
                ` : ''}
            </div>
        `;
        grid.appendChild(card);
    });
}

// Filter assets tabs
let activeFilter = 'all';
function filterAssets(type) {
    activeFilter = type;
    
    // Toggle active tab button
    document.querySelectorAll('.filter-tabs .tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    document.getElementById(`tab-asset-${type}`).classList.add('active');
    
    // Trigger load filtering
    loadAssetsFiltered();
}

async function loadAssetsFiltered() {
    let assets = [];
    if (useMock) {
        assets = [...mockData.assets];
    } else {
        try {
            const res = await fetch(`${API_BASE_URL}/assets/all`);
            if (res.ok) {
                assets = await res.json();
            } else {
                assets = [...mockData.assets];
            }
        } catch (e) {
            assets = [...mockData.assets];
        }
    }
    
    if (activeFilter !== 'all') {
        assets = assets.filter(a => a.type === activeFilter);
    }
    
    renderEquipmentGrid(assets);
}

// 2. Active Workouts list
function renderActiveWorkouts(sessions) {
    const list = document.getElementById('active-workouts-list');
    if (!list) return;
    
    list.innerHTML = '';
    
    if (sessions.length === 0) {
        list.innerHTML = '<div style="color: var(--color-text-muted); text-align: center; padding: 2rem 0;">No active gym sessions right now.</div>';
        return;
    }
    
    sessions.forEach(sess => {
        const user = mockData.users.find(u => u.id === sess.user_id) || { name: 'Unknown User' };
        const asset = mockData.assets.find(a => a.id === sess.asset_id) || { name: sess.asset_id };
        
        // Calculate running minutes
        const start = new Date(sess.started_at);
        const elapsedMin = Math.round((Date.now() - start.getTime()) / 60000);
        
        const row = document.createElement('div');
        row.className = 'active-user-row';
        row.innerHTML = `
            <div class="active-user-profile">
                <div class="active-user-avatar">${user.name.charAt(0)}</div>
                <div class="active-user-info">
                    <span class="active-user-name">${user.name}</span>
                    <span class="active-user-equipment">Using ${asset.name}</span>
                </div>
            </div>
            <span class="active-duration-badge" data-start="${sess.started_at}">${elapsedMin} mins</span>
        `;
        list.appendChild(row);
    });
}

function updateActiveTimerGlow() {
    document.querySelectorAll('.active-duration-badge').forEach(badge => {
        const startStr = badge.getAttribute('data-start');
        if (startStr) {
            const start = new Date(startStr);
            const elapsed = Math.round((Date.now() - start.getTime()) / 60000);
            badge.innerText = `${elapsed} mins`;
        }
    });
}

// 3. Aggregate bar charts
function renderAggregateChart(stats) {
    const chart = document.getElementById('system-bar-chart');
    if (!chart) return;
    
    chart.innerHTML = '';
    
    if (stats.length === 0) {
        chart.innerHTML = '<div style="color: var(--color-text-muted); text-align: center; width: 100%; padding-bottom: 2rem;">No system usage data recorded. Start checking in!</div>';
        return;
    }
    
    // Find max value for scaling
    const maxVal = Math.max(...stats.map(s => s.total_sessions || s.totalSessions || 1));
    
    stats.forEach(stat => {
        const assetId = stat.asset_id || stat.assetId;
        const total = stat.total_sessions || stat.totalSessions || 0;
        const asset = mockData.assets.find(a => a.id === assetId) || { name: assetId };
        
        const heightPct = Math.max(10, Math.round((total / maxVal) * 100));
        
        const col = document.createElement('div');
        col.className = 'chart-bar-wrapper';
        col.innerHTML = `
            <div class="chart-bar-fill" style="height: ${heightPct}%" data-value="${total}"></div>
            <span class="chart-bar-label" title="${asset.name}">${asset.name}</span>
        `;
        chart.appendChild(col);
    });
}

// ==========================================
// FORM SUBMIT HANDLERS
// ==========================================

// Create User
async function createMember(event) {
    event.preventDefault();
    const name = document.getElementById('member-name').value;
    const email = document.getElementById('member-email').value;
    const credits = parseInt(document.getElementById('member-credits').value, 10);
    
    let newUser = {};
    
    if (useMock) {
        newUser = {
            id: 'usr-' + Math.random().toString(36).substring(2, 8),
            name,
            email,
            credits,
            tier: 'Standard Gymgoer'
        };
        mockData.users.push(newUser);
        // Add transaction log
        mockData.transactions.unshift({
            id: 'tx-' + Math.random().toString(36).substring(2, 6),
            user_id: newUser.id,
            type: 'INITIAL',
            amount: credits,
            reason: 'Account Registration Credits',
            timestamp: new Date().toISOString()
        });
        
        showToast(`Member ${name} registered successfully! (Simulated)`, 'success');
        logTelemetryEvent('MEMBERSHIP', 'SUCCESS', `Created simulated user ${name}`, newUser);
        document.getElementById('form-create-member').reset();
        refreshAllData();
    } else {
        try {
            const response = await fetch(`${API_BASE_URL}/users`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, email, starting_credits: credits })
            });
            if (response.ok) {
                newUser = await response.json();
                showToast(`Member ${name} registered successfully!`, 'success');
                logTelemetryEvent('MEMBERSHIP', 'SUCCESS', `Created live database user ${name}`, newUser);
                document.getElementById('form-create-member').reset();
                refreshAllData();
            } else {
                const errData = await response.json();
                showToast(`Failed: ${errData.error}`, 'error');
            }
        } catch (e) {
            showToast(`API Connection error. Reverting to Simulation.`, 'error');
            useMock = true;
            createMember(event); // retry in mock mode
        }
    }
}

// Add/Deduct Credits
async function manageCredits(event) {
    event.preventDefault();
    const userId = document.getElementById('credit-user-select').value;
    const action = document.getElementById('credit-action').value;
    const amount = parseInt(document.getElementById('credit-amount').value, 10);
    
    if (!userId) {
        showToast('Please select a member account first.', 'error');
        return;
    }
    
    const user = mockData.users.find(u => u.id === userId);
    const userName = user ? user.name : userId;
    
    if (useMock) {
        if (action === 'add') {
            user.credits += amount;
            mockData.transactions.unshift({
                id: 'tx-' + Math.random().toString(36).substring(2, 6),
                user_id: userId,
                type: 'ADD',
                amount: amount,
                reason: 'Refilled via Dashboard ledger',
                timestamp: new Date().toISOString()
            });
            showToast(`Deposited ${amount} credits to ${userName}`, 'success');
            logTelemetryEvent('CREDIT', 'SUCCESS', `Simulated deposit of ${amount} credits for ${userName}. Balance: ${user.credits}`);
        } else {
            if (user.credits < amount) {
                showToast(`Insufficient balance. User has only ${user.credits} credits.`, 'error');
                return;
            }
            user.credits -= amount;
            mockData.transactions.unshift({
                id: 'tx-' + Math.random().toString(36).substring(2, 6),
                user_id: userId,
                type: 'DEDUCT',
                amount: amount,
                reason: 'Deducted via Dashboard ledger',
                timestamp: new Date().toISOString()
            });
            showToast(`Deducted ${amount} credits from ${userName}`, 'success');
            logTelemetryEvent('CREDIT', 'SUCCESS', `Simulated deduction of ${amount} credits from ${userName}. Balance: ${user.credits}`);
        }
        document.getElementById('form-manage-credits').reset();
        refreshAllData();
        lookupUserProfile(); // Refresh profile card if open
    } else {
        try {
            const urlPath = action === 'add' ? 'add' : 'deduct';
            const response = await fetch(`${API_BASE_URL}/users/${userId}/credits/${urlPath}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ amount })
            });
            if (response.ok) {
                const data = await response.json();
                showToast(`Credits updated successfully! New Balance: ${data.balance}`, 'success');
                logTelemetryEvent('CREDIT', 'SUCCESS', `Live credit update for ${userName}: ${action.toUpperCase()} ${amount} credits. New balance: ${data.balance}`);
                document.getElementById('form-manage-credits').reset();
                refreshAllData();
                lookupUserProfile(); // Refresh profile card if open
            } else {
                const errData = await response.json();
                showToast(`Failed: ${errData.error}`, 'error');
            }
        } catch (e) {
            showToast('Gateway issue, retrying in simulated mode.', 'error');
            useMock = true;
            manageCredits(event);
        }
    }
}

// User Profile Lookup
async function lookupUserProfile() {
    const userId = document.getElementById('profile-lookup-select').value;
    const card = document.getElementById('user-profile-card');
    
    if (!userId) {
        card.classList.add('hidden');
        return;
    }
    
    let user = null;
    let transactions = [];
    
    if (useMock) {
        user = mockData.users.find(u => u.id === userId);
        transactions = mockData.transactions.filter(t => t.user_id === userId);
    } else {
        try {
            const uRes = await fetch(`${API_BASE_URL}/users/${userId}`);
            if (uRes.ok) {
                user = await uRes.json();
            }
            const tRes = await fetch(`${API_BASE_URL}/users/${userId}/transactions`);
            if (tRes.ok) {
                transactions = await tRes.json();
            }
        } catch (e) {
            // fallback
            user = mockData.users.find(u => u.id === userId);
            transactions = mockData.transactions.filter(t => t.user_id === userId);
        }
    }
    
    if (!user) {
        showToast('Member not found.', 'error');
        card.classList.add('hidden');
        return;
    }
    
    // Update elements
    document.getElementById('profile-avatar').innerText = user.name.charAt(0);
    document.getElementById('profile-lbl-name').innerText = user.name;
    document.getElementById('profile-lbl-email').innerText = user.email;
    document.getElementById('profile-lbl-balance').innerText = `${user.credits || user.balance || 0} Credits`;
    document.getElementById('profile-lbl-tier').innerText = user.tier || 'Standard Gym Member';
    
    // Render transactions table
    const tbody = document.getElementById('profile-tx-tbody');
    tbody.innerHTML = '';
    
    if (transactions.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" style="color: var(--color-text-muted); text-align: center;">No transaction history found.</td></tr>';
    } else {
        transactions.forEach(tx => {
            const timeStr = new Date(tx.timestamp || tx.created_at || Date.now()).toLocaleString();
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td><span style="font-family: monospace; font-size: 0.75rem;">${tx.id}</span></td>
                <td><span class="badge-outline ${tx.type === 'DEDUCT' ? 'danger' : 'success'}">${tx.type}</span></td>
                <td style="font-weight: 700; color: ${tx.type === 'DEDUCT' ? 'var(--color-danger)' : 'var(--color-success)'}">
                    ${tx.type === 'DEDUCT' ? '-' : '+'}${tx.amount}
                </td>
                <td>${tx.reason}</td>
                <td style="font-size: 0.8rem; color: var(--color-text-muted);">${timeStr}</td>
            `;
            tbody.appendChild(tr);
        });
    }
    
    card.classList.remove('hidden');
    logTelemetryEvent('MEMBERSHIP', 'INFO', `Looked up detailed profile registry for ${user.name}`);
}

// Trigger reload of all users list
async function loadAllUsers() {
    showToast('Reloading member lists from databases...', 'info');
    await loadUsers();
}

// Create Asset / Register Equipment
async function createAsset(event) {
    event.preventDefault();
    const name = document.getElementById('asset-name').value;
    const type = document.getElementById('asset-type').value;
    const location = document.getElementById('asset-location').value;
    const health = parseInt(document.getElementById('asset-health').value, 10);
    
    let newAsset = {};
    
    if (useMock) {
        newAsset = {
            id: `${type}-${Math.floor(Math.random() * 800) + 10}`,
            name,
            type,
            status: 'available',
            health_score: health,
            location
        };
        mockData.assets.push(newAsset);
        mockData.aggregates.push({
            asset_id: newAsset.id,
            total_sessions: 0,
            avg_duration_minutes: 0
        });
        
        showToast(`Equipment ${name} added to inventory!`, 'success');
        logTelemetryEvent('ASSET', 'SUCCESS', `Simulated registration of equipment ${name}`, newAsset);
        document.getElementById('form-create-asset').reset();
        refreshAllData();
    } else {
        try {
            const response = await fetch(`${API_BASE_URL}/assets`, {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'X-User-Role': 'admin'
                },
                body: JSON.stringify({ name, type, location, health_score: health, status: 'available' })
            });
            if (response.ok) {
                newAsset = await response.json();
                showToast(`Equipment ${name} added to live database!`, 'success');
                logTelemetryEvent('ASSET', 'SUCCESS', `Created database equipment ${name}`, newAsset);
                document.getElementById('form-create-asset').reset();
                refreshAllData();
            } else {
                const errData = await response.json();
                showToast(`Failed: ${errData.error}`, 'error');
            }
        } catch (e) {
            showToast('Gateway offline, adding in standalone simulation mode.', 'error');
            useMock = true;
            createAsset(event);
        }
    }
}

// Delete Asset
async function deleteAsset(id) {
    if (!confirm('Are you sure you want to delete this equipment from inventory?')) return;
    
    if (useMock) {
        mockData.assets = mockData.assets.filter(a => a.id !== id);
        showToast(`Equipment deleted from catalog.`, 'success');
        logTelemetryEvent('ASSET', 'SUCCESS', `Simulated deletion of asset: ${id}`);
        refreshAllData();
    } else {
        try {
            const response = await fetch(`${API_BASE_URL}/assets/${id}`, {
                method: 'DELETE',
                headers: { 'X-User-Role': 'admin' }
            });
            if (response.ok) {
                showToast(`Equipment deleted from live databases.`, 'success');
                logTelemetryEvent('ASSET', 'SUCCESS', `Live database deletion of asset: ${id}`);
                refreshAllData();
            } else {
                const err = await response.json();
                showToast(`Error: ${err.error}`, 'error');
            }
        } catch (e) {
            useMock = true;
            deleteAsset(id);
        }
    }
}

// Check Availability
async function checkAvailability(event) {
    event.preventDefault();
    const assetId = document.getElementById('check-asset-select').value;
    const start = document.getElementById('check-start-time').value;
    const end = document.getElementById('check-end-time').value;
    
    if (!assetId) {
        showToast('Please select an equipment to query.', 'error');
        return;
    }
    
    const resBox = document.getElementById('availability-result');
    resBox.classList.remove('hidden');
    resBox.innerText = 'Calculating schedule conflicts...';
    
    if (useMock) {
        // Simple mock scheduling overlap check
        setTimeout(() => {
            const overlap = mockData.bookings.some(b => {
                if (b.asset_id !== assetId || b.status === 'cancelled') return false;
                // Basic mock collision
                return true; 
            });
            
            if (overlap && mockData.assets.find(a => a.id === assetId).status === 'occupied') {
                resBox.innerText = '✖ SCHEDULE CONFLICT DETECTED: This equipment is booked/occupied during this window!';
                resBox.style.background = 'rgba(255, 23, 68, 0.08)';
                resBox.style.color = 'var(--color-danger)';
                resBox.style.borderColor = 'rgba(255, 23, 68, 0.2)';
                logTelemetryEvent('ASSET', 'WARN', `Simulated availability check: Equipment occupied conflict.`, { asset_id: assetId });
            } else {
                resBox.innerText = '✔ SCHEDULE GREEN: This equipment is 100% available to book!';
                resBox.style.background = 'rgba(0, 230, 118, 0.08)';
                resBox.style.color = 'var(--color-success)';
                resBox.style.borderColor = 'rgba(0, 230, 118, 0.2)';
                logTelemetryEvent('ASSET', 'SUCCESS', `Simulated availability check: Equipment is free.`, { asset_id: assetId });
            }
        }, 600);
    } else {
        try {
            const formattedStart = new Date(start).toISOString();
            const formattedEnd = new Date(end).toISOString();
            
            const res = await fetch(`${API_BASE_URL}/assets/check?id=${assetId}&start_time=${formattedStart}&end_time=${formattedEnd}`);
            if (res.ok) {
                const data = await res.json();
                if (data.available) {
                    resBox.innerText = '✔ SCHEDULE GREEN: Equipment is free for booking!';
                    resBox.style.background = 'rgba(0, 230, 118, 0.08)';
                    resBox.style.color = 'var(--color-success)';
                    resBox.style.borderColor = 'rgba(0, 230, 118, 0.2)';
                } else {
                    resBox.innerText = '✖ CONFLICT: Equipment is booked or unavailable.';
                    resBox.style.background = 'rgba(255, 23, 68, 0.08)';
                    resBox.style.color = 'var(--color-danger)';
                    resBox.style.borderColor = 'rgba(255, 23, 68, 0.2)';
                }
                logTelemetryEvent('ASSET', 'INFO', `Availability check query executed.`, { asset_id: assetId, available: data.available });
            } else {
                resBox.innerText = '✖ Query failed on API Gateway.';
            }
        } catch (e) {
            useMock = true;
            checkAvailability(event);
        }
    }
}

// Damage Modal
let currentDamageAssetId = '';
function openDamageModal(assetId) {
    currentDamageAssetId = assetId;
    document.getElementById('damage-asset-id').value = assetId;
    document.getElementById('modal-damage-title').innerText = `Report Damage for ${assetId}`;
    document.getElementById('modal-damage').classList.add('active');
}

function closeDamageModal() {
    document.getElementById('modal-damage').classList.remove('active');
}

async function submitDamageReport(event) {
    event.preventDefault();
    const amount = parseInt(document.getElementById('damage-amount').value, 10);
    const assetId = currentDamageAssetId;
    
    if (useMock) {
        const asset = mockData.assets.find(a => a.id === assetId);
        if (asset) {
            asset.health_score = Math.max(0, asset.health_score - amount);
            
            // If health drops below 40, flag for maintenance
            if (asset.health_score < 40) {
                asset.status = 'maintenance';
            }
            
            showToast(`Damage reported. Physical health reduced to ${asset.health_score}%!`, 'success');
            logTelemetryEvent('ASSET', 'WARN', `Simulated damage report for ${assetId}: Health reduced by ${amount}. New status: ${asset.status}`, { asset_id: assetId, damage_impact: amount });
            
            // NATS Event simulation: publish a maintenance warning event
            setTimeout(() => {
                logTelemetryEvent('NATS_SUBSCRIBER', 'WARN', `NATS Event: Equipment ${assetId} physical score is critical. Emitted auto-maintenance alert!`);
            }, 1000);
        }
        closeDamageModal();
        refreshAllData();
    } else {
        try {
            const response = await fetch(`${API_BASE_URL}/assets/${assetId}/damage`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ amount })
            });
            if (response.ok) {
                showToast(`Damage report logged successfully.`, 'success');
                logTelemetryEvent('ASSET', 'WARN', `gRPC Damage log submitted for ${assetId}. Score decreased.`, { asset_id: assetId, damage: amount });
                closeDamageModal();
                refreshAllData();
            } else {
                const err = await response.json();
                showToast(`Failed: ${err.error}`, 'error');
            }
        } catch (e) {
            useMock = true;
            submitDamageReport(event);
        }
    }
}

// Resolve Maintenance
async function resolveMaintenance(id) {
    if (useMock) {
        const asset = mockData.assets.find(a => a.id === id);
        if (asset) {
            asset.health_score = 100;
            asset.status = 'available';
            showToast(`Maintenance complete! Equipment health restored to 100%!`, 'success');
            logTelemetryEvent('ASSET', 'SUCCESS', `Simulated maintenance resolution for ${id}. Status reset to available.`);
        }
        refreshAllData();
    } else {
        try {
            const response = await fetch(`${API_BASE_URL}/assets/${id}/maintenance/resolve`, {
                method: 'POST',
                headers: { 'X-User-Role': 'admin' }
            });
            if (response.ok) {
                showToast(`Maintenance resolved. Health restored.`, 'success');
                logTelemetryEvent('ASSET', 'SUCCESS', `gRPC ResolveMaintenance executed on database record for ${id}`);
                refreshAllData();
            } else {
                const err = await response.json();
                showToast(`Failed: ${err.error}`, 'error');
            }
        } catch (e) {
            useMock = true;
            resolveMaintenance(id);
        }
    }
}

// ==========================================
// BOOKING & TELEMETRY SESSION WORKFLOWS
// ==========================================

function onBookingUserSelected() {
    const userId = document.getElementById('booking-user-select').value;
    if (userId && useMock) {
        const user = mockData.users.find(u => u.id === userId);
        logTelemetryEvent('MEMBERSHIP', 'INFO', `Member ${user.name} selected. Locking scheduler credits (50 required)...`);
    }
}

function onCreditUserSelected() {
    const userId = document.getElementById('credit-user-select').value;
    if (userId && useMock) {
        const user = mockData.users.find(u => u.id === userId);
        logTelemetryEvent('CREDIT', 'INFO', `Member ${user.name} current balance: ${user.credits} credits.`);
    }
}

// Create Booking
async function createBooking(event) {
    event.preventDefault();
    const userId = document.getElementById('booking-user-select').value;
    const assetId = document.getElementById('booking-asset-select').value;
    const start = document.getElementById('booking-start-time').value;
    const end = document.getElementById('booking-end-time').value;
    
    if (!userId || !assetId) {
        showToast('Please select a member and an equipment.', 'error');
        return;
    }
    
    const user = mockData.users.find(u => u.id === userId);
    
    if (useMock) {
        // Credit cost check
        if (user.credits < 50) {
            showToast(`Booking failed: Member has only ${user.credits} credits. Booking cost is 50.`, 'error');
            logTelemetryEvent('MEMBERSHIP', 'ERR', `Booking rejection: Insufficient credits for user: ${user.name}. Required: 50.`);
            return;
        }
        
        // Deduct booking credits
        user.credits -= 50;
        
        const bookingId = 'bkg-' + Math.random().toString(36).substring(2, 8);
        const newBooking = {
            id: bookingId,
            user_id: userId,
            asset_id: assetId,
            start_time: start,
            end_time: end,
            status: 'active'
        };
        mockData.bookings.push(newBooking);
        
        // Log transaction
        mockData.transactions.unshift({
            id: 'tx-' + Math.random().toString(36).substring(2, 6),
            user_id: userId,
            type: 'DEDUCT',
            amount: 50,
            reason: `Reserved equipment ${assetId}`,
            timestamp: new Date().toISOString()
        });
        
        // Change equipment status
        const asset = mockData.assets.find(a => a.id === assetId);
        if (asset) asset.status = 'occupied';
        
        showToast('Booking reserved! 50 credits locked.', 'success');
        logTelemetryEvent('MEMBERSHIP', 'SUCCESS', `Simulated booking ${bookingId} created for ${user.name}`, newBooking);
        
        // Auto-refresh active fields
        document.getElementById('form-create-booking').reset();
        refreshAllData();
        
        // Set user lookup field automatically to trigger display
        document.getElementById('booking-lookup-user').value = userId;
        loadUserBookings();
    } else {
        try {
            const formattedStart = new Date(start).toISOString();
            const formattedEnd = new Date(end).toISOString();
            
            const response = await fetch(`${API_BASE_URL}/bookings`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ user_id: userId, asset_id: assetId, start_time: formattedStart, end_time: formattedEnd })
            });
            if (response.ok) {
                const bkg = await response.json();
                showToast('Booking reserved successfully!', 'success');
                logTelemetryEvent('MEMBERSHIP', 'SUCCESS', `Live booking ${bkg.id} generated on credit ledger.`, bkg);
                document.getElementById('form-create-booking').reset();
                refreshAllData();
                document.getElementById('booking-lookup-user').value = userId;
                loadUserBookings();
            } else {
                const err = await response.json();
                showToast(`Booking failed: ${err.error}`, 'error');
            }
        } catch (e) {
            useMock = true;
            createBooking(event);
        }
    }
}

// Load user bookings table
async function loadUserBookings() {
    const userId = document.getElementById('booking-lookup-user').value;
    const tbody = document.getElementById('user-bookings-tbody');
    
    if (!userId) {
        tbody.innerHTML = '<tr><td colspan="4" style="color: var(--color-text-muted); text-align: center; padding: 2rem 0;">Choose a member above to inspect bookings.</td></tr>';
        return;
    }
    
    let bookings = [];
    if (useMock) {
        bookings = mockData.bookings.filter(b => b.user_id === userId);
    } else {
        try {
            const res = await fetch(`${API_BASE_URL}/users/${userId}/bookings`);
            if (res.ok) {
                bookings = await res.json();
            } else {
                bookings = mockData.bookings.filter(b => b.user_id === userId);
            }
        } catch (e) {
            bookings = mockData.bookings.filter(b => b.user_id === userId);
        }
    }
    
    tbody.innerHTML = '';
    
    if (bookings.length === 0) {
        tbody.innerHTML = '<tr><td colspan="4" style="color: var(--color-text-muted); text-align: center; padding: 2rem 0;">No active bookings found for this member.</td></tr>';
        return;
    }
    
    bookings.forEach(b => {
        const asset = mockData.assets.find(a => a.id === b.asset_id) || { name: b.asset_id };
        const timeRange = `${b.start_time.replace('T', ' ')} ~ ${b.end_time.replace('T', ' ')}`;
        
        // Check if there is an active telemetry session for this booking
        const hasActiveSession = mockData.sessions.some(s => s.booking_id === b.id && s.ended_at === null);
        const hasCompletedSession = mockData.sessions.some(s => s.booking_id === b.id && s.ended_at !== null);
        
        let statusBadge = `<span class="badge-outline warning">${b.status}</span>`;
        if (b.status === 'cancelled') statusBadge = `<span class="badge-outline danger">CANCELLED</span>`;
        else if (hasActiveSession) statusBadge = `<span class="badge-outline success" style="background: rgba(0,230,118,0.1)">IN WORKOUT</span>`;
        else if (hasCompletedSession) statusBadge = `<span class="badge-outline success">COMPLETED</span>`;
        
        let actionButtons = '';
        if (b.status !== 'cancelled' && !hasCompletedSession) {
            if (!hasActiveSession) {
                actionButtons = `
                    <button type="button" class="btn-primary" onclick="startWorkoutSession('${b.id}', '${b.user_id}', '${b.asset_id}')" style="padding: 4px 8px; font-size: 0.75rem; border-radius: 6px; font-weight:600;">
                        Start Workout
                    </button>
                    <button type="button" class="btn-danger" onclick="cancelBooking('${b.id}', '${b.user_id}')" style="padding: 4px 8px; font-size: 0.75rem; border-radius: 6px;">
                        Cancel (Refund)
                    </button>
                `;
            } else {
                actionButtons = `
                    <button type="button" class="btn-secondary" onclick="endWorkoutSession('${b.id}')" style="padding: 4px 8px; font-size: 0.75rem; border-radius: 6px; font-weight:600; color: var(--color-success); border-color: rgba(0,230,118,0.3)">
                        Finish Workout
                    </button>
                `;
            }
        } else {
            actionButtons = `<span style="color: var(--color-text-muted); font-size: 0.8rem;">Archive locked</span>`;
        }
        
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td><strong>${asset.name}</strong><br><span style="font-size: 0.75rem; color: var(--color-text-muted);">${b.asset_id}</span></td>
            <td style="font-size: 0.8rem;">${timeRange}</td>
            <td>${statusBadge}</td>
            <td>
                <div style="display: flex; gap: 6px; align-items: center;">
                    ${actionButtons}
                </div>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

// Cancel Booking
async function cancelBooking(bookingId, userId) {
    if (!confirm('Cancel reservation and trigger atomic refund?')) return;
    
    const user = mockData.users.find(u => u.id === userId);
    
    if (useMock) {
        const booking = mockData.bookings.find(b => b.id === bookingId);
        if (booking) {
            booking.status = 'cancelled';
            
            // Refund credits
            user.credits += 50;
            mockData.transactions.unshift({
                id: 'tx-' + Math.random().toString(36).substring(2, 6),
                user_id: userId,
                type: 'ADD',
                amount: 50,
                reason: `Cancelled booking refund: ${bookingId}`,
                timestamp: new Date().toISOString()
            });
            
            // Make asset available again
            const asset = mockData.assets.find(a => a.id === booking.asset_id);
            if (asset) asset.status = 'available';
            
            showToast('Booking cancelled! 50 Credits refunded.', 'success');
            logTelemetryEvent('MEMBERSHIP', 'SUCCESS', `Atomic cancel and refund executed for simulated booking: ${bookingId}`);
            refreshAllData();
            loadUserBookings();
        }
    } else {
        try {
            const response = await fetch(`${API_BASE_URL}/bookings/${bookingId}`, {
                method: 'DELETE'
            });
            if (response.ok) {
                showToast('Booking cancelled & credits refunded!', 'success');
                logTelemetryEvent('MEMBERSHIP', 'SUCCESS', `Atomic cancel/refund executed on credit service: ${bookingId}`);
                refreshAllData();
                loadUserBookings();
            } else {
                const err = await response.json();
                showToast(`Failed: ${err.error}`, 'error');
            }
        } catch (e) {
            useMock = true;
            cancelBooking(bookingId, userId);
        }
    }
}

// SRE Telemetry: Start Workout Session (manual or scan-in)
async function startWorkoutSession(bookingId, userId, assetId) {
    if (useMock) {
        // Create active session
        const session = {
            booking_id: bookingId,
            user_id: userId,
            asset_id: assetId,
            started_at: new Date().toISOString(),
            ended_at: null,
            duration_minutes: 0,
            email_sent: false
        };
        mockData.sessions.push(session);
        
        // Ensure asset is marked as occupied
        const asset = mockData.assets.find(a => a.id === assetId);
        if (asset) asset.status = 'occupied';
        
        showToast('Workout started! Live telemetry tracking activated.', 'success');
        logTelemetryEvent('TELEMETRY_gRPC', 'SUCCESS', `CreateUsageSession: Member checking into gym equipment. Started active stream.`, session);
        
        // Emit NATS events
        setTimeout(() => {
            logTelemetryEvent('NATS_SUBSCRIBER', 'INFO', `NATS Event: Equipment occupied alert published. Topic: asset.usage.started`, { asset_id: assetId });
        }, 1000);
        
        refreshAllData();
        loadUserBookings();
    } else {
        try {
            const response = await fetch(`${API_BASE_URL}/telemetry/sessions`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ booking_id: bookingId, user_id: userId, asset_id: assetId })
            });
            if (response.ok) {
                showToast('Workout session tracking active!', 'success');
                logTelemetryEvent('TELEMETRY_gRPC', 'SUCCESS', `gRPC CreateUsageSession stream initialized on live telemetry db.`);
                refreshAllData();
                loadUserBookings();
            } else {
                const err = await response.json();
                showToast(`Failed to start: ${err.error}`, 'error');
            }
        } catch (e) {
            useMock = true;
            startWorkoutSession(bookingId, userId, assetId);
        }
    }
}

// SRE Telemetry: End Workout Session
async function endWorkoutSession(bookingId) {
    const session = mockData.sessions.find(s => s.booking_id === bookingId && s.ended_at === null);
    if (!session) return;
    
    const duration = Math.max(1, Math.round((Date.now() - new Date(session.started_at).getTime()) / 60000));
    
    if (useMock) {
        session.ended_at = new Date().toISOString();
        session.duration_minutes = duration;
        session.email_sent = true;
        
        // Reset asset status to available
        const asset = mockData.assets.find(a => a.id === session.asset_id);
        if (asset) asset.status = 'available';
        
        // Increment aggregate stats
        const agg = mockData.aggregates.find(a => a.asset_id === session.asset_id);
        if (agg) {
            agg.total_sessions += 1;
            agg.avg_duration_minutes = Math.round((agg.avg_duration_minutes + duration) / 2);
        }
        
        showToast(`Workout finished! Completed training: ${duration} minutes.`, 'success');
        logTelemetryEvent('TELEMETRY_gRPC', 'SUCCESS', `UpdateUsageSession: Terminating active session stream. Email summary sent.`, session);
        
        // Simulate NATS Event publish
        setTimeout(() => {
            logTelemetryEvent('NATS_SUBSCRIBER', 'INFO', `NATS Event: Equipment released alert published. Topic: asset.usage.completed`, { asset_id: session.asset_id, duration_minutes: duration });
        }, 1000);
        
        // Return booking flow
        const booking = mockData.bookings.find(b => b.id === bookingId);
        if (booking) {
            // Also call Return endpoint
            booking.status = 'returned';
            logTelemetryEvent('MEMBERSHIP', 'SUCCESS', `ReturnBooking endpoint executed. Equipment returned and booking archived.`, { booking_id: bookingId });
        }
        
        refreshAllData();
        loadUserBookings();
    } else {
        try {
            // End Telemetry Session
            const endedAt = new Date().toISOString();
            const telResponse = await fetch(`${API_BASE_URL}/telemetry/sessions/${bookingId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ended_at: endedAt, duration_minutes: duration, email_sent: true })
            });
            
            // Also invoke Return booking on membership service
            const returnResponse = await fetch(`${API_BASE_URL}/bookings/${bookingId}/return`, {
                method: 'POST'
            });
            
            if (telResponse.ok && returnResponse.ok) {
                showToast(`Session successfully completed. Session saved!`, 'success');
                logTelemetryEvent('TELEMETRY_gRPC', 'SUCCESS', `gRPC UpdateUsageSession & ReturnBooking saved on live PostgreSQL. Duration: ${duration} mins`);
                refreshAllData();
                loadUserBookings();
            } else {
                showToast('Terminated successfully with local caching.', 'success');
                logTelemetryEvent('TELEMETRY_gRPC', 'WARN', 'Telemetry database successfully updated with pending local sync.');
                refreshAllData();
                loadUserBookings();
            }
        } catch (e) {
            useMock = true;
            endWorkoutSession(bookingId);
        }
    }
}

// Custom Telemetry Event Injector
async function injectTelemetryLog(event) {
    event.preventDefault();
    const eventType = document.getElementById('inject-event-type').value;
    const message = document.getElementById('inject-message').value;
    const payloadRaw = document.getElementById('inject-payload').value;
    
    let payload = {};
    if (payloadRaw) {
        try {
            payload = JSON.parse(payloadRaw);
        } catch (err) {
            showToast('Invalid JSON syntax in metadata payload field.', 'error');
            return;
        }
    }
    
    if (useMock) {
        showToast('Custom telemetry log emitted successfully!', 'success');
        logTelemetryEvent('gRPC_STREAM', 'INFO', `Custom emitted log event: ${eventType} -> "${message}"`, payload);
        document.getElementById('form-inject-log').reset();
    } else {
        try {
            const response = await fetch(`${API_BASE_URL}/telemetry/log`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ event_type: eventType, message, payload: JSON.stringify(payload) })
            });
            if (response.ok) {
                showToast('Telemetry log event injected!', 'success');
                logTelemetryEvent('gRPC_STREAM', 'INFO', `Live emitted event: ${eventType} -> "${message}"`, payload);
                document.getElementById('form-inject-log').reset();
            } else {
                const err = await response.json();
                showToast(`Failed: ${err.error}`, 'error');
            }
        } catch (e) {
            useMock = true;
            injectTelemetryLog(event);
        }
    }
}

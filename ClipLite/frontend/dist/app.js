let currentTab = 'history';
let searchTimeout = null;
let autoRefreshInterval = null;
let allRecords = [];
let allFavorites = [];

async function loadRecords(search) {
    try {
        const records = await window.go.main.App.GetRecords(search || '');
        allRecords = records || [];
        renderRecords(allRecords, 'historyList');
    } catch (err) {
        console.error('加载记录失败:', err);
        document.getElementById('historyList').innerHTML = '<div class="empty-state">加载失败</div>';
    }
}

async function loadFavorites() {
    try {
        const records = await window.go.main.App.GetFavorites();
        allFavorites = records || [];
        renderRecords(allFavorites, 'favoritesList');
    } catch (err) {
        console.error('加载收藏失败:', err);
        document.getElementById('favoritesList').innerHTML = '<div class="empty-state">加载失败</div>';
    }
}

async function loadConfig() {
    try {
        const config = await window.go.main.App.GetConfig();
        document.getElementById('saveDaysInput').value = config.save_days || 30;
        document.getElementById('storagePathInput').value = config.storage_path || '';
    } catch (err) {
        console.error('加载配置失败:', err);
    }
}

function renderRecords(records, containerId) {
    const container = document.getElementById(containerId);

    if (!records || records.length === 0) {
        container.innerHTML = '<div class="empty-state">暂无记录</div>';
        return;
    }

    let html = '';
    for (let i = 0; i < records.length; i++) {
        const record = records[i];
        const favorited = record.is_favorited ? 'favorited' : '';
        const btnActive = record.is_favorited ? 'active' : '';
        const btnText = record.is_favorited ? '已收藏' : '收藏';
        const isLongContent = record.content.length > 100;
        const displayContent = escapeForDisplay(record.content);
        const recordIndex = i;

        html += '<div class="record-item ' + favorited + '" data-record-index="' + recordIndex + '" data-container="' + containerId + '">';
        html += '<div class="record-content">';
        html += '<div class="record-preview">';
        html += '<span class="preview-text">' + displayContent + '</span>';
        if (isLongContent) {
            html += '<span class="expand-toggle" data-action="expand" data-index="' + recordIndex + '" data-container="' + containerId + '">展开</span>';
        }
        html += '</div>';
        if (isLongContent) {
            html += '<div class="record-full" id="full-' + containerId + '-' + recordIndex + '">';
            html += '<span class="full-text">' + displayContent + '</span>';
            html += '<span class="expand-toggle" data-action="collapse" data-index="' + recordIndex + '" data-container="' + containerId + '">收起</span>';
            html += '</div>';
        }
        html += '</div>';
        html += '<div class="record-actions">';
        html += '<button class="record-btn use-btn" data-action="use" data-index="' + recordIndex + '" data-container="' + containerId + '">使用</button>';
        html += '<button class="record-btn favorite-btn ' + btnActive + '" data-action="favorite" data-id="' + record.id + '">' + btnText + '</button>';
        html += '<button class="record-btn delete-btn" data-action="delete" data-id="' + record.id + '">删除</button>';
        html += '</div>';
        html += '</div>';
    }
    container.innerHTML = html;
}

function getRecordByIndex(containerId, index) {
    if (containerId === 'historyList') {
        return allRecords[index];
    } else if (containerId === 'favoritesList') {
        return allFavorites[index];
    }
    return null;
}

document.addEventListener('click', function (e) {
    const target = e.target;
    const action = target.getAttribute('data-action');
    if (!action) return;

    e.stopPropagation();

    switch (action) {
        case 'use': {
            const index = parseInt(target.getAttribute('data-index'));
            const containerId = target.getAttribute('data-container');
            const record = getRecordByIndex(containerId, index);
            if (record) {
                useRecord(record.content);
            }
            break;
        }
        case 'favorite': {
            const id = parseInt(target.getAttribute('data-id'));
            toggleFavorite(id);
            break;
        }
        case 'delete': {
            const id = parseInt(target.getAttribute('data-id'));
            deleteRecord(id);
            break;
        }
        case 'expand': {
            const index = parseInt(target.getAttribute('data-index'));
            const containerId = target.getAttribute('data-container');
            toggleExpand(containerId, index, true);
            break;
        }
        case 'collapse': {
            const index = parseInt(target.getAttribute('data-index'));
            const containerId = target.getAttribute('data-container');
            toggleExpand(containerId, index, false);
            break;
        }
    }
});

function toggleExpand(containerId, index, expand) {
    const item = document.querySelector('[data-record-index="' + index + '"][data-container="' + containerId + '"]');
    if (!item) return;
    const preview = item.querySelector('.record-preview');
    const full = document.getElementById('full-' + containerId + '-' + index);

    if (preview && full) {
        if (expand) {
            preview.style.display = 'none';
            full.style.display = 'block';
        } else {
            preview.style.display = 'flex';
            full.style.display = 'none';
        }
    }
}

function escapeForDisplay(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

async function useRecord(content) {
    try {
        await window.go.main.App.CopyToClipboard(content);
    } catch (err) {
        try {
            await navigator.clipboard.writeText(content);
        } catch (err2) {
            try {
                const ta = document.createElement('textarea');
                ta.value = content;
                ta.style.position = 'fixed';
                ta.style.left = '-9999px';
                document.body.appendChild(ta);
                ta.select();
                document.execCommand('copy');
                document.body.removeChild(ta);
            } catch (err3) {
                console.error('使用失败:', err3);
            }
        }
    }
}

async function toggleFavorite(id) {
    try {
        const msg = await window.go.main.App.ToggleFavorite(id);
        if (msg && msg !== '已收藏' && msg !== '已取消收藏') {
            alert(msg);
        }
        loadRecords();
        loadFavorites();
    } catch (err) {
        console.error('切换收藏失败:', err);
    }
}

async function deleteRecord(id) {
    try {
        await window.go.main.App.DeleteRecord(id);
        loadRecords();
        loadFavorites();
    } catch (err) {
        console.error('删除失败:', err);
    }
}

async function clearAllRecords() {
    if (confirm('确定要清空所有非收藏记录吗？')) {
        try {
            await window.go.main.App.ClearAllRecords();
            loadRecords();
        } catch (err) {
            console.error('清空失败:', err);
        }
    }
}

function searchRecords() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(function () {
        const search = document.getElementById('searchInput').value;
        loadRecords(search);
    }, 300);
}

function switchTab(tab) {
    currentTab = tab;

    document.querySelectorAll('.tab').forEach(function (t) {
        t.classList.remove('active');
    });
    document.querySelector('[data-tab="' + tab + '"]').classList.add('active');

    document.querySelectorAll('.tab-content').forEach(function (c) {
        c.classList.remove('active');
    });
    document.getElementById(tab + 'Tab').classList.add('active');

    if (tab === 'history') {
        loadRecords();
    } else if (tab === 'favorites') {
        loadFavorites();
    } else if (tab === 'settings') {
        loadConfig();
    }
}

async function saveAllSettings() {
    const saveDays = parseInt(document.getElementById('saveDaysInput').value);
    const storagePath = document.getElementById('storagePathInput').value.trim();

    if (saveDays < 1 || saveDays > 3650) {
        alert('请输入1-3650之间的天数');
        return;
    }

    try {
        await window.go.main.App.UpdateConfig('', saveDays, storagePath);
        alert('设置已保存');
    } catch (err) {
        console.error('保存设置失败:', err);
        alert('保存失败: ' + err.message);
    }
}

async function selectStoragePath() {
    try {
        const path = await window.go.main.App.SelectDirectory();
        if (path) {
            document.getElementById('storagePathInput').value = path;
        }
    } catch (err) {
        console.error('选择路径失败:', err);
    }
}

function minimizeWindow() {
    window.runtime.WindowMinimise();
}

function hideWindow() {
    window.go.main.App.HideWindow();
}

function exitApp() {
    window.go.main.App.ExitApp();
}

document.addEventListener('DOMContentLoaded', function () {
    loadRecords();
    loadConfig();

    autoRefreshInterval = setInterval(function () {
        if (currentTab === 'history') {
            const search = document.getElementById('searchInput').value;
            loadRecords(search);
        } else if (currentTab === 'favorites') {
            loadFavorites();
        }
    }, 500);

    document.addEventListener('keydown', function (e) {
        if (e.key === 'Escape') {
            hideWindow();
        }
    });
});

document.addEventListener('beforeunload', function () {
    if (autoRefreshInterval) {
        clearInterval(autoRefreshInterval);
    }
});

const API_BASE = 'http://localhost:8080/api/v1';

// JWT Token Management
const TokenManager = {
    getToken() {
        return localStorage.getItem('jwt_token');
    },
    
    removeToken() {
        localStorage.removeItem('jwt_token');
    },
    
    isAuthenticated() {
        const token = this.getToken();
        if (!token) return false;
        
        try {
            const parts = token.split('.');
            if (parts.length !== 3) return false;
            const payload = JSON.parse(atob(parts[1]));
            return Date.now() < payload.exp * 1000;
        } catch (e) {
            return false;
        }
    }
};

// Check authentication
if (!TokenManager.isAuthenticated()) {
    window.location.href = '/html/login.html';
}

function getAuthHeaders() {
    const token = TokenManager.getToken();
    return {
        'Authorization': `Bearer ${token}`
    };
}

function getQueryParams() {
    const params = new URLSearchParams(window.location.search);
    return {
        id: params.get('id')
    };
}

async function loadReplay() {
    const { id } = getQueryParams();
    
    if (!id) {
        showError('ID реплея не указан');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/replays/${id}`, {
            headers: getAuthHeaders()
        });

        if (response.status === 401) {
            TokenManager.removeToken();
            window.location.href = '/html/login.html';
            return;
        }

        if (!response.ok) {
            throw new Error('Реплей не найден');
        }

        const replay = await response.json();
        displayReplay(replay);
    } catch (error) {
        console.error('Error loading replay:', error);
        showError('Ошибка загрузки реплея: ' + error.message);
    }
}

function getVideoType(filename) {
    const ext = filename.toLowerCase().split('.').pop();
    const types = {
        'mp4': 'video/mp4',
        'webm': 'video/webm',
        'ogg': 'video/ogg',
        'ogv': 'video/ogg',
        'mov': 'video/quicktime',
        'avi': 'video/x-msvideo',
        'mkv': 'video/x-matroska',
        'm4v': 'video/x-m4v'
    };
    return types[ext] || 'video/mp4';
}

function displayReplay(replay) {
    const headerInfo = document.getElementById('headerInfo');
    headerInfo.innerHTML = `
        <div class="replay-title-header">${replay.title || replay.original_name}</div>
        <div class="replay-meta">${replay.game_name || 'Неизвестная игра'}</div>
    `;

    const token = TokenManager.getToken();
    const videoUrl = `${API_BASE}/replays/${replay.id}/file?token=${encodeURIComponent(token)}`;
    const videoType = getVideoType(replay.original_name);
    const playerContent = document.getElementById('playerContent');
    
    playerContent.innerHTML = `
        <div class="video-wrapper">
            <video id="videoPlayer" controls autoplay preload="metadata">
                <source src="${videoUrl}" type="${videoType}">
                Ваш браузер не поддерживает воспроизведение этого видео формата.
            </video>
        </div>
        <div class="replay-details">
            <div class="detail-row">
                <div class="detail-label">Название:</div>
                <div class="detail-value">${replay.title || 'Без названия'}</div>
            </div>
            <div class="detail-row">
                <div class="detail-label">Файл:</div>
                <div class="detail-value">${replay.original_name}</div>
            </div>
            <div class="detail-row">
                <div class="detail-label">Размер:</div>
                <div class="detail-value">${(replay.size_bytes / 1024 / 1024).toFixed(2)} MB</div>
            </div>
            <div class="detail-row">
                <div class="detail-label">Загружено:</div>
                <div class="detail-value">${new Date(replay.uploaded_at).toLocaleString('ru-RU')}</div>
            </div>
            ${replay.comment ? `
                <div class="detail-row">
                    <div class="detail-label">Комментарий:</div>
                    <div class="detail-value">${replay.comment}</div>
                </div>
            ` : ''}
        </div>
    `;

    // Обработка ошибок загрузки видео
    const video = document.getElementById('videoPlayer');
    video.addEventListener('error', function(e) {
        console.error('Video error:', e);
        showError('Не удалось загрузить видео. Возможно, формат не поддерживается браузером.');
    });
    
    video.addEventListener('loadedmetadata', function() {
        console.log('Video loaded successfully');
    });
}

function showError(message) {
    const playerContent = document.getElementById('playerContent');
    playerContent.innerHTML = `
        <div class="error">
            <div class="error-icon">⚠️</div>
            <p>${message}</p>
        </div>
    `;
}

loadReplay();

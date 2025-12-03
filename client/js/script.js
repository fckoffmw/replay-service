const API_BASE = 'http://localhost:8080/api/v1';
let currentGameId = null;
let currentGameName = null;
let currentReplayId = null;

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
            if (parts.length !== 3) {
                console.warn('Invalid token format');
                this.removeToken();
                return false;
            }
            const payload = JSON.parse(atob(parts[1]));
            const exp = payload.exp * 1000;
            const isValid = Date.now() < exp;
            if (!isValid) {
                console.warn('Token expired');
                this.removeToken();
            }
            return isValid;
        } catch (e) {
            console.error('Token validation error:', e);
            this.removeToken();
            return false;
        }
    },
    
    getUserFromToken() {
        const token = this.getToken();
        if (!token) return null;
        
        try {
            const parts = token.split('.');
            if (parts.length !== 3) {
                console.warn('Invalid token format in getUserFromToken');
                this.removeToken();
                return null;
            }
            const payload = JSON.parse(atob(parts[1]));
            return {
                id: payload.user_id,
                login: payload.login
            };
        } catch (e) {
            console.error('Error parsing token:', e);
            this.removeToken();
            return null;
        }
    }
};

// Check authentication on page load
if (!TokenManager.isAuthenticated()) {
    window.location.href = '/html/login.html';
}

function getAuthHeaders() {
    const token = TokenManager.getToken();
    return {
        'Authorization': `Bearer ${token}`
    };
}

function logout() {
    TokenManager.removeToken();
    window.location.href = '/html/login.html';
}

function getUserId() {
    const user = TokenManager.getUserFromToken();
    return user ? user.id : null;
}

function showCreateGameModal() {
    document.getElementById('createGameModal').classList.add('active');
    document.getElementById('newGameName').value = '';
    document.getElementById('newGameName').focus();
}

function hideCreateGameModal() {
    document.getElementById('createGameModal').classList.remove('active');
}

function showEditGameModal(gameId, gameName) {
    currentGameId = gameId;
    currentGameName = gameName;
    document.getElementById('editGameModal').classList.add('active');
    document.getElementById('editGameName').value = gameName;
    document.getElementById('editGameName').focus();
}

function hideEditGameModal() {
    document.getElementById('editGameModal').classList.remove('active');
}

function showEditReplayModal(replayId, title, comment) {
    currentReplayId = replayId;
    document.getElementById('editReplayModal').classList.add('active');
    document.getElementById('editReplayTitle').value = title || '';
    document.getElementById('editReplayComment').value = comment || '';
    document.getElementById('editReplayTitle').focus();
}

function hideEditReplayModal() {
    document.getElementById('editReplayModal').classList.remove('active');
}

function toggleUploadForm() {
    const form = document.getElementById('uploadForm');
    form.classList.toggle('collapsed');
}

function handleFileSelect(event) {
    const file = event.target.files[0];
    const fileInfo = document.getElementById('fileInfo');
    
    if (!file) {
        fileInfo.style.display = 'none';
        return;
    }

    // Validate file type
    const allowedExtensions = ['.mp4', '.webm', '.ogg', '.ogv', '.mov', '.avi', '.mkv', '.m4v'];
    const fileName = file.name.toLowerCase();
    const isValidExtension = allowedExtensions.some(ext => fileName.endsWith(ext));
    const isVideoMimeType = file.type.startsWith('video/');

    if (!isValidExtension && !isVideoMimeType) {
        fileInfo.style.display = 'block';
        fileInfo.style.borderColor = 'rgba(239, 68, 68, 0.3)';
        fileInfo.style.background = 'rgba(239, 68, 68, 0.1)';
        fileInfo.style.color = '#fca5a5';
        fileInfo.innerHTML = `
            <strong>❌ Неверный формат файла</strong><br>
            Можно загружать только видео файлы
        `;
        event.target.value = '';
        return;
    }

    // Show file info
    const fileSize = (file.size / 1024 / 1024).toFixed(2);
    fileInfo.style.display = 'block';
    fileInfo.style.borderColor = 'rgba(167, 139, 250, 0.3)';
    fileInfo.style.background = 'rgba(167, 139, 250, 0.1)';
    fileInfo.style.color = '#c4b5fd';
    fileInfo.innerHTML = `
        <strong>✅ Файл выбран:</strong><br>
        📁 ${file.name}<br>
        💾 Размер: ${fileSize} MB
    `;
}

// Toast notification system
function showToast(message, type = 'info', title = '') {
    // Create toast container if it doesn't exist
    let container = document.getElementById('toastContainer');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toastContainer';
        container.className = 'toast-container';
        document.body.appendChild(container);
    }

    // Create toast element
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    const icons = {
        success: '✅',
        error: '❌',
        info: 'ℹ️',
        warning: '⚠️'
    };

    toast.innerHTML = `
        <div class="toast-icon">${icons[type] || icons.info}</div>
        <div class="toast-content">
            ${title ? `<div class="toast-title">${title}</div>` : ''}
            <div class="toast-message">${message}</div>
        </div>
    `;

    container.appendChild(toast);

    // Auto remove after 4 seconds
    setTimeout(() => {
        toast.classList.add('hiding');
        setTimeout(() => {
            container.removeChild(toast);
            if (container.children.length === 0) {
                document.body.removeChild(container);
            }
        }, 300);
    }, 4000);
}

function isVideoFile(filename) {
    const videoExtensions = ['.mp4', '.webm', '.ogg', '.ogv', '.mov', '.avi', '.mkv', '.m4v'];
    return videoExtensions.some(ext => filename.toLowerCase().endsWith(ext));
}

function playVideo(replayId) {
    window.location.href = `/html/player.html?id=${replayId}`;
}

async function loadGames() {
    try {
        const response = await fetch(`${API_BASE}/games`, {
            headers: getAuthHeaders()
        });
        
        if (response.status === 401) {
            // Token invalid or expired
            TokenManager.removeToken();
            window.location.href = '/html/login.html';
            return;
        }
        
        const games = await response.json();
        
        const gameList = document.getElementById('gameList');
        
        if (!games || games.length === 0) {
            gameList.innerHTML = '<li class="empty-state" style="padding: 20px; text-align: center; color: #999;">Нет игр</li>';
            return;
        }

        gameList.innerHTML = games.map(game => `
            <li class="game-item" onclick="selectGame('${game.id}', '${game.name}')" id="game-${game.id}">
                <div class="game-name">${game.name}</div>
                <div class="game-count">${game.replay_count} реплеев</div>
            </li>
        `).join('');
    } catch (error) {
        console.error('Error loading games:', error);
        document.getElementById('gameList').innerHTML = '<li style="color: red; padding: 10px;">Ошибка загрузки</li>';
    }
}

async function createGame() {
    const name = document.getElementById('newGameName').value.trim();
    if (!name) {
        showToast('Введите название игры', 'warning', 'Название не указано');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/games`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                ...getAuthHeaders()
            },
            body: JSON.stringify({ name })
        });
        
        if (response.ok) {
            hideCreateGameModal();
            showToast(`Игра "${name}" успешно создана!`, 'success', 'Успешно');
            await loadGames();
        } else {
            showToast('Не удалось создать игру', 'error', 'Ошибка');
        }
    } catch (error) {
        console.error('Error creating game:', error);
        showToast('Ошибка создания игры: ' + error.message, 'error', 'Ошибка');
    }
}

async function selectGame(gameId, gameName) {
    currentGameId = gameId;
    
    document.querySelectorAll('.game-item').forEach(item => {
        item.classList.remove('active');
    });
    document.getElementById(`game-${gameId}`).classList.add('active');

    const contentArea = document.getElementById('contentArea');
    contentArea.innerHTML = '<div class="loading">Загрузка реплеев...</div>';

    try {
        const response = await fetch(`${API_BASE}/games/${gameId}/replays?limit=100`, {
            headers: getAuthHeaders()
        });
        const replays = await response.json() || [];

        contentArea.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                <h2 style="margin: 0;">${gameName}</h2>
                <div style="display: flex; gap: 8px;">
                    <button class="btn btn-primary btn-small" onclick="showEditGameModal('${gameId}', '${gameName}')">Редактировать</button>
                    <button class="btn btn-danger btn-small" onclick="deleteGame('${gameId}', '${gameName}')">Удалить игру</button>
                </div>
            </div>

            <div class="upload-form collapsed" id="uploadForm">
                <div class="upload-form-header" onclick="toggleUploadForm()">
                    <div class="upload-form-title">
                        <span>📤</span>
                        <span>Загрузить реплей</span>
                    </div>
                    <div class="upload-form-toggle">▼</div>
                </div>
                <div class="upload-form-content">
                    <div class="form-group">
                        <label>Файл (только видео)</label>
                        <input type="file" id="replayFile" accept="video/*,.mp4,.webm,.ogg,.mov,.avi,.mkv,.m4v" onchange="handleFileSelect(event)">
                        <small style="color: #a78bfa; font-size: 12px; margin-top: 4px; display: block;">
                            Поддерживаемые форматы: MP4, WebM, OGG, MOV, AVI, MKV, M4V
                        </small>
                        <div id="fileInfo" style="display: none; margin-top: 10px; padding: 10px; background: rgba(167, 139, 250, 0.1); border: 1px solid rgba(167, 139, 250, 0.3); border-radius: 8px; color: #c4b5fd; font-size: 13px;"></div>
                    </div>
                    <div class="form-group">
                        <label>Название (опционально)</label>
                        <input type="text" id="replayTitle" placeholder="Например: Эпичная победа">
                    </div>
                    <div class="form-group">
                        <label>Комментарий (опционально)</label>
                        <textarea id="replayComment" placeholder="Описание реплея..."></textarea>
                    </div>
                    <button class="btn btn-success" onclick="uploadReplay()">Загрузить</button>
                </div>
            </div>

            <h3 style="margin: 20px 0 15px 0;">Реплеи (${replays.length})</h3>
            <div class="replay-list" id="replayList">
                ${replays.length === 0 ? 
                    '<div class="empty-state"><div class="empty-state-icon">📄</div><p>Нет реплеев</p></div>' :
                    replays.map(replay => {
                        const isVideo = isVideoFile(replay.original_name);
                        const videoUrl = `${API_BASE}/replays/${replay.id}/file`;
                        
                        return `
                            <div class="replay-card">
                                <div class="replay-content">
                                    <div class="replay-header">
                                        <div class="replay-title">${replay.title || replay.original_name}</div>
                                    </div>
                                    <div class="replay-info">📁 ${replay.original_name}</div>
                                    <div class="replay-info">💾 ${(replay.size_bytes / 1024).toFixed(2)} KB</div>
                                    <div class="replay-info">📅 ${new Date(replay.uploaded_at).toLocaleString('ru-RU')}</div>
                                    ${replay.comment ? `<div class="replay-info">💬 ${replay.comment}</div>` : ''}
                                    <div class="replay-actions">
                                        ${isVideo ? `<button class="btn btn-success btn-small" onclick="playReplay('${replay.id}')">▶ Проиграть</button>` : ''}
                                        <button class="btn btn-primary btn-small" onclick="downloadReplay('${replay.id}')">⬇ Скачать</button>
                                        <button class="btn btn-primary btn-small" onclick="showEditReplayModal('${replay.id}', '${(replay.title || '').replace(/'/g, "\\'")}', '${(replay.comment || '').replace(/'/g, "\\'")}')">Редактировать</button>
                                        <button class="btn btn-danger btn-small" onclick="deleteReplay('${replay.id}')">Удалить</button>
                                    </div>
                                </div>
                                ${isVideo ? `
                                    <div class="replay-video-preview">
                                        <video src="${videoUrl}" preload="metadata"></video>
                                        <div class="play-overlay" onclick="playVideo('${replay.id}')">
                                            <div class="play-button">
                                                <div class="play-icon"></div>
                                            </div>
                                        </div>
                                    </div>
                                ` : ''}
                            </div>
                        `;
                    }).join('')
                }
            </div>
        `;
    } catch (error) {
        console.error('Error loading replays:', error);
        contentArea.innerHTML = '<div style="color: red; padding: 20px;">Ошибка загрузки реплеев</div>';
    }
}

async function uploadReplay() {
    if (!currentGameId) return;

    const fileInput = document.getElementById('replayFile');
    const title = document.getElementById('replayTitle').value;
    const comment = document.getElementById('replayComment').value;
    const uploadButton = event.target;

    if (!fileInput.files[0]) {
        showToast('Пожалуйста, выберите файл для загрузки', 'warning', 'Файл не выбран');
        return;
    }

    // Validate file type
    const file = fileInput.files[0];
    const allowedExtensions = ['.mp4', '.webm', '.ogg', '.ogv', '.mov', '.avi', '.mkv', '.m4v'];
    const fileNameLower = file.name.toLowerCase();
    const isValidExtension = allowedExtensions.some(ext => fileNameLower.endsWith(ext));
    const isVideoMimeType = file.type.startsWith('video/');

    if (!isValidExtension && !isVideoMimeType) {
        showToast(
            'Можно загружать только видео файлы (MP4, WebM, OGG, MOV, AVI, MKV, M4V)', 
            'error', 
            'Неверный формат файла'
        );
        fileInput.value = ''; // Clear the input
        return;
    }

    // Disable button and show loading state
    uploadButton.disabled = true;
    uploadButton.classList.add('loading');
    uploadButton.setAttribute('data-original-text', uploadButton.textContent);
    
    const fileName = file.name;
    const fileSize = (file.size / 1024 / 1024).toFixed(2);
    
    showToast(`Загрузка файла "${fileName}" (${fileSize} MB)...`, 'info', 'Подождите');

    const formData = new FormData();
    formData.append('file', fileInput.files[0]);
    formData.append('title', title);
    formData.append('comment', comment);

    try {
        const response = await fetch(`${API_BASE}/games/${currentGameId}/replays`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: formData
        });

        if (!response.ok) {
            const error = await response.json();
            console.error('Upload error:', error);
            showToast(error.error || 'Не удалось загрузить файл', 'error', 'Ошибка загрузки');
            return;
        }

        // Success!
        showToast(`Реплей "${title || fileName}" успешно загружен!`, 'success', 'Успешно');
        
        // Clear form
        fileInput.value = '';
        document.getElementById('replayTitle').value = '';
        document.getElementById('replayComment').value = '';
        
        // Hide file info
        const fileInfo = document.getElementById('fileInfo');
        if (fileInfo) {
            fileInfo.style.display = 'none';
        }
        
        // Collapse upload form
        const uploadForm = document.getElementById('uploadForm');
        if (uploadForm) {
            uploadForm.classList.add('collapsed');
        }

        const activeGame = document.querySelector('.game-item.active');
        if (activeGame && currentGameId) {
            const gameName = activeGame.querySelector('.game-name').textContent;
            await selectGame(currentGameId, gameName);
        }
        await loadGames();
    } catch (error) {
        console.error('Error uploading replay:', error);
        showToast('Ошибка загрузки файла: ' + error.message, 'error', 'Ошибка');
    } finally {
        // Re-enable button
        uploadButton.disabled = false;
        uploadButton.classList.remove('loading');
    }
}

function downloadReplay(replayId) {
    const token = TokenManager.getToken();
    window.open(`${API_BASE}/replays/${replayId}/file?download=true&token=${encodeURIComponent(token)}`, '_blank');
}

function playReplay(replayId) {
    // Для видео - открыть страницу плеера, для остальных - просто открыть файл
    window.location.href = `/html/player.html?id=${replayId}`;
}

async function deleteReplay(replayId) {
    if (!confirm('Удалить этот реплей?')) return;

    try {
        const response = await fetch(`${API_BASE}/replays/${replayId}`, {
            method: 'DELETE',
            headers: getAuthHeaders()
        });

        if (!response.ok) {
            const error = await response.json();
            console.error('Delete error:', error);
            showToast('Не удалось удалить реплей', 'error', 'Ошибка');
            return;
        }

        showToast('Реплей успешно удален', 'success', 'Успешно');

        const activeGame = document.querySelector('.game-item.active');
        if (activeGame && currentGameId) {
            const gameName = activeGame.querySelector('.game-name').textContent;
            await selectGame(currentGameId, gameName);
        }
        await loadGames();
    } catch (error) {
        console.error('Error deleting replay:', error);
        showToast('Ошибка удаления: ' + error.message, 'error', 'Ошибка');
    }
}

async function deleteGame(gameId, gameName) {
    if (!confirm(`Удалить игру "${gameName}" и все её реплеи?`)) return;

    try {
        const response = await fetch(`${API_BASE}/games/${gameId}`, {
            method: 'DELETE',
            headers: getAuthHeaders()
        });

        if (response.ok) {
            showToast(`Игра "${gameName}" и все её реплеи удалены`, 'success', 'Успешно');
            currentGameId = null;
            document.getElementById('contentArea').innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📁</div>
                    <p>Выберите игру из списка слева</p>
                </div>
            `;
            await loadGames();
        } else {
            showToast('Не удалось удалить игру', 'error', 'Ошибка');
        }
    } catch (error) {
        console.error('Error deleting game:', error);
        showToast('Ошибка удаления игры: ' + error.message, 'error', 'Ошибка');
    }
}

async function updateGame() {
    const name = document.getElementById('editGameName').value.trim();
    if (!name) {
        showToast('Введите название игры', 'warning', 'Название не указано');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/games/${currentGameId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                ...getAuthHeaders()
            },
            body: JSON.stringify({ name })
        });

        if (response.ok) {
            hideEditGameModal();
            showToast(`Игра переименована в "${name}"`, 'success', 'Успешно');
            await loadGames();
            await selectGame(currentGameId, name);
        } else {
            showToast('Не удалось обновить игру', 'error', 'Ошибка');
        }
    } catch (error) {
        console.error('Error updating game:', error);
        showToast('Ошибка обновления игры: ' + error.message, 'error', 'Ошибка');
    }
}

async function updateReplay() {
    const title = document.getElementById('editReplayTitle').value.trim();
    const comment = document.getElementById('editReplayComment').value.trim();

    try {
        const formData = new FormData();
        if (title) formData.append('title', title);
        if (comment) formData.append('comment', comment);

        const response = await fetch(`${API_BASE}/replays/${currentReplayId}`, {
            method: 'PUT',
            headers: getAuthHeaders(),
            body: formData
        });

        if (response.ok) {
            hideEditReplayModal();
            showToast('Реплей успешно обновлен', 'success', 'Успешно');
            const activeGame = document.querySelector('.game-item.active');
            if (activeGame && currentGameId) {
                const gameName = activeGame.querySelector('.game-name').textContent;
                await selectGame(currentGameId, gameName);
            }
        } else {
            showToast('Не удалось обновить реплей', 'error', 'Ошибка');
        }
    } catch (error) {
        console.error('Error updating replay:', error);
        showToast('Ошибка обновления реплея: ' + error.message, 'error', 'Ошибка');
    }
}

// Display user info
const user = TokenManager.getUserFromToken();
if (user) {
    document.getElementById('userDisplay').textContent = `👤 ${user.login}`;
}

loadGames();

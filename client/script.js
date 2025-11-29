const API_BASE = 'http://localhost:8080/api/v1';
let currentGameId = null;
let currentGameName = null;
let currentReplayId = null;

function getUserId() {
    return document.getElementById('userId').value;
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

function isVideoFile(filename) {
    const videoExtensions = ['.mp4', '.webm', '.ogg', '.mov', '.avi', '.mkv', '.m4v'];
    return videoExtensions.some(ext => filename.toLowerCase().endsWith(ext));
}

function playVideo(replayId) {
    window.location.href = `player.html?id=${replayId}&userId=${getUserId()}`;
}

async function loadGames() {
    try {
        const response = await fetch(`${API_BASE}/games`, {
            headers: { 'X-User-ID': getUserId() }
        });
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
        alert('Введите название игры');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/games`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': getUserId()
            },
            body: JSON.stringify({ name })
        });
        
        if (response.ok) {
            hideCreateGameModal();
            await loadGames();
        } else {
            alert('Ошибка создания игры');
        }
    } catch (error) {
        console.error('Error creating game:', error);
        alert('Ошибка создания игры');
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
            headers: { 'X-User-ID': getUserId() }
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

            <div class="upload-form">
                <h3 style="margin-bottom: 15px;">Загрузить реплей</h3>
                <div class="form-group">
                    <label>Файл</label>
                    <input type="file" id="replayFile">
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
                                        <button class="btn btn-primary btn-small" onclick="downloadReplay('${replay.id}')">Скачать</button>
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

    if (!fileInput.files[0]) {
        alert('Выберите файл');
        return;
    }

    const formData = new FormData();
    formData.append('file', fileInput.files[0]);
    formData.append('title', title);
    formData.append('comment', comment);

    try {
        const response = await fetch(`${API_BASE}/games/${currentGameId}/replays`, {
            method: 'POST',
            headers: { 'X-User-ID': getUserId() },
            body: formData
        });

        if (!response.ok) {
            const error = await response.json();
            console.error('Upload error:', error);
            alert('Ошибка загрузки файла');
            return;
        }

        const activeGame = document.querySelector('.game-item.active');
        if (activeGame && currentGameId) {
            const gameName = activeGame.querySelector('.game-name').textContent;
            await selectGame(currentGameId, gameName);
        }
        await loadGames();
    } catch (error) {
        console.error('Error uploading replay:', error);
        alert('Ошибка загрузки файла: ' + error.message);
    }
}

function downloadReplay(replayId) {
    window.open(`${API_BASE}/replays/${replayId}/file`, '_blank');
}

async function deleteReplay(replayId) {
    if (!confirm('Удалить этот реплей?')) return;

    try {
        const response = await fetch(`${API_BASE}/replays/${replayId}`, {
            method: 'DELETE',
            headers: { 'X-User-ID': getUserId() }
        });

        if (!response.ok) {
            const error = await response.json();
            console.error('Delete error:', error);
            alert('Ошибка удаления');
            return;
        }

        const activeGame = document.querySelector('.game-item.active');
        if (activeGame && currentGameId) {
            const gameName = activeGame.querySelector('.game-name').textContent;
            await selectGame(currentGameId, gameName);
        }
        await loadGames();
    } catch (error) {
        console.error('Error deleting replay:', error);
        alert('Ошибка удаления: ' + error.message);
    }
}

async function deleteGame(gameId, gameName) {
    if (!confirm(`Удалить игру "${gameName}" и все её реплеи?`)) return;

    try {
        const response = await fetch(`${API_BASE}/games/${gameId}`, {
            method: 'DELETE',
            headers: { 'X-User-ID': getUserId() }
        });

        if (response.ok) {
            currentGameId = null;
            document.getElementById('contentArea').innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📁</div>
                    <p>Выберите игру из списка слева</p>
                </div>
            `;
            await loadGames();
        } else {
            alert('Ошибка удаления игры');
        }
    } catch (error) {
        console.error('Error deleting game:', error);
        alert('Ошибка удаления игры');
    }
}

async function updateGame() {
    const name = document.getElementById('editGameName').value.trim();
    if (!name) {
        alert('Введите название игры');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/games/${currentGameId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': getUserId()
            },
            body: JSON.stringify({ name })
        });

        if (response.ok) {
            hideEditGameModal();
            await loadGames();
            await selectGame(currentGameId, name);
        } else {
            alert('Ошибка обновления игры');
        }
    } catch (error) {
        console.error('Error updating game:', error);
        alert('Ошибка обновления игры');
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
            headers: { 'X-User-ID': getUserId() },
            body: formData
        });

        if (response.ok) {
            hideEditReplayModal();
            const activeGame = document.querySelector('.game-item.active');
            if (activeGame && currentGameId) {
                const gameName = activeGame.querySelector('.game-name').textContent;
                await selectGame(currentGameId, gameName);
            }
        } else {
            alert('Ошибка обновления реплея');
        }
    } catch (error) {
        console.error('Error updating replay:', error);
        alert('Ошибка обновления реплея');
    }
}

loadGames();

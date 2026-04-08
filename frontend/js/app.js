// ===== Глобальные переменные =====
let authToken = localStorage.getItem('authToken') || null;
let currentUser = localStorage.getItem('currentUser') || null;
let allTracks = []; // Все треки пользователя
let currentTrackIndex = -1; // Индекс текущего трека
let audioElement = null; // HTML5 Audio элемент
let isPlaying = false; // Состояние воспроизведения

// ===== Инициализация при загрузке страницы =====
document.addEventListener('DOMContentLoaded', () => {
    checkAuth();
    setupEventListeners();
    setupAudioPlayer();
});

// ===== Настройка аудиоплеера =====
function setupAudioPlayer() {
    audioElement = document.getElementById('audioElement');
    
    // Обновление прогресс бара
    audioElement.addEventListener('timeupdate', updateProgress);
    
    // Когда трек заканчивается — включаем следующий
    audioElement.addEventListener('ended', () => {
        playNextTrack();
    });
    
    // Обработка ошибок
    audioElement.addEventListener('error', (e) => {
        console.error('Audio error:', e);
        showPlayerStatus('❌ Ошибка воспроизведения', 'error');
    });
    
    // Когда метаданные загружены — обновляем время
    audioElement.addEventListener('loadedmetadata', () => {
        updatePlayerTime();
        document.getElementById('progressBar').max = Math.floor(audioElement.duration);
    });
}

// ===== Проверка авторизации =====
function checkAuth() {
    if (authToken && currentUser) {
        showDashboard();
    } else {
        showLoginForm();
    }
}

// ===== Переключение между формами =====
function showLoginForm() {
    document.getElementById('loginForm').style.display = 'block';
    document.getElementById('registerForm').style.display = 'none';
    document.getElementById('dashboard').style.display = 'none';
    document.getElementById('userInfo').style.display = 'none';
    document.getElementById('audioPlayer').style.display = 'none';
}

function showRegisterForm() {
    document.getElementById('loginForm').style.display = 'none';
    document.getElementById('registerForm').style.display = 'block';
    document.getElementById('dashboard').style.display = 'none';
    document.getElementById('userInfo').style.display = 'none';
    document.getElementById('audioPlayer').style.display = 'none';
}

function showDashboard() {
    document.getElementById('loginForm').style.display = 'none';
    document.getElementById('registerForm').style.display = 'none';
    document.getElementById('dashboard').style.display = 'block';
    document.getElementById('userInfo').style.display = 'flex';
    document.getElementById('username').textContent = currentUser;
    document.getElementById('audioPlayer').style.display = 'block';
    
    // Загружаем список треков
    loadTracks();
}

// ===== Настройка обработчиков событий =====
function setupEventListeners() {
    // Переключение между входом и регистрацией
    document.getElementById('showRegister').addEventListener('click', (e) => {
        e.preventDefault();
        showRegisterForm();
    });
    
    document.getElementById('showLogin').addEventListener('click', (e) => {
        e.preventDefault();
        showLoginForm();
    });
    
    // Форма входа
    document.getElementById('loginFormEl').addEventListener('submit', async (e) => {
        e.preventDefault();
        await handleLogin();
    });
    
    // Форма регистрации
    document.getElementById('registerFormEl').addEventListener('submit', async (e) => {
        e.preventDefault();
        await handleRegister();
    });
    
    // Форма загрузки трека
    document.getElementById('uploadForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await handleUpload();
    });
    
    // Кнопка выхода
    document.getElementById('logoutBtn').addEventListener('click', handleLogout);
    
    // 🔊 Контролы плеера
    document.getElementById('playBtn').addEventListener('click', togglePlayPause);
    document.getElementById('prevBtn').addEventListener('click', playPrevTrack);
    document.getElementById('nextBtn').addEventListener('click', playNextTrack);
    document.getElementById('progressBar').addEventListener('input', seekAudio);
    document.getElementById('speedSelect').addEventListener('change', changeSpeed);
}

// ===== Вход =====
async function handleLogin() {
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;
    
    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });
        
        const data = await response.json();
        
        if (response.ok) {
            authToken = data.token;
            currentUser = email;
            localStorage.setItem('authToken', authToken);
            localStorage.setItem('currentUser', currentUser);
            showDashboard();
        } else {
            alert(data.error || 'Ошибка входа');
        }
    } catch (error) {
        alert('Ошибка подключения к серверу');
        console.error(error);
    }
}

// ===== Регистрация =====
async function handleRegister() {
    const email = document.getElementById('regEmail').value;
    const username = document.getElementById('regUsername').value;
    const password = document.getElementById('regPassword').value;
    
    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, username, password }),
        });
        
        const data = await response.json();
        
        if (response.ok) {
            alert('Регистрация успешна! Теперь войдите.');
            showLoginForm();
        } else {
            alert(data.error || 'Ошибка регистрации');
        }
    } catch (error) {
        alert('Ошибка подключения к серверу');
        console.error(error);
    }
}

// ===== Выход =====
function handleLogout() {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('authToken');
    localStorage.removeItem('currentUser');
    stopAudio();
    showLoginForm();
}

// ===== Загрузка трека =====
async function handleUpload() {
    const fileInput = document.getElementById('trackFile');
    const statusDiv = document.getElementById('uploadStatus');
    
    if (!fileInput.files[0]) {
        statusDiv.textContent = 'Выберите файл';
        statusDiv.className = 'status error';
        return;
    }
    
    const formData = new FormData();
    formData.append('file', fileInput.files[0]);
    
    try {
        statusDiv.textContent = 'Загрузка...';
        statusDiv.className = 'status';
        
        const response = await fetch('/api/tracks/upload', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authToken}`,
            },
            body: formData,
        });
        
        const data = await response.json();
        
        if (response.ok) {
            statusDiv.textContent = '✅ Трек загружен успешно!';
            statusDiv.className = 'status success';
            fileInput.value = '';
            loadTracks(); // Обновляем список треков
        } else {
            statusDiv.textContent = `❌ ${data.error}`;
            statusDiv.className = 'status error';
        }
    } catch (error) {
        statusDiv.textContent = '❌ Ошибка подключения к серверу';
        statusDiv.className = 'status error';
        console.error(error);
    }
}

// ===== Загрузка списка треков =====
async function loadTracks() {
    const tracksList = document.getElementById('tracksList');
    tracksList.innerHTML = '<p>Загрузка...</p>';
    
    try {
        const response = await fetch('/api/tracks', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${authToken}`,
            },
        });
        
        const data = await response.json();
        
        if (response.ok && data.tracks && data.tracks.length > 0) {
            allTracks = data.tracks;
            tracksList.innerHTML = '';
            data.tracks.forEach((track, index) => {
                tracksList.appendChild(createTrackElement(track, index));
            });
        } else {
            allTracks = [];
            tracksList.innerHTML = '<p>Треков пока нет</p>';
        }
    } catch (error) {
        allTracks = [];
        tracksList.innerHTML = '<p>Ошибка загрузки треков</p>';
        console.error(error);
    }
}

// ===== Создание элемента трека =====
function createTrackElement(track, index) {
    const div = document.createElement('div');
    div.className = 'track-item';
    div.dataset.trackIndex = index;
    
    // Форматируем длительность
    const duration = track.duration ? formatTime(track.duration) : 'N/A';
    
    div.innerHTML = `
        <div class="track-info">
            <span class="track-title">🎵 ${track.title}</span>
            <span class="track-meta">${duration} • ${(track.file_size / 1024 / 1024).toFixed(2)} MB</span>
        </div>
        <div class="waveform-container" data-track-index="${index}">
            <canvas class="waveform-canvas" data-waveform="${JSON.stringify(track.waveform_data || [])}"></canvas>
        </div>
        <div class="track-actions">
            <button class="btn-play-track" data-track-index="${index}">▶ Воспроизвести</button>
        </div>
    `;
    
    // Добавляем обработчик клика на кнопку воспроизведения
    const playBtn = div.querySelector('.btn-play-track');
    playBtn.addEventListener('click', () => {
        playTrack(index);
    });
    
    // Добавляем обработчик клика на waveform
    const waveformContainer = div.querySelector('.waveform-container');
    waveformContainer.addEventListener('click', () => {
        playTrack(index);
    });
    
    // Рисуем волну после добавления в DOM
    setTimeout(() => {
        const canvas = div.querySelector('.waveform-canvas');
        drawWaveform(canvas);
    }, 0);
    
    return div;
}

// ===== 🔊 ВОСПРОИЗВЕДЕНИЕ ТРЕКА 🔊 =====
function playTrack(index) {
    if (index < 0 || index >= allTracks.length) {
        console.error('Invalid track index:', index);
        return;
    }
    
    const track = allTracks[index];
    currentTrackIndex = index;
    
    // Обновляем название в плеере
    document.getElementById('playerTitle').textContent = track.title;
    
    // Устанавливаем источник аудио (путь к файлу на сервере)
    // Важно: нужно настроить раздачу файлов из папки uploads в Go
    audioElement.src = '/' + track.file_path.replace(/\\/g, '/');
    
    // Загружаем и воспроизводим
    audioElement.load();
    audioElement.play()
        .then(() => {
            isPlaying = true;
            updatePlayPauseButton();
            document.getElementById('audioPlayer').classList.add('playing');
        })
        .catch((error) => {
            console.error('Playback error:', error);
            showPlayerStatus('❌ Ошибка воспроизведения', 'error');
        });
}

// ===== Пауза / Воспроизведение =====
function togglePlayPause() {
    if (!audioElement.src) {
        showPlayerStatus('⚠️ Выберите трек для воспроизведения', 'error');
        return;
    }
    
    if (isPlaying) {
        audioElement.pause();
        isPlaying = false;
        document.getElementById('audioPlayer').classList.remove('playing');
    } else {
        audioElement.play();
        isPlaying = true;
        document.getElementById('audioPlayer').classList.add('playing');
    }
    
    updatePlayPauseButton();
}

// ===== Обновление кнопки Play/Pause =====
function updatePlayPauseButton() {
    const playBtn = document.getElementById('playBtn');
    playBtn.textContent = isPlaying ? '⏸' : '▶';
    playBtn.title = isPlaying ? 'Пауза' : 'Воспроизвести';
}

// ===== Предыдущий трек =====
function playPrevTrack() {
    if (allTracks.length === 0) return;
    
    let newIndex = currentTrackIndex - 1;
    if (newIndex < 0) {
        newIndex = allTracks.length - 1; // Зацикливаем
    }
    
    playTrack(newIndex);
}

// ===== Следующий трек =====
function playNextTrack() {
    if (allTracks.length === 0) return;
    
    let newIndex = currentTrackIndex + 1;
    if (newIndex >= allTracks.length) {
        newIndex = 0; // Зацикливаем
    }
    
    playTrack(newIndex);
}

// ===== Перемотка =====
function seekAudio(e) {
    const progressBar = e.target;
    const newTime = parseInt(progressBar.value);
    
    if (!isNaN(newTime) && audioElement.duration) {
        audioElement.currentTime = newTime;
        updatePlayerTime();
    }
}

// ===== Обновление прогресс бара =====
function updateProgress() {
    const progressBar = document.getElementById('progressBar');
    if (audioElement.duration) {
        progressBar.value = Math.floor(audioElement.currentTime);
        updatePlayerTime();
    }
}

// ===== Обновление времени =====
function updatePlayerTime() {
    const currentTime = audioElement.currentTime || 0;
    const duration = audioElement.duration || 0;
    
    document.getElementById('playerTime').textContent = 
        `${formatTime(currentTime)} / ${formatTime(duration)}`;
}

// ===== Форматирование времени =====
function formatTime(seconds) {
    if (!seconds || isNaN(seconds)) return '0:00';
    
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

// ===== Изменение скорости =====
function changeSpeed(e) {
    const speed = parseFloat(e.target.value);
    audioElement.playbackRate = speed;
}

// ===== Остановка воспроизведения =====
function stopAudio() {
    if (audioElement) {
        audioElement.pause();
        audioElement.currentTime = 0;
        audioElement.src = '';
        isPlaying = false;
        updatePlayPauseButton();
        document.getElementById('audioPlayer').classList.remove('playing');
        document.getElementById('playerTitle').textContent = 'Название трека';
        document.getElementById('playerTime').textContent = '0:00 / 0:00';
        document.getElementById('progressBar').value = 0;
    }
}

// ===== Статус плеера =====
function showPlayerStatus(message, type) {
    const title = document.getElementById('playerTitle');
    const originalText = title.textContent;
    
    title.textContent = message;
    title.style.color = type === 'error' ? '#ff0055' : '#00ff88';
    
    setTimeout(() => {
        title.textContent = originalText;
        title.style.color = '#fff';
    }, 3000);
}

// ===== Отрисовка волны =====
function drawWaveform(canvas) {
    const waveform = JSON.parse(canvas.dataset.waveform || '[]');
    
    if (!waveform || waveform.length === 0) {
        return;
    }
    
    const ctx = canvas.getContext('2d');
    const width = canvas.offsetWidth;
    const height = canvas.offsetHeight;
    
    // Устанавливаем размер canvas
    canvas.width = width;
    canvas.height = height;
    
    // Очищаем
    ctx.clearRect(0, 0, width, height);
    
    // Настраиваем стиль
    const gradient = ctx.createLinearGradient(0, 0, width, 0);
    gradient.addColorStop(0, '#00d9ff');
    gradient.addColorStop(1, '#00ff88');
    ctx.fillStyle = gradient;
    
    // Рисуем волну
    const barWidth = width / waveform.length;
    const centerY = height / 2;
    
    waveform.forEach((value, index) => {
        const barHeight = value * height * 0.8; // Масштабируем
        const x = index * barWidth;
        const y = centerY - barHeight / 2;
        
        ctx.fillRect(x, y, barWidth - 1, barHeight);
    });
}
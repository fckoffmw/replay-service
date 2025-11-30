const API_BASE = 'http://localhost:8080/api/v1';

// JWT Token Management
const TokenManager = {
    setToken(token) {
        localStorage.setItem('jwt_token', token);
    },
    
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
            const payload = JSON.parse(atob(token.split('.')[1]));
            const exp = payload.exp * 1000;
            return Date.now() < exp;
        } catch (e) {
            return false;
        }
    },
    
    getUserFromToken() {
        const token = this.getToken();
        if (!token) return null;
        
        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            return {
                id: payload.user_id,
                email: payload.email,
                username: payload.username
            };
        } catch (e) {
            return null;
        }
    }
};

// UI Helpers
function showError(message) {
    const errorDiv = document.getElementById('errorMessage');
    errorDiv.textContent = message;
    errorDiv.classList.add('show');
    
    setTimeout(() => {
        errorDiv.classList.remove('show');
    }, 5000);
}

function setButtonLoading(button, loading) {
    if (loading) {
        button.disabled = true;
        button.classList.add('loading');
    } else {
        button.disabled = false;
        button.classList.remove('loading');
    }
}

// Login Handler
if (document.getElementById('loginForm')) {
    document.getElementById('loginForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const email = document.getElementById('email').value.trim();
        const password = document.getElementById('password').value;
        const rememberMe = document.getElementById('rememberMe').checked;
        const submitButton = e.target.querySelector('button[type="submit"]');
        
        if (!email || !password) {
            showError('Заполните все поля');
            return;
        }
        
        setButtonLoading(submitButton, true);
        
        try {
            const response = await fetch(`${API_BASE}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    email,
                    password,
                    remember_me: rememberMe
                })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Ошибка входа');
            }
            
            // Save JWT token
            TokenManager.setToken(data.token);
            
            // Redirect to main page
            window.location.href = 'index.html';
            
        } catch (error) {
            console.error('Login error:', error);
            showError(error.message || 'Ошибка входа. Проверьте email и пароль.');
        } finally {
            setButtonLoading(submitButton, false);
        }
    });
}

// Register Handler
if (document.getElementById('registerForm')) {
    document.getElementById('registerForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const username = document.getElementById('username').value.trim();
        const email = document.getElementById('email').value.trim();
        const password = document.getElementById('password').value;
        const confirmPassword = document.getElementById('confirmPassword').value;
        const submitButton = e.target.querySelector('button[type="submit"]');
        
        // Validation
        if (!username || !email || !password || !confirmPassword) {
            showError('Заполните все поля');
            return;
        }
        
        if (username.length < 3) {
            showError('Имя пользователя должно быть не менее 3 символов');
            return;
        }
        
        if (password.length < 6) {
            showError('Пароль должен быть не менее 6 символов');
            return;
        }
        
        if (password !== confirmPassword) {
            showError('Пароли не совпадают');
            return;
        }
        
        setButtonLoading(submitButton, true);
        
        try {
            const response = await fetch(`${API_BASE}/auth/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username,
                    email,
                    password
                })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Ошибка регистрации');
            }
            
            // Save JWT token
            TokenManager.setToken(data.token);
            
            // Redirect to main page
            window.location.href = 'index.html';
            
        } catch (error) {
            console.error('Registration error:', error);
            showError(error.message || 'Ошибка регистрации. Возможно, email уже используется.');
        } finally {
            setButtonLoading(submitButton, false);
        }
    });
}

// Check if already authenticated
if (window.location.pathname.includes('login.html') || window.location.pathname.includes('register.html')) {
    if (TokenManager.isAuthenticated()) {
        window.location.href = 'index.html';
    }
}

// Export for use in other scripts
window.TokenManager = TokenManager;

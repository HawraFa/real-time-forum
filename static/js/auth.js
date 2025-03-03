import { showHomePage } from './home.js';

export function showLoginForm() {
    const app = document.getElementById('app');
    app.innerHTML = `
        <div class="container">
            <h2>Login</h2>
            <div id="error" class="error" style="display: none;"></div>
            <form id="loginForm" onsubmit="handleLogin(event)">
                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" required>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required>
                </div>
                <button type="submit">Login</button>
            </form>
            <div class="switch-form">
                <p>Don't have an account? <a href="#" onclick="showRegistrationForm()">Register</a></p>
            </div>
        </div>
    `;
}

export function showRegistrationForm() {
    const app = document.getElementById('app');
    app.innerHTML = `
        <div class="container">
            <h2>Register</h2>
            <div id="error" class="error" style="display: none;"></div>
            <form id="registerForm" onsubmit="handleRegistration(event)">
                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" required>
                </div>
                <div class="form-group">
                    <label for="email">Email</label>
                    <input type="email" id="email" name="email" required>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required>
                </div>
                <div class="form-group">
                    <label for="age">Age</label>
                    <input type="number" id="age" name="age" required min="13">
                </div>
                <div class="form-group">
                    <label for="gender">Gender</label>
                    <select id="gender" name="gender" required>
                        <option value="">Select gender</option>
                        <option value="male">Male</option>
                        <option value="female">Female</option>
                        <option value="other">Other</option>
                    </select>
                </div>
                <button type="submit">Register</button>
            </form>
            <div class="switch-form">
                <p>Already have an account? <a href="#" onclick="showLoginForm()">Login</a></p>
            </div>
        </div>
    `;
}

export async function handleLogin(event) {
    event.preventDefault();
    const form = event.target;
    const error = document.getElementById('error');
    error.style.display = 'none';

    try {
        const response = await fetch('http://localhost:8080/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: form.username.value,
                password: form.password.value,
            }),
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Login failed');
        }

        localStorage.setItem('currentUser', JSON.stringify(data));
        showHomePage(data);
    } catch (err) {
        error.textContent = err.message;
        error.style.display = 'block';
    }
}

export async function handleRegistration(event) {
    event.preventDefault();
    const form = event.target;
    const error = document.getElementById('error');
    error.style.display = 'none';

    try {
        const response = await fetch('http://localhost:8080/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: form.username.value,
                email: form.email.value,
                password: form.password.value,
                age: parseInt(form.age.value),
                gender: form.gender.value,
            }),
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Registration failed');
        }

        alert('Registration successful! Please login.');
        showLoginForm();
    } catch (err) {
        error.textContent = err.message;
        error.style.display = 'block';
    }
}

export function handleLogout() {
    localStorage.removeItem('currentUser');
    showLoginForm();
}

// Make functions globally available for onclick handlers
window.showLoginForm = showLoginForm;
window.showRegistrationForm = showRegistrationForm;
window.handleLogin = handleLogin;
window.handleRegistration = handleRegistration;
window.handleLogout = handleLogout; 
// Make functions globally available
window.showLoginPage = showLoginPage;
window.handleLogin = handleLogin;

function showLoginPage() {
    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar()}
        <div class="container">
            <div class="auth-form">
                <h2>Login</h2>
                <div id="error-container" class="error hidden"></div>
                <form id="loginForm" onsubmit="handleLogin(event)">
                    <div class="form-group">
                        <label for="identifier">Username or Email</label>
                        <input type="text" id="identifier" required>
                    </div>
                    <div class="form-group">
                        <label for="password">Password</label>
                        <input type="password" id="password" required>
                    </div>
                    <button type="submit">Login</button>
                </form>
                <p class="auth-switch">
                    Don't have an account? 
                    <a href="#" onclick="showRegistrationPage(); return false;">Register</a>
                </p>
            </div>
        </div>
    `;
}

async function handleLogin(event) {
    event.preventDefault();
    hideError();

    const identifier = document.getElementById('identifier').value;
    const password = document.getElementById('password').value;

    try {
        const response = await fetch('http://localhost:8080/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                identifier: identifier,
                password: password
            })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.message || 'Login failed');
        }

        // Store user data in localStorage
        const userData = {
            id: data.userID,
            username: data.username
        };
        localStorage.setItem('currentUser', JSON.stringify(userData));

        // Redirect to home page
        showHomePage();
    } catch (error) {
        showError(error.message);
    }
} 
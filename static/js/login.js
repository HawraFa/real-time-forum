// Login page functionality
function showLoginPage() {
    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar(null)}
        <div class="container">
            <h2>Login</h2>
            <div id="error-container" class="error hidden"></div>
            <form id="loginForm" onsubmit="handleLogin(event)">
                <div class="form-group">
                    <label for="identifier">Username or Email</label>
                    <input type="text" id="identifier" name="identifier" required>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required>
                </div>
                <button type="submit">Login</button>
            </form>
            <p>Don't have an account? <a href="#" onclick="showRegistrationPage(); return false;">Register here</a></p>
        </div>
    `;
}

async function handleLogin(event) {
    event.preventDefault();
    hideError();

    const identifier = document.getElementById('identifier').value;
    const password = document.getElementById('password').value;

    console.log('Attempting login...');

    try {
        const response = await fetch('http://localhost:8080/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
            },
            body: JSON.stringify({ identifier, password }),
        });

        console.log('Response status:', response.status);

        const data = await response.json();
        console.log('Server response:', data);

        if (!response.ok) {
            throw new Error(data.message || 'Login failed');
        }

        const userData = {
            id: data.userID,
            username: data.username,
        };
        console.log('Storing user data:', userData);
        localStorage.setItem('currentUser', JSON.stringify(userData));

        showHomePage();
    } catch (error) {
        console.error('Login error:', error);
        if (error.name === 'SyntaxError') {
            showError('Server error. Please try again later.');
        } else if (error.message === 'Failed to fetch') {
            showError('Unable to connect to server. Please make sure the server is running on port 8080.');
        } else {
            showError(error.message || 'Login failed. Please try again.');
        }
    }
}

async function handleLogout() {
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        showLoginPage();
        return;
    }

    try {
        console.log('Attempting to logout user:', currentUser);

        // Always clear local storage and show login page first
        localStorage.removeItem('currentUser');
        showLoginPage();

        const response = await fetch('http://localhost:8080/api/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
            },
            body: JSON.stringify({ userId: currentUser.id }),
        });

        console.log('Logout response status:', response.status);

        if (!response.ok) {
            const data = await response.json();
            console.error('Server logout failed:', data.message);
            // Don't show error to user since they're already logged out locally
        }
    } catch (error) {
        console.error('Logout error:', error);
        // Don't show error to user since they're already logged out locally
    }
} 
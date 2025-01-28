// Make functions globally available
window.showRegistrationPage = showRegistrationPage;
window.handleRegistration = handleRegistration;

function showRegistrationPage() {
    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar()}
        <div class="container">
            <div class="auth-form">
                <h2>Register</h2>
                <div id="error-container" class="error hidden"></div>
                <form id="registrationForm" onsubmit="handleRegistration(event)">
                    <div class="form-group">
                        <label for="username">Username</label>
                        <input type="text" id="username" required>
                    </div>
                    <div class="form-group">
                        <label for="email">Email</label>
                        <input type="email" id="email" required>
                    </div>
                    <div class="form-group">
                        <label for="password">Password</label>
                        <input type="password" id="password" required>
                    </div>
                    <div class="form-group">
                        <label for="firstName">First Name</label>
                        <input type="text" id="firstName" required>
                    </div>
                    <div class="form-group">
                        <label for="lastName">Last Name</label>
                        <input type="text" id="lastName" required>
                    </div>
                    <div class="form-group">
                        <label for="age">Age</label>
                        <input type="number" id="age" required min="13">
                    </div>
                    <div class="form-group">
                        <label for="gender">Gender</label>
                        <select id="gender" required>
                            <option value="">Select Gender</option>
                            <option value="male">Male</option>
                            <option value="female">Female</option>
                            <option value="other">Other</option>
                        </select>
                    </div>
                    <button type="submit">Register</button>
                </form>
                <p class="auth-switch">
                    Already have an account? 
                    <a href="#" onclick="showLoginPage(); return false;">Login</a>
                </p>
            </div>
        </div>
    `;
}

async function handleRegistration(event) {
    event.preventDefault();
    hideError();

    const formData = {
        username: document.getElementById('username').value,
        email: document.getElementById('email').value,
        password: document.getElementById('password').value,
        firstName: document.getElementById('firstName').value,
        lastName: document.getElementById('lastName').value,
        age: parseInt(document.getElementById('age').value),
        gender: document.getElementById('gender').value
    };

    try {
        const response = await fetch('http://localhost:8080/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData)
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.message || 'Registration failed');
        }

        // Show success message
        showError('Registration successful! Please login.', 'success-container');
        
        // Redirect to login page after a short delay
        setTimeout(() => {
            showLoginPage();
        }, 2000);
    } catch (error) {
        showError(error.message);
    }
} 
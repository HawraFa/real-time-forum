function showRegistrationPage() {
    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar(null, false)}
        <div class="container">
            <h2>Register</h2>
            <div id="error-container" class="error hidden"></div>
            <form id="registrationForm" onsubmit="handleRegistration(event)">
                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" required>
                </div>
                <div class="form-group">
                    <label for="firstName">First Name</label>
                    <input type="text" id="firstName" name="firstName" required>
                </div>
                <div class="form-group">
                    <label for="lastName">Last Name</label>
                    <input type="text" id="lastName" name="lastName" required>
                </div>
                <div class="form-group">
                    <label for="email">Email</label>
                    <input type="email" id="email" name="email" required>
                </div>
                <div class="form-group">
                    <label for="age">Age</label>
                    <input type="number" id="age" name="age" required min="13" max="120">
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
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required>
                </div>
                <button type="submit">Register</button>
            </form>
            <p>Already have an account? <a href="#" onclick="showLoginPage(); return false;">Login here</a></p>
        </div>
    `;
}

async function handleRegistration(event) {
    event.preventDefault();
    hideError();

    const formData = {
        username: document.getElementById('username').value,
        firstName: document.getElementById('firstName').value,
        lastName: document.getElementById('lastName').value,
        email: document.getElementById('email').value,
        age: parseInt(document.getElementById('age').value),
        gender: document.getElementById('gender').value,
        password: document.getElementById('password').value,
    };

    console.log('Attempting registration...');

    try {
        const response = await fetch('http://localhost:8080/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
            },
            body: JSON.stringify(formData),
        });

        console.log('Response status:', response.status);

        const data = await response.json();
        console.log('Server response:', data);

        if (!response.ok) {
            throw new Error(data.message || 'Registration failed');
        }

        // Show success message and redirect to login
        showLoginPage();
        showError('Registration successful! Please login.', 'error-container');
    } catch (error) {
        console.error('Registration error:', error);
        if (error.name === 'SyntaxError') {
            showError('Server error. Please try again later.');
        } else if (error.message === 'Failed to fetch') {
            showError('Unable to connect to server. Please make sure the server is running on port 8080.');
        } else {
            showError(error.message || 'Registration failed. Please try again.');
        }
    }
} 
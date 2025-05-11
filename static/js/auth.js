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
                    <label for="firstName">First Name</label>
                    <input type="text" id="firstName" name="firstName" required>
                </div>
                <div class="form-group">
                    <label for="lastName">Last Name</label>
                    <input type="text" id="lastName" name="lastName" required>
                </div>

                <div class="form-group">
                    <label for="age">Age</label>
                    <input type="number" id="age" name="age" required>
                </div>

                 <div class="form-group">
                    <label for="gender">Gender</label>
                    <select id="gender">
                        <option value ="female"> Female </option>
                        <option value ="male"> Male </option>

                    </select>
                </div>

                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required>
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
        const response = await fetch('http://localhost:8082/api/login', {
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
        console.log("Login response data:", data);

        // const text = await response.text();   // Instead of response.json()
        // console.log("Raw response text:", text);


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

    // Log the form data
    const formData = {
        username: form.username.value,
        email: form.email.value,
        firstName: form.firstName.value,
        lastName: form.lastName.value,
        password: form.password.value,
        age: parseInt(form.age.value),
        gender: form.gender.value,
    };
    console.log("Registration form data:", formData);

    try {
        const response = await fetch('http://localhost:8082/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData)
        });

        const data = await response.json();
        console.log("Registration response:", data);

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

export function showProfile() {
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    
    if (!currentUser) {
        showLoginForm();
        return;
    }

    console.log("Current user data:", currentUser);

    const app = document.getElementById('app');
    app.innerHTML = `
        <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <div class="profile-menu">
                    <img src="${currentUser.avatar || '/static/images/default.png'}" alt="Profile" class="profile-icon">
                    <span>${currentUser.username}</span>
                </div>
            </div>
        </nav>
        <div class="container">
            <h2>Profile Information</h2>
            <div class="profile-info">
                <div class="profile-field">
                    <label>Username:</label>
                    <span>${currentUser.username}</span>
                </div>
                <div class="profile-field">
                    <label>Email:</label>
                    <span>${currentUser.email || 'N/A'}</span>
                </div>
                <div class="profile-field">
                    <label>First Name:</label>
                    <span>${currentUser.firstName || 'N/A'}</span>
                </div>
                <div class="profile-field">
                    <label>Last Name:</label>
                    <span>${currentUser.lastName || 'N/A'}</span>
                </div>
                <div class="profile-field">
                    <label>Age:</label>
                    <span>${currentUser.age || 'N/A'}</span>
                </div>
                <div class="profile-field">
                    <label>Gender:</label>
                    <span>${currentUser.gender || 'N/A'}</span>
                </div>
            </div>
            <div class="profile-actions">
                <button onclick="handleLogout()">Logout</button>
                <button onclick="showHomePage(${JSON.stringify(currentUser)})">Back to Home</button>
                <button onclick="showEditProfile()">Edit Profile</button>

            </div>
        </div>
    `;
}

function showEditProfile() {
    const user = JSON.parse(localStorage.getItem("currentUser"));
    const app = document.getElementById("app");

    app.innerHTML = `
        <div class="container">
            <h2>Edit Profile</h2>
            <form id="editProfileForm">
                <div class="form-group">
                    <label>First Name:</label>
                    <input type="text" name="firstName" value="${user.firstName || ''}" required>
                </div>
                <div class="form-group">
                    <label>Last Name:</label>
                    <input type="text" name="lastName" value="${user.lastName || ''}" required>
                </div>
                <div class="form-group">
                    <label>Email:</label>
                    <input type="email" name="email" value="${user.email || ''}" required>
                </div>
                <div class="form-group">
                    <label>Age:</label>
                    <input type="number" name="age" value="${user.age || ''}" required>
                </div>
                <div class="form-group">
                    <label>Gender:</label>
                    <select name="gender">
                        <option value="female" ${user.gender === 'female' ? 'selected' : ''}>Female</option>
                        <option value="male" ${user.gender === 'male' ? 'selected' : ''}>Male</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Profile Picture:</label>
                    <input type="file" name="profilePicture" accept="image/*">
                </div>
                <button type="submit">Save Changes</button>
                <button type="button" onclick="showProfile()">Cancel</button>
            </form>
        </div>
    `;

    // Attach submit listener
    document.getElementById("editProfileForm").addEventListener("submit", handleProfileUpdate);
}
async function handleProfileUpdate(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));

    formData.append("id", currentUser.id); // include user ID

    try {
        const response = await fetch("http://localhost:8082/api/profile/update", {
            method: "POST",
            body: formData
        });

        const updatedUser = await response.json();

        if (!response.ok) {
            throw new Error(updatedUser.error || "Profile update failed");
        }

        alert("Profile updated successfully!");
        localStorage.setItem("currentUser", JSON.stringify(updatedUser));
        showProfile();
    } catch (err) {
        alert(err.message);
    }
}

window.showEditProfile = showEditProfile;
// Make functions globally available for onclick handlers
window.showLoginForm = showLoginForm;
window.showRegistrationForm = showRegistrationForm;
window.handleLogin = handleLogin;
window.handleRegistration = handleRegistration;
window.handleLogout = handleLogout;
window.showProfile = showProfile; 
import { showHomePage } from './home.js';

// ---- Login Form ----
export function showLoginForm() {
    const chatWindow = document.querySelector(".chat-window");
    const chatSidebar = document.querySelector(".chat-sidebar");

    if (chatWindow) chatWindow.style.display = "none";
    if (chatSidebar) chatSidebar.style.display = "none";

    const app = document.getElementById('app');
    app.innerHTML = `
        <div class="container">
            <h2>Login</h2>
            <div id="error" class="error" style="display: none;"></div>
            <form id="loginForm" onsubmit="handleLogin(event)">
                <div class="form-group">
                    <label for="usernameOrEmail">Username or Email</label>
                    <input type="text" id="usernameOrEmail" name="usernameOrEmail" required>
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

// ---- Registration Form ----
export function showRegistrationForm() {
    const chatWindow = document.querySelector(".chat-window");
    const chatSidebar = document.querySelector(".chat-sidebar");

    if (chatWindow) chatWindow.style.display = "none";
    if (chatSidebar) chatSidebar.style.display = "none";

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
                    <select id="gender" name="gender">
                        <option value="female"> Female </option>
                        <option value="male"> Male </option>
                    </select>
                </div>

                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required oninput="validatePasswordStrength()">
                    <div id="password-requirements" class="password-requirements">
                        <div class="requirement" id="req-length"><span class="check">✗</span> At least 8 characters</div>
                        <div class="requirement" id="req-uppercase"><span class="check">✗</span> One uppercase letter</div>
                        <div class="requirement" id="req-lowercase"><span class="check">✗</span> One lowercase letter</div>
                        <div class="requirement" id="req-number"><span class="check">✗</span> One number</div>
                        <div class="requirement" id="req-special"><span class="check">✗</span> One special character</div>
                    </div>
                    <div id="password-strength" class="password-strength">
                        <div class="strength-bar"></div>
                    </div>
                </div>
                 <div class="form-group">
                    <label for="confirmPassword">Confirm Password</label>
                    <input type="password" id="confirmPassword" name="confirmPassword" required>
                </div>
                <button type="submit">Register</button>
            </form>
            <div class="switch-form">
                <p>Already have an account? <a href="#" onclick="showLoginForm()">Login</a></p>
            </div>
        </div>
    `;
}

// ---- Handle Login ----
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
                usernameOrEmail: form.usernameOrEmail.value,
                password: form.password.value,
            }),
            credentials: 'include' // Ensure cookie is saved
        });

        const data = await response.json();
        console.log("Login response data:", data);

        if (!response.ok) {
            throw new Error(data.error || 'Login failed');
        }

        localStorage.setItem('currentUser', JSON.stringify(data));

        // Initialize ChatManager and show home page
        showHomePage(data);
        if (!window.chatManager) {
            window.chatManager = new ChatManager();
        } else {
            // If ChatManager exists, reconnect and send online status
            window.chatManager.setupWebSocket();
            window.chatManager.sendStatusUpdate("online");
        }

    } catch (err) {
        error.textContent = err.message;
        error.style.display = 'block';
    }
}

// ---- Handle Registration ----
export async function handleRegistration(event) {
    event.preventDefault();
    const form = event.target;
    const error = document.getElementById('error');
    error.style.display = 'none';

    // Validate password strength
    const password = form.password.value;
    if (!isPasswordStrong(password)) {
        error.textContent = 'Password does not meet strength requirements. Please ensure your password has at least 8 characters, one uppercase letter, one lowercase letter, one number, and one special character.';
        error.style.display = 'block';
        return;
    }

    const confirmPassword = form.confirmPassword.value;
    if (password !== confirmPassword) {
        error.textContent = 'Passwords do not match.';
        error.style.display = 'block';
        return;
    }

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
        const response = await fetch('http://localhost:8080/api/register', {
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

// ---- Password Validation Functions ----
function validatePasswordStrength() {
    const password = document.getElementById('password').value;
    const requirements = {
        length: password.length >= 8,
        uppercase: /[A-Z]/.test(password),
        lowercase: /[a-z]/.test(password),
        number: /[0-9]/.test(password),
        special: /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)
    };

    // Update requirement indicators
    Object.keys(requirements).forEach(req => {
        const element = document.getElementById(`req-${req}`);
        const checkSpan = element.querySelector('.check');
        if (requirements[req]) {
            checkSpan.textContent = '✓';
            checkSpan.style.color = '#22c55e';
            element.style.color = '#22c55e';
        } else {
            checkSpan.textContent = '✗';
            checkSpan.style.color = '#ef4444';
            element.style.color = '#6b7280';
        }
    });

    // Update strength bar
    const strengthBar = document.querySelector('.strength-bar');
    const metRequirements = Object.values(requirements).filter(Boolean).length;
    const strengthPercentage = (metRequirements / 5) * 100;
    
    strengthBar.style.width = `${strengthPercentage}%`;
    
    if (strengthPercentage <= 40) {
        strengthBar.style.backgroundColor = '#ef4444'; // Red
    } else if (strengthPercentage <= 80) {
        strengthBar.style.backgroundColor = '#f59e0b'; // Orange
    } else {
        strengthBar.style.backgroundColor = '#22c55e'; // Green
    }
}

function isPasswordStrong(password) {
    return (
        password.length >= 8 &&
        /[A-Z]/.test(password) &&
        /[a-z]/.test(password) &&
        /[0-9]/.test(password) &&
        /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)
    );
}

// ---- Handle Logout ----
export async function handleLogout() {
    // Get the chat manager instance if it exists
    const chatManager = window.chatManager;
    if (chatManager) {
        // Send offline status before closing
        chatManager.sendStatusUpdate("offline");
        // Close WebSocket connection
        if (chatManager.ws) {
            chatManager.ws.close();
        }
    }

    // Call backend to destroy session and cookie
    try {
        await fetch('http://localhost:8080/api/logout', {
            method: 'POST',
            credentials: 'include'
        });
    } catch (e) {
        // Ignore errors, proceed with logout
    }

    localStorage.removeItem('currentUser');
    localStorage.removeItem('chatNotifications');
    localStorage.removeItem('onlineUsers');
    showLoginForm();
}

// ---- Show Profile ----
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
                <div class="profile-menu" onclick="toggleProfileMenu(event)">
                    <img src="${currentUser.avatar || '/static/images/profile.png'}" alt="Profile" class="profile-icon">
                    <span>${currentUser.username}</span>
                    <div class="profile-dropdown" id="profileDropdown">
                        <a href="#" onclick="showProfile()">My Profile</a>
                        <a href="#" onclick="handleLogout()">Logout</a>
                    </div>
                </div>
            </div>
        </nav>

        <!-- Chat Sidebar -->
        <div class="chat-sidebar">
            <div class="chat-sidebar-header">
                <h2>Messages</h2>
            </div>
            <div class="chat-users-container">
                <ul id="chat-user-list" class="chat-user-list"></ul>
            </div>
        </div>

        <!-- Main Content Area -->
        <div class="main-content" style="margin-left: 280px; padding: 20px;">
            <div class="container">
                <h2>Profile Information</h2>
                <div class="profile-info">
                    <div class="profile-field"><label>Username:</label><span>${currentUser.username}</span></div>
                    <div class="profile-field"><label>Email:</label><span>${currentUser.email || 'N/A'}</span></div>
                    <div class="profile-field"><label>First Name:</label><span>${currentUser.firstName || 'N/A'}</span></div>
                    <div class="profile-field"><label>Last Name:</label><span>${currentUser.lastName || 'N/A'}</span></div>
                    <div class="profile-field"><label>Age:</label><span>${currentUser.age || 'N/A'}</span></div>
                    <div class="profile-field"><label>Gender:</label><span>${currentUser.gender || 'N/A'}</span></div>
                </div>
                <div class="profile-actions">
                    <button onclick="handleLogout()">Logout</button>
                    <button onclick="backToHome()">Back to Home</button>
                    <button onclick="showEditProfile()">Edit Profile</button>
                </div>
            </div>
        </div>
    `;

    // Initialize ChatManager after DOM is updated
    if (!window.chatManager) {
        window.chatManager = new ChatManager();
    } else {
        window.chatManager.loadAllUsers(); // Reload user list
    }
}

// ---- Show Edit Profile ----
function showEditProfile() {
    const user = JSON.parse(localStorage.getItem("currentUser"));
    const app = document.getElementById("app");

    app.innerHTML = `
        <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <div class="profile-menu" onclick="toggleProfileMenu(event)">
                    <img src="${user.avatar || '/static/images/profile.png'}" alt="Profile" class="profile-icon">
                    <span>${user.username}</span>
                    <div class="profile-dropdown" id="profileDropdown">
                        <a href="#" onclick="showProfile()">My Profile</a>
                        <a href="#" onclick="handleLogout()">Logout</a>
                    </div>
                </div>
            </div>
        </nav>

        <!-- Chat Sidebar -->
        <div class="chat-sidebar">
            <div class="chat-sidebar-header">
                <h2>Messages</h2>
            </div>
            <div class="chat-users-container">
                <ul id="chat-user-list" class="chat-user-list"></ul>
            </div>
        </div>

        <!-- Main Content Area -->
        <div class="main-content" style="margin-left: 280px; padding: 20px;">
            <div class="container">
                <h2>Edit Profile</h2>
                <form id="editProfileForm">
                    <div class="form-group"><label>First Name:</label><input type="text" name="firstName" value="${user.firstName || ''}" required></div>
                    <div class="form-group"><label>Last Name:</label><input type="text" name="lastName" value="${user.lastName || ''}" required></div>
                    <div class="form-group"><label>Email:</label><input type="email" name="email" value="${user.email || ''}" required></div>
                    <div class="form-group"><label>Age:</label><input type="number" name="age" value="${user.age || ''}" required></div>
                    <div class="form-group">
                        <label>Gender:</label>
                        <select name="gender">
                            <option value="female" ${user.gender === 'female' ? 'selected' : ''}>Female</option>
                            <option value="male" ${user.gender === 'male' ? 'selected' : ''}>Male</option>
                        </select>
                    </div>
                    <div class="form-group"><label>Profile Picture:</label><input type="file" name="profilePicture" accept="image/*"></div>
                    <button type="submit">Save Changes</button>
                    <button type="button" onclick="showProfile()">Cancel</button>
                </form>
            </div>
        </div>
    `;

    document.getElementById("editProfileForm").addEventListener("submit", handleProfileUpdate);

    // Initialize ChatManager after DOM is updated
    if (!window.chatManager) {
        window.chatManager = new ChatManager();
    } else {
        window.chatManager.loadAllUsers(); // Reload user list
    }
}

// ---- Handle Profile Update ----
async function handleProfileUpdate(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));

    formData.append("id", String(currentUser.id)); // Always send ID as string
    formData.append("age", String(form.age.value)); // Send age as string

    try {
        const response = await fetch("http://localhost:8080/api/profile/update", {
            method: "POST",
            body: formData
        });

        const rawText = await response.text();
        console.log("🧾 Raw server response:", rawText);

        if (!response.ok) {
            throw new Error(rawText);
        }

        const updatedUser = JSON.parse(rawText);
        localStorage.setItem("currentUser", JSON.stringify(updatedUser));
        showProfile();
    } catch (err) {
        alert("⚠️ Update failed: " + err.message);
    }
}

// ---- Back to Home ----
export function backToHome() {
    const raw = localStorage.getItem("currentUser");
    if (!raw) {
        alert("⚠️ No user found in localStorage.");
        return;
    }

    try {
        const currentUser = JSON.parse(raw);
        console.log("Navigating back to home with user:", currentUser);
        showHomePage(currentUser);
    } catch (err) {
        console.error("⚠️ Failed to parse currentUser:", err.message);
        alert("⚠️ Failed to parse user info. Try logging in again.");
        handleLogout();
    }
}

setInterval(async () => {
    const res = await fetch("http://localhost:8080/api/session", {
      credentials: "include"
    });
  
    if (!res.ok) {
      localStorage.clear();
      handleLogout(); // or showLoginForm();
    }
  }, 60_000); // every 1 minute
  

window.backToHome = backToHome; // Make it global
window.showEditProfile = showEditProfile;
// Make functions globally available for onclick handlers
window.showLoginForm = showLoginForm;
window.showRegistrationForm = showRegistrationForm;
window.handleLogin = handleLogin;
window.handleRegistration = handleRegistration;
window.handleLogout = handleLogout;
window.showProfile = showProfile;
window.validatePasswordStrength = validatePasswordStrength; 
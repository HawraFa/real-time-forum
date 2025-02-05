// Utility functions shared across the application

// Function to create the navigation bar
function createNavbar(currentUser = null) {
    return `
        <nav class="navbar">
            <div class="nav-container">
                <a href="#" class="nav-brand" onclick="showHomePage(); return false;">Forum</a>
                <div class="nav-links">
                    ${currentUser 
                        ? `<button class="nav-button" onclick="showPosts()">Posts</button>
                           <button class="nav-button" onclick="showUserProfile(${currentUser.id})">Profile</button>
                           <button class="nav-button" onclick="handleLogout()">Logout</button>`
                        : `<button class="nav-button" onclick="showLoginPage()">Login</button>
                           <button class="nav-button" onclick="showRegistrationPage()">Register</button>`
                    }
                </div>
            </div>
        </nav>
    `;
}

// Function to show error message
function showError(message, containerId = 'error-container') {
    const errorContainer = document.getElementById(containerId);
    if (errorContainer) {
        errorContainer.textContent = message;
        errorContainer.classList.remove('hidden');
    }
}

// Function to hide error message
function hideError(containerId = 'error-container') {
    const errorContainer = document.getElementById(containerId);
    if (errorContainer) {
        errorContainer.classList.add('hidden');
    }
} 
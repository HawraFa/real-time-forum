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

// Add this to your existing navbar.js file
function handleLogout() {
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        return;
    }

    fetch('http://localhost:8080/api/logout', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            userId: currentUser.id
        })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Logout failed');
        }
        return response.json();
    })
    .then(() => {
        // Clear local storage
        localStorage.removeItem('currentUser');
        // Redirect to login page
        showLoginPage();
    })
    .catch(error => {
        console.error('Error during logout:', error);
        // Still clear local storage and redirect even if server request fails
        localStorage.removeItem('currentUser');
        showLoginPage();
    });
}

// Make sure handleLogout is available globally
window.handleLogout = handleLogout; 
export function showHomePage(user) {
    const app = document.getElementById('app');
    app.innerHTML = `
        <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <div class="profile-menu" onclick="toggleProfileMenu()">
                    <img src="/static/images/default-avatar.png" alt="Profile" class="profile-icon">
                    <span>${user.username}</span>
                    <div class="profile-dropdown" id="profileDropdown">
                        <a href="#" onclick="showProfile()">My Profile</a>
                        <a href="#" onclick="handleLogout()">Logout</a>
                    </div>
                </div>
            </div>
        </nav>
        <div class="container home-container">
            <h2>Welcome, ${user.username}!</h2>
            <div class="welcome-message">
                <p>Welcome to our Forum! Here you can:</p>
                <ul>
                    <li>Create and participate in discussions</li>
                    <li>Share your thoughts and ideas</li>
                    <li>Connect with other users</li>
                </ul>
            </div>
            <div class="action-buttons">
                <button onclick="showCreatePost()" class="primary-button">Create New Post</button>
                <button onclick="showAllPosts()" class="secondary-button">View All Posts</button>
            </div>
        </div>
    `;
}

// Make the function available globally
window.toggleProfileMenu = function() {
    const dropdown = document.getElementById('profileDropdown');
    dropdown.classList.toggle('show');
};

// Close dropdown when clicking outside
document.addEventListener('click', function(event) {
    if (!event.target.matches('.profile-menu') && !event.target.matches('.profile-icon')) {
        const dropdowns = document.getElementsByClassName('profile-dropdown');
        for (const dropdown of dropdowns) {
            if (dropdown.classList.contains('show')) {
                dropdown.classList.remove('show');
            }
        }
    }
}); 
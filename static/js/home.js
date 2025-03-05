export function showHomePage(user) {
    const app = document.getElementById('app');
    app.innerHTML = `
         <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <div class="profile-menu">
                    <img src="/static/images/profile.jpg" alt="Profile" class="profile-icon">
                    <span>${user.username}</span>
                    <div class="profile-dropdown" id="profileDropdown">
                        <a href="#" onclick="showProfile(); return false;">My Profile</a>
                        <a href="#" onclick="handleLogout(); return false;">Logout</a>
                    </div>
                </div>
            </div>
        </nav>
        <div class="container home-container">
            <!-- rest of your home page content -->
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
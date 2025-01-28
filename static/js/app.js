// Function to show the home page
function showHomePage() {
    console.log('Showing home page');
    const currentUserStr = localStorage.getItem('currentUser');
    
    // If no user is logged in, redirect to login page
    if (!currentUserStr) {
        console.log('No user found, showing login page');
        showLoginPage();
        return;
    }
    
    try {
        const currentUser = JSON.parse(currentUserStr);
        console.log('User found:', currentUser);
        
        if (!currentUser || !currentUser.id) {
            console.log('Invalid user data, showing login page');
            localStorage.removeItem('currentUser');
            showLoginPage();
            return;
        }

        const app = document.getElementById('app');
        app.innerHTML = `
            ${createNavbar(currentUser)}
            <div class="container">
                <div class="posts-header">
                    <h2>Recent Posts</h2>
                    <div class="posts-controls">
                        <div class="categories-filter">
                            <!-- Categories will be loaded here -->
                        </div>
                        <button onclick="showNewPostForm()" class="new-post-button">
                            <i class="fas fa-plus"></i> New Post
                        </button>
                    </div>
                </div>
                <div id="posts-container" class="posts-grid">
                    <!-- Posts will be loaded here -->
                </div>
            </div>
            <nav class="bottom-nav">
                <div class="nav-item active" onclick="showHomePage()">
                    <i class="fas fa-home"></i>
                    <span>Home</span>
                </div>
                <div class="nav-item" onclick="showCategories()">
                    <i class="fas fa-list"></i>
                    <span>Categories</span>
                </div>
                <div class="nav-item" onclick="showNewPostForm()">
                    <i class="fas fa-plus"></i>
                    <span>New Post</span>
                </div>
                <div class="nav-item" onclick="showMessages()">
                    <i class="fas fa-envelope"></i>
                    <span>Messages</span>
                </div>
                <div class="nav-item" onclick="showProfile(${currentUser.id})">
                    <i class="fas fa-user"></i>
                    <span>Profile</span>
                </div>
            </nav>
        `;

        loadCategories();
        loadPosts();
    } catch (error) {
        console.error('Error parsing user data:', error);
        localStorage.removeItem('currentUser');
        showLoginPage();
    }
}

// Initialize the application
function init() {
    console.log('Initializing application');
    const currentUserStr = localStorage.getItem('currentUser');
    
    if (!currentUserStr) {
        console.log('No user found, showing login page');
        showLoginPage();
        return;
    }

    try {
        const currentUser = JSON.parse(currentUserStr);
        if (!currentUser || !currentUser.id) {
            throw new Error('Invalid user data');
        }
        console.log('User found, showing home page');
        showHomePage();
    } catch (error) {
        console.error('Error parsing user data:', error);
        localStorage.removeItem('currentUser');
        showLoginPage();
    }
}

// Export the init function to make it globally available
window.init = init;

// Add console logging to check script loading
console.log('App.js loaded'); 
import { ChatManager } from "./chat.js";
import { reactToPost } from './reactions.js';
import { showPostDetails } from "./post.js"; 

export async function showHomePage(user) {
    const app = document.getElementById("app");
    app.innerHTML = `
        <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <button class="theme-toggle" onclick="toggleTheme()">
                    <svg viewBox="0 0 24 24" width="24" height="24">
                        <path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.06-.9-.1-1.36-.1z"/>
                    </svg>
                </button>
                <div class="profile-menu" onclick="toggleProfileMenu(event)">
                    <img src="${user.avatar || '/static/images/profile.png'}" alt="Profile" class="profile-icon">
                    <span>${user.username}</span>
                    <div class="profile-dropdown" id="profileDropdown">
                        <a href="#" onclick="showProfile()">My Profile</a>
                        <a href="#" onclick="handleLogout()">Logout</a>
                    </div>
                </div>
                <img src="/static/images/chat.png" alt="Forum Icon" class="nav-icon">
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
                <h2 class="welcome-message"> Welcome, ${user.username}!</h2>
                
                <!-- Action Buttons at the top -->
                <div class="action-buttons" style="margin-bottom: 20px;">
                    <button onclick="showFilterPage()" class="filter-post-button">
                        <i class="fas fa-filter"></i> Filter Posts
                    </button>
                    <button id="my-posts-btn" class="my-posts-button">
                        <i class="fas fa-user"></i> My Posts
                    </button>
                    <button onclick="showCreatePost()" class="create-post-button">
                        <i class="fas fa-plus"></i> Create Post
                    </button>
                </div>
                
                <div id="posts-container"></div>
            </div>
        </div>
    `;

    // Initialize ChatManager after DOM is updated
    if (!window.chatManager) {
        window.chatManager = new ChatManager();
    } else {
        window.chatManager.loadAllUsers(); // Reload user list
    }

    // Load posts
    showAllPosts();

    document.getElementById("my-posts-btn")?.addEventListener("click", async () => {
        const currentUser = JSON.parse(localStorage.getItem("currentUser"));
        if (!currentUser) {
            alert("You must be logged in to view your posts.");
            return;
        }

        const app = document.getElementById("app");
        app.innerHTML = `
            <nav class="navbar">
                <div class="nav-left">
                    <h1>Forum</h1>
                </div>
                <div class="nav-right">
                    <button class="theme-toggle" onclick="toggleTheme()">
                        <svg viewBox="0 0 24 24" width="24" height="24">
                            <path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.06-.9-.1-1.36-.1z"/>
                        </svg>
                    </button>
                    <div class="profile-menu" onclick="toggleProfileMenu(event)">
                        <img src="${currentUser.avatar || '/static/images/profile.png'}" alt="Profile" class="profile-icon">
                        <span>${currentUser.username}
                        <div class="profile-dropdown" id="profileDropdown">
                            <a href="#" onclick="showProfile()">My Profile</a>
                            <a href="#" onclick="handleLogout()">Logout</a>
                        </div>
                    </div>
                    <img src="/static/images/chat.png" alt="Forum Icon" class="nav-icon">
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
                    <h2>My Posts</h2>
                    <button onclick="backToHome()" class="back-button" style="margin-bottom: 20px;">
                        <i class="fas fa-arrow-left"></i> Back to Home
                    </button>
                    <div id="posts-container"></div>
                </div>
            </div>
        `;

        // Initialize ChatManager if not already initialized
        if (!window.chatManager) {
            window.chatManager = new ChatManager();
        } else {
            window.chatManager.loadAllUsers(); // Reload user list
        }

        const response = await fetch(`http://localhost:8080/api/posts?user_id=${currentUser.id}`);
        const data = await response.json();

        const postsContainer = document.getElementById("posts-container");
        if (!Array.isArray(data) || data.length === 0) {
            postsContainer.innerHTML = "<p>No posts found.</p>";
            return;
        }

        renderPosts(data);
    });
    
    // Show all posts
    document.querySelector(".filter-post-button").addEventListener("click", () => {
        showAllPosts(); // No filter
    });
}

// Profile dropdown toggle (from both versions)
window.toggleProfileMenu = function(event) {
    event.stopPropagation();
    const dropdown = document.getElementById('profileDropdown');
    if (dropdown) {
        dropdown.classList.toggle('show');
    }
};

// Close profile dropdown when clicking outside (from both versions)
document.addEventListener('click', function(event) {
    const dropdown = document.getElementById('profileDropdown');
    if (dropdown && dropdown.classList.contains('show')) {
        dropdown.classList.remove('show');
    }
});

// Show create post form (merged with dynamic category loading from friend)
export async function showCreatePost() {
    const app = document.getElementById("app");
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));

    // Fetch categories from backend (friend's version)
    let categoryOptions = "<option value=''>Loading...</option>";
    try {
        const res = await fetch("http://localhost:8080/api/categories");
        const categories = await res.json();
        categoryOptions = categories.map(cat => `
            <option value="${cat.id}">${cat.name}</option>
        `).join("");
    } catch (e) {
        categoryOptions = "<option disabled>Error loading categories</option>";
    }

    app.innerHTML = `
        <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <button class="theme-toggle" onclick="toggleTheme()">
                    <svg viewBox="0 0 24 24" width="24" height="24">
                        <path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.06-.9-.1-1.36-.1z"/>
                    </svg>
                </button>
                <div class="profile-menu" onclick="toggleProfileMenu(event)">
                    <img src="${currentUser.avatar || '/static/images/profile.png'}" alt="Profile" class="profile-icon">
                    <span>${currentUser.username}</span>
                    <div class="profile-dropdown" id="profileDropdown">
                        <a href="#" onclick="showProfile()">My Profile</a>
                        <a href="#" onclick="handleLogout()">Logout</a>
                    </div>
                </div>
                <img src="/static/images/icon.jpeg" alt="Forum Icon" class="nav-icon">
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
                <h2>Create New Post</h2>
                <button type="button" onclick="backToHome()" class="back-button" style="margin-bottom: 20px;">
                    <i class="fas fa-arrow-left"></i> Back to Home
                </button>
                <form id="createPostForm">
                    <div class="form-group">
                        <label for="title">Title:</label>
                        <input type="text" id="title" name="title" required>
                    </div>
                    <div class="form-group">
                        <label for="content">Content:</label>
                        <textarea id="content" name="content" required></textarea>
                    </div>
                    <div class="form-group">
                        <label for="category">Category:</label>
                        <select id="category" name="category" required>
                            ${categoryOptions}
                        </select>
                    </div>
                    <button type="submit">Post</button>
                </form>
            </div>
        </div>
    `;

    document.getElementById("createPostForm").addEventListener("submit", handleCreatePost);

    // Initialize ChatManager after DOM is updated
    if (!window.chatManager) {
        window.chatManager = new ChatManager();
    } else {
        window.chatManager.loadAllUsers(); // Reload user list
    }
}

export function renderPosts(posts, containerId = "posts-container") {
    const postsContainer = document.getElementById(containerId);
    if (!postsContainer) {
        console.error("❌ Posts container not found");
        return;
    }

    if (!Array.isArray(posts)) {
        console.warn("❗ Posts is not an array:", posts);
        postsContainer.innerHTML = "<p>⚠️ Failed to load posts.</p>";
        return;
    }

    let postHTML = posts.map(post => `
        <div class="post" data-post-id="${post.id}" style="cursor:pointer;">
            <h3>${post.title}</h3>
            <p>${post.content.substring(0, 100)}...</p>
            <div class="post-footer">
                <img src="${post.avatar || '/static/images/profile.png'}" 
                     class="avatar-icon" 
                     style="width: 40px; height: 40px; border-radius: 50%; object-fit: cover;">
                <small>
                    <strong>${post.username}</strong> — ${new Date(post.created_at).toLocaleString()}
                </small>
            </div>
            <div class="reactions">
                <button class="like-button" data-post-id="${post.id}" data-type="like">
                    <img src="/static/images/like.png" alt="Like">
                    <span id="likes-${post.id}">${post.likes_count}</span>
                </button>
                <button class="dislike-button" data-post-id="${post.id}" data-type="dislike">
                    <img src="/static/images/dislike.png" alt="Dislike">
                    <span id="dislikes-${post.id}">${post.dislikes_count}</span>
                </button>
            </div>
        </div>
    `).join("");

    postsContainer.innerHTML = postHTML || "<p>No posts found.</p>";

    // Attach reaction handlers
    document.querySelectorAll(".like-button, .dislike-button").forEach(btn => {
        btn.addEventListener("click", (e) => {
            e.stopPropagation();
            const postId = parseInt(btn.getAttribute("data-post-id"));
            const type = btn.getAttribute("data-type");
            reactToPost(postId, type);
        });
    });

    // Make whole post clickable
    document.querySelectorAll(".post").forEach(postEl => {
        postEl.addEventListener("click", () => {
            const postId = postEl.getAttribute("data-post-id");
            showPostDetails(postId);
        });
    });
}

window.showMyPosts = async function () {
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));
    if (!currentUser) return;

    try {
        const response = await fetch(`http://localhost:8080/api/posts?user_id=${currentUser.id}`);
        const posts = await response.json();

        const postsContainer = document.getElementById("posts-container");
        if (!postsContainer) return;

        if (!posts.length) {
            postsContainer.innerHTML = "<p>You haven't posted anything yet.</p>";
            return;
        }

        postsContainer.innerHTML = posts.map(post => `
            <div class="post" data-post-id="${post.id}" style="cursor:pointer;">
                <h3>${post.title}</h3>
                <p>${post.content.substring(0, 100)}...</p>
                <div class="post-footer">
                    <img src="${post.avatar || '/static/images/profile.png'}"
                    class="avatar-icon"
                    style="width: 40px; height: 40px; border-radius: 50%; object-fit: cover;">
                    <small>
                        <strong>${post.username}</strong> — ${new Date(post.created_at).toLocaleString()}
                    </small>
                </div>
                <div class="reactions">
                    <button class="like-button" data-post-id="${post.id}" data-type="like">
                        <img src="/static/images/like.png" alt="Like">
                        <span id="likes-${post.id}">${post.likes_count}</span>
                    </button>
                    <button class="dislike-button" data-post-id="${post.id}" data-type="dislike">
                        <img src="/static/images/dislike.png" alt="Dislike">
                        <span id="dislikes-${post.id}">${post.dislikes_count}</span>
                    </button>
                </div>
            </div>
        `).join("");

        document.querySelectorAll(".like-button, .dislike-button").forEach(btn => {
            btn.addEventListener("click", (e) => {
                e.stopPropagation();
                const postId = parseInt(btn.getAttribute("data-post-id"));
                const type = btn.getAttribute("data-type");
                reactToPost(postId, type);
            });
        });

        document.querySelectorAll(".post").forEach(postDiv => {
            postDiv.addEventListener("click", (e) => {
                const postId = postDiv.getAttribute("data-post-id");
                const tag = e.target.tagName.toLowerCase();
                const classList = e.target.classList;

                if (
                    !classList.contains("like-button") &&
                    !classList.contains("dislike-button") &&
                    tag !== "img" &&
                    tag !== "span"
                ) {
                    showPostDetails(postId);
                }
            });
        });

    } catch (error) {
        console.error("Failed to load user posts:", error);
    }
};

// Handle post creation (merged, friend's version uses consistent naming and parseInt for IDs)
async function handleCreatePost(event) {
    event.preventDefault();
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));
    const form = event.target;

    const postData = {
        title: form.title.value,
        content: form.content.value,
        category_ids: [parseInt(form.category.value)],
        author_id: parseInt(currentUser.id)
    };

    try {
        const response = await fetch("http://localhost:8080/api/posts/create", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(postData)
        });

        if (!response.ok) {
            throw new Error("Failed to create post");
        }

        alert("Post created successfully!");
        showHomePage(currentUser);
    } catch (error) {
        alert(error.message);
    }
}

// Show all posts or only current user's posts
async function showAllPosts(userId = null) {
    try {
        let url = "http://localhost:8080/api/posts";
        if (userId !== null) {
            url += `?user_id=${userId}`;
        }

        const response = await fetch(url);
        const data = await response.json();

        if (!response.ok) {
            console.error("❌ Server error:", data.error);
            return;
        }

        if (!Array.isArray(data)) {
            console.warn("❗ Expected array, got:", data);
            return;
        }

        const posts = data;
        const postsContainer = document.getElementById("posts-container");

        if (!postsContainer) {
            console.error("Posts container not found");
            return;
        }

        let postHTML = posts.map(post => `
            <div class="post" data-post-id="${post.id}" style="cursor:pointer;">
                <h3>${post.title}</h3>
                <p>${post.content.substring(0, 100)}...</p>
                <div class="post-footer">
                    <img src="${post.avatar || '/static/images/profile.png'}" 
                    class="avatar-icon" 
                    style="width: 40px; height: 40px; border-radius: 50%; object-fit: cover;">
                    <small>
                        <strong>${post.username}</strong> — ${new Date(post.created_at).toLocaleString()}
                    </small>
                </div>

                <div class="reactions">
                    <button class="like-button" data-post-id="${post.id}" data-type="like">
                        <img src="/static/images/like.png" alt="Like">
                        <span id="likes-${post.id}">${post.likes_count}</span>
                    </button>
                    <button class="dislike-button" data-post-id="${post.id}" data-type="dislike">
                        <img src="/static/images/dislike.png" alt="Dislike">
                        <span id="dislikes-${post.id}">${post.dislikes_count}</span>
                    </button>
                </div>
            </div>
        `).join("");

        postsContainer.innerHTML = postHTML || "<p>No posts found.</p>";

        // Attach reaction buttons
        document.querySelectorAll(".like-button, .dislike-button").forEach(btn => {
            btn.addEventListener("click", (e) => {
                e.stopPropagation();
                const postId = parseInt(btn.getAttribute("data-post-id"));
                const type = btn.getAttribute("data-type");
                reactToPost(postId, type);
            });
        });

        // Attach click to open post detail
        document.querySelectorAll(".post").forEach(postDiv => {
            postDiv.addEventListener("click", (e) => {
                const postId = postDiv.getAttribute("data-post-id");
                const tag = e.target.tagName.toLowerCase();
                const classList = e.target.classList;

                if (
                    !classList.contains("like-button") &&
                    !classList.contains("dislike-button") &&
                    tag !== "img" &&
                    tag !== "span"
                ) {
                    showPostDetails(postId);
                }
            });
        });

    } catch (error) {
        console.error("Error loading posts:", error);
    }
}

window.showFilterPage = function () {
    const app = document.getElementById("app");
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));

    app.innerHTML = `
        <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <button class="theme-toggle" onclick="toggleTheme()">
                    <svg viewBox="0 0 24 24" width="24" height="24">
                        <path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.06-.9-.1-1.36-.1z"/>
                    </svg>
                </button>
                <div class="profile-menu" onclick="toggleProfileMenu(event)">
                    <img src="${currentUser.avatar || '/static/images/profile.png'}" alt="Profile" class="profile-icon">
                    <span>${currentUser.username}
                    <div class="profile-dropdown" id="profileDropdown">
                        <a href="#" onclick="showProfile()">My Profile</a>
                        <a href="#" onclick="handleLogout()">Logout</a>
                    </div>
                </div>
                <img src="/static/images/chat.png" alt="Forum Icon" class="nav-icon">
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
                <h2>Filter Posts by Category</h2>
                <div class="filter-controls">
                    <label class="select-label">Select Categories</label>
                    <select id="categoryFilter" multiple>
                        <!-- Categories will be loaded here -->
                    </select>
                    <div class="filter-buttons">
                        <button onclick="filterPosts()" class="primary-button">
                            <i class="fas fa-filter"></i> Apply Filter
                        </button>
                        <button onclick="backToHome()" class="secondary-button">
                            <i class="fas fa-arrow-left"></i> Back to Home
                        </button>
                    </div>
                </div>
                <div id="filtered-posts-container"></div>
            </div>
        </div>
    `;

    // Initialize ChatManager if not already initialized
    if (!window.chatManager) {
        window.chatManager = new ChatManager();
    } else {
        window.chatManager.loadAllUsers(); // Reload user list
    }

    loadCategories();
};

// Navigate back to home page
window.backToHome = function() {
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));
    showHomePage(currentUser);
};

// Logout function 
window.handleLogout = function() {
    localStorage.removeItem("currentUser");
    window.location.href = "/login.html";  
};

// Show user profile (placeholder)
window.showProfile = function() {
    alert("Profile page coming soon!");
};

// Initial call to show home page if needed (you can remove this if you call it from outside)
document.addEventListener("DOMContentLoaded", () => {
    const user = JSON.parse(localStorage.getItem("currentUser"));
    if (user) {
        showHomePage(user);
    }
});

window.showCreatePost = showCreatePost;
window.showAllPosts = showAllPosts;
window.renderPosts = renderPosts;



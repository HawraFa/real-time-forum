import { ChatManager } from "./chat.js";
import { loadComments, submitComment } from './comments.js';
import { reactToPost } from './reactions.js';

export async function showHomePage(user) {
    const chatSidebar = document.querySelector(".chat-sidebar");
    const chatWindow = document.querySelector(".chat-window");
    
    if (chatSidebar) {
        chatSidebar.style.display = "block";
    }
    if (chatWindow) {
        chatWindow.style.display = "none"; // Hide chat window by default
    }
    
    const app = document.getElementById('app');

    // 🔄 Fetch fresh user data from backend (friend's logic)
    let freshUser = user;
    try {
        const res = await fetch(`http://localhost:8080/api/user/${user.id}`);
        if (res.ok) {
            freshUser = await res.json();
        } else {
            console.warn("Could not refresh user info. Using cached.");
        }
    } catch (err) {
        // Ignore fetch errors, use cached user
    }

    const avatar = freshUser.avatar || "/static/images/profile.png";

    app.innerHTML = `
        <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <div class="profile-menu" onclick="toggleProfileMenu(event)">
                    <img src="${avatar}" alt="Profile" class="profile-icon">
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
        </div>
    `;

    // Initialize ChatManager after DOM is updated
    if (!window.chatManager) {
        window.chatManager = new ChatManager();
    } else {
        window.chatManager.loadAllUsers(); // Reload user list
    }
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
        <div class="container">
            <h2>Create New Post</h2>
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
                <button type="button" onclick="backToHome()">Cancel</button>
            </form>
        </div>
    `;

    document.getElementById("createPostForm").addEventListener("submit", handleCreatePost);
}
window.showCreatePost = showCreatePost;

// Handle post creation (merged, friend’s version uses consistent naming and parseInt for IDs)
async function handleCreatePost(event) {
    event.preventDefault();
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));
    const form = event.target;

    const postData = {
        title: form.title.value,
        content: form.content.value,
        category_id: parseInt(form.category.value),
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

// Show all posts (friend’s detailed version with reactions and comments)
async function showAllPosts() {
    const app = document.getElementById("app");

    try {
        const response = await fetch("http://localhost:8080/api/posts");
        const data = await response.json();

        if (!response.ok) {
            console.error("❌ Server error:", data.error);
            app.innerHTML = `<p class="error">⚠️ ${data.error || "Failed to load posts"}</p>`;
            return;
        }

        if (!Array.isArray(data)) {
            console.warn("❗ Expected array, got:", data);
            app.innerHTML = `<p class="error">⚠️ Unexpected response format</p>`;
            return;
        }

        const posts = data;

        let postHTML = posts.map(post => `
            <div class="post">
                <h3>${post.title}</h3>
                <p>${post.content}</p>
                <div class="post-footer">
                    <img src="${post.avatar || '/static/images/profile.png'}" 
                    class="avatar-icon" 
                    style="width: 40px; height: 40px; border-radius: 50%; object-fit: cover;">
                    <small>
                        <strong>${post.username}</strong> — ${new Date(post.created_at).toLocaleString()}
                    </small>
                </div>

                <div class="reactions">
                    <button onclick="reactToPost(${post.id}, 'like')">👍 <span id="likes-${post.id}">${post.likes_count}</span></button>
                    <button onclick="reactToPost(${post.id}, 'dislike')">👎 <span id="dislikes-${post.id}">${post.dislikes_count}</span></button>
                </div>

                <div class="comments-section">
                    <div id="comments-for-${post.id}"></div>
                    <form onsubmit="submitComment(event, ${post.id})">
                        <input id="comment-input-${post.id}" type="text" placeholder="Write a comment..." required>
                        <button type="submit">Send</button>
                    </form>
                </div>

                <hr>
            </div>
        `).join("");

        app.innerHTML = `
            <div class="container">
                <h2>All Posts</h2>
                ${postHTML || "<p>No posts found.</p>"}
                <button onclick="backToHome()">Back to Home</button>
            </div>
        `;

        setTimeout(() => {
            posts.forEach(post => {
                const el = document.getElementById(`comments-for-${post.id}`);
                if (el) {
                    loadComments(post.id);
                } else {
                    console.warn("No container found for post", post.id);
                }
            });
        }, 100);

    } catch (error) {
        console.error("Failed to load posts:", error);
        app.innerHTML = `<p class="error">Failed to load posts</p>`;
    }
}
window.showAllPosts = showAllPosts;

// Navigate back to home page
window.backToHome = function() {
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));
    showHomePage(currentUser);
};

// Logout function (you can add your logout logic here)
window.handleLogout = function() {
    // Clear session/local storage, redirect to login page, etc.
    localStorage.removeItem("currentUser");
    window.location.href = "/login.html";  // Adjust to your login page path
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

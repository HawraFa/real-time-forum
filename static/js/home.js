export function showHomePage(user) {
    const app = document.getElementById('app');
    app.innerHTML = `
        <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <div class="profile-menu" onclick="toggleProfileMenu(event)">
                    <img src="/static/images/profile.png" alt="Profile" class="profile-icon">
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

// Update the toggle function to prevent event bubbling
window.toggleProfileMenu = function(event) {
    event.stopPropagation();
    const dropdown = document.getElementById('profileDropdown');
    if (dropdown) {
        dropdown.classList.toggle('show');
    }
};

// Update the click handler to properly close the dropdown
document.addEventListener('click', function(event) {
    const dropdown = document.getElementById('profileDropdown');
    if (dropdown && dropdown.classList.contains('show')) {
        dropdown.classList.remove('show');
    }
}); 

export function showCreatePost() {
    const app = document.getElementById("app");
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));

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
                        <option value="1">General</option>
                        <option value="2">Announcements</option>
                        <option value="3">Support</option>
                        <!-- You can load categories dynamically later -->
                    </select>
                </div>
                <button type="submit">Post</button>
                <button type="button" onclick="showHomePage(currentUser)">Cancel</button>
            </form>
        </div>
    `;

    document.getElementById("createPostForm").addEventListener("submit", handleCreatePost);
}
async function handleCreatePost(event) {
    event.preventDefault();
    const currentUser = JSON.parse(localStorage.getItem("currentUser"));
    const form = event.target;

    const postData = {
        title: form.title.value,
        content: form.content.value,
        categoryId: parseInt(form.category.value),  // ✅ camelCase to match Go
        authorId: currentUser.id                    // ✅ camelCase to match Go
    };
    

    try {
        const response = await fetch("http://localhost:8082/api/posts/create", {
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

window.showCreatePost = showCreatePost;

window.showAllPosts = showAllPosts;

async function showAllPosts() {
    const app = document.getElementById("app");

    try {
        const response = await fetch("http://localhost:8082/api/posts");
        const posts = await response.json();

        console.log("Received posts:", posts); // 👈 Add this

        let postHTML = posts.map(post => `
            <div class="post">
                <h3>${post.title}</h3>
                <p>${post.content}</p>
                <div class="post-footer">
                    <img src="${post.avatar || '/static/images/profile.png'}" class="avatar-icon">
                    <small>
                        <strong>${post.username}</strong> — ${new Date(post.createdAt).toLocaleString()}
                    </small>
                </div>
                <hr>
            </div>
        `).join("");

        app.innerHTML = `
            <div class="container">
                <h2>All Posts</h2>
                ${postHTML || "<p>No posts found.</p>"}
                <button onclick="showHomePage(JSON.parse(localStorage.getItem('currentUser')))" class="primary-button">Back to Home</button>
            </div>
        `;
    } catch (err) {
        app.innerHTML = `<p>Failed to load posts: ${err.message}</p>`;
    }
}



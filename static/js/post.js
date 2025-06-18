import { loadComments } from './comments.js';
import { reactToPost } from './reactions.js';

// Helper function to render content with images
function renderContentWithImages(content) {
    // Simple regex to detect markdown-style image syntax: ![alt](url)
    const imageRegex = /!\[([^\]]*)\]\(([^)]+)\)/g;
    let lastIndex = 0;
    let result = '';
    let match;

    while ((match = imageRegex.exec(content)) !== null) {
        // Add text before the image
        result += content.slice(lastIndex, match.index);
        // Add the image
        result += `<img src="${match[2]}" alt="${match[1]}" style="max-width: 100%; height: auto; border-radius: 8px; margin: 10px 0;">`;
        lastIndex = match.index + match[0].length;
    }

    // Add remaining text
    result += content.slice(lastIndex);
    return result;
}

export async function showPostDetails(postId) {
  const app = document.getElementById("app");
  const currentUser = JSON.parse(localStorage.getItem("currentUser"));

  try {
    const res = await fetch(`http://localhost:8080/api/posts/${postId}`);
    if (!res.ok) throw new Error("Failed to load post");

    const post = await res.json();

    const avatarSrc = post.avatar ? post.avatar : "static/images/profile.png";

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
            <div class="post-detail">
                <div style="text-align: right; margin-bottom: 20px;">
                    <button onclick="backToHome()" class="back-button">
                        <i class="fas fa-arrow-left"></i> Back to Home
                    </button>
                </div>
                <img src="${avatarSrc}" alt="${post.username}'s avatar" class="avatar">
                <h2>${post.title}</h2>
                <div class="post-content">
                    ${renderContentWithImages(post.content)}
                </div>

                <div class="post-meta">
                    <div><strong>By:</strong> ${post.username}</div>
                    <div><strong>Date:</strong> ${new Date(post.created_at).toLocaleString()}</div>
                </div>

                <div class="reactions">
                    <button id="like-btn-${post.id}">
                        <img src="/static/images/like.png" alt="Like" style="width: 16px; height: 16px;" />
                        <span id="likes-${post.id}">${post.likes_count}</span>
                    </button>
                    <button id="dislike-btn-${post.id}">
                        <img src="/static/images/dislike.png" alt="Dislike" style="width: 16px; height: 16px;" />
                        <span id="dislikes-${post.id}">${post.dislikes_count}</span>
                    </button>
                </div>

                <div class="comments-section">
                    <h3>Comments</h3>
                    <form onsubmit="submitComment(event, ${post.id})" class="comment-form">
                        <input id="comment-input-${post.id}" type="text" placeholder="Write a comment...">
                        <button type="submit">Send Comment</button>
                    </form>
                    <div id="comments-for-${post.id}"></div>
                </div>
            </div>
        </div>
    `;

    // ✅ Wire the like/dislike buttons safely (avoiding onclick global issues)
    document.getElementById(`like-btn-${post.id}`).addEventListener("click", (e) => {
      e.stopPropagation();
      reactToPost(post.id, "like");
    });

    document.getElementById(`dislike-btn-${post.id}`).addEventListener("click", (e) => {
      e.stopPropagation();
      reactToPost(post.id, "dislike");
    });

    loadComments(post.id);

    // Initialize ChatManager if not already initialized
    if (!window.chatManager) {
        window.chatManager = new ChatManager();
    } else {
        window.chatManager.loadAllUsers(); // Reload user list
    }

  } catch (error) {
    console.error("Error loading post:", error);
    app.innerHTML = `<p>Failed to load post.</p>`;
  }
}

//Reaction update happens live when called from backend response
window.onReactionUpdate = async (postId, likes, dislikes) => {
  const inDetailsPage = document.querySelector(".post-detail");
  if (inDetailsPage) {
    await showPostDetails(postId);
  } else {
    // Just update numbers visually
    const likeEl = document.getElementById(`likes-${postId}`);
    const dislikeEl = document.getElementById(`dislikes-${postId}`);
    if (likeEl) likeEl.textContent = likes;
    if (dislikeEl) dislikeEl.textContent = dislikes;
  }
};


window.reactToPost = reactToPost;
window.showPostDetails = showPostDetails;

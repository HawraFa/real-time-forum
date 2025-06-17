import { loadComments } from './comments.js';
import { reactToPost } from './reactions.js';

export async function showPostDetails(postId) {
  const app = document.getElementById("app");
  const currentUser = JSON.parse(localStorage.getItem("currentUser"));

  try {
    const res = await fetch(`http://localhost:8080/api/posts/${postId}`);
    if (!res.ok) throw new Error("Failed to load post");

    const post = await res.json();

    const avatarSrc = post.avatar ? post.avatar : "static/images/profile.png";

    app.innerHTML = `
      <div class="post-detail">
          <img src="${avatarSrc}" alt="${post.username}'s avatar" class="avatar" style="width: 60px; height: 60px; border-radius: 50%; object-fit: cover;">
          <h2>${post.title}</h2>
          <p>${post.content}</p>

          <div><strong>By:</strong> ${post.username}</div>
          <div><strong>Date:</strong> ${new Date(post.created_at).toLocaleString()}</div>
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

          <div id="comments-for-${post.id}"></div>
          <form onsubmit="submitComment(event, ${post.id})">
              <input id="comment-input-${post.id}" type="text" placeholder="Write a comment...">
              <button type="submit">Send</button>
          </form>
          <br/>
          <button onclick="backToHome()">← Back to Home</button>
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

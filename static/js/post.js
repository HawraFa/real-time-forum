import { loadComments } from './comments.js';

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
          <div id="comments-for-${post.id}"></div>
          <form onsubmit="submitComment(event, ${post.id})">
              <input id="comment-input-${post.id}" type="text" placeholder="Write a comment...">
              <button type="submit">Send</button>
          </form>
          <br/>
          <button onclick="backToHome()">← Back to Home</button>
      </div>
    `;

      // Load comments
      loadComments(post.id);

  } catch (error) {
      console.error("Error loading post:", error);
      app.innerHTML = `<p>Failed to load post.</p>`;
  }
}

window.showPostDetails = showPostDetails;


export async function loadComments(postId) {
  const container = document.getElementById(`comments-for-${postId}`);
  // if (!container) return;

  if (!container) {
    console.error("Missing container for post", postId);
    return;
  }

  try {
    const response = await fetch(`http://localhost:8080/api/comments?post_id=${postId}`);
    const comments = await response.json();

    console.log("Comments received for post", postId, comments);

      if (!Array.isArray(comments) || comments.length === 0) {
        container.innerHTML = `<p class="no-comments">Be the first to comment!</p>`;
        return;
      }

      container.innerHTML = comments.map(comment => `
           <div class="comment">
            <img src="${comment.avatar || '/static/images/profile.png'}" 
                class="avatar-icon" 
                style="width: 30px; height: 30px; border-radius: 50%; object-fit: cover;">
            <strong>${comment.username}</strong>:
            <span>${comment.content}</span>
            <small> — ${new Date(comment.created_at).toLocaleString()}</small>
        </div>
      `).join('');

      console.log("🔍 Loading comments for:", postId);

  } catch (err) {
      console.error("Failed to load comments for post", postId, err);
      container.innerHTML = `<p>Error loading comments.</p>`;
  }
}
window.loadComments = loadComments;

export async function submitComment(event, postId) {
  console.log("Submitting comment to post:", postId);

  event.preventDefault();
  const input = document.getElementById(`comment-input-${postId}`);
  const content = input.value.trim();
  if (!content) return;

  const user = JSON.parse(localStorage.getItem('currentUser'));

  await fetch("/api/comments", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
          user_id: user.id,
          post_id: postId,
          content: content
      })
  });

  input.value = "";
  loadComments(postId);
}

window.submitComment = submitComment;


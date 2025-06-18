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
              alt="${comment.username}'s avatar">
          <div class="comment-content">
            <div class="comment-header">
              <span class="comment-author">${comment.username}</span>
              <span class="comment-meta">${new Date(comment.created_at).toLocaleString()}</span>
            </div>
            <div class="comment-text">${comment.content}</div>
          </div>
        </div>
      `).join('');

  } catch (err) {
      console.error("Failed to load comments for post", postId, err);
      container.innerHTML = `<p class="no-comments">Error loading comments.</p>`;
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
  if (!user || !user.id) {
    console.error("No user found in localStorage");
    return;
  }

  try {
    const response = await fetch("http://localhost:8080/api/comments", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        user_id: user.id,
        post_id: postId,
        content: content
      })
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Failed to submit comment");
    }

    input.value = "";
    await loadComments(postId);
  } catch (error) {
    console.error("Error submitting comment:", error);
    const container = document.getElementById(`comments-for-${postId}`);
    if (container) {
      container.innerHTML = `<p class="error-message">Error submitting comment. Please try again.</p>`;
    }
  }
}

window.submitComment = submitComment;


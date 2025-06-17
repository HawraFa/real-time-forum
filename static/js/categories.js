// Load categories into the filter dropdown (run this on page load)
async function loadCategories() {
  try {
    const response = await fetch("/api/categories");
    const categories = await response.json();
    console.log("✅ Categories fetched:", categories); // Debug line

    const select = document.getElementById("categoryFilter");
    if (!select) {
      console.error("❌ Cannot find #categoryFilter");
      return;
    }

    categories.forEach(cat => {
      const option = document.createElement("option");
      option.value = cat.id;
      option.textContent = cat.name;
      select.appendChild(option);
    });

    console.log("✅ Options added.");
  } catch (error) {
    console.error("❌ Failed to load categories:", error);
  }
}

// Trigger filtering when button is clicked
async function filterPosts() {
  const selected = Array.from(document.getElementById("categoryFilter").selectedOptions)
    .map(opt => opt.value)
    .join(",");

  if (!selected) {
    alert("Select at least one category.");
    return;
  }

  try {
    const response = await fetch(`/api/posts/filter?categories=${selected}`);
    const posts = await response.json();
    renderPosts(posts); // Make sure renderPosts() is defined in your main JS
  } catch (error) {
    console.error("Failed to fetch filtered posts:", error);
  }
}

// Shared renderer: Render posts and display them on the page
function renderPosts(posts) {
  const app = document.getElementById("app");

  if (!Array.isArray(posts) || posts.length === 0) {
    app.innerHTML = "<p>No posts found for selected categories.</p>";
    return;
  }

  let postHTML = posts.map(post => {
    const categoriesHTML = (post.categories || [])
      .map(cat => `<span class="category-tag">${cat.name}</span>`)
      .join(" ");

    return `
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

        <div class="post-categories">
          ${categoriesHTML}
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
    `;
  }).join("");

  app.innerHTML = `
    <div class="container">
      <h2>Filtered Posts</h2>
      ${postHTML}
      <button onclick="backToHome()">Back to Home</button>
    </div>
  `;

  // Load comments dynamically after rendering
  setTimeout(() => {
    posts.forEach(post => {
      const el = document.getElementById(`comments-for-${post.id}`);
      if (el) loadComments(post.id);
    });
  }, 100);
}


// Make sure this is globally accessible if needed
window.renderPosts = renderPosts;
window.loadCategories = loadCategories;
window.filterPosts = filterPosts;

// Run loadCategories on page load
window.addEventListener("DOMContentLoaded", loadCategories);

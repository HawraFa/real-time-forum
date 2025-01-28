// Make functions globally available
window.loadCategories = loadCategories;
window.loadPosts = loadPosts;
window.showPosts = showPosts;
window.showNewPostForm = showNewPostForm;
window.filterPostsByCategory = filterPostsByCategory;
window.showPostDetail = showPostDetail;
window.handleNewPost = handleNewPost;
window.handleNewComment = handleNewComment;
window.previewImage = previewImage;
window.handleReaction = handleReaction;

// Load all categories
async function loadCategories() {
    try {
        const response = await fetch('http://localhost:8080/api/categories');
        console.log('Categories response status:', response.status);

        if (!response.ok) {
            throw new Error('Failed to load categories');
        }

        const categories = await response.json();
        console.log('Raw categories response:', categories);

        // Update category filter if it exists
        const categoriesFilter = document.querySelector('.categories-filter');
        if (categoriesFilter) {
            categoriesFilter.innerHTML = `
                <select id="category-select" onchange="filterPostsByCategory(this.value)">
                    <option value="">All Categories</option>
                    ${categories.map(category => 
                        `<option value="${category.id}">${category.name}</option>`
                    ).join('')}
                </select>
            `;
        }

        // Update new post form category select if it exists
        const categorySelect = document.getElementById('category');
        if (categorySelect) {
            categorySelect.innerHTML = `
                <option value="">Select Category</option>
                ${categories.map(category => 
                    `<option value="${category.id}">${category.name}</option>`
                ).join('')}
            `;
        }
    } catch (error) {
        console.error('Error loading categories:', error);
        const categoriesFilter = document.querySelector('.categories-filter');
        if (categoriesFilter) {
            categoriesFilter.innerHTML = `<div class="error">Error loading categories: ${error.message}</div>`;
        }
    }
}

// Show all posts (main posts view)
function showPosts() {
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
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
    `;

    loadCategories();
    loadPosts();
}

// Load all posts or posts by category
async function loadPosts(categoryId = null) {
    try {
        const url = categoryId 
            ? `http://localhost:8080/api/posts/category/${categoryId}`
            : 'http://localhost:8080/api/posts';
            
        console.log('Fetching posts from:', url);
        const response = await fetch(url);
        console.log('Posts response status:', response.status);

        if (!response.ok) {
            const errorData = await response.json();
            console.error('Server error:', errorData);
            throw new Error(errorData.message || 'Failed to load posts');
        }

        const posts = await response.json();
        console.log('Received posts:', posts);
        
        const postsContainer = document.getElementById('posts-container');
        if (!posts || posts.length === 0) {
            postsContainer.innerHTML = '<p class="no-posts">No posts found</p>';
            return;
        }
        
        postsContainer.innerHTML = posts.map(post => `
            <div class="post-card" onclick="showPostDetail(${post.id})">
                <h3>${post.title}</h3>
                ${post.image ? `<img src="${post.image}" alt="Post image" class="post-image">` : ''}
                ${post.content ? `<p class="post-preview">${post.content.substring(0, 150)}...</p>` : ''}
                <div class="post-meta">
                    <span>By ${post.username}</span>
                    <span>${new Date(post.createdAt).toLocaleDateString()}</span>
                </div>
                <div class="post-stats">
                    <span>👍 ${post.likesCount || 0}</span>
                    <span>👎 ${post.dislikesCount || 0}</span>
                    <span>💬 ${post.commentsCount || 0}</span>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading posts:', error);
        const postsContainer = document.getElementById('posts-container');
        if (postsContainer) {
            postsContainer.innerHTML = `<div class="error">Error loading posts: ${error.message}</div>`;
        }
    }
}

// Filter posts by category
function filterPostsByCategory(categoryId) {
    loadPosts(categoryId || null);
}

// Show new post form
function showNewPostForm() {
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        showLoginPage();
        return;
    }

    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar(currentUser)}
        <div class="container">
            <h2>Create New Post</h2>
            <div id="error-container" class="error hidden"></div>
            <form id="newPostForm" onsubmit="handleNewPost(event)">
                <div class="form-group">
                    <label for="title">Title</label>
                    <input type="text" id="title" required>
                </div>
                <div class="form-group">
                    <label for="category">Category</label>
                    <select id="category" required>
                        <option value="">Select Category</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="content">Content</label>
                    <textarea id="content" required></textarea>
                </div>
                <div class="form-group">
                    <label for="postImage">Image (optional)</label>
                    <input type="file" id="postImage" accept="image/*" onchange="previewImage(event)">
                    <div id="imagePreview" class="image-preview"></div>
                </div>
                <div class="form-actions">
                    <button type="submit">Create Post</button>
                    <button type="button" onclick="showPosts()">Cancel</button>
                </div>
            </form>
        </div>
    `;

    // Load categories for the select input
    loadCategoriesForSelect();
}

// Show post detail with comments and reactions
async function showPostDetail(postId) {
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar(currentUser)}
        <div class="container">
            <div id="post-detail">Loading post...</div>
            <div class="comments-section">
                <h3>Comments</h3>
                <form id="comment-form" onsubmit="handleNewComment(event, ${postId})">
                    <div class="form-group">
                        <textarea id="comment-content" required placeholder="Write a comment..."></textarea>
                    </div>
                    <button type="submit">Post Comment</button>
                </form>
                <div id="comments-container">Loading comments...</div>
            </div>
        </div>
    `;

    try {
        const response = await fetch(`http://localhost:8080/api/posts/${postId}`);
        if (!response.ok) {
            throw new Error('Failed to load post');
        }

        const post = await response.json();
        
        document.getElementById('post-detail').innerHTML = `
            <div class="post-full">
                <h2>${post.title}</h2>
                ${post.image ? `<img src="${post.image}" alt="Post image" class="post-image">` : ''}
                <p class="post-content">${post.content}</p>
                <div class="post-meta">
                    <span>By ${post.username}</span>
                    <span>${new Date(post.createdAt).toLocaleDateString()}</span>
                </div>
                <div class="post-reactions">
                    <button onclick="handleReaction(${post.id}, 'like')" class="reaction-btn ${post.userReaction === 'like' ? 'active' : ''}">
                        👍 <span>${post.likesCount || 0}</span>
                    </button>
                    <button onclick="handleReaction(${post.id}, 'dislike')" class="reaction-btn ${post.userReaction === 'dislike' ? 'active' : ''}">
                        👎 <span>${post.dislikesCount || 0}</span>
                    </button>
                </div>
            </div>
        `;

        // Load comments
        loadComments(postId);
    } catch (error) {
        console.error('Error loading post:', error);
        document.getElementById('post-detail').innerHTML = `
            <div class="error">Error loading post: ${error.message}</div>
        `;
    }
}

// Load comments for a post
async function loadComments(postId) {
    try {
        const response = await fetch(`http://localhost:8080/api/posts/${postId}/comments`);
        if (!response.ok) {
            throw new Error('Failed to load comments');
        }

        const comments = await response.json();
        const container = document.getElementById('comments-container');
        
        if (!comments || comments.length === 0) {
            container.innerHTML = '<p class="no-comments">No comments yet</p>';
            return;
        }

        container.innerHTML = comments.map(comment => `
            <div class="comment">
                <div class="comment-header">
                    <span class="comment-author">${comment.username}</span>
                    <span class="comment-date">${new Date(comment.createdAt).toLocaleDateString()}</span>
                </div>
                <p class="comment-content">${comment.content}</p>
                <div class="comment-reactions">
                    <button onclick="handleCommentReaction(${comment.id}, 'like')" class="reaction-btn ${comment.userReaction === 'like' ? 'active' : ''}">
                        👍 <span>${comment.likesCount || 0}</span>
                    </button>
                    <button onclick="handleCommentReaction(${comment.id}, 'dislike')" class="reaction-btn ${comment.userReaction === 'dislike' ? 'active' : ''}">
                        👎 <span>${comment.dislikesCount || 0}</span>
                    </button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading comments:', error);
        document.getElementById('comments-container').innerHTML = `
            <div class="error">Error loading comments: ${error.message}</div>
        `;
    }
}

// Handle new post creation
async function handleNewPost(event) {
    event.preventDefault();
    hideError();

    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        showError('Please login first');
        return;
    }

    const title = document.getElementById('title').value;
    const categoryId = document.getElementById('category').value;
    const content = document.getElementById('content').value;
    const imageFile = document.getElementById('postImage').files[0];

    if (!title || !categoryId || !content) {
        showError('Please fill in all required fields');
        return;
    }

    console.log('Creating post with:', { title, categoryId, content, hasImage: !!imageFile });

    const formData = new FormData();
    formData.append('title', title);
    formData.append('categoryId', categoryId);
    formData.append('content', content);
    if (imageFile) {
        formData.append('image', imageFile);
    }

    try {
        console.log('Sending request with user ID:', currentUser.id);
        const response = await fetch('http://localhost:8080/api/posts', {
            method: 'POST',
            headers: {
                'X-User-ID': currentUser.id.toString()
            },
            credentials: 'include',
            body: formData
        });

        console.log('Response status:', response.status);
        
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.message || 'Failed to create post');
        }

        const data = await response.json();
        console.log('Response data:', data);

        // Show success message
        showError('Post created successfully!', 'success-container');
        
        // Redirect to posts page after a short delay
        setTimeout(() => {
            showPosts();
        }, 1500);
    } catch (error) {
        console.error('Error creating post:', error);
        showError(error.message || 'Failed to create post');
    }
}

// Preview image before upload
function previewImage(event) {
    const file = event.target.files[0];
    const preview = document.getElementById('imagePreview');
    
    if (file) {
        const reader = new FileReader();
        reader.onload = function(e) {
            preview.innerHTML = `<img src="${e.target.result}" alt="Preview">`;
        }
        reader.readAsDataURL(file);
    } else {
        preview.innerHTML = '';
    }
}

// Load categories for select input
async function loadCategoriesForSelect() {
    try {
        const response = await fetch('http://localhost:8080/api/categories');
        if (!response.ok) {
            throw new Error('Failed to load categories');
        }

        const categories = await response.json();
        const select = document.getElementById('category');
        
        categories.forEach(category => {
            const option = document.createElement('option');
            option.value = category.id;
            option.textContent = category.name;
            select.appendChild(option);
        });
    } catch (error) {
        console.error('Error loading categories:', error);
        showError('Failed to load categories');
    }
}

// Handle new comment submission
async function handleNewComment(event, postId) {
    event.preventDefault();
    hideError();

    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        showError('Please login to comment');
        return;
    }

    const content = document.getElementById('comment-content').value;
    if (!content.trim()) {
        showError('Comment cannot be empty');
        return;
    }

    try {
        const response = await fetch(`http://localhost:8080/api/posts/${postId}/comments`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': currentUser.id.toString()
            },
            body: JSON.stringify({ content })
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.message || 'Failed to post comment');
        }

        // Clear comment form and reload comments
        document.getElementById('comment-content').value = '';
        loadComments(postId);
    } catch (error) {
        console.error('Error posting comment:', error);
        showError(error.message);
    }
}

// Handle post/comment reactions
async function handleReaction(id, type, isComment = false) {
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        showError('Please login to react');
        return;
    }

    try {
        const endpoint = isComment 
            ? `/api/comments/${id}/reactions` 
            : `/api/posts/${id}/reactions`;

        const response = await fetch(`http://localhost:8080${endpoint}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': currentUser.id.toString()
            },
            body: JSON.stringify({ type })
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.message || 'Failed to react');
        }

        // Reload the post or comment section
        if (isComment) {
            loadComments(id);
        } else {
            showPostDetail(id);
        }
    } catch (error) {
        console.error('Error handling reaction:', error);
        showError(error.message);
    }
}

// Add other necessary functions... 
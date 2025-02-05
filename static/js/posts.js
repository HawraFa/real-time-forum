// Function to show all posts
function showPosts() {
    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar(JSON.parse(localStorage.getItem('currentUser')))}
        <div class="container">
            <div class="posts-header">
                <h2>Forum Posts</h2>
                <button onclick="showNewPostForm()" class="new-post-button">New Post</button>
            </div>
            <div class="categories-filter">
                <!-- Categories will be loaded here -->
            </div>
            <div id="posts-container">
                <!-- Posts will be loaded here -->
            </div>
        </div>
    `;

    loadCategories();
    loadPosts();
}

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
                    ${categories.map(category => {
                        console.log('Processing category:', category);
                        return `<option value="${category.id}">${category.name}</option>`;
                    }).join('')}
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

// Load all posts or posts by category
async function loadPosts(categoryId = null) {
    try {
        const url = categoryId 
            ? `http://localhost:8080/api/posts/category/${categoryId}`
            : 'http://localhost:8080/api/posts';
            
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Failed to load posts');
        }

        const posts = await response.json();
        
        const postsContainer = document.getElementById('posts-container');
        if (!Array.isArray(posts) || posts.length === 0) {
            postsContainer.innerHTML = '<p class="no-posts">No posts found</p>';
            return;
        }
        
        postsContainer.innerHTML = posts.map(post => `
            <div class="post-card" onclick="showPostDetail(${post.id})">
                <h3>${post.title}</h3>
                <p class="post-preview">${post.content.substring(0, 150)}...</p>
                <div class="post-meta">
                    <span>By ${post.username}</span>
                    <span>${new Date(post.createdAt).toLocaleDateString()}</span>
                </div>
                <div class="post-stats">
                    <span>👍 ${post.likesCount}</span>
                    <span>👎 ${post.dislikesCount}</span>
                    <span>💬 ${post.commentsCount}</span>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading posts:', error);
        document.getElementById('posts-container').innerHTML = `
            <div class="error">Error loading posts: ${error.message}</div>
        `;
    }
}

// Filter posts by category
function filterPostsByCategory(categoryId) {
    loadPosts(categoryId || null);
}

// Show new post form
function showNewPostForm() {
    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar(JSON.parse(localStorage.getItem('currentUser')))}
        <div class="container">
            <h2>Create New Post</h2>
            <form id="newPostForm" onsubmit="handleNewPost(event)">
                <div class="form-group">
                    <label for="title">Title</label>
                    <input type="text" id="title" required>
                </div>
                <div class="form-group">
                    <label for="category">Category</label>
                    <select id="category" required>
                        <!-- Categories will be loaded here -->
                    </select>
                </div>
                <div class="form-group">
                    <label for="content">Content</label>
                    <textarea id="content" required></textarea>
                </div>
                <div class="form-group">
                    <label for="postImage">Image (optional)</label>
                    <input type="file" id="postImage" accept="image/*">
                    <div class="image-preview" id="imagePreview"></div>
                </div>
                <div class="form-actions">
                    <button type="submit">Create Post</button>
                    <button type="button" onclick="showPosts()">Cancel</button>
                </div>
            </form>
        </div>
    `;
    
    // Add image preview functionality
    const imageInput = document.getElementById('postImage');
    const imagePreview = document.getElementById('imagePreview');
    imageInput.addEventListener('change', function(e) {
        const file = e.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = function(e) {
                imagePreview.innerHTML = `<img src="${e.target.result}" alt="Preview">`;
            }
            reader.readAsDataURL(file);
        }
    });
    
    loadCategories();
}

// Handle new post creation
async function handleNewPost(event) {
    event.preventDefault();
    
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        showError('You must be logged in to create a post');
        return;
    }

    // Create post data object
    const postData = {
        userId: currentUser.id,
        categoryId: parseInt(document.getElementById('category').value),
        title: document.getElementById('title').value,
        content: document.getElementById('content').value
    };

    console.log('Post data:', postData);

    try {
        const response = await fetch('http://localhost:8080/api/posts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify(postData)
        });

        console.log('Response status:', response.status);
        const data = await response.json();
        console.log('Response data:', data);

        if (!response.ok) {
            throw new Error(data.message || 'Failed to create post');
        }

        showPosts();
    } catch (error) {
        console.error('Error creating post:', error);
        showError(error.message);
    }
}

// Add click handler to post cards
function showPostDetail(postId) {
    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar(JSON.parse(localStorage.getItem('currentUser')))}
        <div class="container">
            <button onclick="showPosts()" class="back-button">← Back to Posts</button>
            <div id="post-detail">Loading...</div>
            <div class="comments-section">
                <h3>Comments</h3>
                <form id="comment-form" onsubmit="handleNewComment(event, ${postId})">
                    <div class="form-group">
                        <textarea id="comment-content" required placeholder="Write a comment..."></textarea>
                    </div>
                    <button type="submit">Post Comment</button>
                </form>
                <div id="comments-container">
                    <!-- Comments will be loaded here -->
                </div>
            </div>
        </div>
    `;

    loadPostDetail(postId);
    loadComments(postId);
}

// Load post detail
async function loadPostDetail(postId) {
    try {
        const response = await fetch(`http://localhost:8080/api/posts/${postId}`);
        const post = await response.json();

        if (!response.ok) {
            throw new Error(post.message || 'Failed to load post');
        }

        document.getElementById('post-detail').innerHTML = `
            <div class="post-detail">
                <h2>${post.title}</h2>
                <div class="post-meta">
                    <span>By ${post.username}</span>
                    <span>${new Date(post.createdAt).toLocaleDateString()}</span>
                </div>
                <div class="post-content">${post.content}</div>
                <div class="post-stats">
                    <button onclick="handleReaction(${post.id}, 'like')" class="reaction-btn">
                        👍 ${post.likesCount}
                    </button>
                    <button onclick="handleReaction(${post.id}, 'dislike')" class="reaction-btn">
                        👎 ${post.dislikesCount}
                    </button>
                    <span>💬 ${post.commentsCount}</span>
                </div>
            </div>
        `;
    } catch (error) {
        console.error('Error loading post detail:', error);
        document.getElementById('post-detail').innerHTML = `
            <div class="error">Error loading post: ${error.message}</div>
        `;
    }
}

// Load comments for a post
async function loadComments(postId) {
    try {
        const response = await fetch(`http://localhost:8080/api/posts/${postId}/comments`);
        const comments = await response.json();

        const commentsContainer = document.getElementById('comments-container');
        if (comments.length === 0) {
            commentsContainer.innerHTML = '<p class="no-comments">No comments yet</p>';
            return;
        }

        commentsContainer.innerHTML = comments.map(comment => `
            <div class="comment">
                <div class="comment-meta">
                    <span>${comment.username}</span>
                    <span>${new Date(comment.createdAt).toLocaleDateString()}</span>
                </div>
                <div class="comment-content">${comment.content}</div>
                <div class="comment-stats">
                    <button onclick="handleCommentReaction(${comment.id}, 'like')" class="reaction-btn">
                        👍 ${comment.likesCount}
                    </button>
                    <button onclick="handleCommentReaction(${comment.id}, 'dislike')" class="reaction-btn">
                        👎 ${comment.dislikesCount}
                    </button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading comments:', error);
    }
}

// Handle new comment submission
async function handleNewComment(event, postId) {
    event.preventDefault();
    
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        showError('You must be logged in to comment');
        return;
    }

    const content = document.getElementById('comment-content').value;

    try {
        const response = await fetch(`http://localhost:8080/api/posts/${postId}/comments`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                userId: currentUser.id,
                content: content
            }),
        });

        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.message || 'Failed to post comment');
        }

        // Clear form and reload comments
        document.getElementById('comment-form').reset();
        loadComments(postId);
        loadPostDetail(postId); // Reload post to update comment count
    } catch (error) {
        console.error('Error posting comment:', error);
        showError(error.message);
    }
} 
// Make functions globally available
window.showUserProfile = showUserProfile;

async function showUserProfile(userId) {
    const app = document.getElementById('app');
    app.innerHTML = `
        ${createNavbar(JSON.parse(localStorage.getItem('currentUser')))}
        <div class="container">
            <div id="profile-container">Loading profile...</div>
            <div id="user-posts">Loading posts...</div>
        </div>
    `;

    try {
        const response = await fetch(`http://localhost:8080/api/users/${userId}`);
        if (!response.ok) {
            throw new Error('Failed to load profile');
        }

        const profile = await response.json();
        
        document.getElementById('profile-container').innerHTML = `
            <div class="profile-header">
                <img src="${profile.avatar}" alt="Profile" class="profile-avatar">
                <h2>${profile.username}</h2>
            </div>
            <div class="profile-details">
                <p><strong>Name:</strong> ${profile.firstName} ${profile.lastName}</p>
                <p><strong>Email:</strong> ${profile.email}</p>
                <p><strong>Age:</strong> ${profile.age}</p>
                <p><strong>Gender:</strong> ${profile.gender}</p>
                <p><strong>Posts:</strong> ${profile.postCount}</p>
                <p><strong>Joined:</strong> ${new Date(profile.joinDate).toLocaleDateString()}</p>
            </div>
        `;

        // Load user's posts
        const postsResponse = await fetch(`http://localhost:8080/api/posts/user/${userId}`);
        if (!postsResponse.ok) {
            throw new Error('Failed to load user posts');
        }

        const posts = await postsResponse.json();
        const postsContainer = document.getElementById('user-posts');
        
        if (!posts || posts.length === 0) {
            postsContainer.innerHTML = '<p class="no-posts">No posts yet</p>';
            return;
        }

        postsContainer.innerHTML = `
            <h3>Posts</h3>
            <div class="posts-grid">
                ${posts.map(post => `
                    <div class="post-card" onclick="showPostDetail(${post.id})">
                        <h3>${post.title}</h3>
                        ${post.image ? `<img src="${post.image}" alt="Post image" class="post-image">` : ''}
                        <p class="post-preview">${post.content.substring(0, 150)}...</p>
                        <div class="post-meta">
                            <span>${new Date(post.createdAt).toLocaleDateString()}</span>
                        </div>
                        <div class="post-stats">
                            <span>👍 ${post.likesCount || 0}</span>
                            <span>👎 ${post.dislikesCount || 0}</span>
                            <span>💬 ${post.commentsCount || 0}</span>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    } catch (error) {
        console.error('Error loading profile:', error);
        document.getElementById('profile-container').innerHTML = `
            <div class="error">Error loading profile: ${error.message}</div>
        `;
    }
} 
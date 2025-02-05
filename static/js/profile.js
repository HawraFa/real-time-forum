// Show user profile page
function showUserProfile(userID) {
    const app = document.getElementById('app');
    
    // First show loading state
    app.innerHTML = `
        ${createNavbar(JSON.parse(localStorage.getItem('currentUser')))}
        <div class="container">
            <h2>User Profile</h2>
            <div id="profile-content">Loading profile...</div>
        </div>
    `;

    // Then fetch and display the profile
    loadUserProfile(userID);
}

// Load user profile data
async function loadUserProfile(userID) {
    try {
        console.log('Loading profile for user:', userID); // Debug log

        const response = await fetch(`http://localhost:8080/api/profile/${userID}`, {
            method: 'GET',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json',
            },
        });

        console.log('Response status:', response.status); // Debug log

        const data = await response.json();
        console.log('Profile data:', data); // Debug log
        
        if (!response.ok) {
            throw new Error(data.message || 'Failed to load profile');
        }

        displayUserProfile(data);
    } catch (error) {
        console.error('Error loading profile:', error);
        document.getElementById('profile-content').innerHTML = `
            <p class="error">Error loading profile: ${error.message}</p>
        `;
    }
}

// Display user profile
function displayUserProfile(profile) {
    const profileContent = document.getElementById('profile-content');
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    const isOwnProfile = currentUser && currentUser.id === profile.id;

    // Determine which image to display
    let profileImage = profile.avatar;
    if (!profileImage || profileImage === 'pictures/profile.png') {
        profileImage = 'pictures/profile.png';
    }

    console.log('Profile avatar:', profile.avatar); // Debug log
    console.log('Using image:', profileImage); // Debug log

    profileContent.innerHTML = `
        <div class="profile-container">
            <div class="profile-header">
                <img src="${profileImage}" alt="Profile picture" class="profile-avatar">
                <h3>${profile.username}</h3>
            </div>
            <div class="profile-details">
                <p><strong>Name:</strong> ${profile.firstName} ${profile.lastName}</p>
                <p><strong>Email:</strong> ${profile.email}</p>
                <p><strong>Age:</strong> ${profile.age}</p>
                <p><strong>Gender:</strong> ${profile.gender || 'Not specified'}</p>
                <p><strong>Join Date:</strong> ${new Date(profile.joinDate).toLocaleDateString()}</p>
                <p><strong>Posts:</strong> ${profile.postCount || 0}</p>
            </div>
            ${isOwnProfile ? `
                <div class="profile-actions">
                    <button onclick="showEditProfileForm(${JSON.stringify(profile).replace(/"/g, '&quot;')})">Edit Profile</button>
                </div>
            ` : ''}
        </div>
    `;
}

// Show edit profile form
function showEditProfileForm(profile) {
    const profileContent = document.getElementById('profile-content');
    
    profileContent.innerHTML = `
        <div class="edit-profile-container">
            <h3>Edit Profile</h3>
            <form id="editProfileForm" onsubmit="handleProfileUpdate(event)">
                <div class="form-group">
                    <label for="profilePicture">Profile Picture</label>
                    <input type="file" id="profilePicture" name="profilePicture" accept="image/*">
                </div>
                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" value="${profile.username || ''}" required>
                </div>
                <div class="form-group">
                    <label for="email">Email</label>
                    <input type="email" id="email" name="email" value="${profile.email || ''}" required>
                </div>
                <div class="form-group">
                    <label for="firstName">First Name</label>
                    <input type="text" id="firstName" name="firstName" value="${profile.firstName || ''}" required>
                </div>
                <div class="form-group">
                    <label for="lastName">Last Name</label>
                    <input type="text" id="lastName" name="lastName" value="${profile.lastName || ''}" required>
                </div>
                <div class="form-group">
                    <label for="age">Age</label>
                    <input type="number" id="age" name="age" value="${profile.age || ''}" required>
                </div>
                <div class="form-actions">
                    <button type="submit">Save Changes</button>
                    <button type="button" onclick="showUserProfile(${profile.id})">Cancel</button>
                </div>
            </form>

            <h3>Change Password</h3>
            <form id="changePasswordForm" onsubmit="handlePasswordChange(event)">
                <div class="form-group">
                    <label for="currentPassword">Current Password</label>
                    <input type="password" id="currentPassword" required>
                </div>
                <div class="form-group">
                    <label for="newPassword">New Password</label>
                    <input type="password" id="newPassword" required>
                </div>
                <div class="form-group">
                    <label for="confirmPassword">Confirm New Password</label>
                    <input type="password" id="confirmPassword" required>
                </div>
                <div class="form-actions">
                    <button type="submit">Change Password</button>
                </div>
            </form>
        </div>
    `;
}

// Handle profile update
async function handleProfileUpdate(event) {
    event.preventDefault();
    
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        showError('You must be logged in to update your profile');
        return;
    }

    // Create FormData object to handle file upload
    const formData = new FormData();
    const profilePicture = document.getElementById('profilePicture').files[0];
    if (profilePicture) {
        formData.append('profilePicture', profilePicture);
    }

    // Add other form data
    const updates = {
        username: document.getElementById('username').value,
        email: document.getElementById('email').value,
        firstName: document.getElementById('firstName').value,
        lastName: document.getElementById('lastName').value,
        age: parseInt(document.getElementById('age').value),
        userID: currentUser.id
    };

    console.log('Sending updates:', updates); // Debug log

    // Append JSON data
    formData.append('data', JSON.stringify(updates));

    try {
        const response = await fetch('http://localhost:8080/api/profile/update', {
            method: 'POST',
            body: formData,
        });

        console.log('Response status:', response.status); // Debug log

        const data = await response.json();
        console.log('Response data:', data); // Debug log

        if (!response.ok) {
            throw new Error(data.message || 'Failed to update profile');
        }

        // Update local storage with new username if it changed
        if (updates.username !== currentUser.username) {
            currentUser.username = updates.username;
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
        }

        // Refresh the profile display
        showUserProfile(currentUser.id);
    } catch (error) {
        console.error('Profile update error:', error);
        showError('Failed to update profile: ' + error.message);
    }
}

// Add password change handler
async function handlePasswordChange(event) {
    event.preventDefault();
    
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    if (!currentUser) {
        showError('You must be logged in to change your password');
        return;
    }

    const currentPassword = document.getElementById('currentPassword').value;
    const newPassword = document.getElementById('newPassword').value;
    const confirmPassword = document.getElementById('confirmPassword').value;

    if (newPassword !== confirmPassword) {
        showError('New passwords do not match');
        return;
    }

    try {
        const response = await fetch('http://localhost:8080/api/change-password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                userID: currentUser.id,
                currentPassword: currentPassword,
                newPassword: newPassword
            }),
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.message || 'Failed to change password');
        }

        showSuccess('Password changed successfully');
        // Clear the form
        document.getElementById('changePasswordForm').reset();
    } catch (error) {
        console.error('Password change error:', error);
        showError(error.message);
    }
}

// Add success message function
function showSuccess(message) {
    const successDiv = document.createElement('div');
    successDiv.className = 'success';
    successDiv.textContent = message;
    document.querySelector('.edit-profile-container').prepend(successDiv);
    
    // Remove the message after 3 seconds
    setTimeout(() => {
        successDiv.remove();
    }, 3000);
} 
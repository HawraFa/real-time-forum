// Function to show the home page
function showHomePage() {
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    
    // If no user is logged in, redirect to login page
    if (!currentUser) {
        showLoginPage();
        return;
    }
    
    // Show posts page as home page
    showPosts();
}
    

// Initialize the application
function init() {
    // Check if user is logged in
    const currentUser = JSON.parse(localStorage.getItem('currentUser'));
    
    // If no user is logged in, show login page
    if (!currentUser) {
        showLoginPage();
    } else {
        // If user is logged in, show home page
        showHomePage();
    }
}

// Start the application
init(); 
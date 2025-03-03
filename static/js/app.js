import { showLoginForm } from './auth.js';
import { showHomePage } from './home.js';

// Initialize the application
function init() {
    const currentUser = localStorage.getItem('currentUser');
    if (!currentUser) {
        showLoginForm();
    } else {
        showHomePage(JSON.parse(currentUser));
    }
}

// Call init when the page loads
window.addEventListener('load', init); 
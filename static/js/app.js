import { showLoginForm } from './auth.js';
import { showHomePage } from './home.js';

// Initialize the application
function init() {
    const currentUser = localStorage.getItem('currentUser');
    
    if (!currentUser) {
        fetch("/api/session", {
            method: "GET",
            credentials: "include"
        })
        .then(res => res.json())
        .then(data => {
            localStorage.setItem("currentUser", JSON.stringify(data.user));
            showHomePage(data.user);
        })
        .catch(err => {
            showLoginForm();
        });
    } else {
        showHomePage(JSON.parse(currentUser));
    }
}

// Call init when the page loads
window.addEventListener('load', init); 

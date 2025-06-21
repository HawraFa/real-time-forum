import { showLoginForm } from './auth.js';
import { showHomePage } from './home.js';
import { showPostDetails } from "./post.js"; 
import './categories.js';

// Define valid routes for the application
const validRoutes = [
    '/',
    '/home',
    '/login',
    '/register',
    '/profile',
    '/chat',
    '/posts',
    '/create-post',
    '/categories',
    '/settings'
];

// Function to show error page
function showErrorPage() {
    const app = document.getElementById("app");
    const errorContainer = document.getElementById("error-container");
    
    // Hide the main app content
    app.style.display = "none";
    
    // Show the error container
    errorContainer.style.display = "flex";
}

// Function to hide error page
function hideErrorPage() {
    const app = document.getElementById("app");
    const errorContainer = document.getElementById("error-container");
    
    // Show the main app content
    app.style.display = "block";
    
    // Hide the error container
    errorContainer.style.display = "none";
}

// Function to check if current route is valid
function checkRoute() {
    const currentPath = window.location.pathname;
    
    // Check if the path is valid
    if (!validRoutes.includes(currentPath)) {
        showErrorPage();
        return false;
    } else {
        hideErrorPage();
        return true;
    }
}

// Theme management
function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'light';
    document.documentElement.setAttribute('data-theme', savedTheme);
    updateThemeIcon(savedTheme);
}

function toggleTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme');
    const newTheme = currentTheme === 'light' ? 'dark' : 'light';
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    updateThemeIcon(newTheme);
}

function updateThemeIcon(theme) {
    const themeIcon = document.querySelector('.theme-toggle svg');
    if (themeIcon) {
        themeIcon.innerHTML = theme === 'light' 
            ? '<path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.06-.9-.1-1.36-.1z"/>'
            : '<path d="M12 7c-2.76 0-5 2.24-5 5s2.24 5 5 5 5-2.24 5-5-2.24-5-5-5zM2 13h2c.55 0 1-.45 1-1s-.45-1-1-1H2c-.55 0-1 .45-1 1s.45 1 1 1zm18 0h2c.55 0 1-.45 1-1s-.45-1-1-1h-2c-.55 0-1 .45-1 1s.45 1 1 1zM11 2v2c0 .55.45 1 1 1s1-.45 1-1V2c0-.55-.45-1-1-1s-1 .45-1 1zm0 18v2c0 .55.45 1 1 1s1-.45 1-1v-2c0-.55-.45-1-1-1s-1 .45-1 1zM5.99 4.58c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0 .39-.39.39-1.03 0-1.41L5.99 4.58zm12.37 12.37c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0 .39-.39.39-1.03 0-1.41l-1.06-1.06zm1.06-10.96c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41.39.39 1.03.39 1.41 0l1.06-1.06zM7.05 18.36c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41.39.39 1.03.39 1.41 0l1.06-1.06z"/>';
    }
}

// Initialize the application
async function init() {
    // Check route validity first
    if (!checkRoute()) {
        return; // Stop initialization if route is invalid
    }
    
    const currentUser = localStorage.getItem('currentUser');

    if (!currentUser) {
    fetch("/api/session", {
        method: "GET",
        credentials: "include"
    })
    .then(res => {
        if (!res.ok) throw new Error("unauthenticated");
        return res.json();
    })
    .then(data => {
        localStorage.setItem("currentUser", JSON.stringify(data.user));
        showHomePage(data.user); 
    })
    .catch(err => {
        showLoginForm(); // ✅ sidebar/chat will never show if you cleaned HTML
    });
    } else {
    fetch("/api/session", {
        method: "GET",
        credentials: "include"
    })
    .then(res => {
        if (!res.ok) {
        localStorage.removeItem("currentUser");
        showLoginForm();
        return;
        }
        return res.json();
    })
    .then(data => {
        if (data) {
        localStorage.setItem("currentUser", JSON.stringify(data.user));
        showHomePage(data.user);
        }
    })
    .catch(err => {
        localStorage.removeItem("currentUser");
        showLoginForm();
    });
    }

    // Initialize theme
    initTheme();
}

// Add theme toggle to window object
window.toggleTheme = toggleTheme;

// Listen for browser back/forward buttons
window.addEventListener('popstate', function() {
    checkRoute();
});

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', init); 

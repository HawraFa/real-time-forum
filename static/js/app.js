import { showLoginForm } from './auth.js';
import { showHomePage } from './home.js';
import { showPostDetails } from "./post.js"; 
import './categories.js';

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

// Error page component
function showErrorPage() {
    const app = document.getElementById("app");
    app.innerHTML = `
        <nav class="navbar">
            <div class="nav-left">
                <h1>Forum</h1>
            </div>
            <div class="nav-right">
                <button class="theme-toggle" onclick="toggleTheme()">
                    <svg viewBox="0 0 24 24" width="24" height="24">
                        <path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.06-.9-.1-1.36-.1z"/>
                    </svg>
                </button>
            </div>
        </nav>

        <div class="main-content" style="margin-left: 0; padding: 20px;">
            <div class="container">
                <div class="error-page">
                    <div class="error-content">
                        <h1 class="error-title">404</h1>
                        <h2 class="error-subtitle">Page Not Found</h2>
                        <p class="error-message">The page you're looking for doesn't exist or has been moved.</p>
                        <button onclick="goToHome()" class="error-button">Go to Home</button>
                    </div>
                </div>
            </div>
        </div>
    `;

    // Hide all chat elements when showing error page
    const chatUsersContainer = document.querySelector('.chat-users-container');
    const chatWindow = document.querySelector('.chat-window');
    const chatSidebar = document.querySelector('.chat-sidebar');
    
    if (chatUsersContainer) chatUsersContainer.style.display = 'none';
    if (chatWindow) chatWindow.style.display = 'none';
    if (chatSidebar) chatSidebar.style.display = 'none';

    // Stop ChatManager if it exists
    if (window.chatManager) {
        // Close WebSocket connection if it exists
        if (window.chatManager.ws) {
            window.chatManager.ws.close();
        }
        // Clear the chat manager reference
        window.chatManager = null;
    }
}

// Navigation function
window.goToHome = function() {
    // Show chat elements again when navigating back to home
    const chatUsersContainer = document.querySelector('.chat-users-container');
    const chatWindow = document.querySelector('.chat-window');
    const chatSidebar = document.querySelector('.chat-sidebar');
    
    if (chatUsersContainer) chatUsersContainer.style.display = '';
    if (chatWindow) chatWindow.style.display = '';
    if (chatSidebar) chatSidebar.style.display = '';

    const currentUser = localStorage.getItem('currentUser');
    if (currentUser) {
        showHomePage(JSON.parse(currentUser));
    } else {
        showLoginForm();
    }
};

// Simple routing function
function handleRoute() {
    const path = window.location.pathname;
    
    // Valid routes
    const validRoutes = ['/', '/home', '/login', '/register'];
    
    if (validRoutes.includes(path)) {
        // Handle valid routes
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
    } else {
        // Show error page for invalid routes
        showErrorPage();
    }
}

// Initialize the application
function init() {
    // Handle routing
    handleRoute();

    // Initialize theme
    initTheme();
}

// Add theme toggle to window object
window.toggleTheme = toggleTheme;

// Handle browser back/forward buttons
window.addEventListener('popstate', handleRoute);

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', init); 

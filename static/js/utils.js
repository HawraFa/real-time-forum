// Function to show error message
function showError(message, containerClass = 'error') {
    const errorContainer = document.getElementById('error-container');
    if (errorContainer) {
        errorContainer.textContent = message;
        errorContainer.className = containerClass;
    }
}

// Function to hide error message
function hideError() {
    const errorContainer = document.getElementById('error-container');
    if (errorContainer) {
        errorContainer.className = 'error hidden';
    }
}

// Make utility functions globally available
window.showError = showError;
window.hideError = hideError; 
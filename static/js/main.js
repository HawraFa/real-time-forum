// Load all required JavaScript files in order
function loadScript(src) {
  return new Promise((resolve, reject) => {
      const script = document.createElement('script');
      script.src = src;
      script.onload = resolve;
      script.onerror = reject;
      document.body.appendChild(script);
  });
}

// Load scripts in sequence
async function loadAllScripts() {
  try {
      await loadScript('static/js/utils.js');
      await loadScript('static/js/login.js');
      await loadScript('static/js/registration.js');
      await loadScript('static/js/profile.js');
      await loadScript('static/js/posts.js');
      await loadScript('static/js/app.js');
      
      // Initialize the application after all scripts are loaded
      init();
  } catch (error) {
      console.error('Error loading scripts:', error);
  }
}

// Start loading scripts when the page loads
window.onload = loadAllScripts; 
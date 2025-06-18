import { renderPosts } from './home.js';

// Load categories into the filter dropdown (run this on page load)
async function loadCategories() {
  try {
    const response = await fetch("/api/categories");
    const categories = await response.json();
    console.log("✅ Categories fetched:", categories); // Debug line

    const select = document.getElementById("categoryFilter");
    if (!select) {
      return;
    }

    categories.forEach(cat => {
      const option = document.createElement("option");
      option.value = cat.id;
      option.textContent = cat.name;
      select.appendChild(option);
    });

    console.log("✅ Options added.");
  } catch (error) {
    console.error("❌ Failed to load categories:", error);
  }
}

export async function filterPosts() {
    const selectedOptions = Array.from(document.getElementById("categoryFilter").selectedOptions);
    const selectedCategories = selectedOptions.map(opt => opt.value);

    const container = document.getElementById("filtered-posts-container");
    container.innerHTML = "Loading...";

    if (selectedCategories.length === 0) {
        container.innerHTML = "<p>Please select at least one category.</p>";
        return;
    }

    try {
        const res = await fetch(`http://localhost:8080/api/posts/filter?categories=${selectedCategories.join(",")}`);
        const posts = await res.json();
        renderPosts(posts, "filtered-posts-container");
    } catch (err) {
        console.error("❌ Failed to fetch filtered posts:", err);
        container.innerHTML = "<p>⚠️ Error loading filtered posts.</p>";
    }
}


// Make sure this is globally accessible if needed
window.loadCategories = loadCategories;
window.filterPosts = filterPosts;

// Run loadCategories on page load
window.addEventListener("DOMContentLoaded", loadCategories);

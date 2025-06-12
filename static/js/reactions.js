export async function reactToPost(postId, type) {
  const user = JSON.parse(localStorage.getItem("currentUser"));
  const reactionType = type.charAt(0).toUpperCase() + type.slice(1);  // Make it 'Like' or 'Dislike'

  try {
      const res = await fetch("http://localhost:8080/api/react", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
              user_id: user.id,
              post_id: postId,
              comment_id: 0,
              type: reactionType
          })
      });

      if (res.ok) {
        const data = await res.json();
        updateReactionDisplay(postId, data.likes, data.dislikes); // NEW
    } else {
        const msg = await res.text();
        console.error("React failed:", msg);
    }
    

      showAllPosts(); // Re-fetch posts to get updated counts
  } catch (err) {
      console.error("Error in reaction:", err);
  }
}

function updateReactionDisplay(postId, likes, dislikes) {
    document.getElementById(`likes-${postId}`).textContent = likes;
    document.getElementById(`dislikes-${postId}`).textContent = dislikes;
}



window.reactToPost = reactToPost;


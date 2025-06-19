export class ChatManager {
  constructor() {
    console.log("ChatManager: Constructor started");

    const currentUser = JSON.parse(localStorage.getItem("currentUser"));
    if (!currentUser) {
      this.handleNotLoggedIn();
      return;
    }

    this.currentUser = currentUser;
    this.currentUserId = currentUser.id;
    this.ws = null;
    this.messageOffset = 0;
    this.currentChatUser = null;
    this.messageContainer = document.querySelector(".chat-messages");
    this.typingTimeout = null;

    // Add these properties for message pagination
    this.isLoadingMessages = false;       // Track if we're currently loading messages
    this.hasMoreMessages = true;          // Track if there are more messages to load
    this.scrollPositionBeforeLoad = 0;    // Remember scroll position before loading
    this.messagesBatchSize = 10; // Number of messages to load at a time
    this.initialLoadComplete = false;     // Track if initial messages have loaded

    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;

    // Initialize notifications from localStorage
    this.notifications = JSON.parse(localStorage.getItem('chatNotifications')) || {};
    // Initialize online users from localStorage
    this.onlineUsers = new Set(JSON.parse(localStorage.getItem('onlineUsers')) || []);

    // Show chat sidebar
    const chatSidebar = document.querySelector(".chat-sidebar");
    if (chatSidebar) {
      chatSidebar.style.display = "block";
    }

    // Initialize with empty message container
    if (this.messageContainer) {
      this.messageContainer.innerHTML = '';
    }

    // Set up WebSocket and event listeners
    this.setupWebSocket();
    this.setupEventListeners();
  }

  handleNotLoggedIn() {
    const chatSidebar = document.querySelector(".chat-sidebar");
    const chatWindow = document.querySelector(".chat-window");
    if (chatSidebar) chatSidebar.style.display = "none";
    if (chatWindow) chatWindow.style.display = "none";
    alert("You must be logged in to use chat.");
    window.showLoginForm();
  }

  setupWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log("WebSocket connection established");
      this.isConnected = true;
      this.reconnectAttempts = 0;
      
      // Send initial status
      this.sendStatusUpdate("online");
      
      // Request online users list
      this.ws.send(JSON.stringify({
        type: "get_online_users"
      }));

      // Wait until currentUserId is set before loading users
      const waitForCurrentUserId = async () => {
        if (this.currentUserId) {
          this.loadAllUsers();
        } else {
          console.log("Waiting for currentUserId to be set before loading users...");
          setTimeout(waitForCurrentUserId, 200);
        }
      };
      waitForCurrentUserId();
    };

    this.ws.onclose = () => {
      console.log("WebSocket connection closed");
      this.isConnected = false;
      this.handleReconnect();
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (err) {
        console.error("Failed to parse WebSocket message:", err);
      }
    };
  }

  // Add this new helper method
  showConnectionError() {
    const errorElement = document.createElement('div');
    errorElement.className = 'connection-error';
    errorElement.textContent = 'Connection lost. Please refresh the page.';
    document.body.appendChild(errorElement);
    setTimeout(() => errorElement.remove(), 5000);
  }

  setupEventListeners() {
    const form = document.querySelector(".message-input-form");
    if (form) {
      form.addEventListener("submit", (e) => {
        e.preventDefault();
        this.sendMessage();
      });
    }

    // if (this.messageContainer) {
    //   this.messageContainer.addEventListener(
    //     "scroll",
    //     this.throttle(async () => {
    //       if (this.messageContainer.scrollTop === 0 && this.currentChatUser) {
    //         await this.loadMoreMessages();
    //       }
    //     }, 500)
    //   );
    // }

    const input = document.querySelector(".message-input");
    if (input) {
      input.addEventListener(
        "input",
        this.debounce(() => {
          this.sendTypingStatus();
        }, 300)
      );
    }

    if (this.messageContainer) {
      this.messageContainer.addEventListener("scroll", this.throttle(() => {
        // Trigger when within 100px of top
        if (this.messageContainer.scrollTop < 100 && 
            !this.isLoadingMessages && 
            this.hasMoreMessages) {
          this.loadMoreMessages();
        }
      }, 300)); // Throttle to 300ms
    }
  }

  async loadAllUsers() {
    try {
      // Get all users
      const usersRes = await fetch(`/api/users?_=${Date.now()}`);
      if (!usersRes.ok) {
        throw new Error(`Failed to load users: ${usersRes.status}`);
      }
      const users = await usersRes.json();
      console.log("All users from server:", users);

      // Then get interactions
      const interactionsRes = await fetch("/api/chat/last-interactions");
      let interactions = [];
      if (interactionsRes.ok) {
        const contentType = interactionsRes.headers.get("content-type");
        if (contentType && contentType.includes("application/json")) {
          interactions = await interactionsRes.json();
        } else {
          console.warn("⚠️ Interactions response was not JSON");
        }
      }

      // ✅ Ensure it's always an array
      if (!Array.isArray(interactions)) {
        console.warn("⚠️ Interactions is not an array, defaulting to []");
        interactions = [];
      }

      const interactionMap = {};
      interactions.forEach(inter => {
        interactionMap[inter.user2Id] = new Date(inter.lastInteractionTime);
      });

      // Filter out the current user and ensure all required fields are present
      const usersWithTimes = users
        .filter(user => {
          if (!user || !user.id) {
            console.warn("⚠️ Found invalid user object:", user);
            return false;
          }
          // Convert both IDs to numbers for comparison
          const userId = parseInt(user.id);
          const currentUserId = parseInt(this.currentUserId);
          const isCurrentUser = userId === currentUserId;
          console.log(`Checking user ${user.username} (ID: ${userId}) against current user (ID: ${currentUserId}): ${isCurrentUser ? 'is current user' : 'is not current user'}`);
          return !isCurrentUser;
        })
        .map(user => ({
          ...user,
          id: parseInt(user.id), // Ensure ID is a number
          lastInteractionTime: interactionMap[user.id] || null
        }));

      console.log("Filtered users:", usersWithTimes);

      // Sort users by last interaction time and then alphabetically
      usersWithTimes.sort((a, b) => {
        const aTime = a.lastInteractionTime;
        const bTime = b.lastInteractionTime;

        if (aTime && bTime) return bTime - aTime;
        if (aTime) return -1;
        if (bTime) return 1;
        return a.username.localeCompare(b.username);
      });

      console.log("✅ Final sorted users:", usersWithTimes.map(u => u.username));

      // Update the UI with the filtered and sorted users
      this.updateUsersList(usersWithTimes);
    } catch (err) {
      console.error("❌ loadAllUsers failed:", err);
      alert("Something went wrong loading users.");
    }
  }

  updateUsersList(users) {
    const userList = document.querySelector(".chat-user-list");
    if (!userList) {
      console.error("Chat user list element not found");
      return;
    }

    // Clear the current list
    userList.innerHTML = "";

    // Double check to ensure current user is not in the list
    const filteredUsers = users.filter(user => {
      if (!user || !user.id) {
        console.warn("⚠️ Found invalid user in updateUsersList:", user);
        return false;
      }
      const isCurrentUser = parseInt(user.id) === parseInt(this.currentUserId);
      console.log(`updateUsersList: Checking user ${user.username} (ID: ${user.id}) against current user (ID: ${this.currentUserId}): ${isCurrentUser ? 'is current user' : 'is not current user'}`);
      return !isCurrentUser;
    });

    console.log("Final users to display:", filteredUsers);

    filteredUsers.forEach((user) => {
      const userElement = document.createElement("li");
      userElement.className = "chat-user-item";
      userElement.dataset.userId = user.id;

      // Get online status from our Set
      const isOnline = this.onlineUsers.has(user.id);

      userElement.innerHTML = `
        <div class="chat-user-avatar">
          <img src="${user.avatar || "/static/images/profile.png"}" alt="${user.username}">
          <span class="chat-user-status ${isOnline ? "status-online" : "status-offline"}"></span>
        </div>
        <div class="chat-user-info">
          <div class="chat-user-name">${user.username}</div>
          <div class="chat-user-fullname">${user.firstName} ${user.lastName}</div>
        </div>
      `;

      // Add notification dot if user has notifications
      if (this.notifications[user.id]) {
        const userInfo = userElement.querySelector(".chat-user-info");
        const notificationDot = document.createElement("div");
        notificationDot.className = "notification-dot";
        userInfo.appendChild(notificationDot);
      }

      userElement.addEventListener("click", () => {
        // Remove notification dot when clicked
        const notificationDot = userElement.querySelector(".notification-dot");
        if (notificationDot) {
          notificationDot.remove();
          // Remove from notifications state
          delete this.notifications[user.id];
          localStorage.setItem('chatNotifications', JSON.stringify(this.notifications));
        }
        this.selectUser(user.id);
      });
      userList.appendChild(userElement);
    });

    console.log("🧪 Final sidebar order:", filteredUsers.map(u => u.username));
  }

  // Update handleMessage to properly handle status updates
  handleMessage(message) {
    console.log("🔁 handleMessage ran")
    switch (message.type) {
      case "message":
        this.handleIncomingMessage(message);
        break;
      case "typing":
        if (message.receiverId === this.currentUserId) {
          this.showTypingIndicatorUI(message.senderName || "User");
        }
        break;
      case "error":
        // Handle error messages from server
        alert(message.content);
        break;
      case "status":
        console.log(`✅ Updated status for user ${message.senderId}: ${message.content}`);
        this.updateOnlineStatus(message.senderId, message.content === "online");
        break;
      case "online_users":
        // Update our local set of online users
        this.onlineUsers = new Set(message.userIds);
        localStorage.setItem('onlineUsers', JSON.stringify([...this.onlineUsers]));
        this.refreshUserListStatuses();
        break;
      case "user_joined":
        this.onlineUsers.add(message.userId);
        localStorage.setItem('onlineUsers', JSON.stringify([...this.onlineUsers]));
        this.updateSingleUserStatus({
          userId: message.userId,
          isOnline: true
        });
        break;
      case "user_left":
        this.onlineUsers.delete(message.userId);
        localStorage.setItem('onlineUsers', JSON.stringify([...this.onlineUsers]));
        this.updateSingleUserStatus({
          userId: message.userId,
          isOnline: false
        });
        break;
      default:
        console.warn("Unknown message type:", message.type);
    }
  }
    // New method to update online status
  updateOnlineStatus(userId, isOnline) {
    if (isOnline) {
      this.onlineUsers.add(userId);
    } else {
      this.onlineUsers.delete(userId);
    }
    this.updateSingleUserStatus({ userId, isOnline });
  }

updateSingleUserStatus(userStatus) {
    const userElement = document.querySelector(`[data-user-id="${userStatus.userId}"]`);
    if (!userElement) return;

    const statusDot = userElement.querySelector(".chat-user-status");
    if (statusDot) {
        statusDot.classList.toggle("status-online", userStatus.isOnline);
        statusDot.classList.toggle("status-offline", !userStatus.isOnline);
         console.log("Updated status for user:", userStatus.userId, "to", userStatus.isOnline ? "online" : "offline");
    }

}

 displayNewMessage(message) {
    const messageElement = this.createMessageElement(message);
    this.messageContainer.appendChild(messageElement);
    this.messageContainer.scrollTop = this.messageContainer.scrollHeight;
}

 createMessageElement(message) {
    const div = document.createElement("div");
    const isCurrentUser = message.senderId === this.currentUserId;
    
    div.className = `message ${isCurrentUser ? "sent" : "received"}`;  
    
    // Format the date to show date before time in US format
    const messageDate = new Date(message.timestamp);
    const options = { 
        year: 'numeric', 
        month: '2-digit', 
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: true
    };
    const formattedDate = messageDate.toLocaleString('en-US', options);
    
    div.innerHTML = `   
        <div class="message-info"> 
            <span class="message-sender">${isCurrentUser ? "You" : message.username || "User"}</span>
            <span class="message-time">${formattedDate}</span>
        </div>
        <div class="message-content">${message.content}</div>
    `;
    
    return div;
}

sendMessage() {
    const input = document.querySelector(".message-input");
    if (!input) return;
    const content = input.value.trim();
    if (!content || !this.currentChatUser) return;

    // Check for HTML tags in the message
    const htmlTagRegex = /<[^>]*>/g;
    if (htmlTagRegex.test(content)) {
        alert("Error: HTML tags are not allowed in chat messages. Please use plain text only.");
        return;
    }

    // Check if the recipient is online
    if (!this.onlineUsers.has(this.currentChatUser)) {
        alert("Cannot send message: User is offline");
        return;
    }

    // Create the message object
    const message = {
        type: "message",
        senderId: this.currentUserId,
        receiverId: this.currentChatUser,
        content: content,
        timestamp: new Date().toISOString(),
        username: this.currentUser.username,
        avatar: this.currentUser.avatar
    };

    // Immediately display the message in your own chat
    this.displayNewMessage(message);

    try {
        // Send to server
        this.ws.send(JSON.stringify(message));
        input.value = ""; // Clear input field

        //Refetch and resort sidebar after sending immediatly without refreshing
    this.loadAllUsers();
    } catch (err) {
        console.error("Failed to send message", err);
        alert("Failed to send message. Please try again.");
    }
}

  showTypingIndicator(message) {
    const typingIndicator = document.querySelector(".typing-indicator");
    typingIndicator.textContent = "Typing.....";
    typingIndicator.style.display = "block";
    clearTimeout(this.typingTimeout);
    this.typingTimeout = setTimeout(() => {
      typingIndicator.style.display = "none";
    }, 3000);
  }

  sendTypingStatus() {
    if (!this.currentChatUser) return;
    const message = {
      type: "typing",
      senderId: this.currentUserId,
      receiverId: this.currentChatUser,
      timestamp: new Date(),
    };
    try {
      this.ws.send(JSON.stringify(message));
    } catch (err) {
      console.error("Failed to send typing status:", err);
    }
  }

  async selectUser(userId) {
    try {
        const parsedUserId = parseInt(userId);
        const chatWindow = document.querySelector(".chat-window");

        // Toggle if same user
        if (this.currentChatUser === parsedUserId) {
            if (chatWindow) {
                chatWindow.style.display = 
                    chatWindow.style.display === "none" ? "block" : "none";
                // If we're closing the chat window, set currentChatUser to null
                if (chatWindow.style.display === "none") {
                    this.currentChatUser = null;
                }
            }
            return;
        }

        this.currentChatUser = parsedUserId;
        this.messageOffset = 0;

        // Show loading state
        const header = document.getElementById("chat-header");
        if (header) header.textContent = "Loading...";

        // Fetch user and messages in parallel
        const [currentUser] = await Promise.all([
            this.getUserById(userId),
        ]);

        // Update UI
        if (header) {
            header.textContent = currentUser ? 
                `Chat with ${currentUser.username}` : 
                `Chat with User ID: ${userId}`;
        }

        const messageForm = document.getElementById("chat-message-form");
        if (messageForm) messageForm.style.display = "flex";
        if (chatWindow) chatWindow.style.display = "block";

        // Clear and load messages
        if (this.messageContainer) {
            this.messageContainer.innerHTML = "";
            await this.loadInitialChatHistory(userId);
        }

        // Update active state in user list
        document.querySelectorAll(".chat-user-item").forEach((item) => {
            item.classList.toggle("active", item.dataset.userId === String(userId));
        });

    } catch (err) {
        console.error("Error in selectUser:", err);
    }
}

  // async loadChatHistory(userId) {
  //   try {
  //     const response = await fetch(`/api/chat/history?user_id=${userId}&offset=0`);
  //     const messages = await response.json();
  //     messages.reverse().forEach((msg) => {
  //       const msgElement = this.createMessageElement(msg);
  //       this.messageContainer.appendChild(msgElement);
  //     });
  //     this.messageContainer.scrollTop = this.messageContainer.scrollHeight;
  //   } catch (err) {
  //     console.error("Failed to load chat history:", err);
  //   }
  // }

  // async loadMoreMessages() {
  //   // Prevent multiple concurrent loads
  //   if (this.isLoadingMessages || !this.hasMoreMessages || !this.currentChatUser) {
  //     return;
  //   }
  
  //   this.isLoadingMessages = true;
  //   this.showLoadingIndicator();
  
  //   try {
  //     const scrollHeightBefore = this.messageContainer.scrollHeight;
  //     const scrollTopBefore = this.messageContainer.scrollTop;
  
  //     const response = await fetch(
  //       `/api/chat/history?user_id=${this.currentChatUser}&offset=${this.messageOffset}&limit=${this.messagesBatchSize}`
  //     );
      
  //     if (!response.ok) throw new Error(`HTTP ${response.status}`);
      
  //     const messages = await response.json();
      
  //     if (messages.length === 0) {
  //       this.hasMoreMessages = false;
  //       return;
  //     }
  
  //     // Add messages to top (maintain DOM order)
  //     messages.forEach(msg => {
  //       this.addMessageToTop(msg);
  //     });
  
  //     this.messageOffset += messages.length;
  
  //     // Restore scroll position (adjusted for new content)
  //     const scrollHeightAfter = this.messageContainer.scrollHeight;
  //     this.messageContainer.scrollTop = scrollTopBefore + (scrollHeightAfter - scrollHeightBefore);
  
  //     // Check if we've reached the beginning
  //     if (messages.length < this.messagesBatchSize) {
  //       this.hasMoreMessages = false;
  //     }
  
  //   } catch (err) {
  //     console.error("Load more error:", err);
  //   } finally {
  //     this.hideLoadingIndicator();
  //     this.isLoadingMessages = false;
  //   }
  // }

  async loadMoreMessages() {
    // Prevent multiple concurrent loads
    if (this.isLoadingMessages || !this.hasMoreMessages || !this.currentChatUser) {
      return;
    }
  
    this.isLoadingMessages = true;
    this.showLoadingIndicator();
  
    try {
      const scrollHeightBefore = this.messageContainer.scrollHeight;
      const scrollTopBefore = this.messageContainer.scrollTop;
  
      const response = await fetch(
        `/api/chat/history?user_id=${this.currentChatUser}&offset=${this.messageOffset}&limit=${this.messagesBatchSize}`
      );
      
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      
      const messages = await response.json();
      
      if (messages.length === 0) {
        this.hasMoreMessages = false;
        return;
      }
  
      messages.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
      // Add messages to top (maintain DOM order)
      messages.forEach(msg => {
        this.addMessageToTop(msg);
      });
  
      this.messageOffset += messages.length;
  
      // Restore scroll position (adjusted for new content)
      const scrollHeightAfter = this.messageContainer.scrollHeight;
      this.messageContainer.scrollTop = scrollTopBefore + (scrollHeightAfter - scrollHeightBefore);
  
      // Check if we've reached the beginning
      if (messages.length < this.messagesBatchSize) {
        this.hasMoreMessages = false;
      }
  
    } catch (err) {
      console.error("Load more error:", err);
    } finally {
      this.hideLoadingIndicator();
      this.isLoadingMessages = false;
    }
  }


  // async markMessagesAsRead(senderId) {
  //   try {
  //     await fetch(`/api/chat/mark-read?sender_id=${senderId}`, {
  //       method: "POST",
  //       credentials: "include",
  //     });
  //   } catch (error) {
  //     console.error("Failed to mark messages as read:", error);
  //   }
  // }

  updateUserStatus(statusMessage) {
    const userElement = document.querySelector(
      `[data-user-id="${statusMessage.senderId}"]`
    );
    if (!userElement) return;

    const statusDot = userElement.querySelector(".chat-user-status");
    if (statusMessage.content === "1") {
      statusDot.classList.replace("status-offline", "status-online");
    } else {
      statusDot.classList.replace("status-online", "status-offline");
    }
  }

  throttle(func, limit) {
    let inThrottle;
    return (...args) => {
      if (!inThrottle) {
        func.apply(this, args);
        inThrottle = true;
        setTimeout(() => (inThrottle = false), limit);
      }
    };
  }

  debounce(func, wait) {
    let timeout;
    return (...args) => {
      clearTimeout(timeout);
      timeout = setTimeout(() => func.apply(this, args), wait);
    };
  }

async getUserById(userId) {
    try {
        const response = await fetch(`/api/users/${userId}`, {  // Changed endpoint to plural
            method: 'GET',
            credentials: 'include',
            headers: {
                'Accept': 'application/json'
            }
        });

        // Check for non-JSON responses
        const contentType = response.headers.get('content-type');
        if (!contentType || !contentType.includes('application/json')) {
            const text = await response.text();
            throw new Error(`Expected JSON, got: ${text.substring(0, 100)}...`);
        }

        if (!response.ok) {
            const error = await response.json().catch(() => ({}));
            throw new Error(error.message || `HTTP error! Status: ${response.status}`);
        }

        return await response.json();
    } catch (err) {
        console.error("Failed to get user by ID:", err);
        
        // Show user-friendly error
        const errorElement = document.getElementById('chat-error');
        if (errorElement) {
            errorElement.textContent = "Failed to load user. Please try again.";
            errorElement.style.display = 'block';
            setTimeout(() => errorElement.style.display = 'none', 3000);
        }
        
        return null;
    }
}

  showMessageNotification(message) {
    if ("Notification" in window && Notification.permission === "granted") {
      new Notification(`New message from ${message.senderId}`, {
        body: message.content,
        icon: "/static/images/notification.png",
      });
    }
  }

  displayNotification(message) {
    if (message.senderId === this.currentUserId) return;
    this.showMessageNotification(message);
  }

  handleIncomingMessage(message) {
    this.displayNotification(message);
    
    // Show notification if we're not in the current chat with this user
    // or if the chat window is closed
    const chatWindow = document.querySelector(".chat-window");
    const isChatWindowClosed = chatWindow && chatWindow.style.display === "none";
    
    if (message.senderId !== this.currentChatUser || isChatWindowClosed) {
        // Move user to top and add notification dot
        const userList = document.querySelector(".chat-user-list");
        const userElement = document.querySelector(`[data-user-id="${message.senderId}"]`);
        
        if (userElement && userList) {
            // Remove existing notification dot if any
            const existingDot = userElement.querySelector(".notification-dot");
            if (existingDot) {
                existingDot.remove();
            }
            
            // Add new notification dot
            const userInfo = userElement.querySelector(".chat-user-info");
            const notificationDot = document.createElement("div");
            notificationDot.className = "notification-dot";
            userInfo.appendChild(notificationDot);
            
            // Store notification state
            this.notifications[message.senderId] = true;
            localStorage.setItem('chatNotifications', JSON.stringify(this.notifications));
            
            // Move user to top
            userList.insertBefore(userElement, userList.firstChild);
        }
    }
    
    if (message.senderId === this.currentChatUser) {
        this.displayNewMessage(message);
    }
  }

  sendStatusUpdate(status) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({
            type: "status",
            content: status
        }));
    }
  }

  updateUsersListWithLastMessage(users) {
    const userList = document.querySelector(".chat-user-list");
    if (!userList) return;
    userList.innerHTML = "";

    users.forEach((user) => {
      if (user.id === this.currentUserId) return;
      const userElement = document.createElement("li");
      userElement.className = "chat-user-item";
      userElement.dataset.userId = user.id;

      userElement.innerHTML = `
        <div class="chat-user-avatar">
          <img src="${user.avatar || "/static/images/profile.png"}" alt="${user.username}">
          <span class="chat-user-status ${
            user.isOnline ? "status-online" : "status-offline"
          }"></span>
        </div>
        <div class="chat-user-info">
          <div class="chat-user-name">${user.username}</div>
          <div class="chat-user-fullname">${user.firstName} ${user.lastName}</div>
        </div>
      `;

      userElement.addEventListener("click", () => this.selectUser(user.id));
      userList.appendChild(userElement);
    });
  }

  loadAllUsersWithLastMessage() {
    fetch("/api/users")
      .then((res) => res.json())
      .then((users) => {
        this.updateUsersListWithLastMessage(users);
      })
      .catch((err) => console.error("Error fetching users:", err));
  }

  displayOlderMessages(messages) {
    const fragment = document.createDocumentFragment();
    messages.forEach((msg) => {
        const msgElement = this.createMessageElement(msg);
        fragment.insertBefore(msgElement, fragment.firstChild);
    });

    const oldScrollHeight = this.messageContainer.scrollHeight;
    this.messageContainer.insertBefore(fragment, this.messageContainer.firstChild);
    this.messageContainer.scrollTop =
        this.messageContainer.scrollHeight - oldScrollHeight;
  }

  addOldMessageToChat(message) {
    const messageElement = this.createMessageElement(message);
    this.messageContainer.prepend(messageElement);
  }

  showMessage(message) {
    const messageElement = this.createMessageElement(message);
    this.messageContainer.appendChild(messageElement);
    this.messageContainer.scrollTop = this.messageContainer.scrollHeight;
  }

  showTypingIndicatorUI() {
    const typingIndicator = document.querySelector(".typing-indicator");
    if (!typingIndicator) return;
    typingIndicator.textContent = `Typing.....`;
    typingIndicator.style.display = "block";
    clearTimeout(this.typingTimeout);
    this.typingTimeout = setTimeout(() => {
      typingIndicator.style.display = "none";
    }, 2000);
  }

  async loadInitialChatHistory(userId) {
    try {
      this.messageOffset = 0;
      this.hasMoreMessages = true;
      
      const response = await fetch(
        `/api/chat/history?user_id=${userId}&offset=${this.messageOffset}&limit=${this.messagesBatchSize}`
      );
      const messages = await response.json();
      
      if (!Array.isArray(messages)) {
        throw new Error("Expected array of messages");
      }
  
      // Clear existing messages
      this.messageContainer.innerHTML = '';
      messages.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
      
      // Display messages in chronological order (newest at bottom)
      messages.forEach(msg => this.displayNewMessage(msg));
      
      // Auto-scroll to bottom to show newest messages
      this.messageContainer.scrollTop = this.messageContainer.scrollHeight;
      
      // Update offset for next load
      this.messageOffset += messages.length;
      
      // If we got fewer than requested, no more messages
      if (messages.length < this.messagesBatchSize) {
        this.hasMoreMessages = false;
      }
    } catch (err) {
      console.error("Failed to load initial chat history:", err);
    }
  }
  addMessageToTop(message) {
    const messageElement = this.createMessageElement(message);
  
    // Insert message in correct place based on timestamp
    const messageDate = new Date(message.timestamp);
    const children = Array.from(this.messageContainer.children);
  
    for (let i = 0; i < children.length; i++) {
      const el = children[i];
      const timeSpan = el.querySelector('.message-time');
      if (!timeSpan) continue;
      const existingDate = new Date(timeSpan.textContent);
      if (messageDate < existingDate) {
        this.messageContainer.insertBefore(messageElement, el);
        return;
      }
    }
  }


  showLoadingIndicator() {
    const loader = document.createElement('div');
    loader.className = 'message-loader';
    loader.textContent = 'Loading more messages...';
    this.messageContainer.prepend(loader);
  }

  hideLoadingIndicator() {
    const loader = this.messageContainer.querySelector('.message-loader');
    if (loader) loader.remove();
  }

}

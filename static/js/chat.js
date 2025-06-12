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

    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;

    // Show chat sidebar
    const chatSidebar = document.querySelector(".chat-sidebar");
    if (chatSidebar) {
      chatSidebar.style.display = "block";
    }

    this.setupWebSocket();
    this.setupEventListeners();
    this.loadAllUsers();
    this.onlineUsers = new Set();
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
  const wsProtocol = window.location.protocol === "https:" ? "wss://" : "ws://";
  
  try {
    this.ws = new WebSocket(
      `${wsProtocol}${window.location.host}/ws`
    );

    this.ws.onopen = () => {
      console.log("Connected to chat server");
      this.isConnected = true;
      this.reconnectAttempts = 0;

      this.ws.send(JSON.stringify({
          type: "get_online_users"
        }));
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (err) {
        console.error("Invalid WebSocket message", err);
      }
    };

    this.ws.onclose = (event) => {
      console.log("Disconnected from chat server", event.code, event.reason);
      this.isConnected = false;
      
      if (event.code === 4003) {
        this.handleAuthFailure();
        return;
      }

      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        const delay = Math.min(3000 * (this.reconnectAttempts + 1), 15000);
        setTimeout(() => {
          this.reconnectAttempts++;
          this.setupWebSocket();
        }, delay);
      } else {
        console.error("Max reconnection attempts reached");
        this.showConnectionError();
      }
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
      this.isConnected = false;
    };

  } catch (err) {
    console.error("Failed to connect to WebSocket:", err);
    this.isConnected = false;
  }
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

    if (this.messageContainer) {
      this.messageContainer.addEventListener(
        "scroll",
        this.throttle(async () => {
          if (this.messageContainer.scrollTop === 0 && this.currentChatUser) {
            await this.loadMoreMessages();
          }
        }, 500)
      );
    }

    const input = document.querySelector(".message-input");
    if (input) {
      input.addEventListener(
        "input",
        this.debounce(() => {
          this.sendTypingStatus();
        }, 300)
      );
    }
  }

  async loadAllUsers() {
    try {
      const response = await fetch("/api/users");
      const contentType = response.headers.get("content-type");
      if (!contentType || !contentType.includes("application/json")) {
        throw new Error("Received non-json response (likely a redirect)");
      }
      const users = await response.json();
      this.updateUsersList(users);
    } catch (error) {
      console.error("Failed to load users:", error.message);
      alert("You must be logged in to access chat.");
      window.showLoginForm();
    }
  }

// In your ChatManager class, modify the updateUsersList method:
updateUsersList(users) {
    const userList = document.querySelector(".chat-user-list");
    if (!userList) return;
    userList.innerHTML = "";

    users.forEach((user) => {
        if (user.id === this.currentUserId) return;

        const userElement = document.createElement("li");
        userElement.className = "chat-user-item";
        userElement.dataset.userId = user.id;

        // Get online status from WebSocket or initial load
        const isOnline = user.isOnline || false;

        userElement.innerHTML = `
            <div class="chat-user-avatar">
                <img src="${user.avatar || "/static/images/profile.png"}" alt="${user.username}">
                <span class="chat-user-status ${isOnline ? "status-online" : "status-offline"}"></span>
            </div>
            <div class="chat-user-info">
                <div class="chat-user-name">${user.username}</div>
                <div class="chat-user-fullname">${user.firstName} ${user.lastName}</div>
            </div>
            <div class="chat-message-meta">
                <div class="chat-time">Just now</div>
                <div class="chat-unread-count">0</div>
            </div>
        `;

        userElement.addEventListener("click", () => this.selectUser(user.id));
        userList.appendChild(userElement);
    });
}

 // Update handleMessage to properly handle status updates
  handleMessage(message) {
    switch (message.type) {
      case "message":
        this.handleIncomingMessage(message);
        break;
      case "typing":
        if (message.receiverId === this.currentUserId) {
          this.showTypingIndicatorUI(message.senderName || "User");
        }
        break;
      case "status":
        this.updateOnlineStatus(message.senderId, message.content === "online");
        break;
      case "online_users":
        // Update our local set of online users
        message.userIds.forEach(id => this.onlineUsers.add(id));
        this.refreshUserListStatuses();
        break;
      case "user_joined":
        this.onlineUsers.add(message.userId);
        this.updateSingleUserStatus({
          userId: message.userId,
          isOnline: true
        });
        break;
      case "user_left":
        this.onlineUsers.delete(message.userId);
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
    div.innerHTML = `   
        <div class="message-info"> 
            <span class="message-sender">${isCurrentUser ? "You" : message.username || "User"}</span>
            <span class="message-time">${new Date(message.timestamp).toLocaleTimeString()}</span>
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
    } catch (err) {
        console.error("Failed to send message", err);
        // Optional: Show error to user
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
            //this.markMessagesAsRead(userId)
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
            //await this.loadChatHistory(userId);
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

  async loadMoreMessages() {
    if (!this.currentChatUser) return;
    this.messageOffset += 10;

    try {
      const response = await fetch(
        `/api/chat/history?user_id=${this.currentChatUser}&offset=${this.messageOffset}`
      );
      const messages = await response.json();

      const fragment = document.createDocumentFragment();
      messages.reverse().forEach((msg) => {
        const msgElement = this.createMessageElement(msg);
        fragment.insertBefore(msgElement, fragment.firstChild);
      });

      const oldScrollHeight = this.messageContainer.scrollHeight;
      this.messageContainer.insertBefore(
        fragment,
        this.messageContainer.firstChild
      );
      this.messageContainer.scrollTop =
        this.messageContainer.scrollHeight - oldScrollHeight;
    } catch (err) {
      console.error("Failed to load more messages:", err);
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
    if (message.senderId === this.currentChatUser) {
      this.displayNewMessage(message);
      //this.markMessagesAsRead(message.senderId);
    }
  }

  sendStatusUpdate(status) {
    const statusMessage = {
      type: "status",
      senderId: this.currentUserId,
      content: status,
      timestamp: new Date(),
    };
    try {
      this.ws.send(JSON.stringify(statusMessage));
    } catch (err) {
      console.error("Failed to send status update:", err);
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

  createMessageBubble(message) {
    const div = document.createElement("div");
    div.className = `message ${
      message.senderId === this.currentUserId ? "sent" : "received"
    }`;

    div.innerHTML = `
      <div class="message-info">
        <span class="message-sender">${
          message.senderId === this.currentUserId ? "You" : "User"
        }</span>
        <span class="message-time">${new Date(message.timestamp).toLocaleTimeString()}</span>
      </div>
      <div class="message-content">${message.content}</div>
    `;

    return div;
  }

  showMessage(message) {
    const messageElement = this.createMessageBubble(message);
    this.messageContainer.appendChild(messageElement);
    this.messageContainer.scrollTop = this.messageContainer.scrollHeight;
  }

  addOldMessageToChat(message) {
    const messageElement = this.createMessageBubble(message);
    this.messageContainer.prepend(messageElement);
  }

  async loadMoreMessagesHandler() {
    if (!this.currentChatUser) return;
    this.messageOffset += 10;

    try {
      const response = await fetch(
        `/api/chat/history?user_id=${this.currentChatUser}&offset=${this.messageOffset}`
      );
      const messages = await response.json();
      messages.reverse().forEach((msg) => {
        this.addOldMessageToChat(msg);
      });
    } catch (err) {
      console.error("Failed to load older messages:", err);
    }
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
      const response = await fetch(`/api/chat/history?user_id=${userId}&offset=0`);
      const messages = await response.json();
      messages.reverse().forEach((msg) => this.showMessage(msg));
    } catch (err) {
      console.error("Failed to load initial chat history:", err);
    }
  }

}
class ChatManager {
    constructor() {
        this.ws = null;
        this.messageOffset = 0;
        this.currentChatUser = null;
        this.messageContainer = document.querySelector('.chat-messages');
        this.setupWebSocket();
        this.setupEventListeners();
        this.loadOnlineUsers();
    }

    setupWebSocket() {
        this.ws = new WebSocket(`ws://${window.location.host}/ws`);
        
        this.ws.onopen = () => {
            console.log('Connected to chat server');
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.ws.onclose = () => {
            console.log('Disconnected from chat server');
            // Attempt to reconnect after 3 seconds
            setTimeout(() => this.setupWebSocket(), 3000);
        };
    }

    setupEventListeners() {
        // Send message on form submit
        document.querySelector('.message-input-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.sendMessage();
        });

        // Load more messages on scroll
        this.messageContainer.addEventListener('scroll', this.throttle(() => {
            if (this.messageContainer.scrollTop === 0) {
                this.loadMoreMessages();
            }
        }, 500));

        // Handle typing indicator
        document.querySelector('.message-input').addEventListener('input', this.debounce(() => {
            this.sendTypingStatus();
        }, 300));
    }

    async loadOnlineUsers() {
        try {
            const response = await fetch('/api/chat/online-users');
            const users = await response.json();
            this.updateUsersList(users);
        } catch (error) {
            console.error('Error loading online users:', error);
        }
    }

    async loadChatHistory(userId) {
        try {
            const response = await fetch(`/api/chat/history?user_id=${userId}&offset=${this.messageOffset}`);
            const messages = await response.json();
            this.displayMessages(messages);
        } catch (error) {
            console.error('Error loading chat history:', error);
        }
    }

    sendMessage() {
        const input = document.querySelector('.message-input');
        const content = input.value.trim();
        
        if (!content || !this.currentChatUser) return;

        const message = {
            type: 'message',
            senderId: currentUserId, // You'll need to set this from your session
            receiverId: this.currentChatUser,
            content: content,
            timestamp: new Date()
        };

        this.ws.send(JSON.stringify(message));
        input.value = '';
    }

    handleMessage(message) {
        switch (message.type) {
            case 'message':
                this.displayNewMessage(message);
                break;
            case 'status':
                this.updateUserStatus(message);
                break;
            case 'typing':
                this.showTypingIndicator(message);
                break;
        }
    }

    displayMessages(messages) {
        messages.forEach(msg => {
            const messageElement = this.createMessageElement(msg);
            this.messageContainer.appendChild(messageElement);
        });
    }

    createMessageElement(message) {
        const div = document.createElement('div');
        div.className = `message ${message.senderId === currentUserId ? 'sent' : 'received'}`;
        
        div.innerHTML = `
            <div class="message-info">
                <span class="message-sender">${message.senderId === currentUserId ? 'You' : 'User'}</span>
                <span class="message-time">${new Date(message.timestamp).toLocaleTimeString()}</span>
            </div>
            <div class="message-content">${message.content}</div>
        `;
        
        return div;
    }

    displayNewMessage(message) {
        const messageElement = this.createMessageElement(message);
        
        // If this is a new message, append it at the bottom
        this.messageContainer.appendChild(messageElement);
        
        // Scroll to the new message
        this.messageContainer.scrollTop = this.messageContainer.scrollHeight;
        
        // If the message is from someone else and they're the current chat user
        if (message.senderId !== currentUserId && message.senderId === this.currentChatUser) {
            this.markMessagesAsRead(message.senderId);
        }
    }

    updateUserStatus(statusMessage) {
        const userElement = document.querySelector(`[data-user-id="${statusMessage.senderId}"]`);
        if (userElement) {
            const statusDot = userElement.querySelector('.chat-user-status');
            if (statusMessage.data.online) {
                statusDot.classList.remove('status-offline');
                statusDot.classList.add('status-online');
            } else {
                statusDot.classList.remove('status-online');
                statusDot.classList.add('status-offline');
            }
        }
    }

    showTypingIndicator(message) {
        if (message.senderId !== this.currentChatUser) return;
        
        const typingIndicator = document.querySelector('.typing-indicator');
        typingIndicator.textContent = 'User is typing...';
        typingIndicator.style.display = 'block';
        
        // Hide the indicator after 3 seconds
        clearTimeout(this.typingTimeout);
        this.typingTimeout = setTimeout(() => {
            typingIndicator.style.display = 'none';
        }, 3000);
    }

    sendTypingStatus() {
        if (!this.currentChatUser) return;

        const message = {
            type: 'typing',
            senderId: currentUserId,
            receiverId: this.currentChatUser,
            timestamp: new Date()
        };

        this.ws.send(JSON.stringify(message));
    }

    async loadMoreMessages() {
        if (!this.currentChatUser) return;
        
        this.messageOffset += 10;
        try {
            const response = await fetch(`/api/chat/history?user_id=${this.currentChatUser}&offset=${this.messageOffset}`);
            const messages = await response.json();
            
            if (messages.length > 0) {
                const fragment = document.createDocumentFragment();
                messages.reverse().forEach(msg => {
                    const messageElement = this.createMessageElement(msg);
                    fragment.appendChild(messageElement);
                });
                
                // Store current scroll height
                const oldScrollHeight = this.messageContainer.scrollHeight;
                
                // Insert messages at the top
                this.messageContainer.insertBefore(fragment, this.messageContainer.firstChild);
                
                // Maintain scroll position
                this.messageContainer.scrollTop = this.messageContainer.scrollHeight - oldScrollHeight;
            }
        } catch (error) {
            console.error('Error loading more messages:', error);
        }
    }

    updateUsersList(users) {
        const userList = document.querySelector('.chat-user-list');
        userList.innerHTML = ''; // Clear current list
        
        users.forEach(user => {
            const userElement = document.createElement('li');
            userElement.className = 'chat-user-item';
            userElement.dataset.userId = user.id;
            
            userElement.innerHTML = `
                <div class="chat-user-avatar">
                    <img src="${user.avatar || 'default-avatar.png'}" alt="${user.nickname}">
                    <span class="chat-user-status ${user.isOnline ? 'status-online' : 'status-offline'}"></span>
                </div>
                <div class="chat-user-info">
                    <div class="chat-user-name">${user.nickname}</div>
                    <div class="chat-last-message">${user.lastMessage || ''}</div>
                </div>
                <div class="chat-message-meta">
                    ${user.lastMessageTime ? `<span class="chat-time">${new Date(user.lastMessageTime).toLocaleTimeString()}</span>` : ''}
                    ${user.unreadCount ? `<span class="chat-unread-count">${user.unreadCount}</span>` : ''}
                </div>
            `;
            
            userElement.addEventListener('click', () => this.selectUser(user.id));
            userList.appendChild(userElement);
        });
    }

    async selectUser(userId) {
        this.currentChatUser = userId;
        this.messageOffset = 0;
        
        // Clear current messages
        this.messageContainer.innerHTML = '';
        
        // Load chat history
        await this.loadChatHistory(userId);
        
        // Mark messages as read
        await this.markMessagesAsRead(userId);
        
        // Update UI to show selected user
        document.querySelectorAll('.chat-user-item').forEach(item => {
            item.classList.remove('active');
            if (item.dataset.userId === userId.toString()) {
                item.classList.add('active');
            }
        });
    }

    async markMessagesAsRead(senderId) {
        try {
            await fetch(`/api/chat/mark-read?sender_id=${senderId}`, {
                method: 'POST'
            });
        } catch (error) {
            console.error('Error marking messages as read:', error);
        }
    }

    // Utility functions
    throttle(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        }
    }

    debounce(func, wait) {
        let timeout;
        return function() {
            const context = this;
            const args = arguments;
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(context, args), wait);
        }
    }
}

// Initialize chat when the page loads
document.addEventListener('DOMContentLoaded', () => {
    const chat = new ChatManager();
}); 
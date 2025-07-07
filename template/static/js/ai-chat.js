class AIChatWidget {
    constructor() {
        this.isOpen = false;
        this.isTyping = false;
        this.init();
    }
    
    init() {
        this.createWidget();
        this.bindEvents();
        this.addWelcomeMessage();
    }
    
    createWidget() {
        const widget = document.createElement('div');
        widget.innerHTML = `
            <div class="ai-chat-widget" id="aiChatWidget">
                <div class="chat-header">
                    <i class="bi bi-robot"></i>
                    <span>AI购物助手</span>
                    <button class="btn-close" onclick="aiChat.toggle()">
                        <i class="bi bi-x"></i>
                    </button>
                </div>
                
                <div class="chat-messages" id="chatMessages">
                    <!-- 消息将在这里显示 -->
                </div>
                
                <div class="chat-input">
                    <input type="text" id="chatInput" placeholder="有什么可以帮您的吗？" maxlength="500">
                    <button onclick="aiChat.sendMessage()">
                        <i class="bi bi-send"></i>
                    </button>
                </div>
            </div>
            
            <button class="ai-chat-toggle" onclick="aiChat.toggle()">
                <i class="bi bi-chat-dots"></i>
            </button>
        `;
        
        document.body.appendChild(widget);
    }
    
    bindEvents() {
        const input = document.getElementById('chatInput');
        input.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });
    }
    
    toggle() {
        this.isOpen = !this.isOpen;
        const widget = document.getElementById('aiChatWidget');
        widget.classList.toggle('active', this.isOpen);
        
        if (this.isOpen) {
            document.getElementById('chatInput').focus();
        }
    }
    
    addWelcomeMessage() {
        const welcomeMsg = "您好！我是Ryan Mall的AI购物助手🛍️\n\n我可以帮您：\n• 推荐合适的商品\n• 解答商品相关问题\n• 协助购物决策\n• 介绍平台功能\n\n有什么可以帮您的吗？";
        this.addMessage('ai', welcomeMsg);
    }
    
    async sendMessage() {
        const input = document.getElementById('chatInput');
        const message = input.value.trim();
        
        if (!message || this.isTyping) return;
        
        // 检查登录状态
        const token = localStorage.getItem('token');
        if (!token) {
            this.addMessage('ai', '请先登录后再使用AI助手功能哦！');
            return;
        }
        
        // 添加用户消息
        this.addMessage('user', message);
        input.value = '';
        
        // 显示输入中状态
        this.showTyping();
        
        try {
            const response = await fetch('/api/v1/ai/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ message })
            });
            
            const data = await response.json();
            
            if (data.code === 200) {
                this.addMessage('ai', data.data.reply);
            } else {
                this.addMessage('ai', data.message || 'AI助手暂时不可用，请稍后再试');
            }
        } catch (error) {
            console.error('AI聊天错误:', error);
            this.addMessage('ai', '网络连接异常，请检查网络后重试');
        } finally {
            this.hideTyping();
        }
    }
    
    addMessage(type, content) {
        const messagesContainer = document.getElementById('chatMessages');
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${type}`;
        
        const time = new Date().toLocaleTimeString('zh-CN', { 
            hour: '2-digit', 
            minute: '2-digit' 
        });
        
        messageDiv.innerHTML = `
            <div class="message-content">${this.formatMessage(content)}</div>
            <div class="message-time">${time}</div>
        `;
        
        messagesContainer.appendChild(messageDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
    
    formatMessage(content) {
        // 简单的消息格式化
        return content
            .replace(/\n/g, '<br>')
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
            .replace(/\*(.*?)\*/g, '<em>$1</em>');
    }
    
    showTyping() {
        this.isTyping = true;
        const messagesContainer = document.getElementById('chatMessages');
        const typingDiv = document.createElement('div');
        typingDiv.className = 'message ai typing-indicator';
        typingDiv.id = 'typingIndicator';
        typingDiv.innerHTML = `
            <div class="message-content">
                AI助手正在输入
                <div class="typing-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                </div>
            </div>
        `;
        
        messagesContainer.appendChild(typingDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
    
    hideTyping() {
        this.isTyping = false;
        const typingIndicator = document.getElementById('typingIndicator');
        if (typingIndicator) {
            typingIndicator.remove();
        }
    }
}

// 页面加载完成后初始化AI聊天组件
document.addEventListener('DOMContentLoaded', function() {
    window.aiChat = new AIChatWidget();
});
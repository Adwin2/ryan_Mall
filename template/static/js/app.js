// Ryan Mall 前端应用工具函数

// API配置
const API_CONFIG = {
    BASE_URL: 'http://localhost:8081/api/v1',
    TIMEOUT: 10000
};

// 工具函数类
class Utils {
    // 格式化价格
    static formatPrice(price) {
        return `¥${parseFloat(price).toFixed(2)}`;
    }

    // 格式化日期
    static formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    // 防抖函数
    static debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    // 节流函数
    static throttle(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }

    // 生成随机ID
    static generateId() {
        return Math.random().toString(36).substr(2, 9);
    }

    // 验证邮箱格式
    static validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    }

    // 验证手机号格式
    static validatePhone(phone) {
        const re = /^1[3-9]\d{9}$/;
        return re.test(phone);
    }

    // 复制到剪贴板
    static async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            return true;
        } catch (err) {
            console.error('复制失败:', err);
            return false;
        }
    }

    // 显示Toast提示
    static showToast(message, type = 'info', duration = 3000) {
        const toastContainer = document.getElementById('toastContainer') || this.createToastContainer();
        
        const toast = document.createElement('div');
        toast.className = `toast align-items-center text-white bg-${type} border-0`;
        toast.setAttribute('role', 'alert');
        toast.innerHTML = `
            <div class="d-flex">
                <div class="toast-body">${message}</div>
                <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
            </div>
        `;

        toastContainer.appendChild(toast);
        
        const bsToast = new bootstrap.Toast(toast, { delay: duration });
        bsToast.show();

        // 自动移除
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, duration + 500);
    }

    // 创建Toast容器
    static createToastContainer() {
        const container = document.createElement('div');
        container.id = 'toastContainer';
        container.className = 'toast-container position-fixed top-0 end-0 p-3';
        container.style.zIndex = '9999';
        document.body.appendChild(container);
        return container;
    }

    // 显示确认对话框
    static showConfirm(message, title = '确认') {
        return new Promise((resolve) => {
            const modal = document.createElement('div');
            modal.className = 'modal fade';
            modal.innerHTML = `
                <div class="modal-dialog">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">${title}</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <div class="modal-body">
                            <p>${message}</p>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                            <button type="button" class="btn btn-primary" id="confirmBtn">确认</button>
                        </div>
                    </div>
                </div>
            `;

            document.body.appendChild(modal);
            const bsModal = new bootstrap.Modal(modal);
            
            modal.querySelector('#confirmBtn').addEventListener('click', () => {
                bsModal.hide();
                resolve(true);
            });

            modal.addEventListener('hidden.bs.modal', () => {
                document.body.removeChild(modal);
                resolve(false);
            });

            bsModal.show();
        });
    }
}

// API请求类
class ApiClient {
    constructor(baseURL = API_CONFIG.BASE_URL) {
        this.baseURL = baseURL;
    }

    // 获取认证头
    getAuthHeaders() {
        const token = localStorage.getItem('token');
        return token ? { 'Authorization': `Bearer ${token}` } : {};
    }

    // 通用请求方法
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...this.getAuthHeaders(),
                ...options.headers
            },
            ...options
        };

        console.log('API请求:', url, config); // 添加调试日志

        try {
            const response = await fetch(url, config);
            console.log('API响应状态:', response.status, response.statusText); // 添加调试日志

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            console.log('API响应数据:', data); // 添加调试日志

            if (data.code === 401) {
                // Token过期，清除本地存储并跳转到登录页
                localStorage.removeItem('token');
                localStorage.removeItem('username');
                window.location.href = '/views/login.html';
                throw new Error('登录已过期，请重新登录');
            }

            return data;
        } catch (error) {
            console.error('API请求失败:', {
                url: url,
                error: error.message,
                stack: error.stack
            });
            throw error;
        }
    }

    // GET请求
    async get(endpoint, params = {}) {
        const queryString = new URLSearchParams(params).toString();
        const url = queryString ? `${endpoint}?${queryString}` : endpoint;
        return this.request(url);
    }

    // POST请求
    async post(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    // PUT请求
    async put(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    // DELETE请求
    async delete(endpoint) {
        return this.request(endpoint, {
            method: 'DELETE'
        });
    }
}

// 全局API客户端实例
const api = new ApiClient();

// 认证管理类
class AuthManager {
    static isLoggedIn() {
        return !!localStorage.getItem('token');
    }

    static getUsername() {
        return localStorage.getItem('username');
    }

    static logout() {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        window.location.href = '/views/login.html';
    }

    static async checkAuth() {
        if (!this.isLoggedIn()) {
            return false;
        }

        try {
            const response = await api.get('/profile');
            return response.code === 200;
        } catch (error) {
            console.error('验证登录状态失败:', error);
            return false;
        }
    }
}

// 购物车管理类
class CartManager {
    static async getCartCount() {
        if (!AuthManager.isLoggedIn()) {
            return 0;
        }

        try {
            const response = await api.get('/cart');
            if (response.code === 200) {
                return response.data.items?.length || 0;
            }
        } catch (error) {
            console.error('获取购物车数量失败:', error);
        }
        return 0;
    }

    static async updateCartBadge() {
        const count = await this.getCartCount();
        const badge = document.getElementById('cartCount');
        if (badge) {
            badge.textContent = count;
            badge.style.display = count > 0 ? 'inline' : 'none';
        }
    }

    static async addToCart(productId, quantity = 1) {
        if (!AuthManager.isLoggedIn()) {
            Utils.showToast('请先登录', 'warning');
            return false;
        }

        try {
            const response = await api.post('/cart', {
                product_id: productId,
                quantity: quantity
            });

            if (response.code === 200) {
                Utils.showToast('添加到购物车成功', 'success');
                this.updateCartBadge();
                return true;
            } else {
                Utils.showToast(response.message || '添加失败', 'danger');
                return false;
            }
        } catch (error) {
            console.error('添加到购物车失败:', error);
            Utils.showToast('网络错误，请稍后重试', 'danger');
            return false;
        }
    }
}

// 页面加载完成后的初始化
document.addEventListener('DOMContentLoaded', function() {
    // 更新导航栏用户状态
    updateNavbarUserStatus();
    
    // 更新购物车徽章
    CartManager.updateCartBadge();
    
    // 添加全局错误处理
    window.addEventListener('unhandledrejection', function(event) {
        console.error('未处理的Promise错误:', event.reason);
        Utils.showToast('发生了一个错误，请刷新页面重试', 'danger');
    });
});

// 更新导航栏用户状态
function updateNavbarUserStatus() {
    const loginLink = document.getElementById('loginLink');
    const userMenu = document.getElementById('userMenu');
    const usernameEl = document.getElementById('username');

    if (AuthManager.isLoggedIn()) {
        if (loginLink) loginLink.classList.add('d-none');
        if (userMenu) userMenu.classList.remove('d-none');
        if (usernameEl) usernameEl.textContent = AuthManager.getUsername();
    } else {
        if (loginLink) loginLink.classList.remove('d-none');
        if (userMenu) userMenu.classList.add('d-none');
    }
}

// 全局退出登录函数
function logout() {
    AuthManager.logout();
}

// 导出到全局作用域
window.Utils = Utils;
window.api = api;
window.AuthManager = AuthManager;
window.CartManager = CartManager;

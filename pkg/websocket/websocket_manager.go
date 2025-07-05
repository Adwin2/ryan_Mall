package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketManager WebSocket管理器
type WebSocketManager struct {
	clients    map[uint]*Client          // 用户ID -> 客户端连接
	rooms      map[string]map[uint]*Client // 房间 -> 用户连接
	register   chan *Client              // 注册客户端
	unregister chan *Client              // 注销客户端
	broadcast  chan *Message             // 广播消息
	mutex      sync.RWMutex
}

// Client 客户端连接
type Client struct {
	ID       uint                // 用户ID
	Conn     *websocket.Conn     // WebSocket连接
	Send     chan *Message       // 发送消息通道
	Manager  *WebSocketManager   // 管理器引用
	Rooms    map[string]bool     // 加入的房间
	LastPing time.Time           // 最后ping时间
}

// Message 消息结构
type Message struct {
	Type      string      `json:"type"`       // 消息类型
	From      uint        `json:"from"`       // 发送者ID
	To        uint        `json:"to"`         // 接收者ID (0表示广播)
	Room      string      `json:"room"`       // 房间名称
	Content   interface{} `json:"content"`    // 消息内容
	Timestamp time.Time   `json:"timestamp"`  // 时间戳
}

// MessageType 消息类型常量
const (
	MessageTypeNotification = "notification"  // 通知消息
	MessageTypeChat         = "chat"          // 聊天消息
	MessageTypeOrderUpdate  = "order_update"  // 订单更新
	MessageTypeStockUpdate  = "stock_update"  // 库存更新
	MessageTypeSystemAlert  = "system_alert"  // 系统警告
	MessageTypePing         = "ping"          // 心跳
	MessageTypePong         = "pong"          // 心跳响应
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// NewWebSocketManager 创建WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[uint]*Client),
		rooms:      make(map[string]map[uint]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
	}
}

// Run 启动WebSocket管理器
func (wsm *WebSocketManager) Run() {
	// 启动心跳检查
	go wsm.heartbeatChecker()
	
	for {
		select {
		case client := <-wsm.register:
			wsm.registerClient(client)
			
		case client := <-wsm.unregister:
			wsm.unregisterClient(client)
			
		case message := <-wsm.broadcast:
			wsm.broadcastMessage(message)
		}
	}
}

// HandleWebSocket 处理WebSocket连接
func (wsm *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID uint) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	client := &Client{
		ID:       userID,
		Conn:     conn,
		Send:     make(chan *Message, 256),
		Manager:  wsm,
		Rooms:    make(map[string]bool),
		LastPing: time.Now(),
	}
	
	wsm.register <- client
	
	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// registerClient 注册客户端
func (wsm *WebSocketManager) registerClient(client *Client) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	
	// 如果用户已经有连接，关闭旧连接
	if oldClient, exists := wsm.clients[client.ID]; exists {
		close(oldClient.Send)
		oldClient.Conn.Close()
	}
	
	wsm.clients[client.ID] = client
	log.Printf("User %d connected", client.ID)
	
	// 发送欢迎消息
	welcomeMsg := &Message{
		Type:      MessageTypeNotification,
		Content:   "欢迎来到Ryan Mall！",
		Timestamp: time.Now(),
	}
	
	select {
	case client.Send <- welcomeMsg:
	default:
		close(client.Send)
		delete(wsm.clients, client.ID)
	}
}

// unregisterClient 注销客户端
func (wsm *WebSocketManager) unregisterClient(client *Client) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	
	if _, exists := wsm.clients[client.ID]; exists {
		delete(wsm.clients, client.ID)
		close(client.Send)
		
		// 从所有房间中移除
		for room := range client.Rooms {
			wsm.leaveRoom(client, room)
		}
		
		log.Printf("User %d disconnected", client.ID)
	}
}

// broadcastMessage 广播消息
func (wsm *WebSocketManager) broadcastMessage(message *Message) {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()
	
	if message.To > 0 {
		// 发送给特定用户
		if client, exists := wsm.clients[message.To]; exists {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(wsm.clients, message.To)
			}
		}
	} else if message.Room != "" {
		// 发送给房间内所有用户
		if room, exists := wsm.rooms[message.Room]; exists {
			for _, client := range room {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(wsm.clients, client.ID)
				}
			}
		}
	} else {
		// 广播给所有用户
		for userID, client := range wsm.clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(wsm.clients, userID)
			}
		}
	}
}

// SendToUser 发送消息给特定用户
func (wsm *WebSocketManager) SendToUser(userID uint, msgType string, content interface{}) {
	message := &Message{
		Type:      msgType,
		To:        userID,
		Content:   content,
		Timestamp: time.Now(),
	}
	
	wsm.broadcast <- message
}

// SendToRoom 发送消息给房间
func (wsm *WebSocketManager) SendToRoom(room, msgType string, content interface{}) {
	message := &Message{
		Type:      msgType,
		Room:      room,
		Content:   content,
		Timestamp: time.Now(),
	}
	
	wsm.broadcast <- message
}

// BroadcastToAll 广播消息给所有用户
func (wsm *WebSocketManager) BroadcastToAll(msgType string, content interface{}) {
	message := &Message{
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
	}
	
	wsm.broadcast <- message
}

// JoinRoom 加入房间
func (wsm *WebSocketManager) JoinRoom(userID uint, room string) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	
	if client, exists := wsm.clients[userID]; exists {
		if wsm.rooms[room] == nil {
			wsm.rooms[room] = make(map[uint]*Client)
		}
		
		wsm.rooms[room][userID] = client
		client.Rooms[room] = true
		
		log.Printf("User %d joined room %s", userID, room)
	}
}

// LeaveRoom 离开房间
func (wsm *WebSocketManager) LeaveRoom(userID uint, room string) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	
	if client, exists := wsm.clients[userID]; exists {
		wsm.leaveRoom(client, room)
	}
}

// leaveRoom 内部离开房间方法
func (wsm *WebSocketManager) leaveRoom(client *Client, room string) {
	if roomClients, exists := wsm.rooms[room]; exists {
		delete(roomClients, client.ID)
		delete(client.Rooms, room)
		
		// 如果房间为空，删除房间
		if len(roomClients) == 0 {
			delete(wsm.rooms, room)
		}
		
		log.Printf("User %d left room %s", client.ID, room)
	}
}

// GetOnlineUsers 获取在线用户数
func (wsm *WebSocketManager) GetOnlineUsers() int {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()
	
	return len(wsm.clients)
}

// IsUserOnline 检查用户是否在线
func (wsm *WebSocketManager) IsUserOnline(userID uint) bool {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()
	
	_, exists := wsm.clients[userID]
	return exists
}

// heartbeatChecker 心跳检查
func (wsm *WebSocketManager) heartbeatChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			wsm.mutex.RLock()
			var disconnectedClients []*Client
			
			for _, client := range wsm.clients {
				if time.Since(client.LastPing) > 60*time.Second {
					disconnectedClients = append(disconnectedClients, client)
				}
			}
			wsm.mutex.RUnlock()
			
			// 断开超时连接
			for _, client := range disconnectedClients {
				wsm.unregister <- client
			}
		}
	}
}

// readPump 读取消息
func (c *Client) readPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.LastPing = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, messageData, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		var message Message
		if err := json.Unmarshal(messageData, &message); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}
		
		message.From = c.ID
		message.Timestamp = time.Now()
		
		// 处理不同类型的消息
		switch message.Type {
		case MessageTypePing:
			c.LastPing = time.Now()
			pongMsg := &Message{
				Type:      MessageTypePong,
				Timestamp: time.Now(),
			}
			select {
			case c.Send <- pongMsg:
			default:
				return
			}
		case MessageTypeChat:
			// 聊天消息处理
			c.Manager.broadcast <- &message
		default:
			// 其他消息类型处理
			c.Manager.broadcast <- &message
		}
	}
}

// writePump 发送消息
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			messageData, err := json.Marshal(message)
			if err != nil {
				log.Printf("JSON marshal error: %v", err)
				return
			}
			
			if err := c.Conn.WriteMessage(websocket.TextMessage, messageData); err != nil {
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// NotificationService 通知服务
type NotificationService struct {
	wsManager *WebSocketManager
}

// NewNotificationService 创建通知服务
func NewNotificationService(wsManager *WebSocketManager) *NotificationService {
	return &NotificationService{
		wsManager: wsManager,
	}
}

// OrderStatusNotification 订单状态通知
type OrderStatusNotification struct {
	OrderID     uint   `json:"order_id"`
	OrderNo     string `json:"order_no"`
	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	Message     string `json:"message"`
}

// StockUpdateNotification 库存更新通知
type StockUpdateNotification struct {
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
	OldStock    int    `json:"old_stock"`
	NewStock    int    `json:"new_stock"`
	Message     string `json:"message"`
}

// SendOrderStatusNotification 发送订单状态通知
func (ns *NotificationService) SendOrderStatusNotification(userID uint, notification *OrderStatusNotification) {
	ns.wsManager.SendToUser(userID, MessageTypeOrderUpdate, notification)
}

// SendStockUpdateNotification 发送库存更新通知
func (ns *NotificationService) SendStockUpdateNotification(notification *StockUpdateNotification) {
	// 广播给所有在线用户
	ns.wsManager.BroadcastToAll(MessageTypeStockUpdate, notification)
}

// SendSystemAlert 发送系统警告
func (ns *NotificationService) SendSystemAlert(message string) {
	alert := map[string]interface{}{
		"level":   "warning",
		"message": message,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	}
	
	ns.wsManager.BroadcastToAll(MessageTypeSystemAlert, alert)
}

// SendPersonalNotification 发送个人通知
func (ns *NotificationService) SendPersonalNotification(userID uint, title, message string) {
	notification := map[string]interface{}{
		"title":   title,
		"message": message,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	}
	
	ns.wsManager.SendToUser(userID, MessageTypeNotification, notification)
}

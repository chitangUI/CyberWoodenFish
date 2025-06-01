package handlers

import (
	"electronic-muyu-backend/internal/services"
	"electronic-muyu-backend/internal/websocket"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 在生产环境中应该验证来源
	},
}

type LeaderboardWebSocketHandler struct {
	hub                *websocket.Hub
	broadcastService   *services.LeaderboardBroadcastService
	leaderboardService *services.LeaderboardService
	logger             *services.Logger
}

type LeaderboardSubscription struct {
	Type     string   `json:"type"`     // "subscribe" or "unsubscribe"
	Channels []string `json:"channels"` // ["global", "daily", "weekly", "achievements", "rank_changes"]
}

type LeaderboardWebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewLeaderboardWebSocketHandler(
	hub *websocket.Hub,
	broadcastService   *services.LeaderboardBroadcastService,
	leaderboardService *services.LeaderboardService,
	logger *services.Logger,
) *LeaderboardWebSocketHandler {
	return &LeaderboardWebSocketHandler{
		hub:                hub,
		broadcastService:   broadcastService,
		leaderboardService: leaderboardService,
		logger:             logger,
	}
}

func (h *LeaderboardWebSocketHandler) HandleLeaderboardWebSocket(c *gin.Context) {
	// 升级到WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to websocket: %v", err)
		return
	}

	// 从JWT中获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		conn.Close()
		return
	}

	username, _ := c.Get("username")
	userIDUint := userID.(uint)
	usernameStr := ""
	if username != nil {
		usernameStr = username.(string)
	}

	// 创建客户端
	client := &websocket.Client{
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userIDUint,
		Username: usernameStr,
		RoomID:   "leaderboard", // 使用特殊的排行榜房间
		Hub:      h.hub,
	}

	// 注册客户端
	h.hub.Register <- client

	// 启动消息处理
	go h.handleMessages(client)
	go h.writePump(client)
	h.readPump(client)
}

func (h *LeaderboardWebSocketHandler) readPump(client *websocket.Client) {
	defer func() {
		h.hub.Unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				h.logger.Error("WebSocket error: %v", err)
			}
			break
		}

		// 解析客户端消息
		var subscription LeaderboardSubscription
		if err := json.Unmarshal(message, &subscription); err != nil {
			h.logger.Error("Failed to parse subscription message: %v", err)
			continue
		}

		// 处理订阅请求
		h.handleSubscription(client, subscription)
	}
}

func (h *LeaderboardWebSocketHandler) writePump(client *websocket.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 发送队列中的其他消息
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *LeaderboardWebSocketHandler) handleSubscription(client *websocket.Client, subscription LeaderboardSubscription) {
	switch subscription.Type {
	case "subscribe":
		h.subscribeToChannels(client, subscription.Channels)
	case "unsubscribe":
		h.unsubscribeFromChannels(client, subscription.Channels)
	case "get_current":
		h.sendCurrentLeaderboard(client, subscription.Channels)
	default:
		h.logger.Warn("Unknown subscription type: %s", subscription.Type)
	}
}

func (h *LeaderboardWebSocketHandler) subscribeToChannels(client *websocket.Client, channels []string) {
	// 发送确认消息
	response := LeaderboardWebSocketMessage{
		Type: "subscription_confirmed",
		Data: map[string]interface{}{
			"channels": channels,
			"status":   "subscribed",
		},
	}

	h.sendToClient(client, response)

	// 发送当前排行榜数据
	h.sendCurrentLeaderboard(client, channels)
}

func (h *LeaderboardWebSocketHandler) unsubscribeFromChannels(client *websocket.Client, channels []string) {
	response := LeaderboardWebSocketMessage{
		Type: "subscription_confirmed",
		Data: map[string]interface{}{
			"channels": channels,
			"status":   "unsubscribed",
		},
	}

	h.sendToClient(client, response)
}

func (h *LeaderboardWebSocketHandler) sendCurrentLeaderboard(client *websocket.Client, channels []string) {
	for _, channel := range channels {
		switch channel {
		case "global":
			entries, err := h.leaderboardService.GetGlobalLeaderboard(10, 0)
			if err == nil {
				response := LeaderboardWebSocketMessage{
					Type: "current_leaderboard",
					Data: map[string]interface{}{
						"type":    "global",
						"entries": entries,
					},
				}
				h.sendToClient(client, response)
			}

		case "daily":
			entries, err := h.leaderboardService.GetDailyLeaderboard(10, 0)
			if err == nil {
				response := LeaderboardWebSocketMessage{
					Type: "current_leaderboard",
					Data: map[string]interface{}{
						"type":    "daily",
						"entries": entries,
					},
				}
				h.sendToClient(client, response)
			}

		case "weekly":
			entries, err := h.leaderboardService.GetWeeklyLeaderboard(10, 0)
			if err == nil {
				response := LeaderboardWebSocketMessage{
					Type: "current_leaderboard",
					Data: map[string]interface{}{
						"type":    "weekly",
						"entries": entries,
					},
				}
				h.sendToClient(client, response)
			}
		}
	}
}

func (h *LeaderboardWebSocketHandler) sendToClient(client *websocket.Client, message LeaderboardWebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal WebSocket message: %v", err)
		return
	}

	select {
	case client.Send <- data:
	default:
		close(client.Send)
	}
}

func (h *LeaderboardWebSocketHandler) handleMessages(client *websocket.Client) {
	// 如果广播服务可用，订阅Redis频道
	if h.broadcastService != nil {
		channels := []string{
			"leaderboard:updates",
			"leaderboard:global",
			"leaderboard:daily",
			"leaderboard:weekly",
			"leaderboard:achievements",
			"leaderboard:rank_changes",
		}

		pubsub, err := h.broadcastService.SubscribeToLeaderboardUpdates(channels)
		if err != nil {
			h.logger.Error("Failed to subscribe to Redis channels: %v", err)
			return
		}
		defer pubsub.Close()

		// 监听Redis消息并转发给WebSocket客户端
		for msg := range pubsub.GetChannel() {
			var update services.LeaderboardUpdate
			if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
				h.logger.Error("Failed to parse Redis message: %v", err)
				continue
			}

			// 转发给WebSocket客户端
			wsMessage := LeaderboardWebSocketMessage{
				Type: "live_update",
				Data: update,
			}

			h.sendToClient(client, wsMessage)
		}
	}
}

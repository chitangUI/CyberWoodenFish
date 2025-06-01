package handlers

import (
	"electronic-muyu-backend/internal/websocket"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

type GameHandler struct {
	hub *websocket.Hub
}

func NewGameHandler(hub *websocket.Hub) *GameHandler {
	return &GameHandler{
		hub: hub,
	}
}

func (h *GameHandler) HandleWebSocket(c *gin.Context) {
	// 从认证中间件获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found"})
		return
	}

	// 获取房间ID，如果没有提供则使用默认房间
	roomID := c.DefaultQuery("room_id", "default")

	// 升级HTTP连接为WebSocket连接
	upgrader := ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 在生产环境中应该验证来源
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// 创建客户端
	client := &websocket.Client{
		UserID:   userID.(uint),
		Username: username.(string),
		RoomID:   roomID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
	}

	// 注册客户端到hub
	h.hub.Register <- client

	// 启动goroutines处理读写
	go client.WritePump()
	go client.ReadPump()
}

func (h *GameHandler) GetRoomInfo(c *gin.Context) {
	roomID := c.Param("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID is required"})
		return
	}

	roomInfo := h.hub.GetRoomInfo(roomID)
	c.JSON(http.StatusOK, roomInfo)
}

func (h *GameHandler) GetActiveRooms(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	activeRooms := h.hub.GetActiveRooms(page, limit)

	c.JSON(http.StatusOK, gin.H{
		"rooms": activeRooms,
		"page":  page,
		"limit": limit,
	})
}

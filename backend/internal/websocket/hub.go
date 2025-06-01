package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 在生产环境中应该验证来源
	},
}

// Hub 管理所有活跃的连接和房间
type Hub struct {
	// 注册的客户端
	clients map[*Client]bool

	// 按房间分组的客户端
	rooms map[string]map[*Client]bool

	// 来自客户端的消息
	broadcast chan []byte

	// 注册客户端请求
	Register chan *Client

	// 注销客户端请求
	Unregister chan *Client

	// 房间消息
	roomBroadcast chan *RoomMessage

	mutex sync.RWMutex
}

type RoomMessage struct {
	RoomID  string `json:"room_id"`
	Message []byte `json:"message"`
}

type Client struct {
	// WebSocket 连接
	Conn *websocket.Conn

	// 发送消息的通道
	Send chan []byte

	// 用户信息
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoomID   string `json:"room_id"`

	// Hub 引用
	Hub *Hub
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type GameUpdate struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	CurrentScore int64  `json:"current_score"`
	Action       string `json:"action"` // "score_update", "joined", "left"
	Timestamp    int64  `json:"timestamp"`
}

type RoomInfo struct {
	RoomID      string       `json:"room_id"`
	PlayerCount int          `json:"player_count"`
	Players     []PlayerInfo `json:"players"`
}

type PlayerInfo struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	CurrentScore int64  `json:"current_score"`
	IsOnline     bool   `json:"is_online"`
}

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		broadcast:     make(chan []byte),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		roomBroadcast: make(chan *RoomMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAll(message)

		case roomMessage := <-h.roomBroadcast:
			h.broadcastToRoom(roomMessage.RoomID, roomMessage.Message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true

	// 加入房间
	if h.rooms[client.RoomID] == nil {
		h.rooms[client.RoomID] = make(map[*Client]bool)
	}
	h.rooms[client.RoomID][client] = true

	log.Printf("User %s joined room %s", client.Username, client.RoomID)

	// 通知房间内其他玩家有新玩家加入
	joinMessage := Message{
		Type: "player_joined",
		Data: GameUpdate{
			UserID:    client.UserID,
			Username:  client.Username,
			Action:    "joined",
			Timestamp: getCurrentTimestamp(),
		},
	}

	messageBytes, _ := json.Marshal(joinMessage)
	h.broadcastToRoomExcept(client.RoomID, messageBytes, client)

	// 发送房间信息给新加入的客户端
	roomInfo := h.GetRoomInfo(client.RoomID)
	roomInfoMessage := Message{
		Type: "room_info",
		Data: roomInfo,
	}

	roomInfoBytes, _ := json.Marshal(roomInfoMessage)
	select {
	case client.Send <- roomInfoBytes:
	default:
		close(client.Send)
		delete(h.clients, client)
		delete(h.rooms[client.RoomID], client)
	}
}

func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		delete(h.rooms[client.RoomID], client)
		close(client.Send)

		log.Printf("User %s left room %s", client.Username, client.RoomID)

		// 通知房间内其他玩家有玩家离开
		leaveMessage := Message{
			Type: "player_left",
			Data: GameUpdate{
				UserID:    client.UserID,
				Username:  client.Username,
				Action:    "left",
				Timestamp: getCurrentTimestamp(),
			},
		}

		messageBytes, _ := json.Marshal(leaveMessage)
		h.broadcastToRoom(client.RoomID, messageBytes)

		// 如果房间为空，清理房间
		if len(h.rooms[client.RoomID]) == 0 {
			delete(h.rooms, client.RoomID)
		}
	}
}

func (h *Hub) broadcastToAll(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, client)
			delete(h.rooms[client.RoomID], client)
		}
	}
}

func (h *Hub) broadcastToRoom(roomID string, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if room, exists := h.rooms[roomID]; exists {
		for client := range room {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client)
				delete(room, client)
			}
		}
	}
}

func (h *Hub) broadcastToRoomExcept(roomID string, message []byte, except *Client) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if room, exists := h.rooms[roomID]; exists {
		for client := range room {
			if client != except {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
					delete(room, client)
				}
			}
		}
	}
}

func (h *Hub) GetRoomInfo(roomID string) RoomInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	room := h.rooms[roomID]
	players := make([]PlayerInfo, 0, len(room))

	for client := range room {
		players = append(players, PlayerInfo{
			UserID:       client.UserID,
			Username:     client.Username,
			CurrentScore: 0, // 当前分数需要从游戏状态中获取
			IsOnline:     true,
		})
	}

	return RoomInfo{
		RoomID:      roomID,
		PlayerCount: len(players),
		Players:     players,
	}
}

func (h *Hub) GetActiveRooms(page, limit int) []RoomInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	rooms := make([]RoomInfo, 0)
	offset := (page - 1) * limit
	count := 0

	for roomID, clients := range h.rooms {
		if len(clients) > 0 {
			if count >= offset && len(rooms) < limit {
				roomInfo := h.GetRoomInfo(roomID)
				rooms = append(rooms, roomInfo)
			}
			count++
		}
	}

	return rooms
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

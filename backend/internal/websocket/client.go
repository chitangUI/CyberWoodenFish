package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写入消息到对等方的时间限制
	writeWait = 10 * time.Second

	// 从对等方读取下一个 pong 消息的时间限制
	pongWait = 60 * time.Second

	// 在此期间内向对等方发送 ping。必须小于 pongWait
	pingPeriod = (pongWait * 9) / 10

	// 允许的最大消息大小
	maxMessageSize = 512
)

// 处理客户端消息的类型
type ClientMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ScoreUpdateData struct {
	Score     int64 `json:"score"`
	Increment int64 `json:"increment"`
}

// 从 websocket 连接读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// 解析消息
		var msg ClientMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// 处理不同类型的消息
		c.handleMessage(msg)
	}
}

// 向 websocket 连接写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 添加排队的聊天消息到当前的 websocket 消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 处理客户端消息
func (c *Client) handleMessage(msg ClientMessage) {
	switch msg.Type {
	case "score_update":
		c.handleScoreUpdate(msg.Data)
	case "ping":
		c.handlePing()
	case "join_room":
		c.handleJoinRoom(msg.Data)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// 处理分数更新
func (c *Client) handleScoreUpdate(data json.RawMessage) {
	var scoreData ScoreUpdateData
	if err := json.Unmarshal(data, &scoreData); err != nil {
		log.Printf("Failed to parse score update: %v", err)
		return
	}

	// 广播分数更新给房间内的其他玩家
	update := GameUpdate{
		UserID:       c.UserID,
		Username:     c.Username,
		CurrentScore: scoreData.Score,
		Action:       "score_update",
		Timestamp:    getCurrentTimestamp(),
	}

	message := Message{
		Type: "score_update",
		Data: update,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal score update: %v", err)
		return
	}

	// 发送给房间内其他玩家
	roomMessage := &RoomMessage{
		RoomID:  c.RoomID,
		Message: messageBytes,
	}

	select {
	case c.Hub.roomBroadcast <- roomMessage:
	default:
		log.Printf("Failed to send room broadcast")
	}
}

// 处理 ping 消息
func (c *Client) handlePing() {
	pongMessage := Message{
		Type: "pong",
		Data: map[string]interface{}{
			"timestamp": getCurrentTimestamp(),
		},
	}

	messageBytes, err := json.Marshal(pongMessage)
	if err != nil {
		return
	}

	select {
	case c.Send <- messageBytes:
	default:
		close(c.Send)
	}
}

// 处理加入房间请求
func (c *Client) handleJoinRoom(data json.RawMessage) {
	var roomData struct {
		RoomID string `json:"room_id"`
	}

	if err := json.Unmarshal(data, &roomData); err != nil {
		log.Printf("Failed to parse join room request: %v", err)
		return
	}

	// 离开当前房间
	if c.RoomID != "" {
		delete(c.Hub.rooms[c.RoomID], c)
		if len(c.Hub.rooms[c.RoomID]) == 0 {
			delete(c.Hub.rooms, c.RoomID)
		}
	}

	// 加入新房间
	c.RoomID = roomData.RoomID
	if c.Hub.rooms[c.RoomID] == nil {
		c.Hub.rooms[c.RoomID] = make(map[*Client]bool)
	}
	c.Hub.rooms[c.RoomID][c] = true

	// 发送房间信息
	roomInfo := c.Hub.GetRoomInfo(c.RoomID)
	roomInfoMessage := Message{
		Type: "room_info",
		Data: roomInfo,
	}

	messageBytes, _ := json.Marshal(roomInfoMessage)
	select {
	case c.Send <- messageBytes:
	default:
		close(c.Send)
	}
}

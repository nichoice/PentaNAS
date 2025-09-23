package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"pnas/internal/app/dto"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketHub manages WebSocket connections and message broadcasting
type WebSocketHub struct {
	clients         map[string]*WebSocketClient
	broadcast       chan dto.WebSocketMessage
	register        chan *WebSocketClient
	unregister      chan *WebSocketClient
	monitoringData  chan *dto.MonitoringData
	mu              sync.RWMutex
	upgrader        websocket.Upgrader
	ctx             context.Context
	cancel          context.CancelFunc
	monitoringService *MonitoringService
}

// WebSocketClient represents a WebSocket client connection
type WebSocketClient struct {
	ID         string
	Conn       *websocket.Conn
	Send       chan dto.WebSocketMessage
	Hub        *WebSocketHub
	RemoteAddr string
	UserAgent  string
	ConnectedAt time.Time
	LastPing   time.Time
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(monitoringService *MonitoringService) *WebSocketHub {
	ctx, cancel := context.WithCancel(context.Background())

	hub := &WebSocketHub{
		clients:       make(map[string]*WebSocketClient),
		broadcast:     make(chan dto.WebSocketMessage, 256),
		register:      make(chan *WebSocketClient),
		unregister:    make(chan *WebSocketClient),
		monitoringData: make(chan *dto.MonitoringData, 10),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from any origin
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		ctx:               ctx,
		cancel:            cancel,
		monitoringService: monitoringService,
	}

	return hub
}

// Start starts the WebSocket hub
func (hub *WebSocketHub) Start() {
	go hub.run()
	go hub.startMonitoringBroadcast()
}

// Stop stops the WebSocket hub
func (hub *WebSocketHub) Stop() {
	hub.cancel()
	close(hub.broadcast)
	close(hub.register)
	close(hub.unregister)
	close(hub.monitoringData)
}

// run handles the main WebSocket hub loop
func (hub *WebSocketHub) run() {
	for {
		select {
		case client := <-hub.register:
			hub.mu.Lock()
			hub.clients[client.ID] = client
			hub.mu.Unlock()

			log.Printf("WebSocket client connected: %s from %s", client.ID, client.RemoteAddr)

			// Send welcome message
			welcomeMsg := dto.WebSocketMessage{
				Type:      "welcome",
				Data:      map[string]string{"client_id": client.ID, "message": "Connected to monitoring service"},
				Timestamp: time.Now(),
				ClientID:  client.ID,
			}
			client.Send <- welcomeMsg

		case client := <-hub.unregister:
			hub.mu.Lock()
			if _, ok := hub.clients[client.ID]; ok {
				delete(hub.clients, client.ID)
				close(client.Send)
				log.Printf("WebSocket client disconnected: %s", client.ID)
			}
			hub.mu.Unlock()

		case message := <-hub.broadcast:
			hub.mu.RLock()
			for clientID, client := range hub.clients {
				select {
				case client.Send <- message:
				default:
					delete(hub.clients, clientID)
					close(client.Send)
				}
			}
			hub.mu.RUnlock()

		case data := <-hub.monitoringData:
			message := dto.WebSocketMessage{
				Type:      "monitoring_data",
				Data:      data,
				Timestamp: time.Now(),
			}
			hub.broadcast <- message

		case <-hub.ctx.Done():
			return
		}
	}
}

// startMonitoringBroadcast starts broadcasting monitoring data
func (hub *WebSocketHub) startMonitoringBroadcast() {
	ticker := time.NewTicker(5 * time.Second) // Broadcast every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data, err := hub.monitoringService.GetCompleteMonitoringData()
			if err != nil {
				log.Printf("Error getting monitoring data: %v", err)
				continue
			}

			select {
			case hub.monitoringData <- data:
			default:
				// Channel is full, skip this update
			}

		case <-hub.ctx.Done():
			return
		}
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (hub *WebSocketHub) HandleWebSocket(c *gin.Context) {
	conn, err := hub.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WebSocketClient{
		ID:          uuid.New().String(),
		Conn:        conn,
		Send:        make(chan dto.WebSocketMessage, 256),
		Hub:         hub,
		RemoteAddr:  c.Request.RemoteAddr,
		UserAgent:   c.Request.UserAgent(),
		ConnectedAt: time.Now(),
		LastPing:    time.Now(),
	}

	client.Hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// writePump pumps messages from the hub to the websocket connection
func (client *WebSocketClient) writePump() {
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
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (client *WebSocketClient) readPump() {
	defer func() {
		client.Hub.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.LastPing = time.Now()
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message dto.WebSocketMessage
		err := client.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages
		client.handleMessage(message)
	}
}

// handleMessage handles incoming WebSocket messages
func (client *WebSocketClient) handleMessage(message dto.WebSocketMessage) {
	switch message.Type {
	case "ping":
		response := dto.WebSocketMessage{
			Type:      "pong",
			Data:      map[string]interface{}{"timestamp": time.Now()},
			Timestamp: time.Now(),
			ClientID:  client.ID,
		}
		client.Send <- response

	case "subscribe":
		// Handle subscription to specific monitoring data
		if data, ok := message.Data.(map[string]interface{}); ok {
			if subscription, ok := data["subscription"].(string); ok {
				client.handleSubscription(subscription)
			}
		}

	case "request_data":
		// Handle one-time data requests
		if data, ok := message.Data.(map[string]interface{}); ok {
			if dataType, ok := data["type"].(string); ok {
				client.handleDataRequest(dataType)
			}
		}

	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

// handleSubscription handles subscription requests
func (client *WebSocketClient) handleSubscription(subscription string) {
	switch subscription {
	case "system_info":
		go client.subscribeToSystemInfo()
	case "cpu":
		go client.subscribeToCPU()
	case "memory":
		go client.subscribeToMemory()
	case "disk":
		go client.subscribeToDisk()
	case "network":
		go client.subscribeToNetwork()
	case "storage_protocols":
		go client.subscribeToStorageProtocols()
	case "all":
		// Already handled by the main monitoring broadcast
	default:
		response := dto.WebSocketMessage{
			Type:      "error",
			Data:      map[string]string{"error": "Unknown subscription type"},
			Timestamp: time.Now(),
			ClientID:  client.ID,
		}
		client.Send <- response
	}
}

// handleDataRequest handles one-time data requests
func (client *WebSocketClient) handleDataRequest(dataType string) {
	var data interface{}
	var err error

	switch dataType {
	case "system_info":
		data, err = client.Hub.monitoringService.GetSystemInfo()
	case "cpu":
		data, err = client.Hub.monitoringService.GetCPUInfo()
	case "memory":
		data, err = client.Hub.monitoringService.GetMemoryInfo()
	case "disk":
		data, err = client.Hub.monitoringService.GetDiskInfo()
	case "network":
		data, err = client.Hub.monitoringService.GetNetworkInfo()
	case "storage_protocols":
		data, err = client.Hub.monitoringService.GetStorageProtocolInfo()
	case "complete":
		data, err = client.Hub.monitoringService.GetCompleteMonitoringData()
	default:
		err = fmt.Errorf("unknown data type: %s", dataType)
	}

	response := dto.WebSocketMessage{
		Type:      "data_response",
		Timestamp: time.Now(),
		ClientID:  client.ID,
	}

	if err != nil {
		response.Data = map[string]interface{}{
			"error": err.Error(),
			"type":  dataType,
		}
	} else {
		response.Data = map[string]interface{}{
			"type": dataType,
			"data": data,
		}
	}

	client.Send <- response
}

// Subscribe to specific data types with custom intervals

func (client *WebSocketClient) subscribeToSystemInfo() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data, err := client.Hub.monitoringService.GetSystemInfo()
			if err != nil {
				continue
			}

			message := dto.WebSocketMessage{
				Type:      "system_info",
				Data:      data,
				Timestamp: time.Now(),
				ClientID:  client.ID,
			}

			select {
			case client.Send <- message:
			default:
				return
			}

		case <-client.Hub.ctx.Done():
			return
		}
	}
}

func (client *WebSocketClient) subscribeToCPU() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data, err := client.Hub.monitoringService.GetCPUInfo()
			if err != nil {
				continue
			}

			message := dto.WebSocketMessage{
				Type:      "cpu",
				Data:      data,
				Timestamp: time.Now(),
				ClientID:  client.ID,
			}

			select {
			case client.Send <- message:
			default:
				return
			}

		case <-client.Hub.ctx.Done():
			return
		}
	}
}

func (client *WebSocketClient) subscribeToMemory() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data, err := client.Hub.monitoringService.GetMemoryInfo()
			if err != nil {
				continue
			}

			message := dto.WebSocketMessage{
				Type:      "memory",
				Data:      data,
				Timestamp: time.Now(),
				ClientID:  client.ID,
			}

			select {
			case client.Send <- message:
			default:
				return
			}

		case <-client.Hub.ctx.Done():
			return
		}
	}
}

func (client *WebSocketClient) subscribeToDisk() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			diskInfo, err := client.Hub.monitoringService.GetDiskInfo()
			if err != nil {
				continue
			}

			diskIOInfo, err := client.Hub.monitoringService.GetDiskIOInfo()
			if err != nil {
				continue
			}

			data := map[string]interface{}{
				"disks":   diskInfo,
				"disk_io": diskIOInfo,
			}

			message := dto.WebSocketMessage{
				Type:      "disk",
				Data:      data,
				Timestamp: time.Now(),
				ClientID:  client.ID,
			}

			select {
			case client.Send <- message:
			default:
				return
			}

		case <-client.Hub.ctx.Done():
			return
		}
	}
}

func (client *WebSocketClient) subscribeToNetwork() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			networkInfo, err := client.Hub.monitoringService.GetNetworkInfo()
			if err != nil {
				continue
			}

			networkIOInfo, err := client.Hub.monitoringService.GetNetworkIOInfo()
			if err != nil {
				continue
			}

			data := map[string]interface{}{
				"networks":   networkInfo,
				"network_io": networkIOInfo,
			}

			message := dto.WebSocketMessage{
				Type:      "network",
				Data:      data,
				Timestamp: time.Now(),
				ClientID:  client.ID,
			}

			select {
			case client.Send <- message:
			default:
				return
			}

		case <-client.Hub.ctx.Done():
			return
		}
	}
}

func (client *WebSocketClient) subscribeToStorageProtocols() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data, err := client.Hub.monitoringService.GetStorageProtocolInfo()
			if err != nil {
				continue
			}

			message := dto.WebSocketMessage{
				Type:      "storage_protocols",
				Data:      data,
				Timestamp: time.Now(),
				ClientID:  client.ID,
			}

			select {
			case client.Send <- message:
			default:
				return
			}

		case <-client.Hub.ctx.Done():
			return
		}
	}
}

// GetConnectedClients returns information about connected clients
func (hub *WebSocketHub) GetConnectedClients() []map[string]interface{} {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	clients := make([]map[string]interface{}, 0, len(hub.clients))
	for _, client := range hub.clients {
		clientInfo := map[string]interface{}{
			"id":           client.ID,
			"remote_addr":  client.RemoteAddr,
			"user_agent":   client.UserAgent,
			"connected_at": client.ConnectedAt,
			"last_ping":    client.LastPing,
		}
		clients = append(clients, clientInfo)
	}

	return clients
}

// BroadcastMessage broadcasts a message to all connected clients
func (hub *WebSocketHub) BroadcastMessage(message dto.WebSocketMessage) {
	select {
	case hub.broadcast <- message:
	default:
		log.Println("Broadcast channel is full, dropping message")
	}
}

// SendToClient sends a message to a specific client
func (hub *WebSocketHub) SendToClient(clientID string, message dto.WebSocketMessage) error {
	hub.mu.RLock()
	client, exists := hub.clients[clientID]
	hub.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	select {
	case client.Send <- message:
		return nil
	default:
		return fmt.Errorf("client %s channel is full", clientID)
	}
}

// GetClientCount returns the number of connected clients
func (hub *WebSocketHub) GetClientCount() int {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	return len(hub.clients)
}

// BroadcastAlert broadcasts an alert to all connected clients
func (hub *WebSocketHub) BroadcastAlert(alert dto.MonitoringAlert) {
	message := dto.WebSocketMessage{
		Type:      "alert",
		Data:      alert,
		Timestamp: time.Now(),
	}
	hub.BroadcastMessage(message)
}

// BroadcastCustomData broadcasts custom data to all connected clients
func (hub *WebSocketHub) BroadcastCustomData(dataType string, data interface{}) {
	message := dto.WebSocketMessage{
		Type:      dataType,
		Data:      data,
		Timestamp: time.Now(),
	}
	hub.BroadcastMessage(message)
}
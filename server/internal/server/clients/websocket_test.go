package clients

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"server/internal/server"
	"server/internal/server/db"
	"server/internal/server/objects"
	"server/pkg/packets"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// mockHub creates a minimal hub for testing
func mockHub() *server.Hub {
	return &server.Hub{
		Clients:           objects.NewSharedCollection[server.ClientInterfacer](),
		BroadcastChan:     make(chan *packets.Packet, 256),
		RegisterChan:      make(chan server.ClientInterfacer, 256),
		UnregisterChan:    make(chan server.ClientInterfacer, 256),
		SharedGameObjects: &server.SharedGameObjects{
			Players: objects.NewSharedCollection[*objects.Player](),
			Spores:  objects.NewSharedCollection[*objects.Spore](),
		},
	}
}

// mockDbTx creates a mock database transaction for testing
func mockDbTx() *server.DbTx {
	// Create a nil database connection for testing without real DB
	var nilDB *sql.DB
	return &server.DbTx{
		Ctx:     context.Background(),
		Queries: db.New(nilDB),
	}
}

// TestWebSocketCommunication tests WebSocket client-server communication
func TestWebSocketCommunication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	t.Run("Connect and receive ID", func(t *testing.T) {
		hub := mockHub()
		done := make(chan bool)

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			client, err := NewWebSocketClient(hub, w, r)
			if err != nil {
				t.Errorf("Failed to create WebSocket client: %v", err)
				return
			}

			client.Initialize(123)

			// Get the underlying connection to send directly
			wsClient := client.(*WebSocketClient)

			// Create and marshal message
			idMsg := packets.NewId(123)
			packet := &packets.Packet{
				SenderId: 0,
				Msg:      idMsg,
			}
			data, _ := proto.Marshal(packet)

			// Send directly via connection
			wsClient.conn.WriteMessage(websocket.BinaryMessage, data)

			// Wait for test completion
			<-done
		}))
		defer server.Close()

		// Connect as WebSocket client
		wsURL := "ws" + server.URL[4:] // Replace http with ws
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket server: %v", err)
		}
		defer conn.Close()

		// Read ID message with timeout
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}

		// Deserialize
		packet := &packets.Packet{}
		err = proto.Unmarshal(message, packet)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		// Verify ID
		idMsg := packet.GetId()
		if idMsg == nil {
			t.Fatal("Expected ID message, got nil")
		}
		if idMsg.Id != 123 {
			t.Errorf("Expected ID 123, got %d", idMsg.Id)
		}

		close(done)
	})

	t.Run("Send and receive PlayerDirection message", func(t *testing.T) {
		hub := mockHub()
		receivedDirection := make(chan float64, 1)
		done := make(chan bool)

		// Create test server with custom handler
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			client, err := NewWebSocketClient(hub, w, r)
			if err != nil {
				t.Errorf("Failed to create WebSocket client: %v", err)
				return
			}

			client.Initialize(456)

			// Read one message
			conn := client.(*WebSocketClient).conn
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				t.Logf("Error reading message: %v", err)
				return
			}

			// Parse message
			packet := &packets.Packet{}
			if err := proto.Unmarshal(msgBytes, packet); err != nil {
				t.Logf("Error unmarshaling: %v", err)
				return
			}

			// Extract direction
			dirMsg := packet.GetPlayerDirection()
			if dirMsg != nil {
				receivedDirection <- dirMsg.Direction

				// Send confirmation back
				player := &objects.Player{
					Name:      "TestPlayer",
					X:         100.0,
					Y:         200.0,
					Radius:    20.0,
					Direction: dirMsg.Direction,
					Speed:     150.0,
					Color:     0xFF0000,
				}

				// Marshal and send directly
				playerMsg := packets.NewPlayer(456, player)
				responsePacket := &packets.Packet{
					SenderId: 456,
					Msg:      playerMsg,
				}
				data, _ := proto.Marshal(responsePacket)
				conn.WriteMessage(websocket.BinaryMessage, data)
			}

			// Wait for test completion
			<-done
		}))
		defer server.Close()

		// Connect as client
		wsURL := "ws" + server.URL[4:]
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send PlayerDirection message
		direction := 1.57 // ~90 degrees
		dirMsg := &packets.Packet_PlayerDirection{
			PlayerDirection: &packets.PlayerDirectionMessage{
				Direction: direction,
			},
		}
		packet := &packets.Packet{
			SenderId: 456,
			Msg:      dirMsg,
		}

		data, err := proto.Marshal(packet)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}

		err = conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Wait for server to receive direction
		select {
		case received := <-receivedDirection:
			if received != direction {
				t.Errorf("Server received direction %f, expected %f", received, direction)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for server to receive direction")
		}

		// Read response (PlayerMessage)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, responseBytes, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		responsePacket := &packets.Packet{}
		err = proto.Unmarshal(responseBytes, responsePacket)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify PlayerMessage
		playerMsg := responsePacket.GetPlayer()
		if playerMsg == nil {
			t.Fatal("Expected PlayerMessage response, got nil")
		}
		if playerMsg.Id != 456 {
			t.Errorf("Expected player ID 456, got %d", playerMsg.Id)
		}
		if playerMsg.Direction != direction {
			t.Errorf("Player direction %f doesn't match sent direction %f", playerMsg.Direction, direction)
		}

		close(done)
	})

	t.Run("Chat message round-trip", func(t *testing.T) {
		hub := mockHub()
		receivedChat := make(chan string, 1)
		done := make(chan bool)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			client, err := NewWebSocketClient(hub, w, r)
			if err != nil {
				t.Errorf("Failed to create WebSocket client: %v", err)
				return
			}

			client.Initialize(789)

			// Read chat message
			conn := client.(*WebSocketClient).conn
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				return
			}

			packet := &packets.Packet{}
			if err := proto.Unmarshal(msgBytes, packet); err != nil {
				return
			}

			chatMsg := packet.GetChat()
			if chatMsg != nil {
				receivedChat <- chatMsg.Msg

				// Echo back
				echoMsg := packets.NewChat(chatMsg.Msg)
				echoPacket := &packets.Packet{
					SenderId: 789,
					Msg:      echoMsg,
				}
				data, _ := proto.Marshal(echoPacket)
				conn.WriteMessage(websocket.BinaryMessage, data)
			}

			<-done
		}))
		defer server.Close()

		// Connect
		wsURL := "ws" + server.URL[4:]
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send chat message
		chatText := "Hello from test!"
		chatMsg := packets.NewChat(chatText)
		packet := &packets.Packet{
			SenderId: 789,
			Msg:      chatMsg,
		}

		data, err := proto.Marshal(packet)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		err = conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			t.Fatalf("Failed to send: %v", err)
		}

		// Verify server received
		select {
		case received := <-receivedChat:
			if received != chatText {
				t.Errorf("Server received %q, expected %q", received, chatText)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for chat message")
		}

		// Read echo
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, responseBytes, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read echo: %v", err)
		}

		responsePacket := &packets.Packet{}
		err = proto.Unmarshal(responseBytes, responsePacket)
		if err != nil {
			t.Fatalf("Failed to unmarshal echo: %v", err)
		}

		echoMsg := responsePacket.GetChat()
		if echoMsg == nil {
			t.Fatal("Expected chat echo, got nil")
		}
		if echoMsg.Msg != chatText {
			t.Errorf("Echo message %q doesn't match sent %q", echoMsg.Msg, chatText)
		}

		close(done)
	})

	t.Run("Multiple messages in sequence", func(t *testing.T) {
		hub := mockHub()
		done := make(chan bool)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			client, err := NewWebSocketClient(hub, w, r)
			if err != nil {
				t.Errorf("Failed to create WebSocket client: %v", err)
				return
			}

			client.Initialize(999)
			conn := client.(*WebSocketClient).conn

			// Send multiple messages directly
			messages := []packets.Msg{
				packets.NewId(999),
				packets.NewChat("Message 1"),
				packets.NewChat("Message 2"),
				packets.NewChat("Message 3"),
			}

			for _, msg := range messages {
				packet := &packets.Packet{
					SenderId: 999,
					Msg:      msg,
				}
				data, _ := proto.Marshal(packet)
				conn.WriteMessage(websocket.BinaryMessage, data)
				time.Sleep(10 * time.Millisecond) // Small delay between messages
			}

			<-done
		}))
		defer server.Close()

		wsURL := "ws" + server.URL[4:]
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Read 4 messages
		messages := []string{}
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		for i := 0; i < 4; i++ {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				t.Fatalf("Failed to read message %d: %v", i, err)
			}

			packet := &packets.Packet{}
			if err := proto.Unmarshal(msgBytes, packet); err != nil {
				t.Fatalf("Failed to unmarshal message %d: %v", i, err)
			}

			if chatMsg := packet.GetChat(); chatMsg != nil {
				messages = append(messages, chatMsg.Msg)
			}
		}

		if len(messages) != 3 {
			t.Errorf("Expected 3 chat messages, got %d", len(messages))
		}

		expectedMessages := []string{"Message 1", "Message 2", "Message 3"}
		for i, expected := range expectedMessages {
			if i >= len(messages) {
				break
			}
			if messages[i] != expected {
				t.Errorf("Message %d: got %q, expected %q", i, messages[i], expected)
			}
		}

		close(done)
	})
}

// TestWebSocketClientCreation tests client initialization
func TestWebSocketClientCreation(t *testing.T) {
	t.Run("Client ID initialization", func(t *testing.T) {
		hub := mockHub()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			client, err := NewWebSocketClient(hub, w, r)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Initially ID should be 0
			if client.Id() != 0 {
				t.Errorf("Initial ID should be 0, got %d", client.Id())
			}

			// After initialization
			client.Initialize(42)
			if client.Id() != 42 {
				t.Errorf("After initialization, ID should be 42, got %d", client.Id())
			}

			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		wsURL := "ws" + server.URL[4:]
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		time.Sleep(200 * time.Millisecond)
	})
}

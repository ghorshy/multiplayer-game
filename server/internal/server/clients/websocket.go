package clients

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"server/internal/server"
	"server/internal/server/states"
	"server/pkg/packets"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type WebSocketClient struct {
	id         uint64
	conn       *websocket.Conn
	hub        *server.Hub
	dbTx       *server.DbTx
	state      server.ClientStateHandler
	sendChan   chan *packets.Packet
	logger     *log.Logger
	closeOnce  sync.Once
	closeChan  chan struct{}
}

func NewWebSocketClient(hub *server.Hub, writer http.ResponseWriter, request *http.Request) (server.ClientInterfacer, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(_ *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(writer, request, nil)

	if err != nil {
		return nil, err
	}

	c := &WebSocketClient{
		hub:       hub,
		conn:      conn,
		dbTx:      hub.NewDbTx(),
		sendChan:  make(chan *packets.Packet, 256),
		logger:    log.New(log.Writer(), "Client unknown: ", log.LstdFlags),
		closeChan: make(chan struct{}),
	}

	return c, nil
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id
	c.logger.SetPrefix(fmt.Sprintf("Client %d: ", c.id))
	c.SetState(&states.Connected{})
}

func (c *WebSocketClient) DbTx() *server.DbTx {
	return c.dbTx
}

func (c *WebSocketClient) SharedGameObjects() *server.SharedGameObjects {
	return c.hub.SharedGameObjects
}

func (c *WebSocketClient) ProcessMessage(senderId uint64, message packets.Msg) {
	c.state.HandleMessage(senderId, message)
}

func (c *WebSocketClient) SetState(state server.ClientStateHandler) {
	prevStateName := "None"
	if c.state != nil {
		prevStateName = c.state.Name()
		c.state.OnExit()
	}

	newStateName := "None"
	if state != nil {
		newStateName = state.Name()
	}

	c.logger.Printf("Switching from state %s to %s", prevStateName, newStateName)

	c.state = state

	if c.state != nil {
		c.state.SetClient(c)
		c.state.OnEnter()
	}
}

func (c *WebSocketClient) SocketSend(message packets.Msg) {
	c.SocketSendAs(message, c.id)
}

func (c *WebSocketClient) SocketSendAs(message packets.Msg, senderId uint64) {
	select {
	case c.sendChan <- &packets.Packet{SenderId: senderId, Msg: message}:
	default:
		c.logger.Printf("Client %d send channel full, dropping message: %T", c.id, message)
	}
}

func (c *WebSocketClient) PassToPeer(message packets.Msg, peerId uint64) {
	if peer, exists := c.hub.Clients.Get(peerId); exists {
		peer.ProcessMessage(c.id, message)
	}
}

func (c *WebSocketClient) Broadcast(message packets.Msg) {
	select {
	case c.hub.BroadcastChan <- &packets.Packet{SenderId: c.id, Msg: message}:
	default:
		c.logger.Printf("BroadcastChan full, dropping message: %T", message)
	}
}

func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.logger.Println("Closing read pump")
		c.Close("Read pump closed")
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Printf("error: %v", err)
			}
			break
		}

		packet := &packets.Packet{}
		err = proto.Unmarshal(data, packet)
		if err != nil {
			c.logger.Printf("error unmarshalling data: %v", err)
			continue
		}

		// to allow client to lazily not set the sender ID, assume they want to send it as themselves
		if packet.SenderId == 0 {
			packet.SenderId = c.id
		}

		c.ProcessMessage(packet.SenderId, packet.Msg)
	}
}

func (c *WebSocketClient) WritePump() {
	defer func() {
		c.logger.Println("Closing write pump")
		c.Close("write pump closed")
	}()

	for packet := range c.sendChan {
		writer, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			c.logger.Printf("error getting writer for %T packet, closing client: %v", packet.Msg, err)
			return
		}

		data, err := proto.Marshal(packet)
		if err != nil {
			c.logger.Printf("error marshalling %T packet, dropping: %v", packet.Msg, err)
			continue
		}

		_, writeErr := writer.Write(data)

		if writeErr != nil {
			c.logger.Printf("error writing %T packet: %v", packet.Msg, err)
			continue
		}

		writer.Write([]byte{'\n'})

		if closeErr := writer.Close(); closeErr != nil {
			c.logger.Printf("error closing writer, dropping %T packet: %v", packet.Msg, err)
			continue
		}
	}
}

func (c *WebSocketClient) Close(reason string) {
	c.closeOnce.Do(func() {
		c.Broadcast(packets.NewDisconnect(reason))
		c.logger.Printf("Closing client connection because: %s", reason)

		c.SetState(nil)

		// Non-blocking send to UnregisterChan
		select {
		case c.hub.UnregisterChan <- c:
		default:
			c.logger.Printf("UnregisterChan full, forcing unregister in goroutine")
			go func() { c.hub.UnregisterChan <- c }()
		}

		c.conn.Close()
		close(c.closeChan) // Signal that we're closing
		close(c.sendChan)  // Safe because closeOnce guarantees single execution
	})
}

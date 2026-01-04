package packets

import (
	"bytes"
	"server/internal/server/objects"
	"testing"

	"google.golang.org/protobuf/proto"
)

// TestProtocolBuffers tests serialization and deserialization of protobuf messages
func TestProtocolBuffers(t *testing.T) {
	t.Run("Serialize and deserialize PlayerMessage", func(t *testing.T) {
		// Create test player
		player := &objects.Player{
			Name:      "TestPlayer",
			X:         100.5,
			Y:         200.75,
			Radius:    25.0,
			Direction: 1.57, // ~90 degrees
			Speed:     150.0,
			Color:     0xFF0000, // Red
		}

		// Create PlayerMessage packet
		playerMsg := NewPlayer(123, player)
		packet := &Packet{
			Msg: playerMsg,
		}

		// Serialize to bytes
		data, err := proto.Marshal(packet)
		if err != nil {
			t.Fatalf("Failed to serialize PlayerMessage: %v", err)
		}

		if len(data) == 0 {
			t.Fatal("Serialized data is empty")
		}

		// Deserialize from bytes
		decoded := &Packet{}
		err = proto.Unmarshal(data, decoded)
		if err != nil {
			t.Fatalf("Failed to deserialize PlayerMessage: %v", err)
		}

		// Verify values match
		decodedPlayer := decoded.GetPlayer()
		if decodedPlayer == nil {
			t.Fatal("Decoded player message is nil")
		}

		if decodedPlayer.Id != 123 {
			t.Errorf("Player ID mismatch: got %d, want %d", decodedPlayer.Id, 123)
		}
		if decodedPlayer.Name != "TestPlayer" {
			t.Errorf("Player name mismatch: got %s, want %s", decodedPlayer.Name, "TestPlayer")
		}
		if decodedPlayer.X != 100.5 {
			t.Errorf("Player X mismatch: got %f, want %f", decodedPlayer.X, 100.5)
		}
		if decodedPlayer.Y != 200.75 {
			t.Errorf("Player Y mismatch: got %f, want %f", decodedPlayer.Y, 200.75)
		}
		if decodedPlayer.Radius != 25.0 {
			t.Errorf("Player radius mismatch: got %f, want %f", decodedPlayer.Radius, 25.0)
		}
		if decodedPlayer.Direction != 1.57 {
			t.Errorf("Player direction mismatch: got %f, want %f", decodedPlayer.Direction, 1.57)
		}
		if decodedPlayer.Speed != 150.0 {
			t.Errorf("Player speed mismatch: got %f, want %f", decodedPlayer.Speed, 150.0)
		}
		if decodedPlayer.Color != 0xFF0000 {
			t.Errorf("Player color mismatch: got %d, want %d", decodedPlayer.Color, 0xFF0000)
		}
	})

	t.Run("Serialize and deserialize PlayerDirectionMessage", func(t *testing.T) {
		// Create PlayerDirection packet
		direction := 3.14159 // ~180 degrees
		dirMsg := &Packet_PlayerDirection{
			PlayerDirection: &PlayerDirectionMessage{
				Direction: direction,
			},
		}
		packet := &Packet{
			Msg: dirMsg,
		}

		// Serialize
		data, err := proto.Marshal(packet)
		if err != nil {
			t.Fatalf("Failed to serialize PlayerDirectionMessage: %v", err)
		}

		// Deserialize
		decoded := &Packet{}
		err = proto.Unmarshal(data, decoded)
		if err != nil {
			t.Fatalf("Failed to deserialize PlayerDirectionMessage: %v", err)
		}

		// Verify
		decodedDir := decoded.GetPlayerDirection()
		if decodedDir == nil {
			t.Fatal("Decoded player direction message is nil")
		}
		if decodedDir.Direction != direction {
			t.Errorf("Direction mismatch: got %f, want %f", decodedDir.Direction, direction)
		}
	})

	t.Run("Serialize and deserialize ChatMessage", func(t *testing.T) {
		chatMsg := NewChat("Hello, World!")
		packet := &Packet{
			Msg: chatMsg,
		}

		// Serialize
		data, err := proto.Marshal(packet)
		if err != nil {
			t.Fatalf("Failed to serialize ChatMessage: %v", err)
		}

		// Deserialize
		decoded := &Packet{}
		err = proto.Unmarshal(data, decoded)
		if err != nil {
			t.Fatalf("Failed to deserialize ChatMessage: %v", err)
		}

		// Verify
		decodedChat := decoded.GetChat()
		if decodedChat == nil {
			t.Fatal("Decoded chat message is nil")
		}
		if decodedChat.Msg != "Hello, World!" {
			t.Errorf("Chat message mismatch: got %s, want %s", decodedChat.Msg, "Hello, World!")
		}
	})

	t.Run("Serialize and deserialize SporeMessage", func(t *testing.T) {
		spore := &objects.Spore{
			X:      50.5,
			Y:      75.25,
			Radius: 5.0,
		}
		sporeMsg := NewSpore(456, spore)
		packet := &Packet{
			Msg: sporeMsg,
		}

		// Serialize
		data, err := proto.Marshal(packet)
		if err != nil {
			t.Fatalf("Failed to serialize SporeMessage: %v", err)
		}

		// Deserialize
		decoded := &Packet{}
		err = proto.Unmarshal(data, decoded)
		if err != nil {
			t.Fatalf("Failed to deserialize SporeMessage: %v", err)
		}

		// Verify
		decodedSpore := decoded.GetSpore()
		if decodedSpore == nil {
			t.Fatal("Decoded spore message is nil")
		}
		if decodedSpore.Id != 456 {
			t.Errorf("Spore ID mismatch: got %d, want %d", decodedSpore.Id, 456)
		}
		if decodedSpore.X != 50.5 {
			t.Errorf("Spore X mismatch: got %f, want %f", decodedSpore.X, 50.5)
		}
		if decodedSpore.Y != 75.25 {
			t.Errorf("Spore Y mismatch: got %f, want %f", decodedSpore.Y, 75.25)
		}
		if decodedSpore.Radius != 5.0 {
			t.Errorf("Spore radius mismatch: got %f, want %f", decodedSpore.Radius, 5.0)
		}
	})

	t.Run("Serialize and deserialize GameBoundsMessage", func(t *testing.T) {
		boundsMsg := NewGameBounds(-1000.0, 1000.0, -500.0, 500.0)
		packet := &Packet{
			Msg: boundsMsg,
		}

		// Serialize
		data, err := proto.Marshal(packet)
		if err != nil {
			t.Fatalf("Failed to serialize GameBoundsMessage: %v", err)
		}

		// Deserialize
		decoded := &Packet{}
		err = proto.Unmarshal(data, decoded)
		if err != nil {
			t.Fatalf("Failed to deserialize GameBoundsMessage: %v", err)
		}

		// Verify
		decodedBounds := decoded.GetGameBounds()
		if decodedBounds == nil {
			t.Fatal("Decoded game bounds message is nil")
		}
		if decodedBounds.MinX != -1000.0 {
			t.Errorf("MinX mismatch: got %f, want %f", decodedBounds.MinX, -1000.0)
		}
		if decodedBounds.MaxX != 1000.0 {
			t.Errorf("MaxX mismatch: got %f, want %f", decodedBounds.MaxX, 1000.0)
		}
		if decodedBounds.MinY != -500.0 {
			t.Errorf("MinY mismatch: got %f, want %f", decodedBounds.MinY, -500.0)
		}
		if decodedBounds.MaxY != 500.0 {
			t.Errorf("MaxY mismatch: got %f, want %f", decodedBounds.MaxY, 500.0)
		}
	})

	t.Run("Test binary format consistency", func(t *testing.T) {
		// Ensure that same data always serializes to same bytes
		player := &objects.Player{
			Name:      "Consistent",
			X:         10.0,
			Y:         20.0,
			Radius:    15.0,
			Direction: 0.0,
			Speed:     100.0,
			Color:     0xFFFFFF,
		}

		msg1 := NewPlayer(1, player)
		packet1 := &Packet{Msg: msg1}
		data1, _ := proto.Marshal(packet1)

		msg2 := NewPlayer(1, player)
		packet2 := &Packet{Msg: msg2}
		data2, _ := proto.Marshal(packet2)

		if !bytes.Equal(data1, data2) {
			t.Error("Same data should serialize to identical bytes")
		}
	})

	t.Run("Test SporesBatch serialization", func(t *testing.T) {
		// Create batch of spores
		spores := map[uint64]*objects.Spore{
			1: {X: 10.0, Y: 20.0, Radius: 5.0},
			2: {X: 30.0, Y: 40.0, Radius: 5.0},
			3: {X: 50.0, Y: 60.0, Radius: 5.0},
		}

		batchMsg := NewSporeBatch(spores)
		packet := &Packet{Msg: batchMsg}

		// Serialize
		data, err := proto.Marshal(packet)
		if err != nil {
			t.Fatalf("Failed to serialize SporesBatch: %v", err)
		}

		// Deserialize
		decoded := &Packet{}
		err = proto.Unmarshal(data, decoded)
		if err != nil {
			t.Fatalf("Failed to deserialize SporesBatch: %v", err)
		}

		// Verify
		batch := decoded.GetSporesBatch()
		if batch == nil {
			t.Fatal("Decoded spores batch is nil")
		}
		if len(batch.Spores) != 3 {
			t.Errorf("Expected 3 spores in batch, got %d", len(batch.Spores))
		}
	})

	t.Run("Test DisconnectMessage", func(t *testing.T) {
		disconnectMsg := NewDisconnect("Connection timeout")
		packet := &Packet{Msg: disconnectMsg}

		// Serialize
		data, err := proto.Marshal(packet)
		if err != nil {
			t.Fatalf("Failed to serialize DisconnectMessage: %v", err)
		}

		// Deserialize
		decoded := &Packet{}
		err = proto.Unmarshal(data, decoded)
		if err != nil {
			t.Fatalf("Failed to deserialize DisconnectMessage: %v", err)
		}

		// Verify
		decodedDisconnect := decoded.GetDisconnect()
		if decodedDisconnect == nil {
			t.Fatal("Decoded disconnect message is nil")
		}
		if decodedDisconnect.Reason != "Connection timeout" {
			t.Errorf("Disconnect reason mismatch: got %s, want %s", decodedDisconnect.Reason, "Connection timeout")
		}
	})
}

// Benchmark serialization
func BenchmarkPlayerMessageSerialization(b *testing.B) {
	player := &objects.Player{
		Name:      "BenchPlayer",
		X:         100.0,
		Y:         200.0,
		Radius:    25.0,
		Direction: 1.57,
		Speed:     150.0,
		Color:     0xFF0000,
	}
	msg := NewPlayer(123, player)
	packet := &Packet{Msg: msg}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(packet)
	}
}

// Benchmark deserialization
func BenchmarkPlayerMessageDeserialization(b *testing.B) {
	player := &objects.Player{
		Name:      "BenchPlayer",
		X:         100.0,
		Y:         200.0,
		Radius:    25.0,
		Direction: 1.57,
		Speed:     150.0,
		Color:     0xFF0000,
	}
	msg := NewPlayer(123, player)
	packet := &Packet{Msg: msg}
	data, _ := proto.Marshal(packet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoded := &Packet{}
		_ = proto.Unmarshal(data, decoded)
	}
}

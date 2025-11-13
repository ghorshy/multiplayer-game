package states

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"server/internal/server"
	"server/internal/server/db"
	"server/internal/server/objects"
	"server/pkg/packets"
	"time"
)

func (g *InGame) getSpore(sporeId uint64) (*objects.Spore, error) {
	spore, exists := g.client.SharedGameObjects().Spores.Get(sporeId)
	if !exists {
		return nil, fmt.Errorf("spore with ID %d does not exist", sporeId)
	}
	return spore, nil
}

func (g *InGame) validatePlayerCloseToObject(objX, objY, objRadius, buffer float64) error {
	realDX := g.player.X - objX
	realDY := g.player.Y - objY
	realDistSq := realDX*realDX + realDY*realDY

	thresholdDist := g.player.Radius + buffer + objRadius
	thresholdDistSq := thresholdDist * thresholdDist

	if realDistSq > thresholdDistSq {
		return fmt.Errorf("player is too far from the object (distSq: %f, thresholdSq: %f)", realDistSq, thresholdDistSq)
	}
	return nil
}

func (g *InGame) validatePlayerDropCooldown(spore *objects.Spore, buffer float64) error {
	minAcceptableDistance := spore.Radius + g.player.Radius + buffer
	minAcceptableTime := time.Duration(minAcceptableDistance/g.player.Speed*1000) * time.Millisecond
	if spore.DroppedBy == g.player && time.Since(spore.DroppedAt) < minAcceptableTime {
		return fmt.Errorf("player dropped the spore too recently (time since drop: %v, min acceptable time: %v", time.Since(spore.DroppedAt), minAcceptableTime)
	}
	return nil
}

func radToMass(radius float64) float64 {
	return math.Pi * radius * radius
}

func massToRad(mass float64) float64 {
	return math.Sqrt(mass / math.Pi)
}

func (g *InGame) nextRadius(massDiff float64) float64 {
	oldMass := radToMass(g.player.Radius)
	newMass := oldMass + massDiff
	return massToRad(newMass)
}

func (g *InGame) playerUpdateLoop(ctx context.Context) {
	const delta float64 = 0.05
	ticker := time.NewTicker(time.Duration(delta*1000) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.syncPlayer(delta)
		case <-ctx.Done():
			return
		}
	}
}

func (g *InGame) syncPlayer(delta float64) {
	oldX, oldY := g.player.X, g.player.Y

	// Calculate new position based on velocity
	newX := g.player.X + g.player.Speed*math.Cos(g.player.Direction)*delta
	newY := g.player.Y + g.player.Speed*math.Sin(g.player.Direction)*delta

	// Rubber-band boundary enforcement
	// Creates a soft wall that pushes back with increasing force
	buffer := g.player.Radius
	rubberBandZone := 200.0 // Distance from boundary where rubber-band starts

	// X-axis rubber-banding
	minXBound := objects.MinX + buffer
	maxXBound := objects.MaxX - buffer
	if newX < minXBound {
		// Hard clamp at boundary
		newX = minXBound
	} else if newX < minXBound+rubberBandZone {
		// Soft zone: apply resistance
		distanceIntoBoundary := minXBound + rubberBandZone - newX
		resistance := distanceIntoBoundary / rubberBandZone // 0 to 1
		pushBack := resistance * resistance * g.player.Speed * delta * 2
		newX += pushBack
	}

	if newX > maxXBound {
		newX = maxXBound
	} else if newX > maxXBound-rubberBandZone {
		distanceIntoBoundary := newX - (maxXBound - rubberBandZone)
		resistance := distanceIntoBoundary / rubberBandZone
		pushBack := resistance * resistance * g.player.Speed * delta * 2
		newX -= pushBack
	}

	// Y-axis rubber-banding
	minYBound := objects.MinY + buffer
	maxYBound := objects.MaxY - buffer
	if newY < minYBound {
		newY = minYBound
	} else if newY < minYBound+rubberBandZone {
		distanceIntoBoundary := minYBound + rubberBandZone - newY
		resistance := distanceIntoBoundary / rubberBandZone
		pushBack := resistance * resistance * g.player.Speed * delta * 2
		newY += pushBack
	}

	if newY > maxYBound {
		newY = maxYBound
	} else if newY > maxYBound-rubberBandZone {
		distanceIntoBoundary := newY - (maxYBound - rubberBandZone)
		resistance := distanceIntoBoundary / rubberBandZone
		pushBack := resistance * resistance * g.player.Speed * delta * 2
		newY -= pushBack
	}

	// Update player position
	g.player.X = newX
	g.player.Y = newY

	// Debug: Log first position update
	if oldX == 0 && oldY == 0 {
		g.logger.Printf("First position update: (%.2f, %.2f) -> (%.2f, %.2f)", oldX, oldY, newX, newY)
	}

	// Drop spores at the clamped position (after boundary enforcement)
	probability := g.player.Radius / float64(server.MaxSpores*5)
	if rand.Float64() < probability && g.player.Radius > 10 {
		spore := &objects.Spore{
			X:         g.player.X,  // Now uses the clamped position
			Y:         g.player.Y,  // Now uses the clamped position
			Radius:    min(5+g.player.Radius/50, 15),
			DroppedBy: g.player,
			DroppedAt: time.Now(),
		}
		sporeId := g.client.SharedGameObjects().Spores.Add(spore)
		g.client.Broadcast(packets.NewSpore(sporeId, spore))
		go g.client.SocketSend(packets.NewSpore(sporeId, spore))
		g.player.Radius = g.nextRadius(-radToMass(spore.Radius))
	}

	updatePacket := packets.NewPlayer(g.client.Id(), g.player)
	g.client.Broadcast(updatePacket)
	go g.client.SocketSend(updatePacket)
}

func (g *InGame) bestScoreSyncLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.syncPlayerBestScore()
		case <-ctx.Done():
			return
		}
	}
}

func (g *InGame) syncPlayerBestScore() {
	currentScore := int32(math.Round(radToMass(g.player.Radius)))
	if currentScore > g.player.BestScore {
		g.player.BestScore = currentScore
		err := g.client.DbTx().Queries.UpdatePlayerBestScore(g.client.DbTx().Ctx, db.UpdatePlayerBestScoreParams{
			ID:        g.player.DbId,
			BestScore: g.player.BestScore,
		})
		if err != nil {
			g.logger.Printf("Error updating player best score: %v", err)
		}
	}
}

type InGame struct {
	client                    server.ClientInterfacer
	player                    *objects.Player
	logger                    *log.Logger
	cancelPlayerUpdateLoop    context.CancelFunc
	cancelBestScoreSyncLoop   context.CancelFunc
}

func (g *InGame) Name() string {
	return "InGame"
}

func (g *InGame) SetClient(client server.ClientInterfacer) {
	g.client = client
	loggingPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), g.Name())
	g.logger = log.New(log.Writer(), loggingPrefix, log.LstdFlags)
}

func (g *InGame) OnEnter() {
	g.logger.Printf("Adding player %s to the shared collection", g.player.Name)
	go g.client.SharedGameObjects().Players.Add(g.player, g.client.Id())

	// Set the initial properties of the player BEFORE calculating spawn coords
	g.player.Speed = 150.0
	g.player.Radius = 20.0
	g.player.X, g.player.Y = objects.SpawnCoords(g.player.Radius, g.client.SharedGameObjects().Players, nil)

	g.logger.Printf("Player spawned at position (%.2f, %.2f) with radius %.2f", g.player.X, g.player.Y, g.player.Radius)

	// Send game boundaries to the client so it can enforce them locally
	g.client.SocketSend(packets.NewGameBounds(objects.MinX, objects.MaxX, objects.MinY, objects.MaxY))

	// Send the player's initial state to the client
	g.client.SocketSend(packets.NewPlayer(g.client.Id(), g.player))

	go g.sendInitialSpores(50, 50*time.Millisecond)

	// Start background loop to sync best scores to database every 5 seconds
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelBestScoreSyncLoop = cancel
	go g.bestScoreSyncLoop(ctx)
}

func (g *InGame) sendInitialSpores(batchSize int, delay time.Duration) {
	sporesBatch := make(map[uint64]*objects.Spore, batchSize)

	g.client.SharedGameObjects().Spores.ForEach(func(sporeId uint64, spore *objects.Spore) {
		sporesBatch[sporeId] = spore

		if len(sporesBatch) >= batchSize {
			g.client.SocketSend(packets.NewSporeBatch(sporesBatch))
			sporesBatch = make(map[uint64]*objects.Spore, batchSize)
			time.Sleep(delay)
		}
	})

	if len(sporesBatch) > 0 {
		g.client.SocketSend(packets.NewSporeBatch(sporesBatch))
	}

}

func (g *InGame) OnExit() {
	if g.cancelPlayerUpdateLoop != nil {
		g.cancelPlayerUpdateLoop()
	}
	if g.cancelBestScoreSyncLoop != nil {
		g.cancelBestScoreSyncLoop()
	}

	g.client.SharedGameObjects().Players.Remove(g.client.Id())
	// Final sync to ensure best score is saved before exiting
	g.syncPlayerBestScore()
}

func (g *InGame) HandleMessage(senderId uint64, message packets.Msg) {
	switch message := message.(type) {
	case *packets.Packet_Player:
		g.handlePlayer(senderId, message)
	case *packets.Packet_Chat:
		g.handleChat(senderId, message)
	case *packets.Packet_PlayerDirection:
		g.handlePlayerDirection(senderId, message)
	case *packets.Packet_SporeConsumed:
		g.handleSporeConsumed(senderId, message)
	case *packets.Packet_PlayerConsumed:
		g.handlePlayerConsumed(senderId, message)
	case *packets.Packet_Spore:
		g.handleSpore(senderId, message)
	case *packets.Packet_Disconnect:
		g.handleDisconnect(senderId, message)
	}
}

func (g *InGame) handlePlayer(senderId uint64, message *packets.Packet_Player) {
	if senderId == g.client.Id() {
		g.logger.Println("Received player message from out own client, ignoring")
		return
	}
	g.client.SocketSendAs(message, senderId)
}

func (g *InGame) handleChat(senderId uint64, message *packets.Packet_Chat) {
	if senderId == g.client.Id() {
		g.client.Broadcast(message)
	} else {
		g.client.SocketSendAs(message, senderId)
	}
}

func (g *InGame) handlePlayerDirection(senderId uint64, message *packets.Packet_PlayerDirection) {
	if senderId != g.client.Id() {
		return
	}

	g.player.Direction = message.PlayerDirection.Direction

	if g.cancelPlayerUpdateLoop == nil {
		g.logger.Println("Starting player update loop")
		ctx, cancel := context.WithCancel(context.Background())
		g.cancelPlayerUpdateLoop = cancel
		go g.playerUpdateLoop(ctx)
	}
}

func (g *InGame) handleSporeConsumed(senderId uint64, message *packets.Packet_SporeConsumed) {
	if senderId != g.client.Id() {
		g.client.SocketSendAs(message, senderId)
		return
	}

	errMsg := "Could not verify spore consumption: "

	sporeId := message.SporeConsumed.SporeId
	spore, err := g.getSpore(sporeId)
	if err != nil {
		g.logger.Println(errMsg + err.Error())
		return
	}

	// Large buffer to account for network lag + server tick delay
	// Google Cloud: ~50-100ms RTT + 50ms tick + jitter = need generous buffer
	// At speed 150: 100ms = 15 units, 200ms = 30 units
	const validationBuffer = 100.0
	err = g.validatePlayerCloseToObject(spore.X, spore.Y, spore.Radius, validationBuffer)
	if err != nil {
		g.logger.Println(errMsg + err.Error())
		return
	}

	err = g.validatePlayerDropCooldown(spore, validationBuffer)
	if err != nil {
		g.logger.Println(errMsg + err.Error())
		return
	}

	sporeMass := radToMass(spore.Radius)
	g.player.Radius = g.nextRadius(sporeMass)

	go g.client.SharedGameObjects().Spores.Remove(sporeId)

	g.client.Broadcast(message)
}

func (g *InGame) handlePlayerConsumed(senderId uint64, message *packets.Packet_PlayerConsumed) {
	if senderId != g.client.Id() {
		g.client.SocketSendAs(message, senderId)

		if message.PlayerConsumed.PlayerId == g.client.Id() {
			log.Println("Player was consumed, respawning")
			// SetState in goroutine to avoid blocking Hub
			go g.client.SetState(&InGame{
				player: &objects.Player{
					Name: g.player.Name,
				},
			})
		}

		return
	}

	errMsg := "Could not verify player consumption: "

	otherId := message.PlayerConsumed.PlayerId
	other, err := g.getOtherPlayer(otherId)
	if err != nil {
		g.logger.Println(errMsg + err.Error())
		return
	}

	ourMass := radToMass(g.player.Radius)
	otherMass := radToMass(other.Radius)
	if ourMass <= otherMass*1.5 {
		g.logger.Printf(errMsg+"player not massive enough to consume the other player (our radius: %f, other radius: %f)", g.player.Radius, other.Radius)
		return
	}

	const validationBuffer = 100.0
	err = g.validatePlayerCloseToObject(other.X, other.Y, other.Radius, validationBuffer)
	if err != nil {
		g.logger.Println(errMsg + err.Error())
		return
	}

	g.player.Radius = g.nextRadius(otherMass)

	go g.client.SharedGameObjects().Players.Remove(otherId)

	g.client.Broadcast(message)
}

func (g *InGame) handleSpore(senderId uint64, message *packets.Packet_Spore) {
	g.client.SocketSendAs(message, senderId)
}

func (g *InGame) handleDisconnect(senderId uint64, message *packets.Packet_Disconnect) {
	if senderId == g.client.Id() {
		g.client.Broadcast(message)
		// SetState in goroutine to avoid blocking Hub
		go g.client.SetState(&Connected{})
	} else {
		go g.client.SocketSendAs(message, senderId)
	}
}

func (g *InGame) getOtherPlayer(otherId uint64) (*objects.Player, error) {
	other, exists := g.client.SharedGameObjects().Players.Get(otherId)
	if !exists {
		return nil, fmt.Errorf("player with ID %d does not exist", otherId)
	}

	return other, nil
}

package main

import (
	"fmt"
	"server/pkg/packets"
)

func main() {
	packet := &packets.Packet{
		SenderId: 420,
		Msg:      packets.NewChat("Hello World!"),
	}

	fmt.Println(packet)
}

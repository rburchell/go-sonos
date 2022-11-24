package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/caglar10ur/sonos"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	son, err := sonos.NewSonos()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer son.Close()

	ips := []string{}
	for i := 13; i <= 18; i++ {
		ips = append(ips, fmt.Sprintf("192.168.10.%d", i))
	}

	var zps []*sonos.ZonePlayer
	for _, ip := range ips {
		u, err := sonos.FromEndpoint(ip)
		if err != nil {
			log.Fatalf("%s", err)
		}

		zp, err := sonos.NewZonePlayer(
			sonos.WithLocation(u),
		)
		if err != nil {
			log.Fatalf("%s", err)
		}
		fmt.Printf("Trying %s\t%s\t%s (coordinator %t)\n", zp.RoomName(), zp.ModelName(), zp.SerialNum(), zp.IsCoordinator())

		if err := son.Register(zp); err != nil {
			log.Printf("%s", err)
		}

		if zp.IsCoordinator() {
			zps = append(zps, zp)
		}
	}

	for _, zp := range zps {
		fmt.Printf("Connected to %s\t%s\t%s (coordinator %t)\n", zp.RoomName(), zp.ModelName(), zp.SerialNum(), zp.IsCoordinator())

		son.Subscribe(ctx, zp, zp.AVTransport)
	}
	<-ctx.Done()
}

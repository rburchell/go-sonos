package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/caglar10ur/sonos"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s [room name] [media url]\n", os.Args[0])
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	zp, err := sonos.FindRoom(ctx, os.Args[1])
	if err != nil {
		fmt.Printf("FindRoom Error: %v\n", err)
		return
	}

	if err = zp.SetAVTransportURI(os.Args[2]); err != nil {
		fmt.Printf("SetAVTransportURI Error: %v\n", err)
		return
	}

	if err = zp.Play(); err != nil {
		fmt.Printf("Play Error: %v\n", err)
		return
	}
}

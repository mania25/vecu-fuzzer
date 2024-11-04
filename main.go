package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"time"

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

// generateRandomFrame generates a CAN frame with random ID, length, and data.
func generateRandomFrame() can.Frame {
	// Random CAN ID in a standard range
	id := uint32(rand.Intn(0x7FF))    // Standard CAN IDs range from 0x000 to 0x7FF
	length := uint8(rand.Intn(8) + 1) // Data length from 1 to 8
	var data [8]byte
	for i := 0; i < int(length); i++ {
		data[i] = byte(rand.Intn(256)) // Random byte data
	}

	return can.Frame{
		ID:     id,
		Length: length,
		Data:   data,
	}
}

// fuzzECU performs fuzzing by sending random frames at an interval.
func fuzzECU(ctx context.Context, conn net.Conn) {
	tx := socketcan.NewTransmitter(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			frame := generateRandomFrame()
			err := tx.TransmitFrame(ctx, frame)
			if err != nil {
				log.Printf("Failed to transmit frame: %v", err)
			} else {
				log.Printf("Fuzzed Frame: ID: 0x%03x Length: %d Data: %v", frame.ID, frame.Length, frame.Data)
			}

			// Delay to avoid flooding
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func main() {
	log.Println("Starting fuzzing on vCAN interface. . .")

	ctx := context.Background()
	conn, err := socketcan.DialContext(ctx, "can", "vcan0")
	if err != nil {
		log.Fatalf("failed to connect to vcan0: %v", err)
	}
	defer conn.Close()

	log.Println("Fuzzing vCAN interface with random CAN frames...")
	fuzzECU(ctx, conn)
}

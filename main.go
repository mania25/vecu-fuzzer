package main

import (
	"context"
	"encoding/hex"
	"log"
	"math/rand"
	"time"

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

// Define the valid 4-byte commands recognized by the ECU
var validCommands = []string{"ENGS", "ENGO", "TEMP", "INJT", "OXYG", "FUEL", "THRO", "RPMS"}

// Triangular membership function for fuzzy probability
func triangularMembership(x, a, b, c float64) float64 {
	if x <= a || x >= c {
		return 0
	} else if x < b {
		return (x - a) / (b - a)
	} else {
		return (c - x) / (c - b)
	}
}

// Calculate the probability of sending a valid command based on ECU stability
func calculateFuzzyProbability(stability float64) float64 {
	// Using a triangular membership function with parameters for low, medium, and high stability
	low := triangularMembership(stability, 0, 0, 0.5)
	medium := triangularMembership(stability, 0.3, 0.5, 0.7)
	high := triangularMembership(stability, 0.5, 1, 1)

	// Weighted probability: higher stability increases the likelihood of sending valid commands
	return low*0.1 + medium*0.3 + high*0.7
}

// Function to generate random 4-byte frame data
func generateRandomFrameData() []byte {
	data := make([]byte, 4)
	for i := 0; i < 4; i++ {
		data[i] = byte(rand.Intn(256)) // Each byte from 0x00 to 0xFF
	}
	return data
}

// Function to send a CAN frame to the ECU
func sendFrame(tx *socketcan.Transmitter, frame can.Frame) {
	// Send the frame to the ECU
	if err := tx.TransmitFrame(context.Background(), frame); err != nil {
		log.Printf("Error sending frame: %v", err)
	} else {
		log.Printf("Sent frame ID: %03x, Length: %d, Data: %s", frame.ID, frame.Length, hex.EncodeToString(frame.Data[:frame.Length]))
	}
}

// Function to decide if a frame should be valid or random based on fuzzy probability
func sendRandomOrValidFrame(tx *socketcan.Transmitter, stability float64) {
	frame := can.Frame{
		ID:     uint32(rand.Intn(0x7FF)), // Random ID in the standard CAN range
		Length: 4,                        // Fixed length of 4 bytes
	}

	// Calculate probability based on fuzzy logic
	probability := calculateFuzzyProbability(stability)
	if rand.Float64() < probability {
		// Select a random valid 4-byte command
		validCommand := validCommands[rand.Intn(len(validCommands))]
		copy(frame.Data[:], []byte(validCommand))
		log.Printf("Sending valid command '%s' to ECU with fuzzy probability %.2f", validCommand, probability)
	} else {
		// Generate completely random 4-byte data for the frame
		randomData := generateRandomFrameData()
		copy(frame.Data[:], randomData)
		log.Printf("Sending random data frame to ECU with fuzzy probability %.2f", probability)
	}

	sendFrame(tx, frame)
}

func main() {
	log.Println("Opening CAN interface. . .")

	// Connect to the virtual CAN interface
	conn, err := socketcan.DialContext(context.Background(), "can", "vcan0")
	if err != nil {
		log.Fatal("Failed to connect into CAN:", err)
	}
	tx := socketcan.NewTransmitter(conn)

	log.Println("CAN interface has been opened")

	log.Println("Starting fuzzy frame generation with variable probability based on ECU stability...")

	// Assume the ECU stability varies over time or is given as an input
	// Here, we simulate stability changing over time
	for stability := 0.0; stability <= 1.0; stability += 0.1 {
		for i := 0; i < 10; i++ { // Send 10 frames at each stability level
			sendRandomOrValidFrame(tx, stability)
			time.Sleep(200 * time.Millisecond) // Delay between frames for clarity in logging
		}
	}

	log.Println("Completed fuzzy frame generation")
}

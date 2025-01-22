package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/amimof/huego"
	"github.com/holoplot/go-evdev"
)

const (
	// F17 key code
	keyF17 = 67
)

type Controller struct {
	bridge     *huego.Bridge
	lightID    int
	device     *evdev.InputDevice
	devicePath string
	keyCode    int
}

func main() {
	// Load required env vars
	bridgeIP := mustGetenv("HUE_BRIDGE_IP")
	username := mustGetenv("HUE_USERNAME")
	lightID, err := strconv.Atoi(mustGetenv("HUE_LIGHT_ID"))
	if err != nil {
		log.Fatalf("Invalid light ID: %v", err)
	}

	// Get key code with F17 (67) as default
	keyCode, err := strconv.Atoi(getenv("HUE_KEY_CODE", "67"))
	if err != nil {
		log.Fatalf("Invalid key code: %v", err)
	}

	ctrl := &Controller{
		devicePath: getenv("HUE_DEVICE_PATH", "/dev/input/event0"),
		lightID:    lightID,
		keyCode:    keyCode,
	}

	// Create bridge connection
	ctrl.bridge = huego.New(bridgeIP, username)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := ctrl.init(); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Main event loop
	go ctrl.handleEvents()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down...")
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s not set", key)
	}
	return val
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (c *Controller) init() error {
	// Open input device
	device, err := evdev.Open(c.devicePath)
	if err != nil {
		return fmt.Errorf("failed to open input device: %v", err)
	}
	c.device = device

	// Test bridge connection
	_, err = c.bridge.GetLights()
	if err != nil {
		return fmt.Errorf("failed to connect to bridge: %v", err)
	}

	return nil
}

func (c *Controller) handleEvents() {
	log.Printf("Starting event monitoring on device %s for key code %d", c.devicePath, c.keyCode)

	for {
		events, err := c.device.ReadOne()
		if err != nil {
			log.Printf("Error reading events: %v", err)
			continue
		}

		// Log all key events to see what we're getting
		if events.Type == evdev.EV_KEY {
			log.Printf("Key event: Type=%d, Code=%d, Value=%d", events.Type, events.Code, events.Value)
		}

		// Only handle key events when the key is pressed down
		if events.Type == evdev.EV_KEY && events.Value == 1 && events.Code == evdev.EvCode(c.keyCode) {
			log.Printf("Toggling light %d", c.lightID)
			light, err := c.bridge.GetLight(c.lightID)
			if err != nil {
				log.Printf("Error getting light state: %v", err)
				continue
			}

			// Toggle the light
			if light.IsOn() {
				log.Printf("Turning light off")
				err = light.Off()
			} else {
				log.Printf("Turning light on")
				err = light.On()
			}

			if err != nil {
				log.Printf("Error toggling light: %v", err)
			}
		}
	}
}

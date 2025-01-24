package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/amimof/huego"
	"github.com/holoplot/go-evdev"
)

const (
	// Key codes (sudo evtest)
	knobPress  = 187 // Toggle
	knobLeft   = 188 // Dim
	knobRight  = 189 // Brighten
	sceneLeft  = 185 // Previous scene
	sceneRight = 186 // Next scene

	// Light control values
	brightnessIncrement = 25 // Amount to change brightness by
)

type Controller struct {
	bridge     *huego.Bridge
	lightID    int
	device     *evdev.InputDevice
	devicePath string
	keyCode    int
	scenes     []string // List of scene IDs
	sceneIndex int      // Current scene index
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
		scenes:     parseScenes(getenv("HUE_SCENE_IDS", "")),
		sceneIndex: 0,
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

func parseScenes(scenes string) []string {
	if scenes == "" {
		return nil
	}
	return strings.Split(scenes, ",")
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
	log.Printf("Starting event monitoring on device %s for key codes F17=%d, F18=%d, F19=%d, Scene Left=%d, Scene Right=%d",
		c.devicePath, c.keyCode, knobLeft, knobRight, sceneLeft, sceneRight)

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
		if events.Type == evdev.EV_KEY && events.Value == 1 {
			switch events.Code {
			case evdev.EvCode(c.keyCode): // F17 - Toggle
				c.toggleLight()
			case evdev.EvCode(knobLeft): // F18 - Dim
				c.adjustBrightness(-brightnessIncrement)
			case evdev.EvCode(knobRight): // F19 - Brighten
				c.adjustBrightness(brightnessIncrement)
			case evdev.EvCode(sceneLeft): // Previous scene
				c.rotateScene(-1)
			case evdev.EvCode(sceneRight): // Next scene
				c.rotateScene(1)
			}
		}
	}
}

func (c *Controller) toggleLight() {
	log.Printf("Toggling light %d", c.lightID)
	light, err := c.bridge.GetLight(c.lightID)
	if err != nil {
		log.Printf("Error getting light state: %v", err)
		return
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

func (c *Controller) adjustBrightness(delta int) {
	light, err := c.bridge.GetLight(c.lightID)
	if err != nil {
		log.Printf("Error getting light state: %v", err)
		return
	}

	// Only adjust if light is on
	if !light.IsOn() {
		log.Printf("Light is off, not adjusting brightness")
		return
	}

	newBrightness := int(light.State.Bri) + delta
	// Clamp brightness between 1 and 254 (Hue's brightness range)
	if newBrightness < 1 {
		newBrightness = 1
	} else if newBrightness > 254 {
		newBrightness = 254
	}

	log.Printf("Adjusting brightness from %d to %d", light.State.Bri, newBrightness)
	err = light.Bri(uint8(newBrightness))
	if err != nil {
		log.Printf("Error adjusting brightness: %v", err)
	}
}

func (c *Controller) rotateScene(direction int) {
	if len(c.scenes) == 0 {
		log.Printf("No scenes configured")
		return
	}

	// Calculate new index with wrapping
	c.sceneIndex = (c.sceneIndex + direction + len(c.scenes)) % len(c.scenes)
	sceneID := c.scenes[c.sceneIndex]

	log.Printf("Activating scene %s (index %d)", sceneID, c.sceneIndex)
	_, err := c.bridge.RecallScene(sceneID, 0)
	if err != nil {
		log.Printf("Error activating scene: %v", err)
	}
}

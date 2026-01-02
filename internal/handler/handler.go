package handler

import (
	"fmt"
	"log"
	"strings"

	"github.com/amimof/huego"
	"github.com/holoplot/go-evdev"
)

// Bridge defines the interface for Hue bridge operations
type Bridge interface {
	GetGroup(id int) (*huego.Group, error)
	SetGroupState(id int, state huego.State) (*huego.Response, error)
	RecallScene(sceneID string, groupID int) (*huego.Response, error)
}

type Handler struct {
	cfg             *Config
	bridge          Bridge
	device          *evdev.InputDevice
	scenes          []string
	sceneIdx        int
	dynamicsEnabled bool
}

func New(cfg *Config) (*Handler, error) {
	// Connect to Hue bridge
	bridge := huego.New(cfg.BridgeIP, cfg.Username)
	if _, err := bridge.GetLights(); err != nil {
		return nil, fmt.Errorf("failed to connect to bridge: %v", err)
	}

	// Parse scenes if provided
	var scenes []string
	if cfg.SceneIDs != "" {
		scenes = strings.Split(cfg.SceneIDs, ",")
	}

	return &Handler{
		cfg:             cfg,
		bridge:          bridge,
		scenes:          scenes,
		dynamicsEnabled: true, // Start with dynamics enabled
	}, nil
}

func (h *Handler) Init() error {
	device, err := evdev.Open(h.cfg.DevicePath)
	if err != nil {
		return fmt.Errorf("failed to open input device: %v", err)
	}
	h.device = device
	return nil
}

func (h *Handler) HandleEvents() {
	log.Printf("Starting event monitoring on device %s", h.cfg.DevicePath)

	for {
		ev, err := h.device.ReadOne()
		if err != nil {
			log.Printf("Error reading event: %v", err)
			continue
		}

		if ev.Type == evdev.EV_KEY && ev.Value == 1 {
			h.handleKeyEvent(ev.Code)
		}
	}
}

func (h *Handler) handleKeyEvent(code evdev.EvCode) {
	switch code {
	case evdev.EvCode(h.cfg.KeyCode):
		h.toggleLight()
	case evdev.EvCode(KnobLeft):
		h.adjustBrightness(-BrightnessIncrement)
	case evdev.EvCode(KnobRight):
		h.adjustBrightness(BrightnessIncrement)
	case evdev.EvCode(SceneLeft):
		h.rotateScene(-1)
	case evdev.EvCode(SceneRight):
		h.rotateScene(1)
	}
}

func (h *Handler) toggleLight() {
	group, err := h.bridge.GetGroup(h.cfg.GroupID)
	if err != nil {
		log.Printf("Error getting group state: %v", err)
		return
	}

	state := &huego.State{On: !group.State.On}
	_, err = h.bridge.SetGroupState(h.cfg.GroupID, *state)
	if err != nil {
		log.Printf("Error toggling group: %v", err)
	}
}

func (h *Handler) adjustBrightness(delta int) {
	group, err := h.bridge.GetGroup(h.cfg.GroupID)
	if err != nil {
		log.Printf("Error getting group state: %v", err)
		return
	}

	bri := int(group.State.Bri) + delta
	if bri < 0 {
		bri = 0
	} else if bri > 254 {
		bri = 254
	}

	state := &huego.State{On: true, Bri: uint8(bri)}
	_, err = h.bridge.SetGroupState(h.cfg.GroupID, *state)
	if err != nil {
		log.Printf("Error adjusting brightness: %v", err)
	}
}

func (h *Handler) rotateScene(direction int) {
	if len(h.scenes) == 0 {
		log.Printf("No scenes configured")
		return
	}

	h.sceneIdx = (h.sceneIdx + direction + len(h.scenes)) % len(h.scenes)
	sceneID := h.scenes[h.sceneIdx]

	_, err := h.bridge.RecallScene(sceneID, 0) // 0 for all lights
	if err != nil {
		log.Printf("Error activating scene: %v", err)
		return
	}
	log.Printf("Activated scene: %s", sceneID)
}

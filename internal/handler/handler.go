package handler

import (
	"fmt"
	"log"
	"strings"

	"github.com/holoplot/go-evdev"
)

// GroupState represents the state of a group of lights
type GroupState struct {
	On  bool
	Bri uint8
}

// Bridge defines the interface for Hue bridge operations
type Bridge interface {
	GetGroup(id int) (*GroupState, error)
	SetGroupState(id int, on bool, bri uint8) error
	RecallScene(sceneID string, groupID int) error
}

type Handler struct {
	cfg             *Config
	bridge          Bridge
	v2bridge        *V2Bridge
	device          *evdev.InputDevice
	scenes          []string
	sceneIdx        int
	dynamicsEnabled bool
}

func New(cfg *Config) (*Handler, error) {
	// Connect to Hue bridge using v2 API
	v2bridge, err := NewV2Bridge(cfg.BridgeIP, cfg.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bridge: %v", err)
	}

	// Parse scenes if provided
	var scenes []string
	if cfg.SceneIDs != "" {
		scenes = strings.Split(cfg.SceneIDs, ",")
	}

	// Set the group ID for the v2 bridge
	v2bridge.SetGroupID(cfg.GroupID)

	return &Handler{
		cfg:             cfg,
		bridge:          v2bridge,
		v2bridge:        v2bridge,
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
	case evdev.EvCode(SceneNext):
		h.nextScene()
	case evdev.EvCode(ToggleDynamics):
		h.toggleDynamics()
	}
}

func (h *Handler) toggleLight() {
	group, err := h.bridge.GetGroup(0)
	if err != nil {
		log.Printf("Error getting group state: %v", err)
		return
	}

	err = h.bridge.SetGroupState(0, !group.On, group.Bri)
	if err != nil {
		log.Printf("Error toggling group: %v", err)
	}
}

func (h *Handler) adjustBrightness(delta int) {
	group, err := h.bridge.GetGroup(0)
	if err != nil {
		log.Printf("Error getting group state: %v", err)
		return
	}

	bri := int(group.Bri) + delta
	if bri < 0 {
		bri = 0
	} else if bri > 254 {
		bri = 254
	}

	err = h.bridge.SetGroupState(0, true, uint8(bri))
	if err != nil {
		log.Printf("Error adjusting brightness: %v", err)
	}
}

func (h *Handler) nextScene() {
	if len(h.scenes) == 0 {
		log.Printf("No scenes configured")
		return
	}

	h.sceneIdx = (h.sceneIdx + 1) % len(h.scenes)
	sceneID := h.scenes[h.sceneIdx]

	err := h.bridge.RecallScene(sceneID, 0)
	if err != nil {
		log.Printf("Error activating scene: %v", err)
		return
	}
	log.Printf("Activated scene: %s", sceneID)
}

func (h *Handler) toggleDynamics() {
	if len(h.scenes) == 0 {
		log.Printf("No scenes configured for dynamics toggle")
		return
	}

	h.dynamicsEnabled = !h.dynamicsEnabled

	// Get current scene ID
	sceneID := h.scenes[h.sceneIdx]

	if h.v2bridge != nil {
		err := h.v2bridge.SetDynamics(sceneID, h.dynamicsEnabled)
		if err != nil {
			log.Printf("Error toggling dynamics: %v", err)
			return
		}

		if h.dynamicsEnabled {
			log.Printf("Dynamics enabled for scene: %s", sceneID)
		} else {
			log.Printf("Dynamics disabled for scene: %s", sceneID)
		}
	}
}

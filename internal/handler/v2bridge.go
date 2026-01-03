package handler

import (
	"fmt"

	openhue "github.com/openhue/openhue-go"
)

// V2Bridge wraps openhue-go client to implement the Bridge interface
type V2Bridge struct {
	home    *openhue.Home
	groupID string
}

func NewV2Bridge(bridgeIP, username string) (*V2Bridge, error) {
	home, err := openhue.NewHome(bridgeIP, username)
	if err != nil {
		return nil, fmt.Errorf("failed to create v2 home: %w", err)
	}

	return &V2Bridge{
		home: home,
	}, nil
}

// SetGroupID sets the group ID to use for operations
func (b *V2Bridge) SetGroupID(groupID string) {
	b.groupID = groupID
}

// GetGroup returns group state
func (b *V2Bridge) GetGroup(id int) (*GroupState, error) {
	if b.groupID == "" {
		return nil, fmt.Errorf("group ID not set")
	}

	light, err := b.home.GetGroupedLightById(b.groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get grouped light: %w", err)
	}

	// Convert brightness from percentage (0-100) to Hue scale (0-254)
	bri := uint8((*light.Dimming.Brightness / 100.0) * 254.0)

	return &GroupState{
		On:  *light.On.On,
		Bri: bri,
	}, nil
}

// SetGroupState sets the state of a group
func (b *V2Bridge) SetGroupState(id int, on bool, bri uint8) error {
	if b.groupID == "" {
		return fmt.Errorf("group ID not set")
	}

	update := openhue.GroupedLightPut{
		On: &openhue.On{On: &on},
	}

	if bri > 0 {
		brightness := float32(bri) / 254.0 * 100.0
		update.Dimming = &openhue.Dimming{Brightness: &brightness}
	}

	err := b.home.UpdateGroupedLight(b.groupID, update)
	if err != nil {
		return fmt.Errorf("failed to update grouped light: %w", err)
	}

	return nil
}

// RecallScene recalls a scene
func (b *V2Bridge) RecallScene(sceneID string, groupID int) error {
	action := openhue.SceneRecallActionActive
	recall := &openhue.SceneRecall{
		Action: &action,
	}

	err := b.home.UpdateScene(sceneID, openhue.ScenePut{
		Recall: recall,
	})
	if err != nil {
		return fmt.Errorf("failed to recall scene: %w", err)
	}

	return nil
}

// SetDynamics toggles dynamics for a scene (v2-specific functionality)
func (b *V2Bridge) SetDynamics(sceneID string, enabled bool) error {
	update := openhue.ScenePut{
		AutoDynamic: &enabled,
	}

	err := b.home.UpdateScene(sceneID, update)
	if err != nil {
		return fmt.Errorf("failed to update scene dynamics: %w", err)
	}

	return nil
}

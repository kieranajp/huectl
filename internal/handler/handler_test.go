package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBridge is a mock implementation of the Hue bridge
type MockBridge struct {
	mock.Mock
}

func (m *MockBridge) GetGroup(id int) (*GroupState, error) {
	args := m.Called(id)
	return args.Get(0).(*GroupState), args.Error(1)
}

func (m *MockBridge) SetGroupState(id int, on bool, bri uint8) error {
	args := m.Called(id, on, bri)
	return args.Error(0)
}

func (m *MockBridge) RecallScene(sceneID string, groupID int) error {
	args := m.Called(sceneID, groupID)
	return args.Error(0)
}

func TestToggleLight(t *testing.T) {
	tests := []struct {
		name          string
		initialState  bool
		expectedState bool
		expectedError bool
	}{
		{
			name:          "toggle from off to on",
			initialState:  false,
			expectedState: true,
			expectedError: false,
		},
		{
			name:          "toggle from on to off",
			initialState:  true,
			expectedState: false,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBridge := new(MockBridge)
			cfg := &Config{
				BridgeIP: "192.168.1.33",
				Username: "test",
				GroupID:  "1",
			}
			h := &Handler{
				cfg:    cfg,
				bridge: mockBridge,
			}

			// Setup mock expectations
			mockBridge.On("GetGroup", 0).Return(&GroupState{
				On:  tt.initialState,
				Bri: 100,
			}, nil)

			mockBridge.On("SetGroupState", 0, tt.expectedState, uint8(100)).Return(nil)

			// Execute
			h.toggleLight()

			// Verify
			mockBridge.AssertExpectations(t)
		})
	}
}

func TestAdjustBrightness(t *testing.T) {
	tests := []struct {
		name          string
		initialBri    int
		delta         int
		expectedBri   int
		expectedError bool
	}{
		{
			name:          "increase brightness",
			initialBri:    100,
			delta:         25,
			expectedBri:   125,
			expectedError: false,
		},
		{
			name:          "decrease brightness",
			initialBri:    100,
			delta:         -25,
			expectedBri:   75,
			expectedError: false,
		},
		{
			name:          "clamp at minimum",
			initialBri:    10,
			delta:         -20,
			expectedBri:   0,
			expectedError: false,
		},
		{
			name:          "clamp at maximum",
			initialBri:    240,
			delta:         30,
			expectedBri:   254,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBridge := new(MockBridge)
			cfg := &Config{
				BridgeIP: "192.168.1.33",
				Username: "test",
				GroupID:  "1",
			}
			h := &Handler{
				cfg:    cfg,
				bridge: mockBridge,
			}

			// Setup mock expectations
			mockBridge.On("GetGroup", 0).Return(&GroupState{
				On:  true,
				Bri: uint8(tt.initialBri),
			}, nil)

			mockBridge.On("SetGroupState", 0, true, uint8(tt.expectedBri)).Return(nil)

			// Execute
			h.adjustBrightness(tt.delta)

			// Verify
			mockBridge.AssertExpectations(t)
		})
	}
}

func TestNextScene(t *testing.T) {
	tests := []struct {
		name          string
		scenes        []string
		initialIndex  int
		expectedIndex int
		expectedError bool
	}{
		{
			name:          "next scene",
			scenes:        []string{"scene1", "scene2", "scene3"},
			initialIndex:  0,
			expectedIndex: 1,
			expectedError: false,
		},
		{
			name:          "wrap around to first scene",
			scenes:        []string{"scene1", "scene2", "scene3"},
			initialIndex:  2,
			expectedIndex: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBridge := new(MockBridge)
			cfg := &Config{
				BridgeIP: "192.168.1.33",
				Username: "test",
				GroupID:  "1",
			}
			h := &Handler{
				cfg:      cfg,
				bridge:   mockBridge,
				scenes:   tt.scenes,
				sceneIdx: tt.initialIndex,
			}

			// Setup mock expectations
			expectedScene := tt.scenes[tt.expectedIndex]
			mockBridge.On("RecallScene", expectedScene, 0).Return(nil)

			// Execute
			h.nextScene()

			// Verify
			mockBridge.AssertExpectations(t)
			assert.Equal(t, tt.expectedIndex, h.sceneIdx)
		})
	}
}

func TestToggleDynamics(t *testing.T) {
	tests := []struct {
		name             string
		scenes           []string
		initialDynamics  bool
		expectedDynamics bool
	}{
		{
			name:             "disable dynamics",
			scenes:           []string{"scene1"},
			initialDynamics:  true,
			expectedDynamics: false,
		},
		{
			name:             "enable dynamics",
			scenes:           []string{"scene1"},
			initialDynamics:  false,
			expectedDynamics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBridge := new(MockBridge)
			cfg := &Config{
				BridgeIP: "192.168.1.33",
				Username: "test",
				GroupID:  "1",
			}

			// For dynamics test, we need to manually test the SetDynamics call
			// since v2bridge is a concrete type, not an interface
			h := &Handler{
				cfg:             cfg,
				bridge:          mockBridge,
				v2bridge:        nil, // Will test SetDynamics separately
				scenes:          tt.scenes,
				dynamicsEnabled: tt.initialDynamics,
			}

			// Execute - this will be a no-op since v2bridge is nil, just verify state toggle
			h.toggleDynamics()

			// Verify state changed
			assert.Equal(t, tt.expectedDynamics, h.dynamicsEnabled)
		})
	}
}

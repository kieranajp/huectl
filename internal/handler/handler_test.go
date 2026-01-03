package handler

import (
	"testing"

	"github.com/amimof/huego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBridge is a mock implementation of the Hue bridge
type MockBridge struct {
	mock.Mock
}

func (m *MockBridge) GetGroup(id int) (*huego.Group, error) {
	args := m.Called(id)
	return args.Get(0).(*huego.Group), args.Error(1)
}

func (m *MockBridge) SetGroupState(id int, state huego.State) (*huego.Response, error) {
	args := m.Called(id, state)
	return args.Get(0).(*huego.Response), args.Error(1)
}

func (m *MockBridge) RecallScene(sceneID string, groupID int) (*huego.Response, error) {
	args := m.Called(sceneID, groupID)
	return args.Get(0).(*huego.Response), args.Error(1)
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
				GroupID:  1,
			}
			h := &Handler{
				cfg:    cfg,
				bridge: mockBridge,
			}

			// Setup mock expectations
			mockBridge.On("GetGroup", 1).Return(&huego.Group{
				ID: 1,
				State: &huego.State{
					On: tt.initialState,
				},
			}, nil)

			mockBridge.On("SetGroupState", 1, huego.State{On: tt.expectedState}).Return(&huego.Response{}, nil)

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
				GroupID:  1,
			}
			h := &Handler{
				cfg:    cfg,
				bridge: mockBridge,
			}

			// Setup mock expectations
			mockBridge.On("GetGroup", 1).Return(&huego.Group{
				ID: 1,
				State: &huego.State{
					On:  true,
					Bri: uint8(tt.initialBri),
				},
			}, nil)

			mockBridge.On("SetGroupState", 1, huego.State{On: true, Bri: uint8(tt.expectedBri)}).Return(&huego.Response{}, nil)

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
				GroupID:  1,
			}
			h := &Handler{
				cfg:      cfg,
				bridge:   mockBridge,
				scenes:   tt.scenes,
				sceneIdx: tt.initialIndex,
			}

			// Setup mock expectations
			expectedScene := tt.scenes[tt.expectedIndex]
			mockBridge.On("RecallScene", expectedScene, 0).Return(&huego.Response{}, nil)

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
		name            string
		initialDynamics bool
		scenes          []string
		sceneIdx        int
		expectedDynamics bool
		expectedEffect  string
		expectedFlashColor []float32
	}{
		{
			name:            "disable dynamics with scenes",
			initialDynamics: true,
			scenes:          []string{"scene1", "scene2"},
			sceneIdx:        1,
			expectedDynamics: false,
			expectedEffect:  "none",
			expectedFlashColor: []float32{0.55, 0.35},
		},
		{
			name:            "enable dynamics with scenes",
			initialDynamics: false,
			scenes:          []string{"scene1", "scene2"},
			sceneIdx:        0,
			expectedDynamics: true,
			expectedEffect:  "",
			expectedFlashColor: []float32{0.2, 0.7},
		},
		{
			name:            "disable dynamics without scenes",
			initialDynamics: true,
			scenes:          []string{},
			sceneIdx:        0,
			expectedDynamics: false,
			expectedEffect:  "none",
			expectedFlashColor: []float32{0.55, 0.35},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBridge := new(MockBridge)
			cfg := &Config{
				BridgeIP: "192.168.1.33",
				Username: "test",
				GroupID:  1,
			}
			h := &Handler{
				cfg:             cfg,
				bridge:          mockBridge,
				dynamicsEnabled: tt.initialDynamics,
				scenes:          tt.scenes,
				sceneIdx:        tt.sceneIdx,
			}

			// Get current state
			mockBridge.On("GetGroup", 1).Return(&huego.Group{
				ID: 1,
				State: &huego.State{
					On:  true,
					Bri: uint8(150),
				},
			}, nil)

			// Flash feedback
			mockBridge.On("SetGroupState", 1, huego.State{
				On: true,
				Xy: tt.expectedFlashColor,
				Bri: uint8(200),
				TransitionTime: uint16(1),
			}).Return(&huego.Response{}, nil)

			// Scene recall if scenes exist
			if len(tt.scenes) > 0 {
				mockBridge.On("RecallScene", tt.scenes[tt.sceneIdx], 0).Return(&huego.Response{}, nil)
			}

			// Restore brightness and dynamics effect
			mockBridge.On("SetGroupState", 1, huego.State{Bri: uint8(150), Effect: tt.expectedEffect}).Return(&huego.Response{}, nil)

			// Execute
			h.toggleDynamics()

			// Verify
			mockBridge.AssertExpectations(t)
			assert.Equal(t, tt.expectedDynamics, h.dynamicsEnabled)
		})
	}
}

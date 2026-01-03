package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kieranajp/huectl/internal/handler"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "huectl",
		Usage:   "Control Philips Hue lights via keyboard input events",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "bridge-ip",
				EnvVars:  []string{"HUE_BRIDGE_IP"},
				Usage:    "Hue bridge IP address",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "username",
				EnvVars:  []string{"HUE_USERNAME"},
				Usage:    "Hue bridge API username",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "group-id",
				EnvVars:  []string{"HUE_GROUP_ID"},
				Usage:    "ID of the grouped light resource to control",
				Required: true,
			},
			&cli.IntFlag{
				Name:    "key-code",
				EnvVars: []string{"HUE_KEY_CODE"},
				Value:   187,
				Usage:   "Key code to trigger the light (default: 187 for F17)",
			},
			&cli.StringFlag{
				Name:    "device-path",
				EnvVars: []string{"HUE_DEVICE_PATH"},
				Value:   "/dev/input/event0",
				Usage:   "Input device path",
			},
			&cli.StringFlag{
				Name:    "scene-ids",
				EnvVars: []string{"HUE_SCENE_IDS"},
				Usage:   "Comma-separated list of scene IDs to rotate between",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	cfg := &handler.Config{
		BridgeIP:   c.String("bridge-ip"),
		Username:   c.String("username"),
		GroupID:    c.String("group-id"),
		KeyCode:    c.Int("key-code"),
		DevicePath: c.String("device-path"),
		SceneIDs:   c.String("scene-ids"),
	}

	h, err := handler.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create handler: %v", err)
	}

	if err := h.Init(); err != nil {
		return fmt.Errorf("failed to initialize: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go h.HandleEvents()

	<-sigChan
	fmt.Println("\nShutting down...")
	return nil
}

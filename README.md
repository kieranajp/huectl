# huectl

Simple Go service to control Philips Hue lights via keyboard input events. Originally designed for the Pikatea macropad but works with any input device.

## Setup

1. Copy the service template and configure your environment:

```bash
cp huectl.service.template huectl.service
# Edit huectl.service with your Hue bridge details
```

2. Build and deploy to Raspberry Pi (or whatever):

```bash
make install
```

## Configuration

The service expects the following environment variables:
- `HUE_BRIDGE_IP`: Your Hue bridge IP address
- `HUE_USERNAME`: Your Hue bridge API username
- `HUE_GROUP_ID`: ID of the room/group to control
- `HUE_KEY_CODE`: Key code to trigger the light (default: 187 for F17)
- `HUE_DEVICE_PATH`: Input device path (default: /dev/input/event0)
- `HUE_SCENE_IDS`: Comma-separated list of scene IDs to rotate between (optional)

## Makefile Targets

The included Makefile provides several useful targets:
- `make build`: Build for local development/testing
- `make pi`: Cross-compile for Raspberry Pi Zero W
- `make deploy`: Build and copy binary to the Pi
- `make service`: Install and enable systemd service
- `make install`: Full installation (deploy + service)
- `make clean`: Remove build artifacts
- `make test`: Run tests

## Development

For local development, you can run the binary directly with the required environment variables:

```bash
HUE_BRIDGE_IP=192.168.1.40 \
HUE_USERNAME=your_username \
HUE_GROUP_ID=1 \
HUE_KEY_CODE=25 \
go run main.go
```

This sets key code 25 (`p`) to toggle group 1 on the bridge 192.168.1.40. So now you can hammer `p` for `party`.

## License

MIT

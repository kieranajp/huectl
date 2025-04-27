# Basic info
BINARY_NAME=huectl
PI_HOST=rpi-zero2w.local
INSTALL_PATH=/usr/local/bin
SERVICE_NAME=$(BINARY_NAME).service
REMOTE_SERVICE_PATH=/etc/systemd/system/$(SERVICE_NAME)

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Cross compilation flags for Pi Zero W
PI_GOARCH=arm
PI_GOARM=6
PI_GOOS=linux

.PHONY: all build clean test pi deploy service

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

# Build for Raspberry Pi
pi:
	GOARCH=$(PI_GOARCH) GOARM=$(PI_GOARM) GOOS=$(PI_GOOS) $(GOBUILD) -o $(BINARY_NAME) -v

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Deploy binary to Pi
deploy: test pi
	scp $(BINARY_NAME) $(PI_HOST):/tmp/
	ssh $(PI_HOST) "sudo mv /tmp/$(BINARY_NAME) $(INSTALL_PATH) && sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)"

# Install systemd service
service:
	scp $(SERVICE_NAME) $(PI_HOST):/tmp/
	ssh $(PI_HOST) "sudo mv /tmp/$(SERVICE_NAME) $(REMOTE_SERVICE_PATH) && \
		sudo systemctl daemon-reload && \
		sudo systemctl enable $(SERVICE_NAME) && \
		sudo systemctl restart $(SERVICE_NAME)"

# Full installation
install: deploy service

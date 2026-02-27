package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigWithDeviceFiles(t *testing.T) {
	// Create a temporary directory for test configs
	tempDir, err := os.MkdirTemp("", "test-config")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create channels.yaml
	channelsContent := `
- id: test-channel
  name: Test Channel
  protocol: modbus-tcp
  enable: true
  config:
    url: tcp://127.0.0.1:502
  devices:
    - id: test-device
      device_file: "devices/test-device.yaml"
`
	channelsPath := filepath.Join(tempDir, "channels.yaml")
	if err := os.WriteFile(channelsPath, []byte(channelsContent), 0644); err != nil {
		t.Fatalf("Failed to write channels.yaml: %v", err)
	}

	// Create devices directory
	devicesDir := filepath.Join(tempDir, "devices")
	if err := os.Mkdir(devicesDir, 0755); err != nil {
		t.Fatalf("Failed to create devices dir: %v", err)
	}

	// Create device file
	deviceContent := `
id: test-device
name: Test Device
enable: true
interval: 10s
config:
  slave_id: 1
points:
  - id: test-point
    name: Test Point
    address: "0"
    datatype: int16
`
	devicePath := filepath.Join(devicesDir, "test-device.yaml")
	if err := os.WriteFile(devicePath, []byte(deviceContent), 0644); err != nil {
		t.Fatalf("Failed to write test-device.yaml: %v", err)
	}

	// Create other required config files
	createEmptyFile(t, tempDir, "server.yaml")
	createEmptyFile(t, tempDir, "storage.yaml")
	createEmptyFile(t, tempDir, "northbound.yaml")
	createEmptyFile(t, tempDir, "edge_rules.yaml")
	createEmptyFile(t, tempDir, "system.yaml")
	createEmptyFile(t, tempDir, "users.yaml")

	// Test loading config
	cfg, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify channel and device loaded correctly
	if len(cfg.Channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(cfg.Channels))
	}

	channel := cfg.Channels[0]
	if channel.ID != "test-channel" {
		t.Errorf("Expected channel ID 'test-channel', got '%s'", channel.ID)
	}

	if len(channel.Devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(channel.Devices))
	}

	device := channel.Devices[0]
	if device.ID != "test-device" {
		t.Errorf("Expected device ID 'test-device', got '%s'", device.ID)
	}

	if device.Name != "Test Device" {
		t.Errorf("Expected device name 'Test Device', got '%s'", device.Name)
	}

	if len(device.Points) != 1 {
		t.Errorf("Expected 1 point, got %d", len(device.Points))
	}

	point := device.Points[0]
	if point.ID != "test-point" {
		t.Errorf("Expected point ID 'test-point', got '%s'", point.ID)
	}
}

func TestConfigManagerHotReload(t *testing.T) {
	// Create a temporary directory for test configs
	tempDir, err := os.MkdirTemp("", "test-config-manager")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create initial config files
	channelsContent := `
- id: test-channel
  name: Test Channel
  protocol: modbus-tcp
  enable: true
  config:
    url: tcp://127.0.0.1:502
  devices:
    - id: test-device
      device_file: "devices/test-device.yaml"
`
	channelsPath := filepath.Join(tempDir, "channels.yaml")
	if err := os.WriteFile(channelsPath, []byte(channelsContent), 0644); err != nil {
		t.Fatalf("Failed to write channels.yaml: %v", err)
	}

	// Create devices directory
	devicesDir := filepath.Join(tempDir, "devices")
	if err := os.Mkdir(devicesDir, 0755); err != nil {
		t.Fatalf("Failed to create devices dir: %v", err)
	}

	// Create initial device file
	deviceContent := `
id: test-device
name: Test Device
enable: true
interval: 10s
config:
  slave_id: 1
points:
  - id: test-point
    name: Test Point
    address: "0"
    datatype: int16
`
	devicePath := filepath.Join(devicesDir, "test-device.yaml")
	if err := os.WriteFile(devicePath, []byte(deviceContent), 0644); err != nil {
		t.Fatalf("Failed to write test-device.yaml: %v", err)
	}

	// Create other required config files
	createEmptyFile(t, tempDir, "server.yaml")
	createEmptyFile(t, tempDir, "storage.yaml")
	createEmptyFile(t, tempDir, "northbound.yaml")
	createEmptyFile(t, tempDir, "edge_rules.yaml")
	createEmptyFile(t, tempDir, "system.yaml")
	createEmptyFile(t, tempDir, "users.yaml")

	// Create config manager
	cm, err := NewConfigManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Start watcher
	cm.StartWatch(100 * time.Millisecond)
	defer cm.StopWatch()

	// Verify initial config
	initialCfg := cm.GetConfig()
	if len(initialCfg.Channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(initialCfg.Channels))
	}

	// Modify device file
	updatedDeviceContent := `
id: test-device
name: Updated Test Device
enable: true
interval: 15s
config:
  slave_id: 2
points:
  - id: test-point
    name: Test Point
    address: "0"
    datatype: int16
  - id: new-point
    name: New Point
    address: "1"
    datatype: int16
`
	if err := os.WriteFile(devicePath, []byte(updatedDeviceContent), 0644); err != nil {
		t.Fatalf("Failed to update test-device.yaml: %v", err)
	}

	// Wait for reload
	time.Sleep(500 * time.Millisecond)

	// Verify config was reloaded
	updatedCfg := cm.GetConfig()
	if len(updatedCfg.Channels) != 1 {
		t.Errorf("Expected 1 channel after reload, got %d", len(updatedCfg.Channels))
	}

	device := updatedCfg.Channels[0].Devices[0]
	if device.Name != "Updated Test Device" {
		t.Errorf("Expected device name 'Updated Test Device', got '%s'", device.Name)
	}

	if device.Config["slave_id"] != 2 {
		t.Errorf("Expected slave_id 2, got %v", device.Config["slave_id"])
	}

	if len(device.Points) != 2 {
		t.Errorf("Expected 2 points, got %d", len(device.Points))
	}
}

func createEmptyFile(t *testing.T, dir, name string) {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create %s: %v", name, err)
	}
}



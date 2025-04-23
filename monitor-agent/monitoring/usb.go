//go:build windows
// +build windows

package monitoring

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	setupapi = windows.NewLazyDLL("setupapi.dll")

	setupDiGetClassDevsW              = setupapi.NewProc("SetupDiGetClassDevsW")
	setupDiEnumDeviceInfo             = setupapi.NewProc("SetupDiEnumDeviceInfo")
	setupDiGetDeviceRegistryPropertyW = setupapi.NewProc("SetupDiGetDeviceRegistryPropertyW")
	setupDiDestroyDeviceInfoList      = setupapi.NewProc("SetupDiDestroyDeviceInfoList")
)

const (
	DIGCF_PRESENT        = 0x2
	SPDRP_HARDWAREID     = 0x1
	SPDRP_FRIENDLYNAME   = 0x0C
	INVALID_HANDLE_VALUE = ^uintptr(0)
)

type USBMonitor struct {
	lastDevices map[string]bool
	interval    time.Duration
}

func NewUSBMonitor() *USBMonitor {
	interval := 5 // default interval
	if envInterval := os.Getenv("USB_MONITOR_INTERVAL"); envInterval != "" {
		if i, err := strconv.Atoi(envInterval); err == nil {
			interval = i
		}
	}
	return &USBMonitor{
		lastDevices: make(map[string]bool),
		interval:    time.Duration(interval) * time.Second,
	}
}

type SP_DEVINFO_DATA struct {
	CbSize    uint32
	ClassGuid [16]byte
	DevInst   uint32
	Reserved  uintptr
}

func getUSBDevicesWindows() ([]map[string]string, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Get device info set
	var guid [16]byte // USB device class GUID
	h, _, err := setupDiGetClassDevsW.Call(
		uintptr(unsafe.Pointer(&guid)),
		0,
		0,
		DIGCF_PRESENT)
	if h == INVALID_HANDLE_VALUE {
		return nil, fmt.Errorf("SetupDiGetClassDevsW failed: %v", err)
	}
	defer setupDiDestroyDeviceInfoList.Call(h)

	var devices []map[string]string
	var index uint32
	for {
		var data SP_DEVINFO_DATA
		data.CbSize = uint32(unsafe.Sizeof(data))

		ret, _, _ := setupDiEnumDeviceInfo.Call(h, uintptr(index), uintptr(unsafe.Pointer(&data)))
		if ret == 0 {
			break
		}

		// Get hardware ID
		var buf [256]uint16
		var bufSize uint32
		ret, _, _ = setupDiGetDeviceRegistryPropertyW.Call(
			h,
			uintptr(unsafe.Pointer(&data)),
			SPDRP_HARDWAREID,
			0,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Sizeof(buf)),
			uintptr(unsafe.Pointer(&bufSize)))

		if ret != 0 {
			hardwareID := syscall.UTF16ToString(buf[:])
			if strings.Contains(strings.ToLower(hardwareID), "usb") {
				// Get friendly name
				ret, _, _ = setupDiGetDeviceRegistryPropertyW.Call(
					h,
					uintptr(unsafe.Pointer(&data)),
					SPDRP_FRIENDLYNAME,
					0,
					uintptr(unsafe.Pointer(&buf[0])),
					uintptr(unsafe.Sizeof(buf)),
					uintptr(unsafe.Pointer(&bufSize)))

				friendlyName := ""
				if ret != 0 {
					friendlyName = syscall.UTF16ToString(buf[:])
				}

				device := map[string]string{
					"hardware_id": hardwareID,
					"name":        friendlyName,
				}
				devices = append(devices, device)
			}
		}

		index++
	}

	return devices, nil
}

func (m *USBMonitor) GetUSBDevices() ([]map[string]string, error) {
	return getUSBDevicesWindows()
}

func (m *USBMonitor) Monitor(ch chan<- map[string]interface{}) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for range ticker.C {
		devices, err := m.GetUSBDevices()
		if err != nil {
			log.Printf("Error getting USB devices: %v", err)
			continue
		}

		// Create a map of current devices
		currentDevices := make(map[string]bool)
		for _, device := range devices {
			id := device["hardware_id"]
			currentDevices[id] = true

			// Check if this is a new device
			if !m.lastDevices[id] {
				log.Printf("New USB device detected: %s", device["name"])
				ch <- map[string]interface{}{
					"event":     "connected",
					"device_id": id,
					"name":      device["name"],
					"timestamp": time.Now().Format(time.RFC3339),
				}
			}
		}

		// Check for disconnected devices
		for id := range m.lastDevices {
			if !currentDevices[id] {
				log.Printf("USB device disconnected: %s", id)
				ch <- map[string]interface{}{
					"event":     "disconnected",
					"device_id": id,
					"timestamp": time.Now().Format(time.RFC3339),
				}
			}
		}

		// Update last devices
		m.lastDevices = currentDevices
	}
}

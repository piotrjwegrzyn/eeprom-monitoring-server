package templates

import (
	"testing"

	"pi-wegrzyn/ems/storage"
)

func TestNewPageContent(t *testing.T) {
	newEdit := NewPageContent()

	if newEdit.Action != NewAction {
		t.Errorf("Action should be %s, but got %s", NewAction, newEdit.Action)
	}

	if newEdit.Device.ID != 0 {
		t.Errorf("Device ID should be 0, but got %d", newEdit.Device.ID)
	}
}

func TestEditPageContent(t *testing.T) {
	device := storage.Device{
		ID:        1,
		IPAddress: "192.168.1.1",
	}
	errMsg := "error message"

	newEdit := EditPageContent(device, errMsg)

	if newEdit.Action != EditAction {
		t.Errorf("Action should be %s, but got %s", EditAction, newEdit.Action)
	}

	if newEdit.Device.ID != device.ID {
		t.Errorf("Device ID should be %d, but got %d", device.ID, newEdit.Device.ID)
	}

	if newEdit.IPVersion != 4 {
		t.Errorf("IPVersion should be 4, but got %d", newEdit.IPVersion)
	}

	if newEdit.ErrorMessage != errMsg {
		t.Errorf("ErrorMessage should be %s, but got %s", errMsg, newEdit.ErrorMessage)
	}
}

func TestNewPageContentWithError(t *testing.T) {
	device := storage.Device{
		ID:        1,
		IPAddress: "::1",
	}
	errMsg := "error message"

	newEdit := NewPageContentWithError(device, errMsg)

	if newEdit.Action != NewAction {
		t.Errorf("Action should be %s, but got %s", NewAction, newEdit.Action)
	}

	if newEdit.Device.ID != device.ID {
		t.Errorf("Device ID should be %d, but got %d", device.ID, newEdit.Device.ID)
	}

	if newEdit.ErrorMessage != errMsg {
		t.Errorf("ErrorMessage should be %s, but got %s", errMsg, newEdit.ErrorMessage)
	}
}

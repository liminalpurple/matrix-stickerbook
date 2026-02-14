package cli

import (
	"testing"
)

func TestNewLoginCmd(t *testing.T) {
	cmd := NewLoginCmd()

	if cmd.Use != "login" {
		t.Errorf("Expected command use 'login', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Command long description is empty")
	}

	if cmd.RunE == nil {
		t.Error("Command RunE function is nil")
	}
}

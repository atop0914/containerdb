package version

import (
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	info := Info()
	if !strings.Contains(info, "containerdb") {
		t.Error("Info() should contain 'containerdb'")
	}
	if !strings.Contains(info, Version) {
		t.Errorf("Info() should contain version %s", Version)
	}
}

func TestDefaultVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if !strings.HasPrefix(Version, "1.") {
		t.Errorf("expected v1.x.x, got %s", Version)
	}
}

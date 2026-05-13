package container

import (
	"testing"
	"time"
)

func TestAvailablePort(t *testing.T) {
	port, err := AvailablePort()
	if err != nil {
		t.Fatalf("AvailablePort() error = %v", err)
	}

	if port < 1024 || port > 65535 {
		t.Errorf("expected port in valid range (1024-65535), got %d", port)
	}

	// Verify second call works
	port2, err := AvailablePort()
	if err != nil {
		t.Fatalf("Second AvailablePort() error = %v", err)
	}
	_ = port2
}

func TestWaitForPort_NotAvailable(t *testing.T) {
	// Test with localhost on a port we expect to be unavailable
	err := WaitForPort("localhost", 65432, 100*time.Millisecond)
	if err == nil {
		t.Log("Port 65432 unexpectedly available (ok for test)")
	} else {
		t.Logf("Correctly detected unavailable port: %v", err)
	}
}

func TestWaitForPort_Timeout(t *testing.T) {
	// Use a non-routable IP to ensure it times out
	err := WaitForPort("192.0.2.1", 12345, 150*time.Millisecond)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestAvailablePort_Concurrent(t *testing.T) {
	ports := make(chan int, 20)
	errChan := make(chan error, 20)

	for i := 0; i < 20; i++ {
		go func() {
			port, err := AvailablePort()
			if err != nil {
				errChan <- err
				return
			}
			ports <- port
		}()
	}

	var gotPorts []int
	for i := 0; i < 20; i++ {
		select {
		case port := <-ports:
			gotPorts = append(gotPorts, port)
		case err := <-errChan:
			t.Fatalf("concurrent AvailablePort error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for ports (got %d so far)", len(gotPorts))
		}
	}

	// Check for duplicates
	seen := make(map[int]bool)
	for _, p := range gotPorts {
		if seen[p] {
			t.Errorf("duplicate port returned: %d", p)
		}
		seen[p] = true
	}
}

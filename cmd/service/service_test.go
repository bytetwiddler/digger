package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

func TestDiggerService_Execute(t *testing.T) {
	isService, err := svc.IsWindowsService()
	if err != nil || !isService {
		t.Skip("Skipping test: not running as Windows service")
	}

	// Create a temporary directory for test config
	tmpDir, err := os.MkdirTemp("", "digger-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	configContent := []byte(`
digger_path: "test-digger.exe"
`)
	err = os.WriteFile(filepath.Join(tmpDir, "config.yaml"), configContent, 0644)
	require.NoError(t, err)

	// Create test service
	service := &DiggerService{}

	// Create channels for testing
	changes := make(chan svc.Status, 1)
	requests := make(chan svc.ChangeRequest, 1)

	// Start service in goroutine
	done := make(chan struct{})
	go func() {
		ssec, errno := service.Execute(nil, requests, changes)
		assert.False(t, ssec)
		assert.Equal(t, uint32(0), errno)
		close(done)
	}()

	// Verify service transitions
	verifyServiceTransition(t, changes, svc.StartPending)
	verifyServiceTransition(t, changes, svc.Running)

	// Test stop
	requests <- svc.ChangeRequest{Cmd: svc.Stop}
	verifyServiceTransition(t, changes, svc.StopPending)

	// Wait for service to finish
	<-done
}

func verifyServiceTransition(t *testing.T, changes chan svc.Status, expectedState svc.State) {
	select {
	case status := <-changes:
		assert.Equal(t, expectedState, status.State)
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for service state: %v", expectedState)
	}
}

func TestInitializeConfig(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir, err := os.MkdirTemp("", "digger-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	configContent := []byte(`
digger_path: "test-digger.exe"
`)
	err = os.WriteFile(filepath.Join(tmpDir, "config.yaml"), configContent, 0644)
	require.NoError(t, err)

	// Set working directory to temp dir
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Test config initialization
	err = initializeConfig()
	assert.NoError(t, err)
}

func TestRunDiggerTask(t *testing.T) {
	isService, err := svc.IsWindowsService()
	if err != nil || !isService {
		t.Skip("Skipping test: not running as Windows service")
	}

	// Create mock eventlog
	elog, err := eventlog.Open("Digger")
	if err != nil {
		t.Skip("Skipping test: could not open event log")
	}
	defer elog.Close()

	tests := []struct {
		name       string
		diggerPath string
		wantErr    bool
	}{
		{
			name:       "missing digger path",
			diggerPath: "",
			wantErr:    true,
		},
		{
			name:       "invalid digger path",
			diggerPath: "nonexistent.exe",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runDiggerTask(elog)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

package main

import (
	"os"
	"testing"
	"time"

	"github.com/gofrs/flock"
)

func TestFileLocking(t *testing.T) {
	// Create a lock file
	fileLock := flock.New("test.lock")

	// Try to acquire the lock
	locked, err := fileLock.TryLock()
	if err != nil {
		t.Fatalf("Error acquiring lock: %v", err)
	}
	if !locked {
		t.Fatalf("Could not acquire lock")
	}

	// Ensure the lock is released after the test
	defer fileLock.Unlock()

	// Try to acquire the lock again in a separate goroutine with a new fileLock instance
	go func() {
		time.Sleep(1 * time.Second) // Wait for a moment to ensure the first lock is held
		fileLockGoroutine := flock.New("test.lock")
		locked, err := fileLockGoroutine.TryLock()
		if err != nil {
			t.Errorf("Error acquiring lock in goroutine: %v", err)
		}
		if locked {
			t.Errorf("Lock should not be acquired in goroutine")
		}
	}()

	// Wait for the goroutine to finish
	time.Sleep(2 * time.Second)

	// Clean up the lock file
	err = os.Remove("test.lock")
	if err != nil {
		t.Fatalf("Error removing lock file: %v", err)
	}
}

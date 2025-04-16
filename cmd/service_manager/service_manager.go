package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bytetwiddler/digger/cmd/service"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func main() {
	install := flag.Bool("install", false, "Install the service")
	uninstall := flag.Bool("uninstall", false, "Uninstall the service")
	start := flag.Bool("start", false, "Start the service")
	stop := flag.Bool("stop", false, "Stop the service")
	flag.Parse()

	if *install {
		if err := installService("Digger", "Digger IP Monitor Service"); err != nil {
			logrus.Fatalf("Failed to install service: %v", err)
		}
		logrus.Info("Service installed successfully")
		return
	}

	if *uninstall {
		if err := uninstallService("Digger"); err != nil {
			logrus.Fatalf("Failed to uninstall service: %v", err)
		}
		logrus.Info("Service uninstalled successfully")
		return
	}

	if *start {
		if err := startService("Digger"); err != nil {
			logrus.Fatalf("Failed to start service: %v", err)
		}
		logrus.Info("Service started successfully")
		return
	}

	if *stop {
		if err := stopService("Digger"); err != nil {
			logrus.Fatalf("Failed to stop service: %v", err)
		}
		logrus.Info("Service stopped successfully")
		return
	}

	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		logrus.Fatalf("Failed to determine if session is interactive: %v", err)
	}

	if !isInteractive {
		if err := svc.Run("Digger", &service.DiggerService{}); err != nil {
			eventLog, _ := eventlog.Open("Digger")
			if eventLog != nil {
				eventLog.Error(1, fmt.Sprintf("Service failed: %v", err))
				eventLog.Close()
			}
			logrus.Fatalf("Service failed: %v", err)
		}
		return
	}

	logrus.Info("Use -install, -uninstall, -start, or -stop")
}

func installService(name, desc string) error {
	exepath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// Try to open existing service
	s, err := m.OpenService(name)
	if err == nil {
		// Service exists, close handle and wait a bit
		s.Close()
		time.Sleep(2 * time.Second)
		return fmt.Errorf("service %s already exists, please uninstall it first", name)
	}

	// Create new service
	s, err = m.CreateService(name, exepath, mgr.Config{
		DisplayName:  desc,
		StartType:    mgr.StartAutomatic,
		ErrorControl: mgr.ErrorNormal,
	})
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("failed to setup event logger: %w", err)
	}

	// Set up recovery actions
	recovery := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 60 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 120 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 300 * time.Second},
	}

	if err := s.SetRecoveryActions(recovery, uint32(86400)); err != nil {
		logrus.Warnf("Failed to set recovery actions: %v", err)
	}

	return nil
}

func startService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

func stopService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}
	defer s.Close()

	_, err = s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to send stop control: %w", err)
	}

	return nil
}

func uninstallService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}
	defer s.Close()

	// First, try to stop the service
	status, err := s.Query()
	if err == nil && status.State != svc.Stopped {
		_, err = s.Control(svc.Stop)
		if err != nil {
			logrus.Warnf("Failed to stop service: %v", err)
		}
		// Wait for the service to stop
		for i := 0; i < 60; i++ {
			status, err = s.Query()
			if err != nil || status.State == svc.Stopped {
				break
			}
			time.Sleep(time.Second)
		}
	}

	// Try to delete the service
	err = s.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	err = eventlog.Remove(name)
	if err != nil {
		logrus.Warnf("failed to remove event log: %v", err)
		// Don't return error here as the service might still be successfully uninstalled
	}

	// Wait for the service to be fully removed
	for i := 0; i < 10; i++ {
		m2, err := mgr.Connect()
		if err == nil {
			_, err = m2.OpenService(name)
			m2.Disconnect()
			if err != nil {
				// Service no longer exists
				return nil
			}
		}
		time.Sleep(time.Second)
	}

	return nil
}

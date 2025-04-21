package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

type DiggerService struct {
	stopOnce sync.Once
	stopChan chan struct{}
}

func (m *DiggerService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	// Initialize stop channel
	m.stopChan = make(chan struct{})

	// Report start pending
	changes <- svc.Status{
		State:    svc.StartPending,
		WaitHint: 30000,
	}

	// Initialize event log
	elog, err := eventlog.Open("Digger")
	if err != nil {
		return true, 1
	}
	defer elog.Close()

	elog.Info(1, "Digger service starting")

	// Initialize configuration
	if err := initializeConfig(); err != nil {
		elog.Error(1, fmt.Sprintf("Failed to initialize config: %v", err))
		return true, 1
	}

	// Create ticker for periodic execution
	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()

	// Report running status
	changes <- svc.Status{
		State:   svc.Running,
		Accepts: cmdsAccepted,
	}

	elog.Info(1, "Digger service started")

	// Run first task immediately
	go func() {
		if err := runDiggerTask(elog); err != nil {
			elog.Error(1, fmt.Sprintf("Initial digger task failed: %v", err))
		}
	}()

	// Main service loop
	for {
		select {
		case <-m.stopChan:
			changes <- svc.Status{
				State:    svc.StopPending,
				WaitHint: 10000,
			}
			return false, 0
		case <-ticker.C:
			go func() {
				if err := runDiggerTask(elog); err != nil {
					elog.Error(1, fmt.Sprintf("Scheduled digger task failed: %v", err))
				}
			}()
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				m.stopOnce.Do(func() {
					elog.Info(1, "Digger service stopping")
					close(m.stopChan)
				})
			case svc.Pause:
				changes <- svc.Status{
					State:   svc.Paused,
					Accepts: cmdsAccepted,
				}
			case svc.Continue:
				changes <- svc.Status{
					State:   svc.Running,
					Accepts: cmdsAccepted,
				}
			default:
				elog.Error(1, fmt.Sprintf("Unexpected control request: %d", c))
			}
		}
	}
}

func initializeConfig() error {
	// Get the executable's directory
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exe)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(exeDir)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	return nil
}

func runDiggerTask(elog *eventlog.Log) error {
	cmdPath := viper.GetString("digger_path")
	if cmdPath == "" {
		return fmt.Errorf("digger_path not configured")
	}

	// Get the executable's directory
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exe)

	cmd := exec.Command(cmdPath)
	// Set the working directory to the executable's directory
	cmd.Dir = exeDir

	if output, err := cmd.CombinedOutput(); err != nil {

		msg := fmt.Sprintf("digger execution failed: %v, output: %s", err, string(output))
		elog.Error(1, msg)
		return fmt.Errorf(msg)
	}

	elog.Info(1, "Digger task completed successfully")
	return nil
}

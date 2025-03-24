// main.go
package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/bytetwiddler/digger/pkg/logging"
	"github.com/bytetwiddler/digger/pkg/site"
	"github.com/gofrs/flock"
	"github.com/sirupsen/logrus"
)

var (
	version string // These are set by the linker
	build   string //nolint:gochecknoglobals // These are set by the linker
	date    string //nolint:gochecknoglobals // These are set by the linker
	verbose bool   //nolint:gochecknoglobals // Set by the verbose flag
)

// ErrNoIPsFound custom error for no IP addresses found
var ErrNoIPsFound = errors.New("no IP addresses found for the domain")

// dig performs a DNS lookup on the hostnames in the sites slice and updates the IP field if the IP has changed.
func dig(sites *site.Sites) (bool, error) {
	changed := false

	for i := range *sites {
		ip, err := net.LookupIP((*sites)[i].Hostname)
		if err != nil {
			logrus.WithError(err).Errorf("Error resolving host %s", (*sites)[i].Hostname)
			_, _ = fmt.Fprintf(os.Stderr, "Error resolving host %s: %v\n", (*sites)[i].Hostname, err)
			continue
		}

		numberOfIPs := len(ip)
		if numberOfIPs == 0 {
			logrus.Errorf("No IP addresses found for %s", (*sites)[i].Hostname)
			_, _ = fmt.Fprintf(os.Stderr, "No IP addresses found for %s\n", (*sites)[i].Hostname)
			return false, ErrNoIPsFound
		}

		matchFound := false
		for _, addr := range ip {
			if (*sites)[i].IP == addr.String() {
				matchFound = true
				break
			}
		}

		if numberOfIPs > 1 {
			logrus.WithFields(logrus.Fields{
				"hostname":      (*sites)[i].Hostname,
				"number_of_ips": numberOfIPs,
				"match_found":   matchFound,
			}).Info("Multiple IP addresses found")
		}

		if !matchFound {
			logrus.WithFields(logrus.Fields{
				"hostname": (*sites)[i].Hostname,
				"old_ip":   (*sites)[i].IP,
				"new_ip":   ip[0].String(),
			}).Info("IP has changed")

			if verbose {
				_, _ = fmt.Fprintf(os.Stderr, "IP has changed for %s: %s -> %s\n", (*sites)[i].Hostname, (*sites)[i].IP, ip[0].String())
			}

			(*sites)[i].IP = ip[0].String()
			changed = true
		} else {
			logrus.WithFields(logrus.Fields{
				"hostname":      (*sites)[i].Hostname,
				"number_of_ips": numberOfIPs,
				"ip":            (*sites)[i].IP,
			}).Info("IP has not changed")

			if verbose {
				_, _ = fmt.Fprintf(os.Stdout, "IP has not changed for %s: %s\n", (*sites)[i].Hostname, (*sites)[i].IP)
			}
		}
	}

	return changed, nil
}

func run() error {
	// Read configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}

	// Setup logging
	logChannel, err := logging.SetupLogging(cfg)
	if err != nil {
		return fmt.Errorf("error setting up logging: %w", err)
	}
	defer close(logChannel)

	// Lock the sites.csv file
	fileLock := flock.New("sites.csv.lock")
	locked, err := fileLock.TryLock()
	if err != nil {
		return fmt.Errorf("error locking sites file: %w", err)
	}
	if !locked {
		return fmt.Errorf("could not acquire lock on sites file")
	}
	defer func() {
		if err := fileLock.Unlock(); err != nil {
			logrus.Errorf("error unlocking sites file: %v", err)
		}
	}()

	sites := site.Sites{}

	if err := sites.ReadFromFile("sites.csv"); err != nil {
		return fmt.Errorf("error reading sites file: %w", err)
	}

	changed, err := dig(&sites)
	if err != nil {
		return fmt.Errorf("error resolving IP addresses: %w", err)
	}

	if changed {
		if err := sites.WriteToFile("sites.csv"); err != nil {
			return fmt.Errorf("error writing sites file: %w", err)
		}
	}

	return nil
}

func main() {
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	if *versionFlag {
		_, _ = fmt.Fprintf(os.Stderr, "Version: %s, Build: %s, Date: %s\n", version, build, date)
		return
	}

	if err := run(); err != nil {
		logrus.Fatalf("Error: %v", err)
	}
}

package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/bytetwiddler/digger/config"
	"github.com/bytetwiddler/digger/logging"
	"github.com/gofrs/flock"
	"github.com/sirupsen/logrus"
)

var (
	version string //nolint:gochecknoglobals // These are set by the linker
	build   string //nolint:gochecknoglobals // These are set by the linker
	date    string //nolint:gochecknoglobals // These are set by the linker
)

type Site struct {
	Hostname   string `csv:"hostname"`
	Port       int    `csv:"port"`
	EntityName string `csv:"entity_name"`
	IP         string `csv:"ip"`
}

type Sites []Site

// Custom error for no IP addresses found
var ErrNoIPsFound = errors.New("no IP addresses found for the domain")

// writeToFile writes the sites to a CSV file.
func (s *Sites) writeToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Hostname", "Port", "EntityName", "IP"}
	err = writer.Write(header)
	if err != nil {
		return err
	}

	for _, site := range *s {
		record := []string{site.Hostname, strconv.Itoa(site.Port), site.EntityName, site.IP}
		err = writer.Write(record)
		if err != nil {
			return err
		}
	}

	return nil
}

// readFromFile reads the sites from a CSV file and populates the Sites slice.
func (s *Sites) readFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records[1:] { // Skip header row
		port, err := strconv.Atoi(record[1])
		if err != nil {
			return err
		}
		site := Site{
			Hostname:   record[0],
			Port:       port,
			EntityName: record[2],
			IP:         record[3],
		}
		*s = append(*s, site)
	}

	return nil
}

// dig performs a DNS lookup on the hostnames in the sites slice and updates the IP field if the IP has changed.
func dig(sites *Sites) (bool, error) {
	changed := false
	for i := range *sites {
		ip, err := net.LookupIP((*sites)[i].Hostname)
		if err != nil {
			logrus.Errorf("Error resolving %s: %v", (*sites)[i].Hostname, err)
			continue
		}
		numberOfIPs := len(ip)
		if numberOfIPs == 0 {
			logrus.Errorf("No IP addresses found for %s", (*sites)[i].Hostname)
			continue
		}
		matchFound := false
		for _, addr := range ip {
			if (*sites)[i].IP == addr.String() {
				matchFound = true
				break
			}
		}
		if !matchFound {
			fmt.Fprintf(os.Stderr, "%v IP HAS changed: %v -> %v\n", (*sites)[i].Hostname, (*sites)[i].IP, ip[0].String())
			logrus.Infof("%v IP HAS changed: %v -> %v", (*sites)[i].Hostname, (*sites)[i].IP, ip[0].String())
			(*sites)[i].IP = ip[0].String()
			changed = true
		} else {
			fmt.Fprintf(os.Stderr, "%v Number of IP's returned: %v IP was in the set: %v <-> %v\n",
				(*sites)[i].Hostname, numberOfIPs, (*sites)[i].IP, ip[0].String())
			logrus.Infof("%v Number of IP's returned: %v IP has NOT changed: %v <-> %v",
				(*sites)[i].Hostname, numberOfIPs, (*sites)[i].IP, ip[0].String())
		}
	}
	return changed, nil
}

func main() {
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Fprintf(os.Stderr, "Version: %s, Build: %s, Date: %s\n", version, build, date)
		return
	}

	// Read configuration
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		logrus.Fatalf("Error opening config file: %v", err)
	}

	// Setup logging
	logChannel, err := logging.SetupLogging(config)
	if err != nil {
		logrus.Fatalf("Error setting up logging: %v", err)
	}

	// Debug: Log a test message
	logrus.Println("Test log message")

	// Lock the sites.csv file
	fileLock := flock.New("sites.csv.lock")
	locked, err := fileLock.TryLock()
	if err != nil {
		logrus.Fatalf("Error locking sites file: %v", err)
	}
	if !locked {
		logrus.Fatalf("Could not acquire lock on sites file")
	}
	defer fileLock.Unlock()

	sites := Sites{}

	err = sites.readFromFile("sites.csv")
	if err != nil {
		logrus.Fatalf("Error reading sites file: %v", err)
	}

	changed, err := dig(&sites)
	if err != nil {
		logrus.Fatalf("Error in dig function: %v", err)
	}

	if changed {
		err = sites.writeToFile("sites.csv")
		if err != nil {
			logrus.Fatalf("Error writing sites file: %v", err)
		}
	}

	// Close the log channel to stop the goroutine
	close(logChannel)
}

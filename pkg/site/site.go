package site

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/bytetwiddler/digger/pkg/notification"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	"golang.org/x/sys/windows/svc/eventlog"
)

type Site struct {
	Hostname   string
	Port       int
	EntityName string
	IP         string   // Stores semicolon-separated IPs as a single string
	IPs        []string // Stores the split IPs for easier processing
	OldIP      string
	NewIP      string
	Changed    bool
	ChangeTime time.Time
}

type Sites []Site

func (s *Sites) ReadFromCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open csv file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read from csv file: %w", err)
	}

	// Skip the header row
	for _, record := range records[1:] {
		port, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("invalid port value: %w", err)
		}

		// Split the IP field by semicolon and trim spaces
		ips := strings.Split(record[3], ";")
		for i := range ips {
			ips[i] = strings.TrimSpace(ips[i])
		}

		// Filter out empty strings
		validIPs := make([]string, 0, len(ips))
		for _, ip := range ips {
			if ip != "" {
				validIPs = append(validIPs, ip)
			}
		}

		site := Site{
			Hostname:   record[0],
			Port:       port,
			EntityName: record[2],
			IP:         record[3], // Keep original string
			IPs:        validIPs,  // Store split IPs
		}
		*s = append(*s, site)
	}

	return nil
}

func (s *Sites) WriteToCSV(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to write to csv: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row
	err = writer.Write([]string{"Hostname", "Port", "EntityName", "IP", "OldIP", "NewIP", "ChangeTime"})
	if err != nil {
		return fmt.Errorf("failed to write csv header row: %w", err)
	}

	// Write the site records
	for _, site := range *s {
		changeTime := ""
		if !site.ChangeTime.IsZero() {
			changeTime = site.ChangeTime.Format(time.RFC3339)
		}

		err = writer.Write([]string{
			site.Hostname,
			strconv.Itoa(site.Port),
			site.EntityName,
			site.IP,
			site.OldIP,
			site.NewIP,
			changeTime,
		})
		if err != nil {
			return fmt.Errorf("failed to write csv site records: %w", err)
		}
	}

	return nil
}

func (s *Sites) UpdateIPs(cfg *config.Config, db *bbolt.DB, updateFlag bool) error {
	// Initialize Windows event log
	elog, err := eventlog.Open("Digger")
	if err != nil {
		logrus.Warnf("Failed to open Windows event log: %v", err)
		// Continue execution even if event log fails
	} else {
		defer elog.Close()
	}

	for i, site := range *s {
		// Proceed with IP lookup
		dnsIPs, err := net.LookupIP(site.Hostname)
		if err != nil {
			logrus.Errorf("Failed to lookup IP for %s: %v", site.Hostname, err)
			continue
		}

		// Convert DNS IPs to strings and filter for IPv4 only
		dnsIPStrings := make([]string, 0, len(dnsIPs))
		for _, ip := range dnsIPs {
			if ip.To4() != nil {
				dnsIPStrings = append(dnsIPStrings, ip.String())
			}
		}

		if len(dnsIPStrings) == 0 {
			logrus.Warnf("No IPv4 addresses found for %s", site.Hostname)
			continue
		}

		// If update flag is set and we have multiple IPs in the CSV
		if updateFlag && len(site.IPs) > 1 {
			msg := fmt.Sprintf("Multiple IPs in sites.csv for IP field on record %v, manual intervention required, list of IPs %v and resolved IP is %v. The -update flag is set, so the IPs will be updated in the database.",
				site.Hostname, site.IPs, dnsIPStrings)

			// Log to file
			logrus.Warn(msg)

			// Log to Windows Event Log if available
			if elog != nil {
				elog.Warning(1, msg)
			}

			// Skip processing this record
			continue
		}

		// Check if any of the current IPs match any of the resolved IPs
		ipFound := false
		if len(site.IPs) > 0 {
			for _, currentIP := range site.IPs {
				for _, dnsIP := range dnsIPStrings {
					if currentIP == dnsIP {
						ipFound = true
						break
					}
				}
				if ipFound {
					break
				}
			}
		} else {
			// If no IPs array, check the single IP field
			for _, dnsIP := range dnsIPStrings {
				if site.IP == dnsIP {
					ipFound = true
					break
				}
			}
		}

		// If no match was found
		if !ipFound {
			(*s)[i].OldIP = site.IP
			(*s)[i].NewIP = dnsIPStrings[0] // Use first IP if multiple are returned
			(*s)[i].IP = (*s)[i].NewIP
			(*s)[i].Changed = true
			(*s)[i].ChangeTime = time.Now()

			msg := fmt.Sprintf("IP address for %s changed from %s to %s",
				site.Hostname, (*s)[i].OldIP, (*s)[i].NewIP)
			logrus.Info(msg)

			// Send an email notification
			logrus.Infof("Sending email notification to %s", cfg.SMTP.To)
			err = notification.SendIPChangeNotification(cfg, site.Hostname, site.Port,
				site.EntityName, (*s)[i].OldIP, (*s)[i].NewIP)
			if err != nil {
				logrus.Errorf("Failed to send email notification: %v", err)
			}

			// Persist the change in the database
			err = s.persistSiteChange(db, &(*s)[i])
			if err != nil {
				logrus.Errorf("Failed to persist change for %s: %v", site.Hostname, err)
			}
		}
	}

	return nil
}

func (s *Sites) ReadFromDB(db *bbolt.DB) error {
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sites"))
		if b == nil {
			return errors.New("bucket not found")
		}

		return b.ForEach(func(k, v []byte) error {
			var site Site
			err := json.Unmarshal(v, &site)
			if err != nil {
				logrus.Errorf("Failed to unmarshal site data for key %s: %v", k, err)
				return nil // Skip invalid entries
			}

			*s = append(*s, site)
			return nil
		})
	})
}

func (s *Sites) WriteToDB(db *bbolt.DB) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sites"))
		if b == nil {
			return errors.New("bucket not found")
		}

		for _, site := range *s {
			data, err := json.Marshal(site)
			if err != nil {
				return fmt.Errorf("failed to marshal site json: %w", err)
			}

			err = b.Put([]byte(site.Hostname), data)
			if err != nil {
				return fmt.Errorf("failed to put site.Name: %w", err)
			}
		}

		return nil
	})
}

func (s *Sites) persistSiteChange(db *bbolt.DB, site *Site) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sites"))
		if b == nil {
			return errors.New("bucket not found")
		}

		data, err := json.Marshal(site)
		if err != nil {
			return fmt.Errorf("failed to marshal json: %w", err)
		}

		err = b.Put([]byte(site.Hostname), data)
		if err != nil {
			return fmt.Errorf("failed to put site.Name: %w", err)
		}

		// Store the change in the changes bucket
		cb := tx.Bucket([]byte("changes"))
		if cb == nil {
			cb, err = tx.CreateBucket([]byte("changes"))
			if err != nil {
				return fmt.Errorf("failed to store changes in db: %w", err)
			}
		}

		changeKey := fmt.Sprintf("%s-%s", site.Hostname, time.Now().Format(time.RFC3339))
		return cb.Put([]byte(changeKey), data)
	})
}

func (s *Sites) ReportChanges(db *bbolt.DB) error {
	return db.View(func(tx *bbolt.Tx) error {
		cb := tx.Bucket([]byte("changes"))
		if cb == nil {
			return errors.New("changes bucket not found")
		}

		return cb.ForEach(func(k, v []byte) error {
			var site Site
			err := json.Unmarshal(v, &site)
			if err != nil {
				logrus.Errorf("Failed to unmarshal site data for key %s: %v", k, err)
				return nil // Skip invalid entries
			}

			fmt.Printf("Site: %s, Old IP: %s, New IP: %s, Timestamp: %s\n",
				site.Hostname, site.OldIP, site.NewIP, k)
			return nil
		})
	})
}

func (s *Sites) CountRecords(db *bbolt.DB) (int, error) {
	count := 0
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("changes"))
		if b == nil {
			return errors.New("bucket not found")
		}

		return b.ForEach(func(k, v []byte) error {
			count++
			return nil
		})
	})

	return count, err
}

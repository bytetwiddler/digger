package site

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/bytetwiddler/digger/pkg/email"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

type Site struct {
	Hostname   string
	Port       int
	EntityName string
	IP         string
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
		*s = append(*s, Site{Hostname: record[0], Port: port, EntityName: record[2], IP: record[3]})
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
		err = writer.Write([]string{site.Hostname, strconv.Itoa(site.Port), site.EntityName, site.IP, "", "", ""})
		if err != nil {
			return fmt.Errorf("failed to write csv site records: %w", err)
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

func (s *Sites) UpdateIPs(cfg *config.Config, db *bbolt.DB) error {
	for i, site := range *s {
		ips, err := net.LookupIP(site.Hostname)
		if err != nil {
			logrus.Errorf("Failed to lookup IP for %s: %v", site.Hostname, err)

			continue
		}

		ipChanged := true

		for _, ip := range ips {
			if ip.String() == site.IP {
				ipChanged = false

				break
			}
		}

		if ipChanged && len(ips) > 0 {
			(*s)[i].OldIP = site.IP
			(*s)[i].NewIP = ips[0].String()
			(*s)[i].IP = ips[0].String()
			(*s)[i].Changed = true
			msg := fmt.Sprintf("IP address for %s changed from %s to %s", site.Hostname, (*s)[i].OldIP, (*s)[i].NewIP)
			logrus.Infof(msg)

			// Send an email notification
			logrus.Infof("Sending email notification to %s", cfg.SMTP.To)
			err = email.SendEmail(cfg, "IP Address Change", msg)

			if err != nil {
				logrus.Errorf("Failed to send email notification: %v", err)
			}

			// Persist the change in the database
			err = db.Update(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte("sites"))
				if b == nil {
					return errors.New("bucket not found")
				}

				data, err := json.Marshal((*s)[i])
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
			if err != nil {
				logrus.Errorf("Failed to persist change for %s: %v", site.Hostname, err)
			}
		}
	}

	return nil
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

			fmt.Printf("Site: %s, Old IP: %s, New IP: %s, Timestamp: %s\n", site.Hostname, site.OldIP, site.NewIP, k)

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

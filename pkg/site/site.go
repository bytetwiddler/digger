package site

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	"net"
	"os"
	"time"
)

type Site struct {
	Name       string
	Port       string
	EntityName string
	IP         string
	OldIP      string
	NewIP      string
	Changed    bool
}

type Sites []Site

func (s *Sites) ReadFromCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// Skip the header row
	for _, record := range records[1:] {
		*s = append(*s, Site{Name: record[0], Port: record[1], EntityName: record[2], IP: record[3]})
	}

	return nil
}

func (s *Sites) WriteToCSV(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row
	err = writer.Write([]string{"Hostname", "Port", "EntityName", "IP", "OldIP", "NewIP", "ChangeTime"})
	if err != nil {
		return err
	}

	// Write the site records
	for _, site := range *s {
		err = writer.Write([]string{site.Name, site.Port, site.EntityName, site.IP, "", "", ""})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sites) ReadFromDB(db *bbolt.DB) error {
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sites"))
		if b == nil {
			return fmt.Errorf("bucket not found")
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
			return fmt.Errorf("bucket not found")
		}
		for _, site := range *s {
			data, err := json.Marshal(site)
			if err != nil {
				return err
			}
			err = b.Put([]byte(site.Name), data)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Sites) UpdateIPs(db *bbolt.DB) error {
	for i, site := range *s {
		ips, err := net.LookupIP(site.Name)
		if err != nil {
			logrus.Errorf("Failed to lookup IP for %s: %v", site.Name, err)
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
			logrus.Infof("IP address for %s changed from %s to %s", site.Name, (*s)[i].OldIP, (*s)[i].NewIP)

			// Persist the change in the database
			err = db.Update(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte("sites"))
				if b == nil {
					return fmt.Errorf("bucket not found")
				}
				data, err := json.Marshal((*s)[i])
				if err != nil {
					return err
				}
				err = b.Put([]byte(site.Name), data)
				if err != nil {
					return err
				}

				// Store the change in the changes bucket
				cb := tx.Bucket([]byte("changes"))
				if cb == nil {
					cb, err = tx.CreateBucket([]byte("changes"))
					if err != nil {
						return err
					}
				}
				changeKey := fmt.Sprintf("%s-%s", site.Name, time.Now().Format(time.RFC3339))
				return cb.Put([]byte(changeKey), data)
			})
			if err != nil {
				logrus.Errorf("Failed to persist change for %s: %v", site.Name, err)
			}
		}
	}
	return nil
}

func (s *Sites) ReportChanges(db *bbolt.DB) error {
	return db.View(func(tx *bbolt.Tx) error {
		cb := tx.Bucket([]byte("changes"))
		if cb == nil {
			return fmt.Errorf("changes bucket not found")
		}
		return cb.ForEach(func(k, v []byte) error {
			var site Site
			err := json.Unmarshal(v, &site)
			if err != nil {
				logrus.Errorf("Failed to unmarshal site data for key %s: %v", k, err)
				return nil // Skip invalid entries
			}
			fmt.Printf("Site: %s, Old IP: %s, New IP: %s, Timestamp: %s\n", site.Name, site.OldIP, site.NewIP, k)
			return nil
		})
	})
}

func (s *Sites) CountRecords(db *bbolt.DB) (int, error) {
	count := 0
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sites"))
		if b == nil {
			return fmt.Errorf("bucket not found")
		}
		return b.ForEach(func(k, v []byte) error {
			count++
			return nil
		})
	})
	return count, err
}

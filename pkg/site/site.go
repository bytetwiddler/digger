package site

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// Site represents a site with its details.
type Site struct {
	Hostname   string `csv:"hostname"`
	Port       int    `csv:"port"`
	EntityName string `csv:"entity_name"`
	IP         string `csv:"ip"`
}

// Sites is a slice of Site.
type Sites []Site

// WriteToFile writes the sites to a CSV file.
func (s *Sites) WriteToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Hostname", "Port", "EntityName", "IP"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, site := range *s {
		record := []string{site.Hostname, strconv.Itoa(site.Port), site.EntityName, site.IP}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// ReadFromFile reads the sites from a CSV file and populates the Sites slice.
func (s *Sites) ReadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	if err != nil {
		return fmt.Errorf("failed to read records: %w", err)
	}

	for _, record := range records[1:] { // Skip header row
		port, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("failed to parse port: %w", err)
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

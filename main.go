package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/Graylog2/go-gelf/gelf"
	"github.com/gofrs/flock"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Log struct {
		Level string `yaml:"level"`
		Gelf  struct {
			Address string `yaml:"address"`
		} `yaml:"gelf"`
	} `yaml:"log"`
}

type Site struct {
	Hostname   string `csv:"hostname"`
	Port       int    `csv:"port"`
	EntityName string `csv:"entity_name"`
	IP         string `csv:"ip"`
}

type Sites []Site

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

func dig(sites *Sites) (bool, error) {
	changed := false
	for i := range *sites {
		ip, err := net.LookupIP((*sites)[i].Hostname)
		if err != nil {
			log.Printf("Error: %v", err)
			return false, err
		}
		if (*sites)[i].IP != ip[0].String() {
			fmt.Printf("%v IP HAS changed: %v -> %v\n", (*sites)[i].Hostname, (*sites)[i].IP, ip[0].String())
			log.Printf("%v IP HAS changed: %v -> %v", (*sites)[i].Hostname, (*sites)[i].IP, ip[0].String())
			(*sites)[i].IP = ip[0].String()
			changed = true
		} else {
			fmt.Printf("%v IP has NOT changed: %v <-> %v\n", (*sites)[i].Hostname, (*sites)[i].IP, ip[0].String())
			log.Printf("%v IP has NOT changed: %v <-> %v", (*sites)[i].Hostname, (*sites)[i].IP, ip[0].String())
		}
	}
	return changed, nil
}

func main() {
	// Read configuration
	configFile, err := os.Open("config.yaml")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer configFile.Close()

	var config Config
	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding config file: %v", err)
	}

	// Set up GELF logger
	gelfWriter, err := gelf.NewWriter(config.Log.Gelf.Address)
	if err != nil {
		log.Fatalf("Error setting up GELF logger: %v", err)
	}
	log.SetOutput(gelfWriter)

	// Lock the sites.csv file
	fileLock := flock.New("sites.csv.lock")
	locked, err := fileLock.TryLock()
	if err != nil {
		log.Fatalf("Error locking sites file: %v", err)
	}
	if !locked {
		log.Fatalf("Could not acquire lock on sites file")
	}
	defer fileLock.Unlock()

	var sites Sites
	err = sites.readFromFile("sites.csv")
	if err != nil {
		log.Fatalf("Error reading sites file: %v", err)
	}

	changed, err := dig(&sites)
	if err != nil {
		log.Fatalf("Error in dig function: %v", err)
	}

	if changed {
		err = sites.writeToFile("sites.csv")
		if err != nil {
			log.Fatalf("Error writing sites file: %v", err)
		}
	}
}

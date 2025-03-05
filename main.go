package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strconv"
)

type Site struct {
	Hostname   string `csv:"hostname"`
	Port       int    `csv:"port"`
	EntityName string `csv:"entity_name"`
	IP         string `csv:"ip"`
}

type Sites []Site

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
		}
		*s = append(*s, site)
	}

	return nil
}

func dig(sites *Sites) error {
	for i := range *sites {
		ip, err := net.LookupIP((*sites)[i].Hostname)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		(*sites)[i].IP = ip[0].String()
	}
	return nil
}

func main() {
	var sites Sites
	err := sites.readFromFile("sites.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(sites)
	err = dig(&sites)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(sites)
}

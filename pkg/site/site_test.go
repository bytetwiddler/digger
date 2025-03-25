package site

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriteToFile(t *testing.T) {
	sites := Sites{
		{
			Hostname:   "example.com",
			Port:       80,
			EntityName: "ExampleEntity",
			IP:         "192.168.1.1",
			OldIP:      "",
			NewIP:      "",
			ChangeTime: time.Time{},
		},
		{
			Hostname:   "testsite.com",
			Port:       443,
			EntityName: "TestEntity",
			IP:         "192.168.1.2",
			OldIP:      "",
			NewIP:      "",
			ChangeTime: time.Time{},
		},
	}

	err := sites.WriteToCSV("test_sites.csv")
	assert.NoError(t, err)
	defer os.Remove("test_sites.csv")

	expectedContent := "Hostname,Port,EntityName,IP,OldIP,NewIP,ChangeTime\n" +
		"example.com,80,ExampleEntity,192.168.1.1,,,\n" +
		"testsite.com,443,TestEntity,192.168.1.2,,,\n"

	content, err := os.ReadFile("test_sites.csv")
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, string(content))
}

func TestReadFromFile(t *testing.T) {
	content := "Hostname,Port,EntityName,IP,OldIP,NewIP,ChangeTime\n" +
		"example.com,80,ExampleEntity,192.168.1.1,,,\n" +
		"testsite.com,443,TestEntity,192.168.1.2,,,\n"

	err := os.WriteFile("test_sites.csv", []byte(content), 0644)
	assert.NoError(t, err)
	defer os.Remove("test_sites.csv")

	var sites Sites
	err = sites.ReadFromCSV("test_sites.csv")
	assert.NoError(t, err)

	expectedSites := Sites{
		{
			Hostname:   "example.com",
			Port:       80,
			EntityName: "ExampleEntity",
			IP:         "192.168.1.1",
			OldIP:      "",
			NewIP:      "",
		},
		{
			Hostname:   "testsite.com",
			Port:       443,
			EntityName: "TestEntity",
			IP:         "192.168.1.2",
			OldIP:      "",
			NewIP:      "",
		},
	}

	assert.Equal(t, expectedSites, sites)
}

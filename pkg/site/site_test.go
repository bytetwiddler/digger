package site

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteToFile(t *testing.T) {
	sites := Sites{
		{Hostname: "example.com", Port: 80, EntityName: "ExampleEntity", IP: "192.168.1.1"},
		{Hostname: "testsite.com", Port: 443, EntityName: "TestEntity", IP: "192.168.1.2"},
	}

	// Create a temporary file
	file, err := os.CreateTemp("", "sites.csv")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Test WriteToFile
	err = sites.WriteToFile(file.Name())
	assert.NoError(t, err)

	// Verify file content
	expectedContent := "Hostname,Port,EntityName,IP\nexample.com,80,ExampleEntity,192.168.1.1\ntestsite.com,443,TestEntity,192.168.1.2\n"
	content, err := os.ReadFile(file.Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, string(content))
}

func TestReadFromFile(t *testing.T) {
	// Create a temporary file with test data
	file, err := os.CreateTemp("", "sites.csv")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	content := "Hostname,Port,EntityName,IP\nexample.com,80,ExampleEntity,192.168.1.1\ntestsite.com,443,TestEntity,192.168.1.2\n"
	_, err = file.Write([]byte(content))
	assert.NoError(t, err)
	file.Close()

	var sites Sites

	// Test ReadFromFile
	err = sites.ReadFromFile(file.Name())
	assert.NoError(t, err)

	// Verify sites content
	expectedSites := Sites{
		{Hostname: "example.com", Port: 80, EntityName: "ExampleEntity", IP: "192.168.1.1"},
		{Hostname: "testsite.com", Port: 443, EntityName: "TestEntity", IP: "192.168.1.2"},
	}
	assert.Equal(t, expectedSites, sites)
}

func TestReadFromFile_InvalidPort(t *testing.T) {
	// Create a temporary file with invalid port data
	file, err := os.CreateTemp("", "sites.csv")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	content := "Hostname,Port,EntityName,IP\nexample.com,invalid_port,ExampleEntity,192.168.1.1\n"
	_, err = file.Write([]byte(content))
	assert.NoError(t, err)
	file.Close()

	var sites Sites

	// Test ReadFromFile with invalid port
	err = sites.ReadFromFile(file.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse port")
}

package site

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteToFile(t *testing.T) {
	// Test data
	sites := Sites{
		{
			Hostname:   "example.com",
			Port:       80,
			EntityName: "ExampleEntity",
			IP:         "192.168.1.1",
		},
		{
			Hostname:   "testsite.com",
			Port:       443,
			EntityName: "TestEntity",
			IP:         "192.168.1.2",
		},
	}

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "sites_test_*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write to the temporary file
	err = sites.WriteToCSV(tmpfile.Name())
	require.NoError(t, err)

	// Read and verify the content
	var readSites Sites
	err = readSites.ReadFromCSV(tmpfile.Name())
	require.NoError(t, err)

	assert.Equal(t, len(sites), len(readSites))
	for i := range sites {
		assert.Equal(t, sites[i].Hostname, readSites[i].Hostname)
		assert.Equal(t, sites[i].Port, readSites[i].Port)
		assert.Equal(t, sites[i].EntityName, readSites[i].EntityName)
		assert.Equal(t, sites[i].IP, readSites[i].IP)
	}
}

func TestReadFromFile(t *testing.T) {
	// Create a temporary file with test data
	content := `hostname,port,entity_name,ip
example.com,80,ExampleEntity,192.168.1.1
testsite.com,443,TestEntity,192.168.1.2`

	tmpfile, err := os.CreateTemp("", "sites_test_*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	err = os.WriteFile(tmpfile.Name(), []byte(content), 0644)
	require.NoError(t, err)

	// Test reading the file
	var sites Sites
	err = sites.ReadFromCSV(tmpfile.Name())
	require.NoError(t, err)

	// Expected data with IPs slice populated from IP field
	expectedSites := Sites{
		{
			Hostname:   "example.com",
			Port:       80,
			EntityName: "ExampleEntity",
			IP:         "192.168.1.1",
			IPs:        []string{"192.168.1.1"},
			ChangeTime: time.Time{},
		},
		{
			Hostname:   "testsite.com",
			Port:       443,
			EntityName: "TestEntity",
			IP:         "192.168.1.2",
			IPs:        []string{"192.168.1.2"},
			ChangeTime: time.Time{},
		},
	}

	assert.Equal(t, expectedSites, sites)
}

func TestReadFromInvalidFile(t *testing.T) {
	var sites Sites
	err := sites.ReadFromCSV("nonexistent.csv")
	assert.Error(t, err)
}

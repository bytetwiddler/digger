// main_test.go
package main

import (
	"os"
	"testing"

	"github.com/bytetwiddler/digger/pkg/site"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	// Setup temporary config file
	configContent := `
log:
  level: "debug"
  gelf:
    address: "127.0.0.1:12201"
  file:
    filename: "test.log"
    maxsize: 5
    maxbackups: 2
    maxage: 15
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	assert.NoError(t, err)
	defer os.Remove("config.yaml")

	// Setup temporary sites file
	sitesContent := `hostname,port,entity_name,ip
example.com,80,example_entity,1.2.3.4
`
	err = os.WriteFile("sites.csv", []byte(sitesContent), 0644)
	assert.NoError(t, err)
	defer os.Remove("sites.csv")

	// Run the main function
	err = run()
	assert.NoError(t, err)

	// Verify the sites file was updated
	sites := site.Sites{}
	err = sites.ReadFromFile("sites.csv")
	assert.NoError(t, err)
	assert.Equal(t, "example.com", sites[0].Hostname)
	assert.NotEqual(t, "1.2.3.4", sites[0].IP)
}

package main

import (
	"flag"
	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/bytetwiddler/digger/pkg/logging"
	"github.com/bytetwiddler/digger/pkg/site"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	"log"
)

func main() {
	// Define the report and update flags
	report := flag.Bool("report", false, "Report changes from the database")
	update := flag.Bool("update", false, "Update IP addresses in the CSV file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Set up logging
	file, err := logging.SetupLogging(cfg)
	if err != nil {
		log.Fatalf("failed to set up logging: %v", err)
	}
	defer file.Close()

	// Open the bbolt database
	db, err := bbolt.Open(cfg.DB.Path, 0600, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

	// Ensure the buckets are created before reading from them
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("sites"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("changes"))
		return err
	})
	if err != nil {
		logrus.Fatalf("failed to create buckets: %v", err)
	}

	var sites site.Sites

	// Read sites from CSV
	err = sites.ReadFromCSV("sites.csv")
	if err != nil {
		logrus.Fatalf("failed to read from csv: %v", err)
	}

	// If the report flag is set, report changes and exit
	if *report {
		err = sites.ReportChanges(db)
		if err != nil {
			logrus.Fatalf("failed to report changes: %v", err)
		}
		count, err := sites.CountRecords(db)
		if err != nil {
			logrus.Fatalf("failed to count records: %v", err)
		}
		logrus.Infof("Total number of records: %d", count)
		logrus.Info("Report completed successfully")
		return
	}

	// Update IPs and log changes
	err = sites.UpdateIPs(db)
	if err != nil {
		logrus.Fatalf("failed to update IPs: %v", err)
	}

	// Write sites to the database
	err = sites.WriteToDB(db)
	if err != nil {
		logrus.Fatalf("failed to write to db: %v", err)
	}

	// If the update flag is set, update the CSV file with the new IPs
	if *update {
		err = sites.WriteToCSV("sites.csv")
		if err != nil {
			logrus.Fatalf("failed to write to csv: %v", err)
		}
		logrus.Info("CSV file updated successfully")
	}

	logrus.Info("Operation completed successfully")
}

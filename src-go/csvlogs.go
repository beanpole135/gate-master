package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// toCSV converts a LogEntry struct into a slice of strings
// suitable for writing to a CSV file.

var csvheader []string = []string{"Timestamp", "GateOpened", "OpenedBy", "OpenedHow", "AccountID"}

func toCSV(entry GateLog) []string {
	log := []string{
		entry.TimeOpened.Format(time.RFC3339),
		boolToString(entry.Success),
		entry.OpenedName,
	}
	if entry.UsedWeb {
		log = append(log, "Website")
	} else {
		log = append(log, fmt.Sprintf("PIN:%s", entry.UsedCode))
	}
	log = append(log, fmt.Sprintf("%d", entry.AccountID))
	return log
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// SaveLogEntry saves a log entry to a CSV file organized by Year/Month/Day.
func SaveCSVLog(entry GateLog, basedir string) error {
	// Get the current year, month, and day
	year := strconv.Itoa(entry.TimeOpened.Year())
	month := fmt.Sprintf("%2.f-", float32(entry.TimeOpened.Month())) + entry.TimeOpened.Month().String()
	day := strconv.Itoa(entry.TimeOpened.Day())

	// Construct the directory path
	dirPath := filepath.Join(basedir, year, month)

	// Create the directory structure if it doesn't exist
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	picdir := dirPath + "/pictures"
	// Create the directory structure if it doesn't exist
	if err := os.MkdirAll(picdir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Construct the full file path
	filePath := filepath.Join(dirPath, day)

	// See if the file exists first
	newfile := true
	if _, err := os.Stat(filePath); err == nil {
		newfile = false
	}

	// Open the file in append mode, or create it if it doesn't exist
	file, err := os.OpenFile(filePath+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if newfile {
		//Write the header line to the CSV file
		_ = writer.Write(csvheader)
	}
	// Convert the LogEntry to a CSV record
	record := toCSV(entry)

	// Write the record to the CSV file
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write to CSV: %w", err)
	}

	//Also write the picture to a jpg file
	if entry.HasImage() {
		picfile := picdir + "/" + entry.TimeOpened.Format("2006-01-02_03_04PM") + ".jpg"
		_ = os.WriteFile(picfile, entry.GatePicture, 0644)
	}
	return nil
}

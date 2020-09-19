package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os/exec"
	"strings"
)

const (
	updateJobSQL = `UPDATE jobs
    SET status = ?
    WHERE id = ?`
)

// Worker that run nmap scan form queue in channel
// TODO check is jobChan needed as param?
func worker(i int) {
	// Open connection to DB
	db, err := sql.Open("sqlite3", sqliteConnStr)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Start nmap every time when job appears in channel
	for job := range jobChan {
		// TODO send error to logger chanel
		if err := startNmap(job, i, db); err != nil {
			logChan <- logMessage(fmt.Sprintf("Scan %s failed by worker %d: %s", job.Id, i, err))
		} else {
			logChan <- logMessage(fmt.Sprintf("Scan %s done by worker %d", job.Id, i))
		}
	}
}

// Run nmap and save result to XML file
func startNmap(job Job, i int, db *sql.DB) error {
	options, err := jobOptions(job.Options, job.Id)
	if err != nil {
		return err
	}
	cmd := exec.Command("sudo", options...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return updateFailed(db, job.Id, err)
	}
	// TODO remove it
	fmt.Println("Scan done!!!!")
	return updateFinished(db, job.Id)
}

func jobOptions(options string, id string) ([]string, error) {
	nmapPath, err := exec.LookPath("nmap")
	if err != nil {
		return nil, fmt.Errorf("Nmap path lookup error: %s", err)
	}
	outputPath := getPath(id)
	o := fmt.Sprintf("%s %s -oX %s", nmapPath, options, outputPath)
	return strings.Split(o, " "), nil
}

func updateRecord(db *sql.DB, id string, status string) error {
	mutex.Lock()
	defer mutex.Unlock()
	_, err := db.Exec(updateJobSQL, status, id)
	if err != nil {
		return err
	}
	return nil
}

func getPath(id string) string {
	outputPath := fmt.Sprintf("%s.xml", id)
	return outputPath
}

func updateFailed(db *sql.DB, id string, e error) error {
	err := updateRecord(db, id, "1")
	if err != nil {
		return fmt.Errorf("%s; DB update error: %s", e, err)
	} else {
		return e
	}
}

func updateFinished(db *sql.DB, id string) error {
	err := updateRecord(db, id, "3")
	if err != nil {
		return fmt.Errorf("DB update error: %s", err)
	}
	return nil
}

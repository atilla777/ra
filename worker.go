package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// Worker that run nmap scan form queue in channel
// TODO check is jobChan needed as param?
func worker(i int) {
	// Start nmap every time when job appears in channel
	for job := range jobChan {
		// TODO send error to logger chanel
		if err := startNmap(job, i); err != nil {
			logChan <- logMessage(fmt.Sprintf("Scan %s failed by worker %d: %s", job.Id, i, err))
		} else {
			logChan <- logMessage(fmt.Sprintf("Scan %s done by worker %d", job.Id, i))
		}
	}
}

// Run nmap and save result to XML file
func startNmap(job Job, i int) error {
	options, err := jobOptions(job.Options, job.Id)
	if err != nil {
		return err
	}
	cmd := exec.Command("sudo", options...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return updateFailed(job.Id, err)
	}
	// TODO remove it
	fmt.Println("Scan done!!!!")
	return updateFinished(job.Id)
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

func updateRecord(id string, status int) error {
	updateJobSQL := "UPDATE jobs SET status = ? WHERE id = ?"
	err := execSQL(updateJobSQL, nil, status, id)
	if err != nil {
		return err
	}
	return nil
}

func getPath(id string) string {
	outputPath := fmt.Sprintf("%s.xml", id)
	return outputPath
}

func updateFailed(id string, e error) error {
	err := updateRecord(id, 1)
	if err != nil {
		return fmt.Errorf("%s; DB update error: %s", e, err)
	} else {
		return e
	}
}

func updateFinished(id string) error {
	err := updateRecord(id, 3)
	if err != nil {
		return fmt.Errorf("DB update error: %s", err)
	}
	return nil
}

package main

// Command for response test
// while true; do { echo -e "HTTP/1.1 200 OK\r\n$(date)\r\n\r\n<h1>hello world from $(hostname) on $(date)</h1>" |  nc -vl 3000; } done

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"os"
)

func sendResults() error {
	jobs, err := finishedJobs()
	if err != nil {
		return fmt.Errorf("Search finished jobs error: %s", err)
	}
	for _, j := range jobs {
		err := sendOneResult(j.Id)
		// TODO Add retry for send result
		if err != nil {
			logChan <- logMessage(fmt.Sprintf("Can`t send result of job %s: %s", j.Id, err))
		} else {
			logChan <- logMessage(fmt.Sprintf("Result %s sent.", j.Id))
			deleteJob(j.Id)
		}
	}
	return nil
}

func sendOneResult(id string) error {
	outputPath := getPath(id)
	_, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("Can`t find result file: %s", err)
	}
	resultJSON, err := nmapJSON(outputPath)
	if err != nil {
		return fmt.Errorf("Can`t convert result to JSON: %s", err)
	}
	rism_url := fmt.Sprintf(
		"%s://%s:%s%s",
		viper.GetString("rism.protocol"),
		viper.GetString("rism.host"),
		viper.GetString("rism.port"),
		viper.GetString("rism.path"),
	)
	client := &http.Client{}
	req, err := http.NewRequest("POST", rism_url, bytes.NewBuffer(resultJSON))
	if err != nil {
		return fmt.Errorf("Can`t make new request: %s", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", viper.GetString("ra.secret")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Can`t make request: %s", err)
	} else {
		// TODO remove it
		fmt.Println("Result sent")
		defer resp.Body.Close()
		return nil
	}
}

func deleteJob(id string) {
	if err := deleteFinishedJob(id); err != nil {
		logChan <- logMessage(fmt.Sprintf("Can`t delete job: %s", err))
	}
	if err := deleteFile(id); err != nil {
		logChan <- logMessage(fmt.Sprintf("Can`t delete file: %s", err))
	}
}

func deleteFile(id string) error {
	outputPath := getPath(id)
	_, err := os.Stat(outputPath)
	if err != nil {
		return err
	}
	if err := os.Remove(outputPath); err != nil {
		return err
	}
	// TODO remove it
	fmt.Println("File deleted")
	return nil
}

func deleteFinishedJob(id string) error {
	deleteJobSQL := `DELETE FROM jobs
    WHERE
    id = ?
    AND status = 3`
	err := execSQL(deleteJobSQL, nil, id)
	if err != nil {
		return err
	}
	return nil
}

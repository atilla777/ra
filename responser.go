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

const (
	finishedJobsSQL = `SELECT id
    FROM jobs
    WHERE status = 3
    ORDER BY created_at`
	deleteJobSQL = `DELETE FROM jobs
    WHERE
    id = ?
    AND status = 3`
)

func sendResults() error {
	//	tx, err := db.Begin()
	//	if err != nil {
	//		return fmt.Errorf("Can`t start DB transaction: %s", err)
	//	}
	//	rows, err := tx.Query(finishedJobsSQL)
	//	if err != nil {
	//		return fmt.Errorf("Can`t make DB query: %s", err)
	//	}
	//	defer rows.Close()
	//	for rows.Next() {
	//		var id string
	//		err = rows.Scan(&id)
	//		if err != nil {
	//			return fmt.Errorf("Can`t scan DB query result: %s", err)
	//		}
	//		err = sendResult(id)
	//		// TODO Add retry for send result
	//		if err != nil {
	//			return fmt.Errorf("Can`t send result: %s", err)
	//		} else {
	//			return deleteJob(tx, id)
	//		}
	//	}
	//	tx.Commit()
	return nil
}

func sendResult(id string) error {
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

func deleteJob(id string) error {
	if err := deleteRecord(id); err != nil {
		return fmt.Errorf("Can`t delete job: %s", err)
	}
	if err := deleteFile(id); err != nil {
		return fmt.Errorf("Can`t delete file: %s", err)
	}
	return nil
}

func deleteRecord(id string) error {
	mutex.Lock()
	defer mutex.Unlock()
	err := execSQL(deleteJobSQL, nil)
	if err != nil {
		return err
	}
	// TODO remove it
	fmt.Println("Job deleted")
	return nil
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

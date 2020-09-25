package main

// Command for response test:
// while true; do { echo -e "HTTP/1.1 200 OK\r\n$(date)\r\n\r\n<h1>hello world from $(hostname) on $(date)</h1>" |  nc -vl 3000; } done
// Command for request test:
// curl -v -H "Authorization: Token token=011a54f3922298708bdf2677fcfb8829" -H "Content-Type: application/json" -X POST "http://127.0.0.1:3000/api/v1/ra_api" -d '{"args": "as"}'

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
)

func sendResults() error {
	jobs, err := finishedJobs()
	if err != nil {
		return fmt.Errorf("Search finished jobs error: %s", err)
	}
	for _, j := range jobs {
		resp, err := sendOneResult(j.Id)
		if err != nil {
			// TODO Add retry for send result
			logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("Can`t send result of job %s: %s", j.Id, err)}
			retryResponserJob(j.Id, err, j.Attempts)
		} else {
			if resp["message"] == "accepted" {
				logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("Result job %s sent.", j.Id)}
				deleteJob(j.Id)
				logChan <- raLog{Lev: "info", Mes: fmt.Sprintf("Result job %s sent\n", j.Id)}
			} else {
				err := fmt.Errorf("RISM don`t accept result.")
				retryResponserJob(j.Id, err, j.Attempts)
				// TODO Add retry for send result
				logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("Result job %s not accepted by RISM", j.Id)}
			}
		}
	}
	return nil
}

func sendOneResult(id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	outputPath := getPath(id)
	_, err := os.Stat(outputPath)
	if err != nil {
		return result, fmt.Errorf("Can`t find result file: %s", err)
	}
	resultJSON, err := nmapJSON(outputPath, id)
	if err != nil {
		return result, fmt.Errorf("Can`t convert result to JSON: %s", err)
	}
	rism_url := fmt.Sprintf(
		"%s://%s:%s%s",
		viper.GetString("rism.protocol"),
		viper.GetString("rism.host"),
		viper.GetString("rism.port"),
		viper.GetString("rism.path"),
	)
	//client := &http.Client{Timeout: 20 * time.Second}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", rism_url, bytes.NewBuffer(resultJSON))
	if err != nil {
		return result, fmt.Errorf("Can`t make new request: %s", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", viper.GetString("rism.secret")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("Can`t send request: %s", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("Can`t read response: %s", err)
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("Can`t read response: %s", err)
	}
	return result, nil
}

func retryResponserJob(id string, err error, att int) {
	if att+1 == viper.GetInt("ra.workers.responser_attempts") {
		deleteJob(id)
		logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("Max attempts reached. Response job %s was killed.", id)}
		return
	}
	if err := updateRecord(id, 3, att+1); err != nil {
		logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("Responer DB update error: %s. Response job %s was killed.", err, id)}
		deleteJob(id)
	}
}

func deleteJob(id string) {
	if err := deleteJobInDB(id); err != nil {
		logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("Can`t delete job: %s", err)}
	}
	if err := deleteFile(id); err != nil {
		logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("Can`t delete file: %s", err)}
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
	logChan <- raLog{Lev: "info", Mes: fmt.Sprintf("File deleted")}
	return nil
}

package main

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os/exec"
	"strings"
	"time"
)

// Worker that run nmap scan form queue in channel
// TODO check is jobChan needed as param?
func worker(i int) {
	// Start nmap every time when job appears in channel
	for job := range jobChan {
		// TODO send error to logger chanel
		if err := startNmap(job, i); err != nil {
			logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("Scan %s failed by worker %d: %s", job.Id, i, err)}
		} else {
			logChan <- raLog{Lev: "info", Mes: fmt.Sprintf("Scan %s done by worker %d", job.Id, i)}
		}
	}
}

// Run nmap and save result to XML file
func startNmap(job Job, i int) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(viper.GetInt("ra.workers.timeout"))*time.Hour,
	)
	defer cancel()
	options, err := jobOptions(job.Options, job.Id)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "sudo", options...)
	if _, err := cmd.CombinedOutput(); err != nil {
		updateFailed(job.Id, job.Attempts)
		return err
	}
	if ctx.Err() != nil {
		updateFailed(job.Id, job.Attempts)
		return fmt.Errorf("Scan timeout exceeded")
	}
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

func getPath(id string) string {
	outputPath := fmt.Sprintf("%s/%s.xml", viper.GetString("ra.nmapxml"), id)
	return outputPath
}

func updateFailed(id string, att int) {
	if att+1 == viper.GetInt("ra.workers.attempts") {
		deleteJob(id)
		logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("Max attempts reached. Scan job %s was killed.", id)}
		return
	}
	err := updateRecord(id, 1, att+1)
	if err != nil {
		logChan <- raLog{Lev: "err", Mes: fmt.Sprintf("DB failed job %s update error: %s", id, err)}
		return
	}
}

func updateFinished(id string) error {
	err := updateRecord(id, 3, 0)
	if err != nil {
		return fmt.Errorf("DB update error: %s", err)
	}
	return nil
}

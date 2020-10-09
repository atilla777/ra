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
	command, options, err := jobOptions(job.Options, job.Id)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, command, options...)
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

func jobOptions(options string, id string) (string, []string, error) {
	var command string
	var opt []string
	var opt_str string
	outputPath := getPath(id)
	nmapPath, err := exec.LookPath("nmap")
	if err != nil {
		return command, opt, fmt.Errorf("Nmap path lookup error: %s", err)
	}
	//o = fmt.Sprintf("%s %s -oX %s", "/usr/bin/nmap", options, outputPath)
	if viper.GetString("ra.mode") == "sudo" {
		command = "/usr/bin/sudo"
		opt_str = fmt.Sprintf("%s %s -oX %s", nmapPath, options, outputPath)
	} else {
		command = "/usr/bin/nmap"
		opt_str = fmt.Sprintf("%s -oX %s", options, outputPath)
	}
	return command, strings.Split(opt_str, " "), nil
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

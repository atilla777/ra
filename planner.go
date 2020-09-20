package main

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

func startPlanner() {
	ticker := time.NewTicker(time.Second * viper.GetDuration("ra.workers.tick"))
	for _ = range ticker.C {
		if err := sendJobsToQueue(); err != nil {
			logChan <- logMessage(fmt.Sprintf("Planner error: %s", err))
		}
		if err := sendResults(); err != nil {
			logChan <- logMessage(fmt.Sprintf("Responser error: %s", err))
		}
	}
}

// Select not finished jobs from sqlite database end sent it to workers via queue (channel)
// jobs.statuses:
// 0 - planned (saved in tadabase by controller)
// 1 - failed (by worker)
// 2 - in queue (placed to chanel by planner)
// 3 - finished (taken from queue by worker)
// X (deleted) - sent (result sent by responser)
func sendJobsToQueue() error {
	jobs, err := waitingJobs()
	if err != nil {
		return fmt.Errorf("Search new jobs error: %s", err)
	}

	for _, j := range jobs {
		jobChan <- j
		if err := markQueueJob(j.Id, j.Attempts); err != nil {
			logChan <- logMessage(fmt.Sprintf("Mark job %s as planned error: %s", j.Id, err))
		}
	}
	return nil
}

func markQueueJob(id string, a int) error {
	// TODO add retry period to jobs field
	queueJobSQL := `UPDATE jobs SET status = 2, attempts = ? WHERE id = ?`
	err := execSQL(
		queueJobSQL,
		nil,
		a+1,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

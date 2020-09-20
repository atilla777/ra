package main

import (
	"database/sql"
	//	"fmt"
	_ "github.com/mattn/go-sqlite3"
	//	"github.com/spf13/viper"
)

const (
	newJobsSQL = `SELECT id, options, attempts FROM jobs
    WHERE status < 2
    AND attempts < ?
    ORDER BY created_at
    LIMIT ?`
	queueJobSQL = `UPDATE jobs SET status = 2, attempts = ? WHERE id = ?`
)

// Select not finished jobs from sqlite database end sent it to workers via queue (channel)
// jobs.statuses:
// 0 - planned (saved in tadabase by controller)
// 1 - failed (by worker)
// 2 - in queue (placed to chanel by planner)
// 3 - finished (taken from queue by worker)
// X (deleted) - sent (result sent by responser)
func sendJobsToQueue() error {
	//	rows, err := tx.Query(
	//		newJobsSQL,
	//		viper.GetInt("ra.workers.attempts"),
	//		viper.GetInt("ra.workers.queue"),
	//	)
	//	if err != nil {
	//		return fmt.Errorf("Can`t make DB query: %s", err)
	//	}
	//	defer rows.Close()
	//	for rows.Next() {
	//		var id string
	//		var options string
	//		var attempts int
	//		err = rows.Scan(&id, &options, &attempts)
	//		if err != nil {
	//			return fmt.Errorf("Can`t scan DB query result: %s", err)
	//		}
	//		job := Job{Id: id, Options: options, Attempmts: attempts}
	//		jobChan <- job
	//		if err := queueJob(tx, job, attempts); err != nil {
	//			return fmt.Errorf("Can`t update DB to set job queued : %s", err)
	//		}
	//	}
	//	tx.Commit()
	return nil
}

func queueJob(tx *sql.Tx, j Job, a int) error {
	mutex.Lock()
	defer mutex.Unlock()
	// TODO add retry period to jobs field
	_, err := tx.Exec(queueJobSQL, a+1, j.Id)
	if err != nil {
		return err
	}
	return nil
}

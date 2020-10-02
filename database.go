package main

import (
	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"strconv"
)

const (
	dbSchemaSQL = `CREATE TABLE IF NOT EXISTS jobs (
      id text PRIMARY KEY,
      options text NOT NULL,
      status integer NOT NULL DEFAULT 0,
      attempts integer NOT NULL DEFAULT 0,
      created_at text NOT NULL
    );`
)

func createPool() *sqlitex.Pool {
	sqliteConnStr := fmt.Sprintf(
		"file:%s?cache=shared&mode=rwc&_journal_mode=WAL",
		viper.GetString("ra.sqlite"),
	)
	db, err := sqlitex.Open(sqliteConnStr, 0, 16)
	if err != nil {
		log.Fatalf("Error open database: %v", err)
	}
	return db
}

func createDatabse() {
	err := execSQL(dbSchemaSQL, nil)
	if err != nil {
		log.Fatalf("Error create/check existence DB: %v", err)
	}
}

func waitingJobs() ([]Job, error) {
	var res []Job
	conn := pool.Get(nil)
	defer pool.Put(conn)
	newJobsSQL := `SELECT id, options, attempts FROM jobs
    WHERE status < 2
    AND attempts < $att
    ORDER BY created_at
    LIMIT $lim`
	stmt := conn.Prep(newJobsSQL)
	defer stmt.Reset()
	stmt.SetText("$att", strconv.Itoa(viper.GetInt("ra.workers.attempts")))
	stmt.SetText("$lim", strconv.Itoa(viper.GetInt("ra.workers.queue")))
	for {
		if hasRow, err := stmt.Step(); err != nil {
			return nil, err
		} else if !hasRow {
			break
		}
		job := Job{
			Id:       stmt.GetText("id"),
			Options:  stmt.GetText("options"),
			Attempts: int(stmt.GetInt64("attempts")),
		}
		res = append(res, job)
	}
	return res, nil
}

func updateRecord(id string, status int, att int) error {
	var err error
	updateJobSQL := "UPDATE jobs SET status = ?, attempts = ? WHERE id = ?"
	err = execSQL(updateJobSQL, nil, status, att, id)
	return err
}

func finishedJobs() ([]Job, error) {
	var res []Job
	conn := pool.Get(nil)
	defer pool.Put(conn)
	finishedJobsSQL := `SELECT id, attempts
    FROM jobs
    WHERE status = 3
    AND attempts < $att
    ORDER BY created_at`
	stmt := conn.Prep(finishedJobsSQL)
	defer stmt.Reset()
	stmt.SetText("$att", strconv.Itoa(viper.GetInt("ra.responser.attempts")))
	for {
		if hasRow, err := stmt.Step(); err != nil {
			return nil, err
		} else if !hasRow {
			break
		}
		job := Job{
			Id:       stmt.GetText("id"),
			Options:  stmt.GetText("options"),
			Attempts: int(stmt.GetInt64("attempts")),
		}
		res = append(res, job)
	}
	return res, nil
}

func cehckIdExistence(id string) (bool, error) {
	conn := pool.Get(nil)
	defer pool.Put(conn)
	stmt := conn.Prep("SELECT EXISTS (SELECT 1 FROM jobs WHERE id = $id);")
	defer stmt.Reset()
	stmt.SetText("$id", id)
	var res bool
	for {
		if hasRow, err := stmt.Step(); err != nil {
			return false, err
		} else if !hasRow {
			break
		}
		res = stmt.ColumnInt(0) == 1
		break
	}
	return res, nil
}

func execSQL(sql string, resultFn func(stmt *sqlite.Stmt) error, values ...interface{}) error {
	conn := pool.Get(nil)
	defer pool.Put(conn)
	err := sqlitex.Exec(conn, sql, resultFn, values...)
	if err != nil {
		return err
	}
	return nil
}

func deleteJobInDB(id string) error {
	deleteJobSQL := `DELETE FROM jobs
    WHERE
    id = ?`
	err := execSQL(deleteJobSQL, nil, id)
	if err != nil {
		return err
	}
	return nil
}

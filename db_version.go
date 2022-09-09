package makoto

import (
	"database/sql"
	"errors"
	"log"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

const (
	_sqlFind = `
SELECT 
	id,
	version,
	filename,
	checksum,
	statement,
	created_at
FROM schema_version
`
)

func createSchemaVersionTable(db *sql.DB) error {
	sql := `
	CREATE TABLE IF NOT EXISTS schema_version (
		id serial PRIMARY KEY,
		version integer,
		filename text,
		checksum text,
		statement text,
		created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
	)
	`
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Rollback migration, Error: ", r)
		}
	}()

	stmt, err := tx.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return tx.Commit()
}

func addRecord(tx *sql.Tx, version int, filename, checksum, statement string) error {
	sql := `
	INSERT INTO schema_version (version, filename, checksum, statement) 
	VALUES (?, ?, ?, ?)
	`
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(version, filename, checksum, statement)
	if err != nil {
		return err
	}

	return nil
}

func getLastRecord(db *sql.DB) (*MigrationRecord, error) {
	query := _sqlFind + `
ORDER by id desc
LIMIT 1
	`
	row, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	record := MigrationRecord{}
	if row.Next() {
		err = row.Scan(&record.ID, &record.Version, &record.Filename, &record.Checksum, &record.Statement, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		return &record, nil
	}
	return nil, ErrRecordNotFound
}

func GetAllRecords(db *sql.DB) ([]MigrationRecord, error) {
	rows, err := db.Query(_sqlFind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []MigrationRecord{}
	for rows.Next() {
		record := MigrationRecord{}
		err = rows.Scan(&record.ID, &record.Version, &record.Filename, &record.Checksum, &record.Statement, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

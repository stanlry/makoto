package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/stanlry/makoto"
)

const SQLFileExtension = ".sql"

func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func processMigrationCollection(path string) *makoto.MigrationCollection {
	files, err := readSQLMigrationScript(path)
	logError(err)

	collection := &makoto.MigrationCollection{}
	for _, f := range files {
		fullPath := filepath.Join(path, f.Name())
		file, err := os.Open(fullPath)
		logError(err)

		migration := makoto.ParseMigrationStatement(file.Name(), file)

		// skip invalid file
		if migration.Version == 0 {
			continue
		}

		collection.Add(migration)
	}

	return collection
}

func readSQLMigrationScript(path string) ([]os.FileInfo, error) {
	dir, err := os.Open(path)
	logError(err)

	files, err := dir.Readdir(0)
	logError(err)

	result := []os.FileInfo{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if filepath.Ext(f.Name()) != SQLFileExtension {
			continue
		}
		result = append(result, f)
	}
	return result, nil
}

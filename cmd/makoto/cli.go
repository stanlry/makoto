package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/stanlry/makoto"
)

const (
	migrationDir = "migration"
	sqlDir       = "sql"
	seedDir      = "seed"
)

func initMigrationDir() {
	dir := currentDir()

	// create migration folder
	migrationPath := filepath.Join(dir, migrationDir)
	mkdir(migrationPath)

	// create sql script folder
	sqlPath := filepath.Join(dir, migrationDir, sqlDir)
	mkdir(sqlPath)

	// create sql seed folder
	seedPath := filepath.Join(dir, migrationDir, seedDir)
	mkdir(seedPath)
}

func mkdir(path string) {
	if exists(path) {
		log.Printf("Directory '%v' already exists\n", path)
		return
	}
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		log.Printf("Created directory '%v'\n", path)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	log.Fatal(err)
	return true
}

func getSQLScriptDir() string {
	dir := currentDir()
	if strings.HasSuffix(dir, sqlDir) {
		return dir
	}
	fullPath := filepath.Join(dir, migrationDir, sqlDir)
	if exists(fullPath) {
		return fullPath
	}
	log.Fatal("Unknow sql script directory")
	return ""
}

func getMigrationDir() string {
	dir := currentDir()
	if strings.HasSuffix(dir, migrationDir) {
		return dir
	}
	fullPath := filepath.Join(dir, migrationDir)
	if exists(fullPath) {
		return fullPath
	}
	log.Fatal("Unknow migration directory")
	return ""
}

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func createNewScript(name string, useSequence bool) {
	dir := getSQLScriptDir()
	version := time.Now().Local().Format("20060102150405")
	if useSequence {
		version = getNewScriptSequence()
	}

	filename := fmt.Sprintf("%v_%s.sql", version, name)
	fullPath := filepath.Join(dir, filename)
	log.Println("Create new migration script: ", filename)
	os.Create(fullPath)
}

func getNewScriptSequence() string {
	collection := initCollection()
	if st := collection.LastStatement(); st != nil {
		return strconv.Itoa(st.Version + 1)
	}

	return "1"
}

// func displayMigrati

func initCollection() *makoto.MigrationCollection {
	dir := getMigrationDir()
	return processMigrationCollection(dir)
}

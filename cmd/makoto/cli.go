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

const migrationPath = "migration"

func initMigrationDir() {
	dir := currentDir()
	path := filepath.Join(dir, migrationPath)
	if exists(path) {
		fmt.Println("Migration directory already exists")
		return
	}
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		fmt.Println("Created migration directory")
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

func getMigrationDir() string {
	dir := currentDir()
	if strings.HasSuffix(dir, migrationPath) {
		return dir
	}
	fullPath := filepath.Join(dir, migrationPath)
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
	dir := getMigrationDir()
	version := time.Now().Local().Format("20060201150405")
	if useSequence {
		version = getNewScriptSequence()
	}

	filename := fmt.Sprintf("%v_%s.sql", version, name)
	fullPath := filepath.Join(dir, filename)
	fmt.Println("Create new migration script: ", filename)
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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

const collectionFilename = "collection.go"

func GenerateStringCollection(path string) {
	fmt.Println("Collect migration scripts:")

	buffer := bytes.NewBuffer(nil)
	fmt.Fprint(buffer, `// this is generated by makoto cli, do not modify this file.
package migration

import "github.com/cororoGrap/makoto"

func GetCollection() makoto.MigrateCollection {
	statements := []makoto.MigrateStatement{
	`)

	collection := processMigrationCollection(path)
	migration := collection.Head()
	for {
		st := migration.Statement()
		upSt, _ := json.Marshal(st.UpStatement)
		downSt, _ := json.Marshal(st.DownStatement)

		fmt.Fprintf(buffer, `
		{"%v", "%v", %v, %v, "%v"},
		`, st.Version, st.Filename, string(upSt), string(downSt), st.Checksum)

		fmt.Printf("%v\n", st.Filename)

		if migration.Next() != nil {
			migration = migration.Next()
			continue
		}
		break
	}

	fmt.Fprint(buffer, `
	}

	collection := makoto.MigrationCollection{}
	for _, st := range statements {
		collection.Add(statement)
	}
	return &collection
}`)

	dest := filepath.Join(path, collectionFilename)
	if err := ioutil.WriteFile(dest, buffer.Bytes(), 0644); err != nil {
		panic(err)
	}
}

func GenerateEmbedCollection(path string) {
	fmt.Println("Generating go embed file")

	buffer := bytes.NewBuffer(nil)
	fmt.Fprint(buffer, `// this is generated by makoto cli, do not modify this file.
package migration

import "embed"

//go:embed sql/*.sql
var Content embed.FS

`)

	dest := filepath.Join(path, collectionFilename)
	if err := ioutil.WriteFile(dest, buffer.Bytes(), 0644); err != nil {
		panic(err)
	}
}

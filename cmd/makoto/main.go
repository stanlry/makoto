package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/olekukonko/tablewriter"
	"github.com/stanlry/makoto"
	"github.com/stanlry/makoto/cmd/makoto/db"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	database   string
	configPath string
)

func main() {

	app := cli.NewApp()
	app.Name = "makoto"
	app.Version = makoto.VERSION
	app.Usage = "minimalist migration tool for PostgreSQL"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "database",
			Destination: &database,
		},
		cli.StringFlag{
			Name:        "config",
			Destination: &configPath,
		},
	}

	app.Commands = []cli.Command{
		{
			Name: "version",
			Action: func(c *cli.Context) error {
				fmt.Println("makoto version: ", makoto.VERSION)
				return nil
			},
		},
		{
			Name: "init",
			Action: func(c *cli.Context) error {
				initMigrationDir()
				return nil
			},
		},
		{
			Name: "collect",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "no-embed",
				},
			},
			Action: func(c *cli.Context) error {
				migrationPath := getMigrationDir()
				if c.Bool("no-embed") {
					GenerateStringCollection(migrationPath)
				} else {
					GenerateEmbedCollection(migrationPath)
				}
				return nil
			},
		},
		{
			Name:  "new",
			Usage: "Create new migration sql script",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "seq",
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() == 1 {
					name := c.Args()[0]
					createNewScript(name, c.Bool("seq"))
				} else {
					fmt.Println("Missing file name")
				}
				return nil
			},
		},
		{
			Name: "status",
			Action: func(c *cli.Context) error {
				configureDBUri()
				db := db.ConnectPostgres(database)
				defer db.Close()
				r, err := makoto.GetAllRecords(db)
				if err != nil {
					panic(err)
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Version", "Script", "Create Date"})
				for _, record := range r {
					date := record.CreatedAt.Format(time.RFC3339)
					table.Append([]string{strconv.Itoa(record.Version), record.Filename, date})
				}
				table.Render()
				return nil
			},
		},
		{
			Name: "up",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name: "version",
				},
			},
			Action: func(c *cli.Context) error {
				configureDBUri()
				db := db.ConnectPostgres(database)
				defer db.Close()
				collection := processMigrationCollection(getMigrationDir())
				migrator := makoto.GetMigrator(db, collection)

				version := c.Int("version")
				if version == 0 {
					migrator.Up()
				} else {
					migrator.EnsureSchema(version)
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}

func configureDBUri() {
	if len(database) == 0 {
		err := loadDBConfig()
		if err != nil {
			panic(err)
		}
	}
}

func getConfigPath() string {
	if len(strings.TrimSpace(configPath)) == 0 {
		return filepath.Join(currentDir(), "config.toml")
	}
	return configPath
}

func loadDBConfig() error {
	path := getConfigPath()
	fmt.Println("load config path: ", path)

	file, err := os.Open(path)
	logError(err)

	config := dbConfig{}
	configSt, err := ioutil.ReadAll(file)
	err = toml.Unmarshal(configSt, &config)
	logError(err)

	pg := config.Postgres
	database = fmt.Sprintf("user=%v password=%v host=%v port=%v dbname=%v sslmode=disable",
		pg.User, pg.Password, pg.Host, pg.Port, pg.DBName)

	return nil
}

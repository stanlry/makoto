package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/olekukonko/tablewriter"
	"github.com/stanlry/makoto"
	"github.com/stanlry/makoto/cmd/makoto/db"
	cli "github.com/urfave/cli/v2"
)

const keyMigrator = "migrator"

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
		&cli.StringFlag{
			Name:        "database",
			Usage:       "Database connection URL",
			Destination: &database,
		},
		&cli.StringFlag{
			Name:        "config",
			Usage:       "Specify config path",
			Destination: &configPath,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:  "version",
			Usage: "Version of makoto",
			Action: func(c *cli.Context) error {
				fmt.Println("makoto version: ", makoto.VERSION)
				return nil
			},
		},
		{
			Name:  "init",
			Usage: "Initialize migration directory",
			Action: func(c *cli.Context) error {
				initMigrationDir()
				return nil
			},
		},
		{
			Name:  "pack",
			Usage: "Generate a go file that packs all the sql migration scripts with it",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "no-embed",
					Usage: "Do not use embed to pack the sql migration scripts",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("no-embed") {
					GenerateStringCollection(getSQLScriptDir())
				} else {
					GenerateEmbedCollection(getMigrationDir())
				}
				return nil
			},
		},
		{
			Name:  "new",
			Usage: "Create new migration sql script",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:     "seq",
					Usage:    "Use incremental sequence instead of datetime to generate file version",
					Required: false,
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() == 1 {
					name := c.Args().Get(0)
					createNewScript(name, c.Bool("seq"))
				} else {
					fmt.Println("Missing file name")
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "List existing sql migration scripts in the directory",
			Action: func(c *cli.Context) error {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Version", "Script Name"})

				collection := processMigrationCollection(getSQLScriptDir())
				item := collection.Head()
				for {
					if item == nil {
						break
					}
					if item.Statement() == nil {
						break
					}
					table.Append([]string{strconv.Itoa(item.Statement().Version), item.Statement().Filename})
					item = item.Next()
				}
				table.Render()
				return nil
			},
		},
		{
			Name:  "status",
			Usage: "Return the migration table from database",
			Action: func(c *cli.Context) error {
				configureDBUri()
				db := db.ConnectPostgres(database)
				defer db.Close()
				r, err := makoto.GetAllRecords(db)
				if err != nil {
					panic(err)
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Version", "Execution Type", "Script", "Create Date"})
				for _, record := range r {
					date := record.CreatedAt.Format(time.RFC3339)
					table.Append([]string{strconv.Itoa(record.Version), record.Exectype, record.Filename, date})
				}
				table.Render()
				return nil
			},
		},
		{
			Name:  "migrate",
			Usage: "Run migration scripts",
			Before: func(ctx *cli.Context) error {
				configureDBUri()
				db := db.ConnectPostgres(database)
				collection := processMigrationCollection(getSQLScriptDir())
				migrator := makoto.GetMigrator(db, collection)
				migrator.SetCollection(collection)

				ctx.Context = context.WithValue(ctx.Context, keyMigrator, migrator)
				return nil
			},
			After: func(ctx *cli.Context) error {
				migrator := ctx.Context.Value(keyMigrator).(*makoto.Migrator)
				migrator.Close()
				return nil
			},
			Subcommands: cli.Commands{
				{
					Name:  "drop",
					Usage: "drop all migrations",
					Action: func(ctx *cli.Context) error {
						migrator := ctx.Context.Value(keyMigrator).(*makoto.Migrator)
						migrator.DropAll()
						return nil
					},
				},
				{
					Name:  "up",
					Usage: "migrate up",
					Flags: []cli.Flag{
						&cli.IntFlag{
							Name:  "version",
							Usage: "Specify the migration version",
						},
					},
					Action: func(ctx *cli.Context) error {
						migrator := ctx.Context.Value(keyMigrator).(*makoto.Migrator)
						version := ctx.Int("version")
						if version == 0 {
							migrator.EnsureHead()
						} else {
							migrator.EnsureSchema(version)
						}
						return nil
					},
				},
				{
					Name:  "down",
					Usage: "migrate down",
					Flags: []cli.Flag{
						&cli.IntFlag{
							Name:     "version",
							Usage:    "Specify the migration version",
							Required: true,
						},
					},
					Action: func(ctx *cli.Context) error {
						migrator := ctx.Context.Value(keyMigrator).(*makoto.Migrator)
						version := ctx.Int("version")
						migrator.Down(version)
						return nil
					},
				},
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
	log.Println("Load config: ", path)

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

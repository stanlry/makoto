# makoto

Simple migration tool for PostgreSQL

## Install

Install makoto CLI

```bash
go get -u github.com/stanlry/makoto/cmd/makoto
```

Install makoto migrator

```bash
go get github.com/stanlry/makoto
```

## Structure

makoto will create a directory named migration under your project. All sql migration sql will placed under this directory.
Migration scripts should be named as

```bash
[numeric version number]_[script name].sql
e.g.
1_basic.sql
```

## CLI

Init migration directory

```bash
makoto init
```

Create new migration sql script

```bash
makoto new [script_name]
```

Generate golang migration collection, a golang file 'collection.go' will be created under the migration directory

```bash
makoto collect
```

Check current migration status

```bash
makoto status
```

Migrate to latest version

```bash
makoto up
```

Database connection uri format

```
makoto -database postgres://[username]:[password]@[host]:5432/[dbname]?sslmode=[enable|disable] [command]
```

Custom config file

```
makoto -config [file path] [command]
```

If no custom config file or database uri is given, makoto will search for "config.toml"

Config file format

```toml
[postgres]
  host="localhost"
  port="5432"
  user="postgres"
  password="123456"
  name="database name"
```

## Integrate with Golang

First generate the collection file with CLI.

Perform migration

```go
migrator.Up() // migrate to latest version
// or
migrator.EnsureSchema(202201011233) // migrate to a given version
```

#### Example

```go
func startMigration(db *sql.DB) {
    migrator := migration.New(db)
    migrator.EnsureSchema(202201011233)
}
```

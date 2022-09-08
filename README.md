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

If no custom config file or database uri is given, makoto will search for "config.json" placed inside migration directory

Config file format

```json
{
  "database": "PostgreSQL",
  "PostgreSQL": {
    "Host": "localhost",
    "Port": "5432",
    "DBName": "xxx",
    "User": "postgres",
    "Password": "123456"
  }
}
```

## Integrate with Golang

First generate the collection file with CLI.

Initialize the makoto migrator

```go
migrator := makoto.New(db) // pass the DB pointer
```

Pass the migration collection to migrator

```go
migrator.SetCollection(migration.GetCollection())
```

Perform migration

```go
migrator.Up() // migrate to latest version
// or
migrator.EnsureSchema(202201011233) // migrate to a given version
```

#### Example

```go
func startMigration(db *sql.DB) {
    migrator := makoto.New(db)
    migrator.SetEmbedCollection(migration.Content)
    migrator.EnsureSchema(202201011233)
}
```

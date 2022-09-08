package makoto

import (
	"bytes"
	"database/sql"
	"embed"
	"log"
	"path"
)

type migrator struct {
	db         *sql.DB
	collection *MigrationCollection
}

func GetMigrator(db *sql.DB, collection *MigrationCollection) *migrator {
	return &migrator{
		db:         db,
		collection: collection,
	}
}

func New(db *sql.DB) *migrator {
	err := createSchemaVersionTable(db)
	if err != nil {
		log.Fatal(err)
	}

	return &migrator{
		db: db,
	}
}

func (m *migrator) GetCollection() *MigrationCollection {
	if m.collection != nil {
		return m.collection
	}
	log.Fatal("Migration collection not found")
	return nil
}

func (m *migrator) SetCollection(c *MigrationCollection) {
	m.collection = c
}

func (m *migrator) SetEmbedCollection(fs embed.FS) {
	m.collection.Reset()

	fnames, err := getAllFilenames(&fs, "")
	if err != nil {
		panic(err)
	}

	for _, fname := range fnames {
		data, err := fs.ReadFile(fname)
		if err != nil {
			panic(err)
		}
		reader := bytes.NewReader(data)
		statement := ParseMigrationStatement(fname, reader)
		m.collection.Add(statement)
	}
}

func (m *migrator) EnsureSchema(targetVersion int) {
	currentNode, err := m.getCurrentNode()

	if err != nil && err != ErrRecordNotFound {
		log.Fatal(err)
	}

	if err == ErrRecordNotFound {
		currentNode = m.GetCollection().Head()
		m.upto(currentNode, targetVersion)
		return
	}

	st := currentNode.Statement()
	if st.Version == targetVersion {
		return
	}
	if st.Version < targetVersion {
		log.Println("start migrate")
		m.upto(currentNode.nextNode, targetVersion)
	}
}

func (m *migrator) getCurrentNode() (*migrationItem, error) {
	record, err := getLastRecord(m.db)
	if err != nil {
		return nil, err
	}
	if record.Version > m.GetCollection().LastStatement().Version {
		return m.GetCollection().Tail(), nil
	}
	return m.GetCollection().Find(record.Version), nil
}

func (m *migrator) upto(currentNode *migrationItem, targetVersion int) {
	tx, err := m.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Rollback migration, Error: ", r)
		}
	}()

	upTo(tx, currentNode, targetVersion)
	tx.Commit()
}

func (m *migrator) Up() {
	lastVersion := m.GetCollection().LastStatement().Version
	m.EnsureSchema(lastVersion)
}

func upTo(tx *sql.Tx, node *migrationItem, targetVersion int) {
	currentNode := node
	for {
		statement := currentNode.statement
		if statement.Version <= targetVersion {
			_, err := tx.Exec(statement.UpStatement)
			if err != nil {
				log.Println("Fail to run migration script: ", statement.Filename)
				log.Fatal(err)
			}
			log.Println("Migrated script: ", statement.Filename)
			err = addRecord(tx, statement.Version, statement.Filename, statement.Checksum, statement.UpStatement)
			if err != nil {
				log.Fatal(err)
			}
			if currentNode.nextNode == nil {
				break
			}
			currentNode = currentNode.nextNode
		} else {
			break
		}
	}
}

func getAllFilenames(fs *embed.FS, dir string) (out []string, err error) {
	if len(dir) == 0 {
		dir = "."
	}

	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fp := path.Join(dir, entry.Name())
		if entry.IsDir() {
			res, err := getAllFilenames(fs, fp)
			if err != nil {
				return nil, err
			}

			out = append(out, res...)

			continue
		}

		out = append(out, fp)
	}

	return
}

package makoto

import (
	"bytes"
	"database/sql"
	"embed"
	"log"
	"path"
)

type Migrator struct {
	db         *sql.DB
	collection *MigrationCollection
}

func GetMigrator(db *sql.DB, collection *MigrationCollection) *Migrator {
	return &Migrator{
		db:         db,
		collection: collection,
	}
}

func New(db *sql.DB) *Migrator {
	return &Migrator{
		db: db,
	}
}

func (m *Migrator) GetCollection() *MigrationCollection {
	if m.collection != nil {
		return m.collection
	}
	log.Fatal("Migration collection not found")
	return nil
}

func (m *Migrator) SetCollection(c *MigrationCollection) {
	m.collection = c
}

func (m *Migrator) SetEmbedCollection(fs embed.FS) {
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

func (m *Migrator) EnsureSchema(targetVersion int) {
	currentNode, err := m.getCurrentNode()
	if err != nil && err != ErrRecordNotFound {
		log.Fatal(err)
	}

	targetNode := m.collection.Find(targetVersion)
	if targetNode == nil {
		log.Fatal("Target version not exists")
	}

	if err == ErrRecordNotFound {
		currentNode = m.GetCollection().Head()
		m.upto(currentNode, targetVersion)
		return
	}

	st := currentNode.Statement()
	if st.Version == targetVersion {
		log.Println("Schema version is already up to date")
		return
	}
	if st.Version < targetVersion {
		log.Println("Start migration")
		m.upto(currentNode.nextNode, targetVersion)
	} else {
		log.Println("Database schema version is ahead of migration script")
	}
}

func (m *Migrator) DropAll() {
	currentNode, err := m.getCurrentNode()
	if err != nil && err != ErrRecordNotFound {
		log.Fatal(err)
	}
	m.downTo(currentNode, 0, true)
}

func (m *Migrator) Down(targetVersion int) {
	currentNode, err := m.getCurrentNode()
	if err != nil && err != ErrRecordNotFound {
		log.Fatal(err)
	}

	targetNode := m.collection.Find(targetVersion)
	if targetNode == nil {
		log.Fatal("Target version not exists")
	}

	st := currentNode.Statement()
	if st.Version <= targetVersion {
		log.Println("Database schema version is behind target version")
	} else {
		m.downTo(currentNode, targetVersion, false)
	}
}

func (m *Migrator) getCurrentNode() (*migrationItem, error) {
	// ensure schema version table exists
	if err := createSchemaVersionTable(m.db); err != nil {
		panic(err)
	}

	record, err := getLastRecord(m.db)
	if err != nil {
		return nil, err
	}

	lastStatement := m.GetCollection().LastStatement()
	if lastStatement != nil && record.Version > lastStatement.Version {
		return m.GetCollection().Tail(), nil
	}
	return m.GetCollection().Find(record.Version), nil
}

func (m *Migrator) upto(currentNode *migrationItem, targetVersion int) {
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

func (m *Migrator) downTo(currentNode *migrationItem, targetVersion int, dropAll bool) {
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

	downTo(tx, currentNode, targetVersion, dropAll)
	tx.Commit()
}

func (m *Migrator) EnsureHead() {
	lastStatement := m.GetCollection().LastStatement()
	if lastStatement != nil {
		m.EnsureSchema(lastStatement.Version)
	}
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
			log.Println("Migrate script: ", statement.Filename)
			err = addRecord(tx, statement.Version, statement.Filename, statement.Checksum, ExecUP, statement.UpStatement)
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

func downTo(tx *sql.Tx, node *migrationItem, targetVersion int, dropAll bool) {
	currentNode := node
	for {
		statement := currentNode.statement
		if statement.Version > targetVersion || dropAll {
			_, err := tx.Exec(statement.DownStatement)
			if err != nil {
				log.Println("Fail to run migration script: ", statement.Filename)
				log.Fatal(err)
			}
			log.Println("Migrate script: ", statement.Filename)
			err = addRecord(tx, statement.Version, statement.Filename, statement.Checksum, ExecDOWN, statement.UpStatement)
			if err != nil {
				log.Fatal(err)
			}
			if currentNode.nextNode == nil {
				break
			}
			currentNode = currentNode.previousNode
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

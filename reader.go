package makoto

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

func ParseMigrationStatement(fname string, r io.Reader) *MigrateStatement {
	migration := newStatementFromReader(r)
	migration.Filename = fname
	migration.Version = parseFilenameVersion(fname)

	return migration
}

func parseFilenameVersion(filename string) int {
	r, err := regexp.Compile("[0-9]_")
	if err != nil {
		log.Fatal(err)
	}

	st := r.FindString(filename)
	if st == "" {
		log.Fatal("invalid file, empty version number")
	}
	st = st[:len(st)-1]
	v, err := strconv.Atoi(st)
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func newStatementFromReader(r io.Reader) *MigrateStatement {
	var buf bytes.Buffer
	isDown := false

	migration := MigrateStatement{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		// line cannot be longer than 65536 characters
		line := scanner.Text()

		buf.WriteString(line)

		if strings.HasPrefix(line, "-- Down") {
			isDown = true
			continue
		}
		if strings.HasPrefix(line, "-- Up") {
			isDown = false
			continue
		}

		if isDown {
			migration.DownStatement += line + "\n"
		} else {
			migration.UpStatement += line + "\n"
		}
	}

	migration.Checksum = getMD5SumString(buf.Bytes())

	return &migration
}

func getMD5SumString(b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}

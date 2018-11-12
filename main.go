package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	internal "github.com/david-torres/skrapa/internal"
	"github.com/docopt/docopt-go"
	bolt "go.etcd.io/bbolt"
)

var arguments docopt.Opts

func init() {
	var usage = `Skrapa.

Usage:
	skrapa collect <script>
	skrapa export (csv|json) <database>
	skrapa -h | --help

Options:
	-h --help     Show this screen.
`
	var err error
	arguments, err = docopt.ParseDoc(usage)

	if err != nil {
		panic(fmt.Sprintf("could not parse commandline arguments: %s", err))
	}
}

func main() {
	// collect command
	if arguments["collect"] == true {
		// get script name from args
		scriptName, err := arguments.String("<script>")
		if err != nil {
			panic("could not parse script filename")
		}

		// get database name
		databaseName := filepath.Base(strings.TrimSuffix(scriptName, filepath.Ext(scriptName)) + ".db")

		// collect
		collect(scriptName, databaseName)
		return
	}

	// export command
	if arguments["export"] == true {
		// get database name from args
		databaseName, err := arguments.String("<database>")
		if err != nil {
			panic("could not parse database filename")
		}

		var exportType string
		if arguments["csv"] == true {
			exportType = "csv"
		} else if arguments["json"] == true {
			exportType = "json"
		} else {
			panic("unsupported export type")
		}

		// export
		export(databaseName, exportType)
		return
	}
}

func collect(scriptName string, databaseName string) {
	// parse script
	script := parseScript(scriptName)

	// initialize a db
	db := newDB(databaseName)

	// initialize Collector
	c := internal.NewCollector(script, db)

	// run collection
	c.Run()
}

func export(databaseName string, exportType string) {
	// open db
	db := openDB(databaseName)

	// init exporter
	e := internal.NewExporter(db)

	// get export filename
	filename := filepath.Base(strings.TrimSuffix(databaseName, filepath.Ext(databaseName)))

	switch exportType {
	case "csv":
		filename = filename + ".csv"
		e.ExportCSV(filename)
		break
	case "json":
		filename = filename + ".json"
		e.ExportJSON(filename)
		break
	}
}

func parseScript(scriptName string) *internal.Script {
	scriptParser := internal.NewScriptParser()
	script, err := scriptParser.Parse(scriptName)
	if err != nil {
		panic(fmt.Sprintf("could not parse script: %s", scriptName))
	}

	return script
}

func newDB(databaseName string) *bolt.DB {
	// open db
	db, err := bolt.Open(databaseName, 0600, nil)
	if err != nil {
		panic(fmt.Sprintf("could not open database: %s", databaseName))
	}

	// init buckets
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists(internal.DB)
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists(internal.Entries)
		if err != nil {
			return fmt.Errorf("could not create entries bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("could not initialize database: %v", err))
	}

	log.Printf("Database initialized: %s", databaseName)

	return db
}

func openDB(databaseName string) *bolt.DB {
	// open db
	db, err := bolt.Open(databaseName, 0600, nil)
	if err != nil {
		panic(fmt.Sprintf("could not open database: %s", databaseName))
	}

	return db
}

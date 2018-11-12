package internal

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"
)

// NewExporter return a new script parser
func NewExporter(db *bolt.DB) *Exporter {
	return &Exporter{
		db,
	}
}

// Exporter parse scripts
type Exporter struct {
	db *bolt.DB
}

// ExportJSON export database as JSON
func (e Exporter) ExportJSON(filename string) error {
	data, err := e.getData()
	if err != nil {
		return err
	}

	j, err := json.Marshal(collateJSON(data))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, j, 0644)
	if err != nil {
		return err
	}

	log.Printf("Exported JSON: %s", filename)
	return nil
}

// ExportCSV export database as CSV
func (e Exporter) ExportCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	data, err := e.getData()
	if err != nil {
		return err
	}

	for _, value := range collateCSV(data) {
		err := writer.Write(value)
		if err != nil {
			return err
		}
	}

	log.Printf("Exported CSV: %s", filename)
	return nil
}

// pull the data blobs out of the database
func (e Exporter) getData() (map[string][]string, error) {
	data := make(map[string][]string)

	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(DB).Bucket(Entries)
		b.ForEach(func(k, v []byte) error {
			s, err := gobDecodeStringSlice(v)
			if err != nil {
				return err
			}

			data[string(k)] = s
			return nil
		})
		return nil
	})
	if err != nil {
		return data, err
	}

	return data, nil
}

// util

// collateJSON will massage the data into maps
func collateJSON(data map[string][]string) []map[string]string {
	collated := make([]map[string]string, 0)
	keys := make([]string, 0)

	for k := range data {
		keys = append(keys, k)
	}

	if len(keys) <= 0 {
		log.Fatal("No data collected")
	}

	for i := 0; i <= len(data[keys[0]])-1; i++ {
		c := make(map[string]string)
		for _, k := range keys {
			c[k] = data[k][i]
		}
		collated = append(collated, c)
	}

	return collated
}

// collateCSV will massage the data into rows
func collateCSV(data map[string][]string) [][]string {
	collated := make([][]string, 0)
	keys := make([]string, 0)

	for k := range data {
		keys = append(keys, k)
	}

	if len(keys) <= 0 {
		log.Fatal("No data collected")
	}

	collated = append(collated, keys)

	for i := 0; i <= len(data[keys[0]])-1; i++ {
		c := make([]string, 0)
		for _, k := range keys {
			c = append(c, data[k][i])
		}

		collated = append(collated, c)
	}

	return collated
}

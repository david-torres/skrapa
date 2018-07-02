package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gocolly/colly"
)

var saved map[string][]string
var c *colly.Collector
var config Config

const defaultUA = "Skrapa"

func init() {
	fileName := flag.String("file", "./examples/example.toml", "The Skrapa config file to load")
	flag.Parse()

	// marshal config
	if _, err := toml.DecodeFile(*fileName, &config); err != nil {
		log.Fatal(err)
		return
	}

	// primitive in-memory storage
	saved = make(map[string][]string)
}

func main() {
	// initalize Colly
	initCollector()

	// load up the pipeline
	initPipeline()

	// run!
	err := c.Visit(config.Main.URL)
	if err != nil {
		log.Fatal(err)
	}

	// wait until done...
	c.Wait()

	// save the datas
	save()
}

// initCollector initalize Colly and some standard outputs
func initCollector() {
	ua := defaultUA
	if config.Main.UserAgent != "" {
		ua = config.Main.UserAgent
	}

	c = colly.NewCollector(
		colly.UserAgent(ua),
		colly.AllowedDomains(config.Main.AllowedDomains...),
	)

	if config.Main.Delay != 0 {
		c.Limit(&colly.LimitRule{
			DomainGlob: ".*",
			Delay:      time.Duration(config.Main.Delay) * time.Second,
		})
	}

	if config.Main.RandomDelay != 0 {
		c.Limit(&colly.LimitRule{
			DomainGlob:  ".*",
			RandomDelay: time.Duration(config.Main.RandomDelay) * time.Second,
		})
	}

	c.OnRequest(func(r *colly.Request) {
		log.Println("Attempting to load:", r.URL.String())
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Error:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Println("Loaded page from:", r.Request.URL)
	})
}

// initPipeline initalize the pipeline as defined in the config
func initPipeline() {
	for _, p := range config.Pipeline {
		i := *p
		switch p.Action {
		case "follow":
			c.OnHTML(p.Selector, func(e *colly.HTMLElement) {
				log.Printf("Triggering follow pipeline: %q\n", i.Selector)
				follow(e, i)
			})
		case "collect":
			c.OnHTML(p.Selector, func(e *colly.HTMLElement) {
				log.Printf("Triggering collect pipeline: %q\n", i.Selector)
				collect(e, i)
			})
		}
	}
}

// save writes data to disk
func save() {
	switch config.Main.Format {
	case "json":
		saveJSON()
	case "csv":
		saveCSV()
	}
}

// saveJSON write JSON to disk
func saveJSON() {
	j, err := json.Marshal(collateJSON(saved))
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(config.Main.File, j, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Saved", config.Main.File)
}

// saveCSV write CSV to disk
func saveCSV() {
	file, err := os.Create(config.Main.File)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range collateCSV(saved) {
		err := writer.Write(value)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Saved", config.Main.File)
}

// actions

// follow instructs Skrapa to follow a link
func follow(e *colly.HTMLElement, p PipelineItem) {
	link := e.Attr(p.Attribute)
	u := e.Request.AbsoluteURL(link)
	log.Println("Following link", u)

	// VisitOnce flag used to avoid looping over a common link that might be followed
	if p.VisitOnce {
		if u == e.Request.URL.String() {
			log.Println("Revisit encountered but visit-once enabled, skipping:", u)
			return
		}
	}

	err := c.Visit(u)
	if err != nil {
		log.Fatal(err)
	}
}

// collect instructs Skrapa to save data
func collect(e *colly.HTMLElement, p PipelineItem) {
	var data string
	if p.Attribute == "text" {
		data = strings.TrimSpace(e.Text)
	} else {
		data = e.Attr(p.Attribute)
	}

	saved[p.Column] = append(saved[p.Column], data)
	log.Printf("Collecting data: %q -> %s\n", p.Column, data)
}

// util

// collateJSON will massage the data into maps
func collateJSON(data map[string][]string) []map[string]string {
	collated := make([]map[string]string, 0)
	keys := make([]string, 0)

	for k := range data {
		keys = append(keys, k)
	}

	if len(keys) < 1 {
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

	if len(keys) < 1 {
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

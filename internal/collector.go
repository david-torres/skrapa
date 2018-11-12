package internal

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	bolt "go.etcd.io/bbolt"
)

// DB the root bucket name
var DB = []byte("DB")

// Entries the data bucket name
var Entries = []byte("ENTRIES")

const defaultUA = "Skrapa"

// Collector data collector
type Collector struct {
	*colly.Collector                     // extend colly.Collector
	db               *bolt.DB            // database output
	script           *Script             // script used to instruct Skrapa
	data             map[string][]string // temp data storage
}

// NewCollector return an extended colly.Collector configured just the way Skrapa likes it based on the given script
func NewCollector(script *Script, db *bolt.DB) *Collector {
	// init colly
	c := initColly(script)

	// db
	scriptBytes, err := gobEncode(script)
	if err != nil {
		panic(fmt.Sprintf("could not encode script: %v", err))
	}

	// write script to db
	err = db.Update(func(tx *bolt.Tx) error {
		err = tx.Bucket([]byte("DB")).Put([]byte("SCRIPT"), scriptBytes)
		if err != nil {
			return fmt.Errorf("could not write script settings to db: %v", err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	log.Println("Wrote script settings to db")

	// init Collector
	collector := &Collector{
		c,
		db,
		script,
		make(map[string][]string),
	}

	// load the pipeline
	for _, p := range script.Pipeline {
		item := *p // copy for closure
		switch p.Action {
		case "follow":
			collector.OnHTML(p.Selector, func(el *colly.HTMLElement) {
				log.Printf("Triggering follow pipeline: %q\n", item.Selector)
				collector.follow(el, item)
			})
		case "collect":
			collector.OnHTML(p.Selector, func(el *colly.HTMLElement) {
				log.Printf("Triggering collect pipeline: %q\n", item.Selector)
				collector.collect(el, item)
			})
		}
	}

	return collector
}

// Run execute the data collection
func (c Collector) Run() {
	// start running
	log.Printf("Running %s", c.script.Main.URL)
	err := c.Visit(c.script.Main.URL)
	if err != nil {
		log.Fatal(err)
	}

	// wait until done...
	c.Wait()
	log.Println("Run complete, saving data")

	// save to database
	c.save()
	log.Println("Data saved")
}

func initColly(script *Script) *colly.Collector {
	ua := defaultUA
	if script.Main.UserAgent != "" {
		ua = script.Main.UserAgent
	}

	c := colly.NewCollector(
		colly.Async(false),
		colly.UserAgent(ua),
		colly.AllowedDomains(script.Main.AllowedDomains...),
	)

	if script.Main.Delay != 0 {
		log.Println("Adding delay: " + strconv.Itoa(script.Main.Delay))
		c.Limit(&colly.LimitRule{
			DomainGlob: ".*",
			Delay:      time.Duration(script.Main.Delay) * time.Second,
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

	return c
}

// persist data to the entries bucket
func (c Collector) save() {
	for key, item := range c.data {
		err := c.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(DB).Bucket(Entries)
			if bucket == nil {
				return errors.New("entries bucket not found")
			}

			enc, err := gobEncode(item)
			if err != nil {
				return err
			}

			err = bucket.Put([]byte(key), enc)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}
	}
}

// Skrapa actions

// follow instructs Skrapa to follow a link
func (c Collector) follow(e *colly.HTMLElement, p PipelineItem) {
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
func (c Collector) collect(e *colly.HTMLElement, p PipelineItem) {
	var data string
	if p.Attribute == "text" {
		data = strings.TrimSpace(e.Text)
	} else {
		data = e.Attr(p.Attribute)
	}

	c.data[p.Column] = append(c.data[p.Column], data)
	log.Printf("Collecting data: %q -> %s\n", p.Column, data)
}

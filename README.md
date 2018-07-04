# Skrapa: Web Scraping Utility

Skrapa is a web scraping tool designed to be as easy to use for non-technical folk as possible. It combines the powerful [Colly](http://go-colly.org/) library with a simple configuration format. Simply write out a pipeline of commands to instruct Skrapa to follow links and collect data from pages.

To use Skrapa, download the [latest release](https://github.com/david-torres/skrapa/releases) and create a configuration file for it to follow. Check out the [examples folder](https://github.com/david-torres/skrapa/tree/master/examples) for inspiration.

Run Skrapa from the command line:

    $ skrapa --file github_stars.toml

## Skrapa Configuration Documentation

Skrapa configuration is in [TOML](https://github.com/toml-lang/toml#toml) format. It has two primary parts, the main configuration block and the pipeline. The main block tells Skrapa what URL to scrape and where to save data. The pipeline is a repeatable configuration block that consists of commands for Skrapa to follow.

```toml
# primary configuration block
[main]
url = "https://example.com" # the url to scrape
format = "csv" # the output format, json or csv
file = "test.csv" # the output file
user_agent = "Skrapa" # the user agent sent to websites
allowed_domains = ["example.com"] # restrict any follow actions to these domains
delay = 1 # introduce a delay in seconds

# multiple pipeline blocks instruct Skrapa what to do
[[pipeline]]
selector = "a.link-class" # the 'selector' field allows Skrapa to use css selectors to find elements
action = "follow" # the 'action' field tells Skrapa what action to perform, in this case, follow a link
attr = "href" # the 'attr' field tells Skrapa which attribute of this element to use as a url to follow
visit_once = true # the "visit_once" field is used when the link you are following could appear again on subsequent pages, triggering a looping pipeline, this flag instructs Skrapa to only visit a given URL once

[[pipeline]]
selector = "span.title"
action = "collect" # the collect action tells Skrapa this is data we want to save
column = "title" # the 'column' field tells Skrapa what column/field we should save this data under
attr = "text" # the 'attr' field tells Skrapa which attribute of this element we want to save

[[pipeline]]
selector = "span.name"
action = "collect"
column = "name"
attr = "text"
```

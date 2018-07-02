package main

// Config models the Skrapa TOML config
type Config struct {
	Main struct {
		URL            string   `toml:"url"`
		File           string   `toml:"file"`
		Format         string   `toml:"format"`
		UserAgent      string   `toml:"user_agent,omitempty"`
		Delay          int      `toml:"delay,omitempty"`
		RandomDelay    int      `toml:"random_delay,omitempty"`
		AllowedDomains []string `toml:"allowed_domains,omitempty"`
	} `toml:"main"`
	Pipeline []*PipelineItem `toml:"pipeline"`
}

// PipelineItem models a Skrapa pipeline item
type PipelineItem struct {
	Selector  string `toml:"selector"`
	Action    string `toml:"action"`
	Attribute string `toml:"attr"`
	Column    string `toml:"column,omitempty"`
	VisitOnce bool   `toml:"visit_once,omitempty"`
}

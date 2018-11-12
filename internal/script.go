package internal

// Script models the Skrapa TOML config
type Script struct {
	Main struct {
		URL            string   `toml:"url"`
		File           string   `toml:"file"`
		Format         string   `toml:"format"`
		UserAgent      string   `toml:"user_agent,omitempty"`
		Delay          int      `toml:"delay,omitempty"`
		AllowedDomains []string `toml:"allowed_domains,omitempty"`
	} `toml:"main"`
	Pipeline []*PipelineItem `toml:"pipeline"`
}

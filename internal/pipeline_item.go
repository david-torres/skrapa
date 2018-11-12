package internal

// PipelineItem models a Skrapa pipeline item
type PipelineItem struct {
	Selector  string `toml:"selector"`
	Action    string `toml:"action"`
	Attribute string `toml:"attr"`
	Column    string `toml:"column,omitempty"`
	VisitOnce bool   `toml:"visit_once,omitempty"`
}

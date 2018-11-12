package internal

import (
	"github.com/BurntSushi/toml"
)

// ScriptParser parse scripts
type ScriptParser struct{}

// NewScriptParser return a new script parser
func NewScriptParser() *ScriptParser {
	return &ScriptParser{}
}

// Parse do the parsing of scripts
func (p ScriptParser) Parse(fileName string) (*Script, error) {
	script := &Script{}
	if _, err := toml.DecodeFile(fileName, script); err != nil {
		return script, err
	}

	return script, nil
}

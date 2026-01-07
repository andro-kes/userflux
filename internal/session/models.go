package session

import (
	"os"
	"time"
)

type Session struct {
	ResultFile *os.File
	Users      int
	Time       time.Duration // время работы агента
	Data       ScriptYAML
}

type ScriptYAML struct {
	Config struct {
		Users int    `yaml:"users"`
		Time  string `yaml:"time"`
	} `yaml:"config"`

	Script struct {
		Name string     `yaml:"name"`
		Flow []FlowStep `yaml:"flow"`
	} `yaml:"script"`
}

type FlowStep struct {
	Name    string  `yaml:"name"`
	URL     string  `yaml:"url"`
	Request Request `yaml:"request"`
}

type Request struct {
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Headers map[string]string `yaml:"headers"`
}

func NewSessionStruct() *Session {
	return &Session{}
}

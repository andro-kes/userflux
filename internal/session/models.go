package session

import (
	"os"
	"time"
)

// Logger interface to avoid import cycles
type Logger interface {
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
}

type Session struct {
	ResultFile *os.File
	Time       time.Duration // время работы агента
	Data       ScriptYAML
	Logger     Logger
}

type ScriptYAML struct {
	Config struct {
		Generated []Gen `yaml:"generated"` // request data
		Time  string `yaml:"time"`
	} `yaml:"config"`

	Script struct {
		Name string     `yaml:"name"`
		Flow []FlowStep `yaml:"flow"`
	} `yaml:"script"`
}

type FlowStep struct {
	Name    string   `yaml:"name"`
	URL     string   `yaml:"url"`
	Body    []string `yaml:"body"`
	Request Request  `yaml:"request"`
}

type Gen struct {
	Model  string
	Fields []string
}

type Request struct {
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Headers map[string]string `yaml:"headers"`
}

func NewSessionStruct() *Session {
	return &Session{}
}

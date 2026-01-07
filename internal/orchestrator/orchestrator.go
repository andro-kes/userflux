package orchestrator

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Session struct {
	ResultFile *os.File
	Users      int
	Time       time.Duration // время работы агента
	Data       ScriptYAML
}

// This matches script
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

func newSessionStruct() *Session {
	return &Session{}
}

// Orchester:
// - reads YAML script file
// - decodes it
// - fills Session fields
// - creates session/result files
func Orchester(scriptName string) error {
	s := newSessionStruct()

	// 1) Read + decode YAML script
	raw, err := os.ReadFile("../scripts/" + scriptName)
	if err != nil {
		return err
	}

	var sc ScriptYAML
	if err := yaml.Unmarshal(raw, &sc); err != nil {
		return err
	}

	// 2) Map script data into Session
	s.Data = sc
	s.Users = sc.Config.Users

	dur, err := time.ParseDuration(sc.Config.Time)
	if err != nil {
		return err
	}
	s.Time = dur

	// 3) Create session + result files
	fS, n, err := createNewSessionFile()
	if err != nil {
		return err
	}
	defer fS.Close()

	fR, err := createNewResult(n)
	if err != nil {
		return err
	}
	s.ResultFile = fR
	// No defer

	// 4) save the script into the session file (encoded YAML)
	enc := yaml.NewEncoder(fS)
	enc.SetIndent(2)
	if err := enc.Encode(sc); err != nil {
		return err
	}
	_ = enc.Close()

	// runAgent

	return nil
}

// Returns: descriptor on new session file, session number, error
func createNewSessionFile() (*os.File, string, error) {
	cnt := 0
	_ = filepath.Walk("../sessions", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		cnt++
		return nil
	})

	number := strconv.Itoa(cnt)
	file, err := os.Create("../sessions/session_" + number)
	if err != nil {
		return nil, "", err
	}

	return file, number, nil
}

func createNewResult(n string) (*os.File, error) {
	f, err := os.Create("../results/result_" + n)
	if err != nil {
		return nil, err
	}
	return f, nil
}

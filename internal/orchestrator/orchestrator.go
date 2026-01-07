package orchestrator

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/andro-kes/userflux/internal/agent"
	"github.com/andro-kes/userflux/internal/session"
	"gopkg.in/yaml.v3"
)

// Orchester:
// - reads YAML script file
// - decodes it
// - fills Session fields
// - creates session/result files
func Orchestrator(scriptName string) error {
	s := session.NewSessionStruct()

	// 1) Read + decode YAML script
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(exe)
	p := filepath.Join(dir, "..", "scripts", scriptName)
	p = filepath.Clean(p)
	raw, err := os.ReadFile(p)
	if err != nil {
		return err
	}

	var sc session.ScriptYAML
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

	agent.RunAgent(s)

	return nil
}

// Returns: descriptor on new session file, session number, error
func createNewSessionFile() (*os.File, string, error) {
	cnt := 0
	exe, err := os.Executable()
	if err != nil {
		return nil, "", err
	}
	dir := filepath.Dir(exe)
	p := filepath.Join(dir, "..", "sessions")
	p = filepath.Clean(p)
	_ = filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		cnt++
		return nil
	})

	number := strconv.Itoa(cnt)
	p = filepath.Join(p, "session_" + number)
	p = filepath.Clean(p)
	file, err := os.Create(p)
	if err != nil {
		return nil, "", err
	}

	return file, number, nil
}

func createNewResult(n string) (*os.File, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(exe)
	p := filepath.Join(dir, "..", "results", "result_" + n)
	p = filepath.Clean(p)
	f, err := os.Create(p)
	if err != nil {
		return nil, err
	}
	return f, nil
}

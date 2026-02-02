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

func baseDir() string {
	wd, err := os.Getwd()
	if err == nil {
		return wd
	}

	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// Orchester:
// - reads YAML script file
// - decodes it
// - fills Session fields
// - creates session/result files
func Orchestrator(scriptName string, logger session.Logger) error {
	logger.Info("Orchestrator starting")
	s := session.NewSessionStruct()
	s.Logger = logger

	// 1) Read + decode YAML script
	root := baseDir()
	p := filepath.Join(root, "scripts", scriptName)
	p = filepath.Clean(p)
	raw, err := os.ReadFile(p)
	if err != nil {
		logger.Errorf("Failed to read script file: %v", err)
		return err
	}

	logger.Info("Parsing YAML script")
	var sc session.ScriptYAML
	if err := yaml.Unmarshal(raw, &sc); err != nil {
		logger.Errorf("Failed to parse YAML: %v", err)
		return err
	}

	// 2) Map script data into Session
	s.Data = sc

	dur, err := time.ParseDuration(sc.Config.Time)
	if err != nil {
		logger.Errorf("Failed to parse duration: %v", err)
		return err
	}
	s.Time = dur
	logger.Infof("Configured for duration: %s", s.Time)

	// 3) Create session + result files
	logger.Info("Creating session file")
	fS, n, err := createNewSessionFile(logger)
	if err != nil {
		logger.Errorf("Failed to create session file: %v", err)
		return err
	}
	defer fS.Close()

	logger.Infof("Creating result file for session %s", n)
	fR, err := createNewResult(n, logger)
	if err != nil {
		logger.Errorf("Failed to create result file: %v", err)
		return err
	}
	s.ResultFile = fR
	// No defer

	// 4) save the script into the session file (encoded YAML)
	logger.Info("Encoding script to session file")
	enc := yaml.NewEncoder(fS)
	enc.SetIndent(2)
	if err := enc.Encode(sc); err != nil {
		logger.Errorf("Failed to encode script: %v", err)
		return err
	}
	_ = enc.Close()

	agent.RunAgent(s)

	return nil
}

// Returns: descriptor on new session file, session number, error
func createNewSessionFile(logger session.Logger) (*os.File, string, error) {
	cnt := 0
	dir := baseDir()

	p := filepath.Join(dir, "sessions")
	p = filepath.Clean(p)

	if err := os.MkdirAll(p, 0755); err != nil {
		return nil, "", err
	}

	_ = filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		cnt++
		return nil
	})

	number := strconv.Itoa(cnt)
	p = filepath.Join(p, "session_"+number)
	p = filepath.Clean(p)
	logger.Infof("Creating session file: %s", p)
	file, err := os.Create(p)
	if err != nil {
		return nil, "", err
	}

	return file, number, nil
}

func createNewResult(n string, logger session.Logger) (*os.File, error) {
	dir := baseDir()

	p := filepath.Join(dir, "results")
	p = filepath.Clean(p)

	if err := os.MkdirAll(p, 0755); err != nil {
		return nil, err
	}

	p = filepath.Join(p, "result_"+n)
	p = filepath.Clean(p)
	logger.Infof("Creating result file: %s", p)

	f, err := os.Create(p)
	if err != nil {
		return nil, err
	}
	return f, nil
}

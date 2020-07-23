package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	JSON = "json"
	YAML = "yaml"
)

type Config struct {
	OutDir      string
	OutDirBib   string
	DefaultRepo string
	ExecCmd     string
	TermUi      bool
}

var UserConfig *Config

func Init() (err error) {
	homeDir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	UserConfig = &Config{
		OutDir:      homeDir,
		OutDirBib:   homeDir,
		DefaultRepo: "libgen",
		TermUi:      true,
	}
	return
}

func (c *Config) Parse(fp string) error {
	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()
	ext := filepath.Ext(fp)
	if ext == "" {
		return errors.New("config: filepath without extension")
	}
	switch strings.ToLower(ext[1:]) {
	case JSON:
		return c.parseJson(f)
	case YAML:
		return c.parseYaml(f)
	}
	return errors.New("config: parser not supported")
}

func (c *Config) parseJson(r io.Reader) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(c)
}

func (c *Config) parseYaml(r io.Reader) error {
	decoder := yaml.NewDecoder(r)
	return decoder.Decode(c)
}

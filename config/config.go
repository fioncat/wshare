package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fioncat/wshare/pkg/osutil"
	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var Default []byte

var defaultInstance = func() *Config {
	var cfg Config
	err := yaml.Unmarshal(Default, &cfg)
	if err != nil {
		panic("internal: parse default config failed: " + err.Error())
	}
	err = validate(&cfg)
	if err != nil {
		panic("internal: validate default config failed: " + err.Error())
	}
	return &cfg
}()

func validate(cfg *Config) error {
	err := validator.New().Struct(cfg)
	if err == nil {
		return nil
	}
	if _, ok := err.(*validator.InvalidValidationError); ok {
		// This only occur with code error such as interface with nil.
		// When you find this kind of error, please check the validate
		// caller code and Config struct.
		return fmt.Errorf("internal: validatation error: %v", err)
	}

	// convert the field name from validator to yaml style.
	// Such as "Config.Name" to "name"
	convertFieldName := func(ns string) string {
		parts := strings.Split(ns, ".")
		if len(parts) <= 1 {
			return ns
		}

		parts = parts[1:]
		for i, part := range parts {
			parts[i] = strcase.ToSnake(part)
		}

		return strings.Join(parts, ".")
	}

	errs := err.(validator.ValidationErrors)
	lines := make([]string, len(errs))
	for i, err := range errs {
		name := convertFieldName(err.StructNamespace())
		tag := err.Tag()
		line := fmt.Sprintf(" * validate tag %q for field %q failed", tag, name)
		lines[i] = line
	}

	// The error output example:
	// validate config failed:
	//  * validate tag "required" for field "name" failed
	//  * validate tag "required" for field "server" failed
	//  ... more
	return fmt.Errorf("validate config failed:\n%s", strings.Join(lines, "\n"))
}

// parse yaml data to config instance. Include validating.
func parse(data []byte) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("parse yaml failed: %v", err)
	}

	// Use mergo to replace empty required fields with default values.
	// For example, field `Server` is required, but the user does not
	// fill it in, we replace it with the value in defaultInstance.
	err = mergo.Merge(&cfg, defaultInstance)
	if err != nil {
		// the mergo should not fail in normal case.
		// If occur, it means that we have type error, such as
		// two params' type is mismatch.
		return nil, fmt.Errorf("internal: merge default config failed: %v", err)
	}

	err = validate(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

type Config struct {
	Name   string `yaml:"name" json:"name"`
	Server string `yaml:"server" validate:"required" json:"server"`

	Clipboard *Clipboard `yaml:"clipboard" json:"clipboard"`

	Log *Log `yaml:"log" validate:"dive" json:"log"`
}

type Clipboard struct {
	Readonly bool `yaml:"readonly" json:"readonly"`
}

type Log struct {
	Level string `yaml:"level" validate:"required" json:"level"`
}

var (
	path     string
	homeDir  string
	instance *Config

	initOnce sync.Once
)

func Init() error {
	var err error
	initOnce.Do(func() {
		err = doInit()
	})
	return err
}

func doInit() error {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %v", err)
	}

	path = os.Getenv("WSHARE_CONFIG")
	if path == "" {
		path = filepath.Join(homeDir, ".config", "wshare", "daemon.yaml")
	}

	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// User does not create config file, use a default one
			// The default config fields are defined in `default.yaml`
			instance = defaultInstance
			return nil
		}
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("config file %q is a directory", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	instance, err = parse(data)
	if err != nil {
		return err
	}
	return nil
}

func Get() *Config {
	if instance == nil {
		panic("internal: please call config.Init before using Get()")
	}
	return instance
}

func HomeDir() string {
	if homeDir == "" {
		panic("internal: please call config.Init before using HomeDir()")
	}
	return homeDir
}

func Path() string {
	if homeDir == "" {
		panic("internal: please call config.Init before using Path()")
	}
	return path
}

func LocalFile(name string) (string, error) {
	localPath := filepath.Join(HomeDir(), ".local", "share", "wshare")
	err := osutil.EnsureDir(localPath)
	if err != nil {
		return "", err
	}

	return filepath.Join(localPath, name), nil
}

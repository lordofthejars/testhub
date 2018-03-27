package hub

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var configOrder *yaml.MapSlice

type RepositoryConfig struct {
	Path string `yaml:"path"`
}

type Config struct {
	Repository RepositoryConfig `yaml:"repository"`
	Port       int              `yaml:"port"`
}

func NewConfig(configurationFile string) (*Config, error) {
	configurationPath := configFile(configurationFile)
	return readConfig(configurationPath)
}

func configFile(path string) string {
	if len(path) > 0 {
		return path
	}
	return "testhub.yml"
}

func readConfig(filename string) (*Config, error) {
	var config *Config
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		Info("No configuration file found, relaying to defaults")
		config, err = defaultConfig(), nil
	}
	config, err = unmarshal(data)

	config = applyDefaults(config)
	os.Chdir(filepath.Dir(config.Repository.Path))

	return config, err
}

func applyDefaults(config *Config) *Config {
	if config.Port == 0 {
		config.Port = 8000
	}

	if len(config.Repository.Path) == 0 {
		config.Repository.Path = resolveStorageDirectory()
	}

	return config
}

func defaultConfig() *Config {
	return applyDefaults(&Config{})
}

func unmarshal(data []byte) (*Config, error) {
	var config *Config
	err := yaml.Unmarshal(data, &config)

	return config, err
}

func resolveStorageDirectory() string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	return filepath.Join(dir, ".hub")
}

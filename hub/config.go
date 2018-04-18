package hub

import (
	"io/ioutil"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var configOrder *yaml.MapSlice

type RepositoryConfig struct {
	Path string `yaml:"path"`
}

type AuthenticationConfig struct {
	Secret    string `yaml:"secret"`
	UsersPath string `yaml:"userspath"`
}

type Config struct {
	Repository     RepositoryConfig     `yaml:"repository"`
	Authentication AuthenticationConfig `yaml:"authentication"`
	Port           int                  `yaml:"port"`
	Cert           string               `yaml:"cert"`
	Key            string               `yaml:"key"`
}

func (config Config) isSSLConfigured() bool {
	return len(config.Cert) > 0 && len(config.Key) > 0
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
		return defaultConfig(), nil
	}
	config, err = unmarshal(data)

	config = applyDefaults(config)

	return config, err
}

func applyDefaults(config *Config) *Config {
	if config.Port == 0 {
		if config.isSSLConfigured() {
			config.Port = 443
		} else {
			config.Port = 8000
		}
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

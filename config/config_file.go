// Copyright (c) OpenFaaS Author(s) 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package config

import (
	"bytes"
	"encoding/base64"

	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

// AuthType auth type
type AuthType string

const (
	//BasicAuthType basic authentication type
	BasicAuthType = "basic"
	//Oauth2AuthType oauth2 authentication type
	Oauth2AuthType = "oauth2"

	// ConfigLocationEnv is the name of he env variable used
	// to configure the location of the faas-cli config folder.
	// When not set, DefaultDir location is used.
	ConfigLocationEnv string = "OPENFAAS_CONFIG"

	DefaultDir         string      = "~/.openfaas"
	DefaultFile        string      = "config.yml"
	DefaultPermissions os.FileMode = 0700

	// DefaultCIDir creates the 'openfaas' directory in the current directory
	// if running in a CI environment.
	DefaultCIDir string = ".openfaas"
	// DefaultCIPermissions creates the config file with elevated permissions
	// for it to be read by multiple users when running in a CI environment.
	DefaultCIPermissions os.FileMode = 0744
)

// ConfigFile for OpenFaaS CLI exclusively.
type ConfigFile struct {
	AuthConfigs []AuthConfig `yaml:"auths"`
	FilePath    string       `yaml:"-"`
}

type AuthConfig struct {
	Gateway string   `yaml:"gateway,omitempty"`
	Auth    AuthType `yaml:"auth,omitempty"`
	Token   string   `yaml:"token,omitempty"`
	Options []Option `yaml:"options,omitempty"`
}

type Option struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// New initializes a config file for the given file path
func New(filePath string) (*ConfigFile, error) {
	if filePath == "" {
		return nil, fmt.Errorf("can't create config with empty filePath")
	}
	conf := &ConfigFile{
		AuthConfigs: make([]AuthConfig, 0),
		FilePath:    filePath,
	}

	return conf, nil
}

// ConfigDir returns the path to the faas-cli config directory.
// When
// 1. CI = "true" and OPENFAAS_CONFIG="", then it will return `.openfaas`, which is located in the current working directory.
// 2. CI = "true" and OPENFAAS_CONFIG="<path>", then it will return the path value in  OPENFAAS_CONFIG
// 3. CI = "" and OPENFAAS_CONFIG="", then it will return the default location ~/.openfaas
func ConfigDir() string {
	override := os.Getenv(ConfigLocationEnv)
	ci := isRunningInCI()

	switch {
	// case (1) from docs string
	case ci && override == "":
		return DefaultCIDir
	// case (2) from the doc string
	case override != "":
		// case (3) from the doc string
		return override
	default:
		return DefaultDir
	}
}

// isRunningInCI checks the ENV var CI and returns true if it's set to true or 1
func isRunningInCI() bool {
	if env, ok := os.LookupEnv("CI"); ok {
		if env == "true" || env == "1" {
			return true
		}
	}
	return false
}

// EnsureFile creates the root dir and config file
func EnsureFile() (string, error) {
	permission := DefaultPermissions
	dir := ConfigDir()
	if isRunningInCI() {
		permission = DefaultCIPermissions
	}
	dirPath, err := homedir.Expand(dir)
	if err != nil {
		return "", err
	}

	filePath := path.Clean(filepath.Join(dirPath, DefaultFile))
	if err := os.MkdirAll(filepath.Dir(filePath), permission); err != nil {
		return "", err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return "", err
		}
		defer file.Close()
	}

	return filePath, nil
}

// FileExists returns true if the config file is located at the default path
func fileExists() bool {
	dir := ConfigDir()
	dirPath, err := homedir.Expand(dir)
	if err != nil {
		return false
	}

	filePath := path.Clean(filepath.Join(dirPath, DefaultFile))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

// Save writes the config to disk
func (configFile *ConfigFile) save() error {
	file, err := os.OpenFile(configFile.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	var buff bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buff)
	yamlEncoder.SetIndent(2) // this is what you're looking for
	if err := yamlEncoder.Encode(&configFile); err != nil {
		return err
	}

	_, err = file.Write(buff.Bytes())
	return err
}

// Load reads the yml file from disk
func (configFile *ConfigFile) load() error {
	conf := &ConfigFile{}

	if _, err := os.Stat(configFile.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("can't load config from non existent filePath")
	}

	data, err := ioutil.ReadFile(configFile.FilePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, conf); err != nil {
		return err
	}

	if len(conf.AuthConfigs) > 0 {
		configFile.AuthConfigs = conf.AuthConfigs
	}
	return nil
}

// EncodeAuth encodes the username and password strings to base64
func EncodeAuth(username string, password string) string {
	input := username + ":" + password
	msg := []byte(input)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
	base64.StdEncoding.Encode(encoded, msg)
	return string(encoded)
}

// DecodeAuth decodes the input string from base64 to username and password
func DecodeAuth(input string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", "", err
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", fmt.Errorf("invalid auth config file")
	}
	return arr[0], arr[1], nil
}

// UpdateAuthConfig creates or updates the username and password for a given gateway
func UpdateAuthConfig(authConfig AuthConfig) error {
	gateway := authConfig.Gateway

	_, err := url.ParseRequestURI(gateway)
	if err != nil || len(gateway) < 1 {
		return fmt.Errorf("invalid gateway URL")
	}

	configPath, err := EnsureFile()
	if err != nil {
		return err
	}

	cfg, err := New(configPath)
	if err != nil {
		return err
	}

	if err := cfg.load(); err != nil {
		return err
	}

	index := -1
	for i, v := range cfg.AuthConfigs {
		if gateway == v.Gateway {
			index = i
			break
		}
	}

	if index == -1 {
		cfg.AuthConfigs = append(cfg.AuthConfigs, authConfig)
	} else {
		cfg.AuthConfigs[index] = authConfig
	}

	if err := cfg.save(); err != nil {
		return err
	}

	return nil
}

// LookupAuthConfig returns the username and password for a given gateway
func LookupAuthConfig(gateway string) (AuthConfig, error) {
	var authConfig AuthConfig

	if !fileExists() {
		return authConfig, fmt.Errorf("config file not found")
	}

	configPath, err := EnsureFile()
	if err != nil {
		return authConfig, err
	}

	cfg, err := New(configPath)
	if err != nil {
		return authConfig, err
	}

	if err := cfg.load(); err != nil {
		return authConfig, err
	}

	for _, v := range cfg.AuthConfigs {
		if gateway == v.Gateway {
			authConfig = v
			return authConfig, nil
		}
	}

	return authConfig, fmt.Errorf("no auth config found for %s", gateway)
}

// RemoveAuthConfig deletes the username and password for a given gateway
func RemoveAuthConfig(gateway string) error {
	if !fileExists() {
		return fmt.Errorf("config file not found")
	}

	configPath, err := EnsureFile()
	if err != nil {
		return err
	}

	cfg, err := New(configPath)
	if err != nil {
		return err
	}

	if err := cfg.load(); err != nil {
		return err
	}

	index := -1
	for i, v := range cfg.AuthConfigs {
		if gateway == v.Gateway {
			index = i
			break
		}
	}

	if index > -1 {
		cfg.AuthConfigs = removeAuthByIndex(cfg.AuthConfigs, index)
		if err := cfg.save(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("gateway %s not found in config", gateway)
	}

	return nil
}

func removeAuthByIndex(s []AuthConfig, index int) []AuthConfig {
	return append(s[:index], s[index+1:]...)
}

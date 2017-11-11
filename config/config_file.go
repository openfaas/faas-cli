// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package config

import (
	"encoding/base64"

	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

var (
	DefaultDir  = "~/.openfaas"
	DefaultFile = "config.yml"
)

// ConfigFile for OpenFaaS CLI exclusively.
type ConfigFile struct {
	AuthConfigs []AuthConfig `yaml:"auths"`
	FilePath    string       `yaml:"-"`
}

type AuthConfig struct {
	Gateway string `yaml:"gateway,omitempty"`
	Auth    string `yaml:"auth,omitempty"`
	Token   string `yaml:"token,omitempty"`
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

// EnsureFile creates the root dir and config file
func EnsureFile() (string, error) {
	dirPath, err := homedir.Expand(DefaultDir)
	if err != nil {
		return "", err
	}

	filePath := path.Clean(filepath.Join(dirPath, DefaultFile))
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
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
	dirPath, err := homedir.Expand(DefaultDir)
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

	data, err := yaml.Marshal(configFile)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
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
func UpdateAuthConfig(gateway string, username string, password string) error {
	_, err := url.ParseRequestURI(gateway)
	if err != nil || len(gateway) < 1 {
		return fmt.Errorf("invalid gateway URL")
	}

	if len(username) < 1 {
		return fmt.Errorf("username can't be an empty string")
	}

	if len(password) < 1 {
		return fmt.Errorf("password can't be an empty string")
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

	auth := AuthConfig{
		Gateway: gateway,
		Auth:    "basic",
		Token:   EncodeAuth(username, password),
	}

	index := -1
	for i, v := range cfg.AuthConfigs {
		if gateway == v.Gateway {
			index = i
			break
		}
	}

	if index == -1 {
		cfg.AuthConfigs = append(cfg.AuthConfigs, auth)
	} else {
		cfg.AuthConfigs[index] = auth
	}

	if err := cfg.save(); err != nil {
		return err
	}

	return nil
}

// LookupAuthConfig returns the username and password for a given gateway
func LookupAuthConfig(gateway string) (string, string, error) {
	if !fileExists() {
		return "", "", fmt.Errorf("config file not found")
	}

	configPath, err := EnsureFile()
	if err != nil {
		return "", "", err
	}

	cfg, err := New(configPath)
	if err != nil {
		return "", "", err
	}

	if err := cfg.load(); err != nil {
		return "", "", err
	}

	for _, v := range cfg.AuthConfigs {
		if gateway == v.Gateway {
			user, pass, err := DecodeAuth(v.Token)
			if err != nil {
				return "", "", err
			}
			return user, pass, nil
		}
	}

	return "", "", fmt.Errorf("no auth config found for %s", gateway)
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

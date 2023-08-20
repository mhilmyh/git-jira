package main

import (
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
)

type Config struct {
	Version  string            `yaml:"version"`
	Settings map[string]string `yaml:"settings"`
}

func (c *Config) GetUserEmail() string {
	if v, ok := c.Settings["user_email"]; ok {
		return v
	}
	return ""
}

func (c *Config) GetPersonalAccessToken() string {
	if v, ok := c.Settings["personal_access_token"]; ok {
		return v
	}
	return ""
}

func (c *Config) GetOrganization() string {
	if v, ok := c.Settings["organization"]; ok {
		return v
	}
	return ""
}

func (c *Config) GetProjectKey() string {
	if v, ok := c.Settings["project_key"]; ok {
		return v
	}
	return ""
}

func (c *Config) GetJiraApiVersion() string {
	if v, ok := c.Settings["jira_api_version"]; ok {
		return v
	}
	return ""
}

func ReadConfigFile(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(file)

	var content []byte
	content, err = io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config *Config
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

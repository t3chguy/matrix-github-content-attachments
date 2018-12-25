package main

import (
	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
)

type Config struct {
	Github githubapp.Config   `yaml:"github"`
	Server baseapp.HTTPConfig `yaml:"server"`
	Matrix struct {
		HomeServerURL string `yaml:"hs_url"`
		UserID        string `yaml:"user_id"`
		AccessToken   string `yaml:"access_token"`
	} `yaml:"matrix"`
	Regexes struct {
		Rooms []string `yaml:"rooms"`
	} `yaml:"regexes"`
}

func (c *Config) GetRoomRegexes() []*regexp.Regexp {
	list := c.Regexes.Rooms
	if list == nil {
		list = []string{
			`https://matrix\.to/(?:#/)?(!.+?)(?:/.*)?$`,
			`https://view\.matrix\.org/room/(!.+?)(?:/.*)?$`,
		}
	}

	regexps := make([]*regexp.Regexp, len(list))
	for i := range list {
		regexps[i] = regexp.MustCompile(list[i])
	}

	return regexps
}

func ReadConfig(path string) (*Config, error) {
	var c Config

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading server config file: %s", path)
	}

	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		return nil, errors.Wrap(err, "failed parsing configuration file")
	}

	return &c, nil
}

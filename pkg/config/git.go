package config

import (
	"errors"
	"main/pkg/constants"
)

type GitConfig struct {
	Repository string `default:""   toml:"repository"`
	Token      string `toml:"token"`
}

func (c *GitConfig) Validate() error {
	if c.Repository == "" {
		return nil
	}

	if !constants.GithubRegexp.Match([]byte(c.Repository)) && !constants.GitopiaRegexp.Match([]byte(c.Repository)) {
		return errors.New("repository is not valid")
	}

	return nil
}

package twitterapi

import (
	"os"
	"strings"
	"io/ioutil"
	"github.com/pkg/errors"
)

type ApiKeyPair struct {
	ApiKey       string
	ApiSecretKey string
}

func LoadApiKey(apiKeyFile string) (*ApiKeyPair, error) {
	fileInfo, err := os.Stat(apiKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "not exists twitter api key file (%v)", apiKeyFile)
	}
	if fileInfo.Mode().Perm() != 0600 {
		return nil, errors.Errorf("twitter api key file have insecure permission (e.g. !=  0600) (%v)", apiKeyFile)
	}
	apiKeyPair, err := ioutil.ReadFile(apiKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "can not read twitter api key file (%v)", apiKeyFile)
	}
	s := strings.SplitN(string(apiKeyPair), "\n", 2)
	if len(s) < 2 {
		return nil, errors.Wrapf(err, "can not parse twitter api key file (%v)", apiKeyFile)
	}
	return &ApiKeyPair {
		ApiKey: strings.TrimSpace(s[0]),
		ApiSecretKey: strings.TrimSpace(s[1]),
	}, nil
}

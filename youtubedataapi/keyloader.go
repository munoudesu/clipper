package youtubedataapi

import (
	"os"
	"strings"
	"io/ioutil"
	"github.com/pkg/errors"
)

func LoadApiKey(apiKeyFile string) ([]string, error) {
	fileInfo, err := os.Stat(apiKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "not exists youtube data api key file (%v)", apiKeyFile)
	}
	if fileInfo.Mode().Perm() != 0600 {
		return nil, errors.Errorf("youtube data api key file have insecure permission (e.g. !=  0600) (%v)", apiKeyFile)
	}
	apiKeysBytes, err := ioutil.ReadFile(apiKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "can not read youtube data api key file (%v)", apiKeyFile)
	}
	apiKeysStrings := strings.Split(string(apiKeysBytes), "\n")
	apiKeys := make([]string, 0, len(apiKeysStrings))
	for _, s := range apiKeysStrings {
		apiKey := strings.TrimSpace(s)
		if strings.HasPrefix(apiKey, "#") {
			continue
		}
		if apiKey == "" {
			continue
		}
		apiKeys = append(apiKeys, apiKey)
	}
	return apiKeys, nil
}

package youtubedataapi

import (
	"os"
	"strings"
	"io/ioutil"
	"github.com/pkg/errors"
)

func LoadApiKey(apiKeyFile string) (string, error) {
	fileInfo, err := os.Stat(apiKeyFile)
	if err != nil {
		return "", errors.Wrapf(err, "not exists youtube data api key file (%v)", apiKeyFile)
	}
	if fileInfo.Mode().Perm() != 0600 {
		return "", errors.Errorf("youtube data api key file have insecure permission (e.g. !=  0600) (%v)", apiKeyFile)
	}
	apiKey, err := ioutil.ReadFile(apiKeyFile)
	if err != nil {
		return "", errors.Wrapf(err, "can not read youtube data api key file (%v)", apiKeyFile)
	}
	return strings.TrimSpace(string(apiKey)), nil
}

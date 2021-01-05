package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func GetPort(endpoint string) int {
	parts := strings.Split(endpoint, ":")
	if len(parts) != 2 {
		return 0
	}
	i, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	return i
}

func loadConfigFile(cfgPath string, ptr interface{}) error {
	if ptr == nil {
		return fmt.Errorf("ptr of type (%T) is nil", ptr)
	}
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s, err: %v", cfgPath, err)
	}
	if err := json.Unmarshal(data, ptr); err != nil {
		return fmt.Errorf("failed to unmarshal file %s, err: %v", cfgPath, err)
	}
	return nil
}

func LoadConfigFileOrPanic(cfgPath string, ptr interface{}) {
	if err := loadConfigFile(cfgPath, ptr); err != nil {
		log.Fatal().Err(err).Msgf("failed to load config %s", cfgPath)
	}
}

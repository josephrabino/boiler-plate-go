package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

var envConfigMap map[string]string

func GetEnvVariable(envKey, defaultValue string) string {
	if envConfigMap == nil {
		envConfigMap = make(map[string]string)
	}
	if _, exist := envConfigMap[envKey]; exist {
		return envConfigMap[envKey]
	} else if v, exist := os.LookupEnv(envKey); exist && v != "" {
		envValue := os.Getenv(envKey)
		logrus.Printf("Saving new value in cache map for key: %s", envKey)
		envConfigMap[envKey] = envValue
		return envValue
	} else {
		logrus.Printf("Saving default value in cache map for key: %s", envKey)
		return defaultValue
	}
}

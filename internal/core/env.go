package core

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	// переменные окружения для подключения к постгрес
	pgEnvs = map[string]string{
		"host":        "ASD_POSTGRES_HOST",
		"port":        "ASD_POSTGRES_PORT",
		"user":        "SERVICE_PG_ILOGIC_USERNAME",
		"password":    "SERVICE_PG_ILOGIC_PASSWORD",
		"db_name":     "ASD_POSTGRES_DBNAME",
		"numPullConn": "SERVICE_PG_NUM_PULL",
	}

	// переменные окружения для подключения к rabbitMQ
	rbEnvs = map[string]string{
		"host":     "ASD_RMQ_HOST",
		"port":     "ASD_RMQ_PORT",
		"user":     "SERVICE_RMQ_ENOTIFY_USERNAME",
		"password": "SERVICE_RMQ_ENOTIFY_PASSWORD",
		"v_host":   "ASD_RMQ_VHOST",
		// "heartbeat":     "ASD_RMQ_HEARTBEAT", // пока не используется
		"name_queue":    "SERVICE_RMQ_QUEUE",
		"name_consumer": "SERVICE_RMQ_NAME_CONSUMER",
	}
)

type envs struct {
	pgEnvs map[string]string
	rbEnvs map[string]string
}

func getEnvs() (envs, error) {
	var err error
	var envs envs
	envLoader := InitEnvLoader()
	envs.pgEnvs = envLoader.Load(pgEnvs)
	envs.rbEnvs = envLoader.Load(rbEnvs)

	if envLoader.CheckUnloadEnvs() {
		return envs, err
	}

	err = errors.New("not all environment variables are loaded")
	return envs, err
}

// метод загрузки переменных окружения
type EnvLoader struct {
	unloadedVariables []string
}

func InitEnvLoader() *EnvLoader {
	res := &EnvLoader{
		unloadedVariables: make([]string, 0),
	}
	return res
}

// метод загрузки переменных окружения
func (e *EnvLoader) Load(keys map[string]string) map[string]string {
	res := make(map[string]string, len(keys))

	for key, envKey := range keys {
		value, exists := os.LookupEnv(envKey)
		if exists && value != "" {
			res[key] = value
		} else {
			e.unloadedVariables = append(e.unloadedVariables, envKey)
			log.Warningf("the %s variable is not loaded", envKey)
		}
	}
	return res
}

// метод проверки, были ли незагруженные переменные
func (e *EnvLoader) CheckUnloadEnvs() bool {
	return len(e.unloadedVariables) <= 0
}

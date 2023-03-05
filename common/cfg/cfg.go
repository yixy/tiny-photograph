package cfg

import (
	"github.com/spf13/viper"
	"github.com/yixy/tiny-photograph/internal/log"
	"go.uber.org/zap"
)

const (
	PORT             = "PORT"
	READ_TIMEOUT     = "READ_TIMEOUT"
	WRITE_TIMEOUT    = "WRITE_TIMEOUT"
	SHUTDOWN_TIMEOUT = "SHUTDOWN_TIMEOUT"
)

const TIMEOUT = 20000

var Port string
var Rtimeout int64
var Wtimeout int64
var ShutTimeout int64

func CfgCheck() error {
	log.Logger.Info("========= print config file =========")
	for _, key := range viper.AllKeys() {
		log.Logger.Info(zap.Field(key, viper.Get(key)))
	}
	log.Logger.Info("=========       end         =========")
	Port = viper.GetString(PORT)
	Rtimeout = viper.GetInt64(READ_TIMEOUT)
	Wtimeout = viper.GetInt64(WRITE_TIMEOUT)
	ShutTimeout = viper.GetInt64(SHUTDOWN_TIMEOUT)
	if Rtimeout == 0 {
		Rtimeout = TIMEOUT
	}
	if Wtimeout == 0 {
		Wtimeout = TIMEOUT
	}
	if ShutTimeout == 0 {
		ShutTimeout = TIMEOUT * 2
	}
	return nil
}

func isEmpty(keys ...string) (result bool) {
	result = false
	for _, key := range keys {
		if key == "" {
			result = true
		}
	}
	return result
}

package cfg

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/yixy/tiny-photograph/common/env"
	"github.com/yixy/tiny-photograph/internal/log"
	"go.uber.org/zap"
)

const (
	//viper load key is lower case
	PORT             = "port"
	READ_TIMEOUT     = "read_timeout"
	WRITE_TIMEOUT    = "write_timeout"
	SHUTDOWN_TIMEOUT = "shutdown_timeout"
	FILE_TYPE        = "file_type"
)

const TIMEOUT = 20000

func init() {
	conFileStr := fmt.Sprintf("%s/conf/config.yml", env.Workdir)
	fmt.Printf("config file is :%s\n", conFileStr)
	viper.SetConfigFile(conFileStr)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func CfgCheck() error {
	log.Logger.Info("========= print config file =========")
	for _, key := range viper.AllKeys() {
		log.Logger.Info("config item", zap.Any(key, viper.Get(key)))
	}
	log.Logger.Info("=========       end         =========")
	env.Port = viper.GetString(PORT)
	env.Rtimeout = viper.GetInt64(READ_TIMEOUT)
	env.Wtimeout = viper.GetInt64(WRITE_TIMEOUT)
	env.ShutTimeout = viper.GetInt64(SHUTDOWN_TIMEOUT)
	if env.Rtimeout == 0 {
		env.Rtimeout = TIMEOUT
	}
	if env.Wtimeout == 0 {
		env.Wtimeout = TIMEOUT
	}
	if env.ShutTimeout == 0 {
		env.ShutTimeout = TIMEOUT * 2
	}
	s := viper.GetString(FILE_TYPE)
	env.FileTypeList = strings.Split(s, ",")
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

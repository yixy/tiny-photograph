package env

import (
	"math/rand"
	"time"

	"github.com/yixy/golang-util/path"
	"github.com/yixy/golang-util/random"
)

const AppName = "tiny-photograph"

var (
	Version      string
	Date         string
	BuildEnv     string
	Workdir      string
	Secret       []byte
	Port         string
	Rtimeout     int64
	Wtimeout     int64
	ShutTimeout  int64
	FileTypeList []string
)

func init() {

	var err error
	Workdir, err = path.GetProgramPath()
	if err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UnixNano())
	SecretStr := random.RandomString(32)
	Secret = []byte(SecretStr)
}

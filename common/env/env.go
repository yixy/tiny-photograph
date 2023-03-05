package env

import (
	"github.com/yixy/golang-util/path"
)

const AppName = "tiny-photograph"

var Workdir string

func init() {

	var err error
	Workdir, err = path.GetProgramPath()
	if err != nil {
		panic(err)
	}

}

/*
Copyright © 2023 yixy <youzhilane01@gmail.com>
*/
package main

import (
	"github.com/yixy/tiny-photograph/cmd"
	"github.com/yixy/tiny-photograph/common/cfg"
)

func init() {
	err := cfg.CfgCheck()
	if err != nil {
		panic(err)
	}
}

func main() {
	cmd.Execute()
}

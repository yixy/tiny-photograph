package internal

import (
	"github.com/yixy/tiny-photograph/common/env"
)

func IsTypeMatched(fileType string) (result bool) {
	result = false
	if fileType == "" {
		return
	}
	for _, v := range env.FileTypeList {
		if fileType == v {
			result = true
			break
		}
	}
	return
}

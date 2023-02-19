package internal

func IsTypeMatched(fileType string) (result bool) {
	result = false
	if fileType == "jpg" {
		result = true
	} else if fileType == "cr2" {
		result = true
	}
	return result
}

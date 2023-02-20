package utils

import "strconv"

func BoolToFloat64(value bool) float64 {
	if value {
		return 1
	}

	return 0
}

func StringToFloat64(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func StringToInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

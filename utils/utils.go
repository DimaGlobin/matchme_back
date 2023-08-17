package utils

import "github.com/lib/pq"

func Contains(arr pq.Int64Array, target int64) bool {
	for _, val := range arr {
		if val == target {
			return true
		}
	}
	return false
}

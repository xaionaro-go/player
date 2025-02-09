package builtin

import "time"

func abs(in time.Duration) time.Duration {
	if in < 0 {
		return -in
	}
	return in
}

package lib

import (
	"fmt"
	"time"
)

func ParseTime(str string) (time.Duration, error) {
	var h, m, s int
	if _, err := fmt.Sscanf(str, "%d:%d:%d", &h, &m, &s); err != nil {
		return 0, fmt.Errorf("parse time: %w", err)
	}
	return time.Duration(h*3600+m*60+s) * time.Second, nil
}

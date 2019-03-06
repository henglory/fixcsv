package fixcsv

import (
	"fmt"
	"strconv"
	"strings"
)

func parseTag(tag string) (pos, length int, err error) {
	parts := strings.Split(tag, ":")
	if len(parts) != 2 {
		return pos, length, fmt.Errorf("miss format")
	}

	if pos, err = strconv.Atoi(parts[0]); err != nil {
		return pos, length, fmt.Errorf("position must be numberic")
	}
	if length, err = strconv.Atoi(parts[1]); err != nil {
		return pos, length, fmt.Errorf("length must be numberic")
	}

	return pos, length, nil
}

package request

import (
	"fmt"
	"strconv"
)

func ValidatePathID(pathValue string, fieldName string) (int32, error) {
	if pathValue == "" {
		return 0, fmt.Errorf("missing %s", fieldName)
	}

	id, err := strconv.ParseInt(pathValue, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", fieldName, err)
	}

	if id < 1 {
		return 0, fmt.Errorf("%s must be positive", fieldName)
	}

	return int32(id), nil
}

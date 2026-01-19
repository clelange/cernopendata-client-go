package updater

import (
	"fmt"
	"strings"
)

func CompareVersions(current, latest string) int {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	if current == "dev" || strings.HasPrefix(current, "dev") {
		return -1
	}
	if latest == "dev" || strings.HasPrefix(latest, "dev") {
		return 1
	}

	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for i := 0; i < maxLen; i++ {
		var currentNum, latestNum int

		if i < len(currentParts) {
			part := strings.Split(currentParts[i], "-")[0]
			_, _ = fmt.Sscanf(part, "%d", &currentNum)
		}
		if i < len(latestParts) {
			part := strings.Split(latestParts[i], "-")[0]
			_, _ = fmt.Sscanf(part, "%d", &latestNum)
		}

		if currentNum < latestNum {
			return -1
		}
		if currentNum > latestNum {
			return 1
		}
	}

	return 0
}

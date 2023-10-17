package util

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ParsePageIndices(indices string, totalPages int) ([]int, error) {
	resultMap := make(map[int]bool)

	parts := strings.Split(indices, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check if the part is a single page index
		if idx, err := strconv.Atoi(part); err == nil {
			// Validate the single page index
			if idx < 1 || idx > totalPages {
				return nil, fmt.Errorf("invalid page index: %d, from input: %s, max supported page: %d", idx, indices, totalPages)
			}

			resultMap[idx] = true
			continue
		}

		// Check if the part is a range of page indices
		rangeParts := strings.Split(part, "-")
		if len(rangeParts) == 2 {
			start := 1
			if rangeParts[0] != "" {
				var err error
				start, err = strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid start page range: %s", part)
				}
				if start < 1 || start > totalPages {
					return nil, fmt.Errorf("invalid start page range: %s", part)
				}
			}

			end := totalPages
			if rangeParts[1] != "" {
				var err error
				end, err = strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				if err != nil {
					return nil, fmt.Errorf("invalid end page range: %s", part)
				}
				if end < 1 || end > totalPages {
					return nil, fmt.Errorf("invalid end page range: %s", part)
				}
			}

			// Validate the range of page indices
			if start > end {
				return nil, fmt.Errorf("invalid page range: %s, start index is after ending index", part)
			}

			for i := start; i <= end; i++ {
				resultMap[i] = true
			}

			continue
		}

		return nil, fmt.Errorf("invalid page index or range: %s", part)
	}

	var result []int
	for idx := range resultMap {
		result = append(result, idx)
	}

	sort.Ints(result)

	return result, nil
}

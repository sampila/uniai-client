package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func ParsePageRange(pageRange string) (pageNumbers []int, err error) {
	if pageRange == "" {
		return nil, nil
	}

	if strings.Contains(pageRange, "-") {
		parts := strings.Split(pageRange, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid page range format: %s", pageRange)
		}

		// Parse the start page
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid start page: %s", parts[0])
		}

		// Parse the end page if provided
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid end page: %s", parts[1])
		}

		for i := start; i <= end; i++ {
			pageNumbers = append(pageNumbers, i)
		}
	}

	// Split the range by ',' to handle multiple ranges
	if strings.Contains(pageRange, ",") {
		ranges := strings.Split(pageRange, ",")

		pageNumbers = make([]int, 0, len(ranges))
		for i, r := range ranges {
			pageNumbers[i], err = strconv.Atoi(r)
			if err != nil {
				return nil, fmt.Errorf("invalid page number: %s", r)
			}
		}
	}

	return pageNumbers, nil
}

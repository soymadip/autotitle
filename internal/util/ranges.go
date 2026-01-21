package util

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ParseRanges parses "1-3, 5, 7-9" into a sorted slice of integers
func ParseRanges(s string) ([]int, error) {
	var results []int
	parts := strings.Split(s, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))

			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("invalid numbers in range: %s", part)
			}

			// Normalize reversed ranges
			if start > end {
				start, end = end, start
			}

			for i := start; i <= end; i++ {
				results = append(results, i)
			}
		} else {
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}
			results = append(results, num)
		}
	}

	sort.Ints(results)
	results = removeDuplicates(results)

	return results, nil
}

// removeDuplicates removes duplicate integers from a sorted slice
func removeDuplicates(nums []int) []int {
	if len(nums) == 0 {
		return nums
	}

	var result []int
	result = append(result, nums[0])

	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[i-1] {
			result = append(result, nums[i])
		}
	}

	return result
}

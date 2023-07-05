package main

import (
	"fmt"
	"testing"
)

func TestParsePageIndices(t *testing.T) {
	testCases := []struct {
		input          string
		totalPages     int
		expectedOutput []int
		expectedError  error
	}{
		{
			input:          "8",
			totalPages:     10,
			expectedOutput: []int{8},
			expectedError:  nil,
		},
		{
			input:          "0",
			totalPages:     10,
			expectedOutput: []int{},
			expectedError:  fmt.Errorf("invalid page index: 0, from input: 0, max supported page: 10"),
		},
		{
			input:          "1,2",
			totalPages:     10,
			expectedOutput: []int{1, 2},
			expectedError:  nil,
		},
		{
			input:          "1,2,5",
			totalPages:     10,
			expectedOutput: []int{1, 2, 5},
			expectedError:  nil,
		},
		{
			input:          "1-3,6-8",
			totalPages:     10,
			expectedOutput: []int{1, 2, 3, 6, 7, 8},
			expectedError:  nil,
		},
		{
			input:          "4,6-",
			totalPages:     10,
			expectedOutput: []int{4, 6, 7, 8, 9, 10}, // Assuming there are 10 pages in total
			expectedError:  nil,
		},
		{
			input:          "4,1-6",
			totalPages:     10,
			expectedOutput: []int{1, 2, 3, 4, 5, 6}, // Assuming there are 10 pages in total
			expectedError:  nil,
		},
		{
			input:          "1,2,5-3",
			totalPages:     10,
			expectedOutput: nil,
			expectedError:  fmt.Errorf("invalid page range: 5-3, start index is after ending index"),
		},
		{
			input:          "a,b,c",
			totalPages:     10,
			expectedOutput: nil,
			expectedError:  fmt.Errorf("invalid page index or range: a"),
		},
	}

	for _, tc := range testCases {
		output, err := parsePageIndices(tc.input, tc.totalPages)

		// Check if the returned error matches the expected error
		if (err == nil && tc.expectedError != nil) || (err != nil && tc.expectedError == nil) || (err != nil && tc.expectedError != nil && err.Error() != tc.expectedError.Error()) {
			t.Errorf("Error mismatch for input '%s'. Expected: '%v', Got: '%v'", tc.input, tc.expectedError, err)
		}

		// Check if the returned output matches the expected output
		if len(output) != len(tc.expectedOutput) {
			t.Errorf("Output length mismatch for input '%s'. Expected: %v, Got: %v", tc.input, tc.expectedOutput, output)
		} else {
			for i := range output {
				if output[i] != tc.expectedOutput[i] {
					t.Errorf("Output value mismatch at index %d for input '%s'. Expected: %v, Got: %v", i, tc.input, tc.expectedOutput, output)
				}
			}
		}
	}
}

package services

import (
	"testing"
)

func TestFormatChangeDetails(t *testing.T) {
	service := &ChangeLogService{}

	tests := []struct {
		name     string
		changes  []string
		expected string
	}{
		{
			name:     "No changes",
			changes:  []string{},
			expected: "",
		},
		{
			name:     "Single change",
			changes:  []string{"changed title to 'New Title'"},
			expected: "changed title to 'New Title'",
		},
		{
			name:     "Two changes",
			changes:  []string{"changed title to 'New Title'", "updated description"},
			expected: "changed title to 'New Title' and updated description",
		},
		{
			name:     "Three changes",
			changes:  []string{"changed title to 'New Title'", "updated description", "changed status to 'Done'"},
			expected: "changed title to 'New Title', updated description, and changed status to 'Done'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.FormatChangeDetails(tt.changes)
			if result != tt.expected {
				t.Errorf("FormatChangeDetails() = %v, want %v", result, tt.expected)
			}
		})
	}
}

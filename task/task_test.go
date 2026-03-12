package task

import "testing"

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		states []string
		state  string
		want   bool
	}{
		{
			name:   "item present",
			states: []string{"Pending", "Running", "Completed"},
			state:  "Running",
			want:   true,
		},
		{
			name:   "item absent",
			states: []string{"Pending", "Running", "Completed"},
			state:  "Failed",
			want:   false,
		},
		{
			name:   "empty slice",
			states: []string{},
			state:  "Running",
			want:   false,
		},
		{
			name:   "nil slice",
			states: nil,
			state:  "Running",
			want:   false,
		},
		{
			name:   "empty string present",
			states: []string{"Pending", "", "Completed"},
			state:  "",
			want:   true,
		},
		{
			name:   "empty string absent",
			states: []string{"Pending", "Completed"},
			state:  "",
			want:   false,
		},
		{
			name:   "case sensitive mismatch",
			states: []string{"Pending", "Running"},
			state:  "running",
			want:   false,
		},
		{
			name:   "multiple items present",
			states: []string{"Running", "Running"},
			state:  "Running",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.states, tt.state); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

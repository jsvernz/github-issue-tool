package cli

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags_LabelOnly(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantLabelOnly bool
	}{
		{
			name:       "label-only flag set",
			args:       []string{"--file", "test.txt", "--label-only"},
			wantErr:    false,
			wantLabelOnly: true,
		},
		{
			name:       "label-only flag not set",
			args:       []string{"--file", "test.txt"},
			wantErr:    false,
			wantLabelOnly: false,
		},
		{
			name:       "label-only with dry-run",
			args:       []string{"--file", "test.txt", "--label-only", "--dry-run"},
			wantErr:    false,
			wantLabelOnly: true,
		},
		{
			name:       "missing required file flag",
			args:       []string{"--label-only"},
			wantErr:    true,
			wantLabelOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			
			// Simulate command line arguments
			os.Args = append([]string{"cmd"}, tt.args...)
			
			opts, err := ParseFlags()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err == nil && opts.LabelOnly != tt.wantLabelOnly {
				t.Errorf("ParseFlags() LabelOnly = %v, want %v", opts.LabelOnly, tt.wantLabelOnly)
			}
		})
	}
}
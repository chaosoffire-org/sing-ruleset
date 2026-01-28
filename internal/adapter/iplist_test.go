package adapter_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sing-ruleset/internal/adapter"
	"testing"
)

func TestIPListProcessor_Process(t *testing.T) {
	processor := adapter.NewIPListProcessor()
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		expectedIPs int
		ctx         context.Context
	}{
		{
			name: "Valid IPs",
			input: `
192.168.1.0/24
10.0.0.1
# Comment
  
2001:db8::/32`,
			wantErr:     false,
			expectedIPs: 3,
			ctx:         context.Background(),
		},
		{
			name: "Invalid IPs",
			input: `
invalid-ip
192.168.1.1.1
`,
			wantErr:     true, // Code returns error if no valid IPs found
			expectedIPs: 0,
			ctx:         context.Background(),
		},
		{
			name:        "Empty File",
			input:       "",
			wantErr:     true,
			expectedIPs: 0,
			ctx:         context.Background(),
		},
		{
			name: "Context Cancelled",
			input: `
192.168.1.1
`,
			wantErr:     true,
			expectedIPs: 0,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFile := filepath.Join(tmpDir, "input.txt")
			outputFile := filepath.Join(tmpDir, "output.json")

			err := os.WriteFile(inputFile, []byte(tt.input), 0644)
			if err != nil {
				t.Fatal(err)
			}

			err = processor.Process(tt.ctx, inputFile, outputFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify output
				data, err := os.ReadFile(outputFile)
				if err != nil {
					t.Fatal(err)
				}

				if len(data) == 0 {
					// Empty output
					if tt.expectedIPs > 0 {
						t.Errorf("Expected output, got empty file")
					}

					return
				}

				var result struct {
					Rules []struct {
						IPCidr []string `json:"ip_cidr"`
					} `json:"rules"`
				}

				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatal(err)
				}

				// Handle empty rules case
				actualIPs := 0
				if len(result.Rules) > 0 {
					actualIPs = len(result.Rules[0].IPCidr)
				}

				if actualIPs != tt.expectedIPs {
					t.Errorf("Expected %d IPs, got %d", tt.expectedIPs, actualIPs)
				}
			}
		})
	}
}

package imagio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParam(t *testing.T) {
	uri := "/show/s120/bdouymx4a7ro.jpg"
	p, err := ParseFromPath(uri)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	t.Logf("param: %+v", p)
	assert.Equal(t, "bdouymx4a7ro.jpg", p.Name)
	assert.Equal(t, "bd/ou/ymx4a7ro.jpg", p.Path)
	assert.Equal(t, 120, int(p.Width))
	assert.Equal(t, 120, int(p.Height))
	assert.Equal(t, "show", p.Roof)
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMode   string
		wantWidth  uint
		wantHeight uint
		wantErr    bool
	}{
		{
			name:       "square size",
			input:      "s100",
			wantMode:   "s",
			wantWidth:  100,
			wantHeight: 100,
			wantErr:    false,
		},
		{
			name:       "width and height",
			input:      "s800x600",
			wantMode:   "s",
			wantWidth:  800,
			wantHeight: 600,
			wantErr:    false,
		},
		{
			name:       "height only",
			input:      "h500",
			wantMode:   "h",
			wantWidth:  500,
			wantHeight: 500,
			wantErr:    false,
		},
		{
			name:       "width only",
			input:      "w600",
			wantMode:   "w",
			wantWidth:  600,
			wantHeight: 600,
			wantErr:    false,
		},
		{
			name:    "invalid format",
			input:   "x100",
			wantErr: true,
		},
		{
			name:    "invalid size",
			input:   "w10",
			wantErr: true,
		},
		{
			name:    "too large size",
			input:   "w10000",
			wantErr: true,
		},
		{
			name:    "too short size",
			input:   "w10",
			wantErr: true,
		},
		{
			name:    "too short string",
			input:   "c",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, width, height, err := ParseSize(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseSize(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSize(%q) unexpected error: %v", tt.input, err)
				return
			}

			if mode != tt.wantMode {
				t.Errorf("mode = %q, want %q", mode, tt.wantMode)
			}

			if width != tt.wantWidth {
				t.Errorf("width = %d, want %d", width, tt.wantWidth)
			}

			if height != tt.wantHeight {
				t.Errorf("height = %d, want %d", height, tt.wantHeight)
			}
		})
	}
}

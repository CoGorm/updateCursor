package version

import (
	"testing"
)

func TestSemverFromName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "valid cursor filename",
			filename: "Cursor-1.0.0-x86_64.AppImage",
			expected: "1.0.0",
		},
		{
			name:     "valid cursor filename with patch version",
			filename: "Cursor-1.2.3-x86_64.AppImage",
			expected: "1.2.3",
		},
		{
			name:     "invalid filename format",
			filename: "invalid-name.AppImage",
			expected: "",
		},
		{
			name:     "empty filename",
			filename: "",
			expected: "",
		},
		{
			name:     "cursor without version",
			filename: "Cursor-x86_64.AppImage",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SemverFromName(tt.filename)
			if result != tt.expected {
				t.Errorf("SemverFromName(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestVersionLessThan(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{
			name:     "v1 < v2",
			v1:       "1.0.0",
			v2:       "1.1.0",
			expected: true,
		},
		{
			name:     "v1 = v2",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: false,
		},
		{
			name:     "v1 > v2",
			v1:       "1.1.0",
			v2:       "1.0.0",
			expected: false,
		},
		{
			name:     "patch version comparison",
			v1:       "1.0.0",
			v2:       "1.0.1",
			expected: true,
		},
		{
			name:     "minor version comparison",
			v1:       "1.0.0",
			v2:       "1.1.0",
			expected: true,
		},
		{
			name:     "major version comparison",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: true,
		},
		{
			name:     "empty version handling",
			v1:       "",
			v2:       "1.0.0",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LessThan(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("LessThan(%q, %q) = %v, want %v", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectError bool
		major       int
		minor       int
		patch       int
	}{
		{
			name:        "valid semver",
			version:     "1.2.3",
			expectError: false,
			major:       1,
			minor:       2,
			patch:       3,
		},
		{
			name:        "valid semver with single digits",
			version:     "0.1.0",
			expectError: false,
			major:       0,
			minor:       1,
			patch:       0,
		},
		{
			name:        "invalid semver format",
			version:     "1.2",
			expectError: true,
		},
		{
			name:        "empty version",
			version:     "",
			expectError: true,
		},
		{
			name:        "non-numeric version",
			version:     "a.b.c",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, err := ParseSemver(tt.version)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseSemver(%q) expected error but got none", tt.version)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSemver(%q) unexpected error: %v", tt.version, err)
				return
			}

			if major != tt.major || minor != tt.minor || patch != tt.patch {
				t.Errorf("ParseSemver(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.version, major, minor, patch, tt.major, tt.minor, tt.patch)
			}
		})
	}
}

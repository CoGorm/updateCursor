package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemverFromName extracts the semantic version from a Cursor filename
// Example: "Cursor-1.0.0-x86_64.AppImage" -> "1.0.0"
func SemverFromName(filename string) string {
	if filename == "" {
		return ""
	}

	// Regex to match Cursor-<version>-x86_64.AppImage pattern
	re := regexp.MustCompile(`^Cursor-([0-9]+(\.[0-9]+)*)-x86_64\.AppImage$`)
	matches := re.FindStringSubmatch(filename)

	if len(matches) < 2 {
		return ""
	}

	return matches[1]
}

// LessThan compares two semantic versions and returns true if v1 < v2
func LessThan(v1, v2 string) bool {
	if v1 == "" || v2 == "" {
		return false
	}

	// Parse both versions
	major1, minor1, patch1, err1 := ParseSemver(v1)
	if err1 != nil {
		return false
	}

	major2, minor2, patch2, err2 := ParseSemver(v2)
	if err2 != nil {
		return false
	}

	// Compare major version
	if major1 < major2 {
		return true
	}
	if major1 > major2 {
		return false
	}

	// Compare minor version
	if minor1 < minor2 {
		return true
	}
	if minor1 > minor2 {
		return false
	}

	// Compare patch version
	return patch1 < patch2
}

// ParseSemver parses a semantic version string and returns major, minor, patch components
func ParseSemver(version string) (major, minor, patch int, err error) {
	if version == "" {
		return 0, 0, 0, fmt.Errorf("empty version string")
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver format: expected 3 parts, got %d", len(parts))
	}

	// Parse major version
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %v", err)
	}

	// Parse minor version
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %v", err)
	}

	// Parse patch version
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %v", err)
	}

	return major, minor, patch, nil
}

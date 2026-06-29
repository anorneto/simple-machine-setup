package app

import "runtime"

// DetectOS returns the current OS as reported by Go's runtime
// (one of: "linux", "darwin", "windows", ...).
func DetectOS() string {
	return runtime.GOOS
}

// GetOSConfigFile returns the platform-specific manifest filename
// (e.g. "linux.toml" on Linux, "darwin.toml" on macOS).
func GetOSConfigFile() string {
	return DetectOS() + ".toml"
}

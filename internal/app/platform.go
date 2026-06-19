package app

import "runtime"

// Detect returns the current OS as reported by Go's runtime
// (one of: "linux", "darwin", "windows", "freebsd", ...).
func Detect() string {
	return runtime.GOOS
}

// ConfigFileName returns the platform-specific manifest filename
// (e.g. "linux.toml" on Linux, "darwin.toml" on macOS).
func ConfigFileName() string {
	return Detect() + ".toml"
}

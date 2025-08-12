package deploy

import (
	"runtime"
	"strings"
)

// translateLocalhostForDocker translates localhost URLs to host.docker.internal
// for Docker containers on macOS. This is necessary because containers on macOS
// cannot reach services on localhost directly.
func translateLocalhostForDocker(url string) string {
	// Only apply translation on macOS
	if runtime.GOOS != "darwin" {
		return url
	}

	// Simple string replacements for common localhost patterns
	result := url
	
	// Replace http://localhost with http://host.docker.internal
	result = strings.ReplaceAll(result, "http://localhost", "http://host.docker.internal")
	result = strings.ReplaceAll(result, "https://localhost", "https://host.docker.internal")
	result = strings.ReplaceAll(result, "ws://localhost", "ws://host.docker.internal")
	result = strings.ReplaceAll(result, "wss://localhost", "wss://host.docker.internal")
	
	// Replace http://127.0.0.1 with http://host.docker.internal
	result = strings.ReplaceAll(result, "http://127.0.0.1", "http://host.docker.internal")
	result = strings.ReplaceAll(result, "https://127.0.0.1", "https://host.docker.internal")
	result = strings.ReplaceAll(result, "ws://127.0.0.1", "ws://host.docker.internal")
	result = strings.ReplaceAll(result, "wss://127.0.0.1", "wss://host.docker.internal")
	
	return result
}
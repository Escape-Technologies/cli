package private

import "os"

func sshTarget() string {
	if target := os.Getenv("ESCAPE_PRIVATE_LOCATION_URL"); target != "" {
		return target
	}
	return "private-location.escape.tech:2222"
}

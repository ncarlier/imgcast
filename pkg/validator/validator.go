package validator

import "regexp"

var roomNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// IsValidRoomName checks if a room name is valid
// Room names must be alphanumeric with dash and underscore allowed
func IsValidRoomName(name string) bool {
	return roomNameRegex.MatchString(name)
}

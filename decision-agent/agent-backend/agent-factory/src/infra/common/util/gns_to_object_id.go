package util

import "strings"

// GNS2ObjectID gns 转objectID
func GNS2ObjectID(gns string) string {
	idSlice := strings.Split(gns, "/")
	return idSlice[len(idSlice)-1]
}

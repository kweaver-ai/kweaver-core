package cconstant

import "slices"

const (
	PmsAllFlag = "*"
)

func IsContainsPmsAllFlag(s []string) (isPmsAllFlag bool) {
	return slices.Contains(s, PmsAllFlag)
}

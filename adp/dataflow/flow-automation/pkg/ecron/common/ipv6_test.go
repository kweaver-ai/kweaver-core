package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIpv6(t *testing.T) {
	ipv4 := "10.10.2.1"
	ipv6_1 := "[0:0:0:0:10:10:2:1]"
	ipv6_2 := "0:0:0:0:10:10:2:1"
	ipv6_3 := "0:0:0:0:10:10:2:1]"
	ipv6_4 := "[0:0:0:0:10:10:2:1"

	rIpv4 := ParseHost(ipv4)
	rIpv6_1 := ParseHost(ipv6_1)
	rIpv6_2 := ParseHost(ipv6_2)
	rIpv6_3 := ParseHost(ipv6_3)
	rIpv6_4 := ParseHost(ipv6_4)

	assert.Equal(t, rIpv4, "10.10.2.1")
	assert.Equal(t, rIpv6_1, "[0:0:0:0:10:10:2:1]")
	assert.Equal(t, rIpv6_2, "[0:0:0:0:10:10:2:1]")
	assert.Equal(t, rIpv6_3, "[0:0:0:0:10:10:2:1]")
	assert.Equal(t, rIpv6_4, "[0:0:0:0:10:10:2:1]")
}

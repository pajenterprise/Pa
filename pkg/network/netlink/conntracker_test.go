// +build linux
// +build !android

package netlink

import (
	"crypto/rand"
	"net"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/network"
	"github.com/DataDog/datadog-agent/pkg/process/util"
	ct "github.com/florianl/go-conntrack"
	"github.com/stretchr/testify/assert"
)

func TestIsNat(t *testing.T) {
	src := net.ParseIP("1.1.1.1")
	dst := net.ParseIP("2.2.2..2")
	tdst := net.ParseIP("3.3.3.3")
	var srcPort uint16 = 42
	var dstPort uint16 = 8080

	t.Run("not nat", func(t *testing.T) {

		c := Con{
			ct.Con{
				Origin: &ct.IPTuple{
					Src: &src,
					Dst: &dst,
					Proto: &ct.ProtoTuple{
						SrcPort: &srcPort,
						DstPort: &dstPort,
					},
				},
				Reply: &ct.IPTuple{
					Src: &dst,
					Dst: &src,
					Proto: &ct.ProtoTuple{
						SrcPort: &dstPort,
						DstPort: &srcPort,
					},
				},
			},
			0,
		}
		assert.False(t, IsNAT(c))
	})

	t.Run("nil proto field", func(t *testing.T) {
		c := Con{
			ct.Con{
				Origin: &ct.IPTuple{
					Src: &src,
					Dst: &dst,
				},
				Reply: &ct.IPTuple{
					Src: &dst,
					Dst: &src,
				},
			},
			0,
		}
		assert.False(t, IsNAT(c))
	})

	t.Run("nat", func(t *testing.T) {

		c := Con{
			ct.Con{
				Origin: &ct.IPTuple{
					Src: &src,
					Dst: &dst,
					Proto: &ct.ProtoTuple{
						SrcPort: &srcPort,
						DstPort: &dstPort,
					},
				},
				Reply: &ct.IPTuple{
					Src: &tdst,
					Dst: &src,
					Proto: &ct.ProtoTuple{
						SrcPort: &dstPort,
						DstPort: &srcPort,
					},
				},
			},
			0,
		}
		assert.True(t, IsNAT(c))
	})
}

func TestRegisterNonNat(t *testing.T) {
	rt := newConntracker()
	c := makeUntranslatedConn(net.ParseIP("10.0.0.0"), net.ParseIP("50.30.40.10"), 6, 8080, 12345)

	rt.register(c)
	translation := rt.GetTranslationForConn(
		network.ConnectionStats{
			Source: util.AddressFromString("10.0.0.0"),
			SPort:  8080,
			Dest:   util.AddressFromString("50.30.40.10"),
			DPort:  12345,
			Type:   network.TCP,
		},
	)
	assert.Nil(t, translation)
}

func TestRegisterNat(t *testing.T) {
	rt := newConntracker()
	c := makeTranslatedConn(net.ParseIP("10.0.0.0"), net.ParseIP("20.0.0.0"), net.ParseIP("50.30.40.10"), 6, 12345, 80, 80)

	rt.register(c)
	translation := rt.GetTranslationForConn(
		network.ConnectionStats{
			Source: util.AddressFromString("10.0.0.0"),
			SPort:  12345,
			Dest:   util.AddressFromString("50.30.40.10"),
			DPort:  80,
			Type:   network.TCP,
		},
	)
	assert.NotNil(t, translation)
	assert.Equal(t, &network.IPTranslation{
		ReplSrcIP:   util.AddressFromString("20.0.0.0"),
		ReplDstIP:   util.AddressFromString("10.0.0.0"),
		ReplSrcPort: 80,
		ReplDstPort: 12345,
	}, translation)

	udpTranslation := rt.GetTranslationForConn(
		network.ConnectionStats{
			Source: util.AddressFromString("10.0.0.0"),
			SPort:  12345,
			Dest:   util.AddressFromString("50.30.40.10"),
			DPort:  80,
			Type:   network.UDP,
		},
	)
	assert.Nil(t, udpTranslation)

}

func TestRegisterNatUDP(t *testing.T) {
	rt := newConntracker()
	c := makeTranslatedConn(net.ParseIP("10.0.0.0"), net.ParseIP("20.0.0.0"), net.ParseIP("50.30.40.10"), 17, 12345, 80, 80)

	rt.register(c)
	translation := rt.GetTranslationForConn(
		network.ConnectionStats{
			Source: util.AddressFromString("10.0.0.0"),
			SPort:  12345,
			Dest:   util.AddressFromString("50.30.40.10"),
			DPort:  80,
			Type:   network.UDP,
		},
	)
	assert.NotNil(t, translation)
	assert.Equal(t, &network.IPTranslation{
		ReplSrcIP:   util.AddressFromString("20.0.0.0"),
		ReplDstIP:   util.AddressFromString("10.0.0.0"),
		ReplSrcPort: 80,
		ReplDstPort: 12345,
	}, translation)

	translation = rt.GetTranslationForConn(
		network.ConnectionStats{
			Source: util.AddressFromString("10.0.0.0"),
			SPort:  12345,
			Dest:   util.AddressFromString("50.30.40.10"),
			DPort:  80,
			Type:   network.TCP,
		},
	)
	assert.Nil(t, translation)
}

func TestTooManyEntries(t *testing.T) {
	rt := newConntracker()
	rt.maxStateSize = 1

	rt.register(makeTranslatedConn(net.ParseIP("10.0.0.0"), net.ParseIP("20.0.0.0"), net.ParseIP("50.30.40.10"), 6, 12345, 80, 80))
	rt.register(makeTranslatedConn(net.ParseIP("10.0.0.1"), net.ParseIP("20.0.0.1"), net.ParseIP("50.30.40.10"), 6, 12345, 80, 80))
	rt.register(makeTranslatedConn(net.ParseIP("10.0.0.2"), net.ParseIP("20.0.0.2"), net.ParseIP("50.30.40.10"), 6, 12345, 80, 80))
}

// Run this test with -memprofile to get an insight of how much memory is
// allocated/used by Conntracker to store maxStateSize entries.
// Example: go test -run TestConntrackerMemoryAllocation -memprofile mem.prof .
func TestConntrackerMemoryAllocation(t *testing.T) {
	rt := newConntracker()
	ipGen := randomIPGen()

	for i := 0; i < rt.maxStateSize; i++ {
		c := makeTranslatedConn(ipGen(), ipGen(), ipGen(), 6, 12345, 80, 80)
		rt.register(c)
	}
}

func newConntracker() *realConntracker {
	return &realConntracker{
		state:                make(map[connKey]*network.IPTranslation),
		maxStateSize:         10000,
		exceededSizeLogLimit: util.NewLogLimit(1, time.Minute),
	}
}

func makeUntranslatedConn(src, dst net.IP, proto uint8, srcPort, dstPort uint16) Con {
	return makeTranslatedConn(src, dst, dst, proto, srcPort, dstPort, dstPort)
}

// makes a translation where from -> to is shows as transFrom -> from
func makeTranslatedConn(from, transFrom, to net.IP, proto uint8, fromPort, transFromPort, toPort uint16) Con {

	return Con{
		ct.Con{
			Origin: &ct.IPTuple{
				Src: &from,
				Dst: &to,
				Proto: &ct.ProtoTuple{
					Number:  &proto,
					SrcPort: &fromPort,
					DstPort: &toPort,
				},
			},
			Reply: &ct.IPTuple{
				Src: &transFrom,
				Dst: &from,
				Proto: &ct.ProtoTuple{
					Number:  &proto,
					SrcPort: &transFromPort,
					DstPort: &fromPort,
				},
			},
		},
		0,
	}
}

func randomIPGen() func() net.IP {
	b := make([]byte, 4)
	return func() net.IP {
		for {
			if _, err := rand.Read(b); err != nil {
				continue
			}

			return net.IPv4(b[0], b[1], b[2], b[3])
		}
	}
}

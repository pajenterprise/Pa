package encoding

import (
	"github.com/DataDog/datadog-agent/pkg/ebpf"
	"github.com/DataDog/datadog-agent/pkg/ebpf/netlink"
	agent "github.com/DataDog/datadog-agent/pkg/process/model"
	"github.com/DataDog/datadog-agent/pkg/process/util"
)

// FormatConnection converts a ConnectionStats into an agent.Connection
func FormatConnection(conn ebpf.ConnectionStats) *agent.Connection {
	return &agent.Connection{
		Pid:                int32(conn.Pid),
		Laddr:              formatAddr(conn.Source, conn.SPort),
		Raddr:              formatAddr(conn.Dest, conn.DPort),
		Family:             formatFamily(conn.Family),
		Type:               formatType(conn.Type),
		TotalBytesSent:     conn.MonotonicSentBytes,
		TotalBytesReceived: conn.MonotonicRecvBytes,
		TotalRetransmits:   conn.MonotonicRetransmits,
		LastBytesSent:      conn.LastSentBytes,
		LastBytesReceived:  conn.LastRecvBytes,
		LastRetransmits:    conn.LastRetransmits,
		Direction:          agent.ConnectionDirection(conn.Direction),
		NetNS:              conn.NetNS,
		IpTranslation:      formatIPTranslation(conn.IPTranslation),
	}
}

func formatAddr(addr util.Address, port uint16) *agent.Addr {
	if addr == nil {
		return nil
	}

	return &agent.Addr{Ip: addr.String(), Port: int32(port)}
}

func formatFamily(f ebpf.ConnectionFamily) agent.ConnectionFamily {
	switch f {
	case ebpf.AFINET:
		return agent.ConnectionFamily_v4
	case ebpf.AFINET6:
		return agent.ConnectionFamily_v6
	default:
		return -1
	}
}

func formatType(f ebpf.ConnectionType) agent.ConnectionType {
	switch f {
	case ebpf.TCP:
		return agent.ConnectionType_tcp
	case ebpf.UDP:
		return agent.ConnectionType_udp
	default:
		return -1
	}
}

func formatDirection(d ebpf.ConnectionDirection) agent.ConnectionDirection {
	switch d {
	case ebpf.INCOMING:
		return agent.ConnectionDirection_incoming
	case ebpf.OUTGOING:
		return agent.ConnectionDirection_outgoing
	case ebpf.LOCAL:
		return agent.ConnectionDirection_local
	default:
		return agent.ConnectionDirection_unspecified
	}
}

func formatIPTranslation(ct *netlink.IPTranslation) *agent.IPTranslation {
	if ct == nil {
		return nil
	}

	return &agent.IPTranslation{
		ReplSrcIP:   ct.ReplSrcIP,
		ReplDstIP:   ct.ReplDstIP,
		ReplSrcPort: int32(ct.ReplSrcPort),
		ReplDstPort: int32(ct.ReplDstPort),
	}
}

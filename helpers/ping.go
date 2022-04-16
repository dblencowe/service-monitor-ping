package helpers

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	ProtocolICMP = 1
)

var listenAddr = "0.0.0.0"

func Ping(addr string) (*net.IPAddr, time.Duration, error) {
	conn, err := icmp.ListenPacket("ip4:icmp", listenAddr)
	if err != nil {
		return nil, 0, err
	}
	defer conn.Close()

	dest, err := net.ResolveIPAddr("ip4", addr)
	if err != nil {
		return nil, 0, err
	}

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte(""),
		},
	}
	b, err := msg.Marshal(nil)
	if err != nil {
		return dest, 0, err
	}

	start := time.Now()
	n, err := conn.WriteTo(b, dest)
	if err != nil {
		return dest, 0, err
	}

	if n != len(b) {
		return dest, 0, fmt.Errorf("go %v; want %v", n, len(b))
	}

	reply := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return dest, 0, err
	}
	n, peer, err := conn.ReadFrom(reply)
	if err != nil {
		return dest, 0, err
	}
	duration := time.Since(start)

	rm, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
	if err != nil {
		return dest, 0, err
	}

	if rm.Type == ipv4.ICMPTypeEchoReply {
		return dest, duration, nil
	}

	return dest, 0, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
}

package main

import (
	"encoding/base64"
	"encoding/hex"
	"net"
	"net/netip"
	"strconv"
	"time"
)

type ipT netip.Addr

func (o *ipT) UnmarshalFlag(value string) error {
	ip, err := netip.ParseAddr(value)
	*o = ipT(ip)
	return err
}

func (o ipT) String() string {
	return netip.Addr(o).String()
}

type hostPortT struct {
	host string
	port uint16
}

func (o *hostPortT) UnmarshalFlag(value string) error {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return err
	}
	port16, err := strconv.ParseUint(port, 10, 16)
	*o = hostPortT{host, uint16(port16)}
	return err
}

type keyT string

func (o *keyT) UnmarshalFlag(value string) error {
	key, err := base64.StdEncoding.DecodeString(value)
	*o = keyT(hex.EncodeToString(key))
	return err
}

type timeT int64

func (o *timeT) UnmarshalFlag(value string) error {
	i, err := strconv.ParseInt(value, 10, 32)
	if err == nil {
		*o = timeT(i)
		return nil
	}
	d, err := time.ParseDuration(value)
	*o = timeT(d.Seconds())
	return err
}

type options struct {
	ClientIPs  []ipT  `long:"client-ip" env:"CLIENT_IP" env-delim:"," required:"true" description:"[Interface].Address\tfor WireGuard client (can be set multiple times)"`
	ClientPort int    `long:"client-port" env:"CLIENT_PORT" description:"[Interface].ListenPort\tfor WireGuard client (optional)"`
	PrivateKey keyT   `long:"private-key" env:"PRIVATE_KEY" required:"true" description:"[Interface].PrivateKey\tfor WireGuard client (format: base64)"`
	DNS        string `long:"dns" env:"DNS" description:"[Interface].DNS\tfor WireGuard network (format: protocol://ip:port)\nProtocol includes udp(default), tcp, tls(DNS over TLS) and https(DNS over HTTPS)"`
	MTU        int    `long:"mtu" env:"MTU" default:"1280" description:"[Interface].MTU\tfor WireGuard network"`

	PeerEndpoint      hostPortT `long:"peer-endpoint" env:"PEER_ENDPOINT" required:"true" description:"[Peer].Endpoint\tfor WireGuard server (format: host:port)"`
	PeerKey           keyT      `long:"peer-key" env:"PEER_KEY" required:"true" description:"[Peer].PublicKey\tfor WireGuard server (format: base64)"`
	PresharedKey      keyT      `long:"preshared-key" env:"PRESHARED_KEY" description:"[Peer].PresharedKey\tfor WireGuard network (optional, format: base64)"`
	KeepaliveInterval timeT     `long:"keepalive-interval" env:"KEEPALIVE_INTERVAL" description:"[Peer].PersistentKeepalive\tfor WireGuard network (optional)"`

	ResolveDNS      string `long:"resolve-dns" env:"RESOLVE_DNS" description:"DNS for resolving WireGuard server address (optional, format: protocol://ip:port)\nProtocol includes udp(default), tcp, tls(DNS over TLS) and https(DNS over HTTPS)"`
	ResolveInterval timeT  `long:"resolve-interval" env:"RESOLVE_INTERVAL" default:"1m" description:"Interval for resolving WireGuard server address (set 0 to disable)"`

	Listen   string `long:"listen" env:"LISTEN" default:"localhost:8080" description:"HTTP & SOCKS5 server address"`
	ExitMode string `long:"exit-mode" env:"EXIT_MODE" choice:"remote" choice:"local" default:"remote" description:"Exit mode"`
	Verbose  bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	S1       uint   `long:"s1" env:"S1" default:"0" description:"[Interface].S1\t\tInit Packet Padding. Amnezia: 15-150 recommended (max 1280)"`
	S2       uint   `long:"s2" env:"S2" default:"0" description:"[Interface].S2\t\tResponse Packet Padding. Amnezia: 15-150 recommended (max 1280)"`
	H1       uint32 `long:"h1" env:"H1" default:"1" description:"[Interface].H1\t\tInit Packet ID. Amnezia: 5-4294967295 recommended"`
	H2       uint32 `long:"h2" env:"H2" default:"2" description:"[Interface].H2\t\tResponse Packet ID. Amnezia: 5-4294967295 recommended"`
	H3       uint32 `long:"h3" env:"H3" default:"3" description:"[Interface].H3\t\tCookie Packet ID. Amnezia: 5-4294967295 recommended"`
	H4       uint32 `long:"h4" env:"H4" default:"4" description:"[Interface].H4\t\tData Packet ID. Amnezia: 5-4294967295 recommended"`
	JC       uint8  `long:"junk-count" env:"JC" default:"0" description:"[Interface].JC\t\tNumber of junk packets to send before sending Init. More or larger Junk Packets will delay (re-)connections. (Compatible with Wireguard)"`
	Jmin     int    `long:"junk-min" env:"JMIN" default:"50" description:"[Interface].Jmin\tMin size of junk packets. (max 1279)"`
	Jmax     int    `long:"junk-max" env:"JMAX" default:"1000" description:"[Interface].Jmax\tMax size of junk packets. (max 1280)"`
	ClientID string `long:"client-id" env:"CLIENT_ID" hidden:"true"`
}

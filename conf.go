package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/bobpaul/awghttp/internal/resolver"
	"github.com/amnezia-vpn/amneziawg-go/device"
)

type peer struct {
	resolver *resolver.Resolver

	pubKey keyT
	psk    keyT

	host string
	ip   netip.Addr
	port uint16
}

func newPeerEndpoint() (*peer, error) {
	p := &peer{
		pubKey: opts.PeerKey,
		psk:    opts.PresharedKey,
		host:   opts.PeerEndpoint.host,
		port:   opts.PeerEndpoint.port,
	}
	var err error
	p.ip, err = netip.ParseAddr(p.host)
	if err == nil {
		return p, nil
	}

	p.resolver = resolver.New(
		opts.ResolveDNS,
		func(ctx context.Context, network, address string) (net.Conn, error) {
			netConn, err := (&net.Dialer{}).DialContext(ctx, network, address)
			logger.Verbosef("Using %s to resolve peer endpoint: %v", opts.ResolveDNS, err)
			return netConn, err
		},
	)

	p.ip, err = p.resolveHost()
	if err != nil {
		return nil, fmt.Errorf("resolve peer endpoint ip: %w", err)
	}

	return p, err
}

func (p *peer) initConf() string {
	conf := fmt.Sprintf("public_key=%s\n", p.pubKey)
	conf += fmt.Sprintf("endpoint=%s\n", netip.AddrPortFrom(p.ip, p.port))
	conf += "allowed_ip=0.0.0.0/0\n"
	conf += "allowed_ip=::/0\n"

	if opts.KeepaliveInterval > 0 {
		conf += fmt.Sprintf("persistent_keepalive_interval=%d\n", opts.KeepaliveInterval)
	}
	if p.psk != "" {
		conf += fmt.Sprintf("preshared_key=%s\n", p.psk)
	}

	return conf
}

func (p *peer) updateConf() (string, bool) {
	newIP, err := p.resolveHost()
	if err != nil {
		logger.Verbosef("Resolve peer endpoint: %v", err)
		return "", false
	}
	if p.ip == newIP {
		return "", false
	}
	p.ip = newIP
	logger.Verbosef("PeerEndpoint is changed to: %s", p.ip)

	conf := fmt.Sprintf("public_key=%s\n", p.pubKey)
	conf += "update_only=true\n"
	conf += fmt.Sprintf("endpoint=%s\n", netip.AddrPortFrom(p.ip, p.port))
	return conf, true
}

func (p *peer) resolveHost() (netip.Addr, error) {
	ips, err := p.resolver.LookupNetIP(context.Background(), "ip", p.host)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("resolve ip for %s: %w", p.host, err)
	}
	for _, ip := range ips {
		// netstack doesn't seem to understand IPv4-mapped IPv6 addresses.
		ip = ip.Unmap()
		conn, err := net.DialUDP("udp", nil, net.UDPAddrFromAddrPort(netip.AddrPortFrom(ip, p.port)))
		if err == nil {
			conn.Close()
			return ip, nil
		} else {
			logger.Verbosef("Dial %s: %s", ip, err)
		}
	}
	return netip.Addr{}, fmt.Errorf("no available ip for %s", p.host)
}

func ipcSet(dev *device.Device) error {
	conf := fmt.Sprintf("private_key=%s\n", opts.PrivateKey)
	if opts.ClientPort != 0 {
		conf += fmt.Sprintf("listen_port=%d\n", opts.ClientPort)
	}

	if (opts.H1 == opts.H2) || (opts.H1 == opts.H3) || (opts.H1 == opts.H4) || 
		(opts.H2 == opts.H3) || (opts.H2 == opts.H4) || (opts.H3 == opts.H4) {
			logger.Errorf("H1, H2, H3, H4 must all be unique")
	}
	if (opts.H1 != 1) || (opts.H1 != 2) || (opts.H3 != 3) || (opts.H4 != 4) {
		logger.Verbosef("Using custom message headers for AmneziaWG")
	}
	conf += fmt.Sprintf("h1=%d\nh2=%d\nh3=%d\nh4=%d\n", opts.H1, opts.H2, opts.H3, opts.H4)
	
	if (opts.S1 != 0) || (opts.S2 != 0) {
		logger.Verbosef("Padding Initializion and Resposne messages for AmneziaWG")
	}
	if (opts.S1 + 56 == opts.S2) {
		logger.Errorf("S1 and S2 must be set so Initiator and Response packages have different data lengths!")
	}
	conf += fmt.Sprintf("s1=%d\ns2=%d\n", opts.S1, opts.S2)

	if (opts.Jmin > opts.Jmax) {
		logger.Errorf("Max junk packet size cannot be less than min junk packet size!")
	}
	conf += fmt.Sprintf("jc=%d\njmin=%d\njmax=%d\n", opts.JC, opts.Jmin, opts.Jmax)

	peer, err := newPeerEndpoint()
	if err != nil {
		return err
	}
	conf += peer.initConf()
	logger.Verbosef("Device config:\n%s", conf)

	if opts.H1 == 1 && opts.H2 == 2 && opts.H3 == 3 && opts.H4 == 4 && opts.S1 == 0 && opts.S2 == 0 {
		logger.Verbosef("Configuration compatible with wireguard")
	}

	if err := dev.IpcSet(conf); err != nil {
		return err
	}

	if peer.resolver != nil {
		go func() {
			c := time.Tick(time.Duration(opts.ResolveInterval) * time.Second)

			for range c {
				conf, needUpdate := peer.updateConf()
				if !needUpdate {
					continue
				}

				if err := dev.IpcSet(conf); err != nil {
					logger.Errorf("Config device: %v", err)
				}
			}
		}()
	}
	return nil
}

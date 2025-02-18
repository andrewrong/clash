package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"

	"github.com/phuslu/log"

	"github.com/Dreamacro/clash/component/dialer"
	C "github.com/Dreamacro/clash/constant"
)

type Base struct {
	name  string
	addr  string
	iface string
	tp    C.AdapterType
	udp   bool
	dns   bool
	rmark int
}

// Name implements C.ProxyAdapter
func (b *Base) Name() string {
	return b.name
}

// Type implements C.ProxyAdapter
func (b *Base) Type() C.AdapterType {
	return b.tp
}

// StreamConn implements C.ProxyAdapter
func (b *Base) StreamConn(c net.Conn, _ *C.Metadata) (net.Conn, error) {
	return c, errors.New("no support")
}

// StreamPacketConn implements C.ProxyAdapter
func (b *Base) StreamPacketConn(c net.Conn, _ *C.Metadata) (net.Conn, error) {
	return c, errors.New("no support")
}

// ListenPacketContext implements C.ProxyAdapter
func (b *Base) ListenPacketContext(_ context.Context, _ *C.Metadata, _ ...dialer.Option) (C.PacketConn, error) {
	return nil, errors.New("no support")
}

// SupportUDP implements C.ProxyAdapter
func (b *Base) SupportUDP() bool {
	return b.udp
}

// DisableDnsResolve implements C.DisableDnsResolve
func (b *Base) DisableDnsResolve() bool {
	return !b.dns
}

// MarshalJSON implements C.ProxyAdapter
func (b *Base) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type": b.Type().String(),
	})
}

// Addr implements C.ProxyAdapter
func (b *Base) Addr() string {
	return b.addr
}

// Unwrap implements C.ProxyAdapter
func (b *Base) Unwrap(_ *C.Metadata) C.Proxy {
	return nil
}

// Cleanup implements C.ProxyAdapter
func (b *Base) Cleanup() {}

// DialOptions return []dialer.Option from struct
func (b *Base) DialOptions(opts ...dialer.Option) []dialer.Option {
	if b.iface != "" {
		opts = append(opts, dialer.WithInterface(b.iface))
	}

	if b.rmark != 0 {
		opts = append(opts, dialer.WithRoutingMark(b.rmark))
	}

	return opts
}

type BasicOption struct {
	Interface   string `proxy:"interface-name,omitempty" group:"interface-name,omitempty"`
	RoutingMark int    `proxy:"routing-mark,omitempty" group:"routing-mark,omitempty"`
}

type BaseOption struct {
	Name        string
	Addr        string
	Type        C.AdapterType
	UDP         bool
	Interface   string
	RoutingMark int
}

func NewBase(opt BaseOption) *Base {
	return &Base{
		name:  opt.Name,
		addr:  opt.Addr,
		tp:    opt.Type,
		udp:   opt.UDP,
		iface: opt.Interface,
		rmark: opt.RoutingMark,
	}
}

type conn struct {
	net.Conn
	chain C.Chain
}

// Chains implements C.Connection
func (c *conn) Chains() C.Chain {
	return c.chain
}

// AppendToChains implements C.Connection
func (c *conn) AppendToChains(a C.ProxyAdapter) {
	c.chain = append(c.chain, a.Name())
}

// SetChains implements C.Connection
func (c *conn) SetChains(chains []string) {
	c.chain = chains
}

func (c *conn) MarshalObject(e *log.Entry) {
	e.Str("proxy", c.Chains().String())
}

func NewConn(c net.Conn, a C.ProxyAdapter) C.Conn {
	return &conn{c, []string{a.Name()}}
}

type packetConn struct {
	net.PacketConn
	chain C.Chain
}

// Chains implements C.Connection
func (c *packetConn) Chains() C.Chain {
	return c.chain
}

// AppendToChains implements C.Connection
func (c *packetConn) AppendToChains(a C.ProxyAdapter) {
	c.chain = append(c.chain, a.Name())
}

// SetChains implements C.Connection
func (c *packetConn) SetChains(chains []string) {
	c.chain = chains
}

func (c *packetConn) MarshalObject(e *log.Entry) {
	e.Str("proxy", c.Chains().String())
}

func NewPacketConn(pc net.PacketConn, a C.ProxyAdapter) C.PacketConn {
	return &packetConn{pc, []string{a.Name()}}
}

type wrapConn struct {
	net.PacketConn
}

func (*wrapConn) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (*wrapConn) Write([]byte) (int, error) {
	return 0, io.EOF
}

func (*wrapConn) RemoteAddr() net.Addr {
	return nil
}

func WrapConn(packetConn net.PacketConn) net.Conn {
	return &wrapConn{
		PacketConn: packetConn,
	}
}

func IsPacketConn(c net.Conn) bool {
	if _, ok := c.(net.PacketConn); !ok {
		return false
	}

	if ua, ok := c.LocalAddr().(*net.UnixAddr); ok {
		return ua.Net == "unixgram"
	}

	return true
}

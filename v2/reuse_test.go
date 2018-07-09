package reuseport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

func ResolveAddr(network, address string) (net.Addr, error) {
	switch network {
	default:
		return nil, net.UnknownNetworkError(network)
	case "ip", "ip4", "ip6":
		return net.ResolveIPAddr(network, address)
	case "tcp", "tcp4", "tcp6":
		return net.ResolveTCPAddr(network, address)
	case "udp", "udp4", "udp6":
		return net.ResolveUDPAddr(network, address)
	case "unix", "unixgram", "unixpacket":
		return net.ResolveUnixAddr(network, address)
	}
}

func Listen(network, addr string) (net.Listener, error) {
	l := net.ListenConfig{
		Control: Control,
	}
	return l.Listen(context.Background(), network, addr)
}
func ListenPacket(network, addr string) (net.PacketConn, error) {
	l := net.ListenConfig{
		Control: Control,
	}
	return l.ListenPacket(context.Background(), network, addr)
}

func Dial(network, laddr, raddr string) (net.Conn, error) {
	var d net.Dialer
	if laddr != "" {
		netladdr, err := ResolveAddr(network, laddr)
		if err != nil {
			return nil, err
		}
		d.LocalAddr = netladdr
	}
	d.Control = Control
	return d.Dial(network, raddr)
}

type readOnly struct {
	r io.Reader
}

func (r readOnly) Read(buf []byte) (n int, err error) {
	return r.r.Read(buf)
}

func echo(c net.Conn) {
	io.Copy(c, readOnly{c})
	c.Close()
}

func packetEcho(c net.PacketConn) {
	defer c.Close()
	buf := make([]byte, 65536)
	for {
		n, addr, err := c.ReadFrom(buf)
		if err != nil {
			return
		}
		if _, err := c.WriteTo(buf[:n], addr); err != nil {
			return
		}
	}
}

func acceptAndEcho(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			return
		}
		go echo(c)
	}
}

func CI() bool {
	return os.Getenv("TRAVIS") == "true"
}

func TestStreamListenSamePort(t *testing.T) {

	// any ports
	any := [][]string{
		[]string{"tcp", "0.0.0.0:0"},
		[]string{"tcp4", "0.0.0.0:0"},
		[]string{"tcp6", "[::]:0"},

		[]string{"tcp", "127.0.0.1:0"},
		[]string{"tcp", "[::1]:0"},
		[]string{"tcp4", "127.0.0.1:0"},
		[]string{"tcp6", "[::1]:0"},
	}

	// specific ports. off in CI
	specific := [][]string{
		[]string{"tcp", "127.0.0.1:5556"},
		[]string{"tcp", "[::1]:5557"},
		[]string{"tcp4", "127.0.0.1:5558"},
		[]string{"tcp6", "[::1]:5559"},
	}

	testCases := any
	if !CI() {
		testCases = append(testCases, specific...)
	}

	for _, tcase := range testCases {
		network := tcase[0]
		addr := tcase[1]
		t.Log("testing", network, addr)

		l1, err := Listen(network, addr)
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l1.Close()
		t.Log("listening 1", l1.Addr())

		l2, err := Listen(l1.Addr().Network(), l1.Addr().String())
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l2.Close()
		t.Log("listening 2", l2.Addr())

		l3, err := Listen(l2.Addr().Network(), l2.Addr().String())
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l3.Close()
		t.Log("listening 3", l3.Addr())

		if !addrEqual(l1.Addr(), l2.Addr()) {
			t.Fatal("addrs should match 1", l1.Addr(), l2.Addr())
		}

		if !addrEqual(l1.Addr(), l3.Addr()) {
			t.Fatal("addrs should match 2", l1.Addr(), l3.Addr())
		}
	}
}

func addrEqual(l1, l2 net.Addr) bool {

	l1A, l1P, err := net.SplitHostPort(l1.String())
	if err != nil {
		panic(err)
	}
	l2A, l2P, err := net.SplitHostPort(l2.String())
	if err != nil {
		panic(err)
	}
	l1Aip := net.ParseIP(l1A)
	l2Aip := net.ParseIP(l2A)
	return (l1Aip.Equal(l2Aip) || (l1Aip.IsUnspecified() && l2Aip.IsUnspecified())) && l1P == l2P
}

func TestDialSelf(t *testing.T) {
	t.SkipNow()
	l, err := Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Dial("tcp4", l.Addr().String(), l.Addr().String())
	if err == nil {
		t.Fatal("should have gotten an error for dialing self")
	}
}

func TestPacketListenSamePort(t *testing.T) {

	// any ports
	any := [][]string{
		[]string{"udp", "0.0.0.0:0"},
		[]string{"udp4", "0.0.0.0:0"},
		[]string{"udp6", "[::]:0"},

		[]string{"udp", "127.0.0.1:0"},
		[]string{"udp", "[::1]:0"},
		[]string{"udp4", "127.0.0.1:0"},
		[]string{"udp6", "[::1]:0"},
	}

	// specific ports. off in CI
	specific := [][]string{
		[]string{"udp", "127.0.0.1:5560"},
		[]string{"udp", "[::1]:5561"},
		[]string{"udp4", "127.0.0.1:5562"},
		[]string{"udp6", "[::1]:5563"},
	}

	testCases := any
	if !CI() {
		testCases = append(testCases, specific...)
	}

	for _, tcase := range testCases {
		network := tcase[0]
		addr := tcase[1]
		t.Log("testing", network, addr)

		l1, err := ListenPacket(network, addr)
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l1.Close()
		t.Log("listening", l1.LocalAddr())

		l2, err := ListenPacket(l1.LocalAddr().Network(), l1.LocalAddr().String())
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l2.Close()
		t.Log("listening", l2.LocalAddr())

		l3, err := ListenPacket(l2.LocalAddr().Network(), l2.LocalAddr().String())
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l3.Close()
		t.Log("listening", l3.LocalAddr())

		if !addrEqual(l1.LocalAddr(), l2.LocalAddr()) {
			t.Fatal("addrs should match 1", l1.LocalAddr(), l2.LocalAddr())
		}

		if !addrEqual(l1.LocalAddr(), l3.LocalAddr()) {
			t.Fatal("addrs should match 2", l1.LocalAddr(), l3.LocalAddr())
		}
	}
}

func TestStreamListenDialSamePort(t *testing.T) {

	any := [][]string{
		[]string{"tcp", "0.0.0.0:0", "0.0.0.0:0"},
		[]string{"tcp4", "0.0.0.0:0", "0.0.0.0:0"},
		[]string{"tcp6", "[::]:0", "[::]:0"},

		[]string{"tcp", "127.0.0.1:0", "127.0.0.1:0"},
		[]string{"tcp4", "127.0.0.1:0", "127.0.0.1:0"},
		[]string{"tcp6", "[::1]:0", "[::1]:0"},
	}

	specific := [][]string{
		[]string{"tcp", "127.0.0.1:0", "127.0.0.1:5571"},
		[]string{"tcp4", "127.0.0.1:0", "127.0.0.1:5573"},
		[]string{"tcp6", "[::1]:0", "[::1]:5574"},
		[]string{"tcp", "127.0.0.1:5570", "127.0.0.1:0"},
		[]string{"tcp4", "127.0.0.1:5572", "127.0.0.1:0"},
		[]string{"tcp6", "[::1]:5573", "[::1]:0"},
	}

	testCases := any
	if !CI() {
		testCases = append(testCases, specific...)
	}

	for _, tcase := range testCases {
		t.Log("testing", tcase)
		network := tcase[0]
		addr1 := tcase[1]
		addr2 := tcase[2]

		l1, err := Listen(network, addr1)
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l1.Close()
		t.Log("listening", l1.Addr())

		l2, err := Listen(network, addr2)
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l2.Close()
		t.Log("listening", l2.Addr())

		go acceptAndEcho(l1)
		go acceptAndEcho(l2)

		c1, err := Dial(network, l1.Addr().String(), l2.Addr().String())
		if err != nil {
			t.Fatal(err, network, l1.Addr().String(), l2.Addr().String())
			continue
		}
		defer c1.Close()
		t.Log("dialed", c1, c1.LocalAddr(), c1.RemoteAddr())

		if getPort(l1.Addr()) != getPort(c1.LocalAddr()) {
			t.Fatal("addrs should match", l1.Addr(), c1.LocalAddr())
		}

		if getPort(l2.Addr()) != getPort(c1.RemoteAddr()) {
			t.Fatal("addrs should match", l2.Addr(), c1.RemoteAddr())
		}

		hello1 := []byte("hello world")
		hello2 := make([]byte, len(hello1))
		if _, err := c1.Write(hello1); err != nil {
			t.Fatal(err)
			continue
		}

		if _, err := c1.Read(hello2); err != nil {
			t.Fatal(err)
			continue
		}

		if !bytes.Equal(hello1, hello2) {
			t.Fatal("echo failed", string(hello1), "!=", string(hello2))
		}
		t.Log("echoed", string(hello2))
		c1.Close()
	}
}

func TestStreamListenDialSamePortStressManyMsgs(t *testing.T) {
	testCases := [][]string{
		[]string{"tcp", "127.0.0.1:0"},
		[]string{"tcp4", "127.0.0.1:0"},
		[]string{"tcp6", "[::]:0"},
	}

	for _, tcase := range testCases {
		t.Run(tcase[0], func(t *testing.T) {
			subestStreamListenDialSamePortStress(t, tcase[0], tcase[1], 2, 1000)
		})
	}
}

func TestStreamListenDialSamePortStressManyNodes(t *testing.T) {
	testCases := [][]string{
		[]string{"tcp", "127.0.0.1:0"},
		[]string{"tcp4", "127.0.0.1:0"},
		[]string{"tcp6", "[::]:0"},
	}

	for _, tcase := range testCases {
		t.Run(tcase[0], func(t *testing.T) {
			subestStreamListenDialSamePortStress(t, tcase[0], tcase[1], 50, 1)
		})
	}
}

func TestStreamListenDialSamePortStressManyMsgsManyNodes(t *testing.T) {
	testCases := [][]string{
		[]string{"tcp", "127.0.0.1:0"},
		[]string{"tcp4", "127.0.0.1:0"},
		[]string{"tcp6", "[::]:0"},
	}
	for _, tcase := range testCases {
		t.Run(tcase[0], func(t *testing.T) {
			subestStreamListenDialSamePortStress(t, tcase[0], tcase[1], 50, 50)
		})
	}
}

func subestStreamListenDialSamePortStress(t *testing.T, network, addr string, nodes int, msgs int) {

	var ls []net.Listener
	for i := 0; i < nodes; i++ {
		l, err := Listen(network, addr)
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		go acceptAndEcho(l)
		ls = append(ls, l)
	}

	// connect them all
	k := 0
	var cs []net.Conn
	for i := 0; i < nodes; i++ {
		for j := 0; j < i; j++ {
			if i == j {
				continue // cannot do self.
			}

			ia := ls[i].Addr().String()
			ja := ls[j].Addr().String()
			c, err := Dial(network, ia, ja)
			if err != nil {
				t.Fatal(network, ia, ja, err, k)
			}
			k++
			defer c.Close()
			cs = append(cs, c)
		}
	}

	errs := make(chan error)

	send := func(c net.Conn, buf []byte) {
		if _, err := c.Write(buf); err != nil {
			errs <- err
		}
	}

	recv := func(c net.Conn, buf []byte) {
		buf2 := make([]byte, len(buf))
		if _, err := c.Read(buf2); err != nil {
			errs <- err
		}
		if !bytes.Equal(buf, buf2) {
			errs <- fmt.Errorf("recv failure: %s <--> %s -- %s %s", c.RemoteAddr(), c.LocalAddr(), buf, buf2)
		}
	}

	t.Logf("sending %d msgs per conn", msgs)
	go func() {
		var wg sync.WaitGroup
		for _, c := range cs {
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				for i := 0; i < msgs; i++ {
					msg := []byte(fmt.Sprintf("message %d", i))
					send(c, msg)
					recv(c, msg)
				}
			}(c)
		}
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestPacketListenDialSamePort(t *testing.T) {
	t.Skip("these don't pass reliably.")

	any := [][]string{
		[]string{"udp", "0.0.0.0:0", "0.0.0.0:0"},
		[]string{"udp4", "0.0.0.0:0", "0.0.0.0:0"},
		[]string{"udp6", "[::]:0", "[::]:0"},

		[]string{"udp", "127.0.0.1:0", "127.0.0.1:0"},
		[]string{"udp4", "127.0.0.1:0", "127.0.0.1:0"},
		[]string{"udp6", "[::1]:0", "[::1]:0"},
	}

	specific := [][]string{
		[]string{"udp", "127.0.0.1:5670", "127.0.0.1:5671"},
		[]string{"udp4", "127.0.0.1:5672", "127.0.0.1:5673"},
		[]string{"udp6", "[::1]:5673", "[::1]:5674"},
	}

	testCases := any
	if !CI() {
		testCases = append(testCases, specific...)
	}

	for _, tcase := range testCases {
		t.Run(tcase[0]+"/"+tcase[1], func(t *testing.T) {
			network := tcase[0]
			addr1 := tcase[1]
			addr2 := tcase[2]

			l1, err := ListenPacket(network, addr1)
			if err != nil {
				t.Fatal(err)
			}
			defer l1.Close()
			t.Log("listening", l1.LocalAddr())

			l2, err := ListenPacket(network, addr2)
			if err != nil {
				t.Fatal(err)
			}
			defer l2.Close()
			t.Log("listening", l2.LocalAddr())

			go packetEcho(l1)
			go packetEcho(l2)

			c1, err := Dial(network, l1.LocalAddr().String(), l2.LocalAddr().String())
			if err != nil {
				t.Fatal(err)
			}
			defer c1.Close()
			t.Log("dialed", c1.LocalAddr(), c1.RemoteAddr())

			if getPort(l1.LocalAddr()) != getPort(c1.LocalAddr()) {
				t.Fatal("addrs should match", l1.LocalAddr(), c1.LocalAddr())
			}

			if getPort(l2.LocalAddr()) != getPort(c1.RemoteAddr()) {
				t.Fatal("addrs should match", l2.LocalAddr(), c1.RemoteAddr())
			}

			hello1 := []byte("hello world")
			hello2 := make([]byte, len(hello1))
			if _, err := c1.Write(hello1); err != nil {
				t.Fatal(err)
			}

			if err := c1.SetReadDeadline(time.Now().Add(time.Second * 2)); err != nil {
				t.Fatal(err)
			}

			if _, err := c1.Read(hello2); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(hello1, hello2) {
				t.Fatal("echo failed", string(hello1), "!=", string(hello2))
			}
			t.Log("echoed", string(hello2))
		})
	}
}

func TestOpenFDs(t *testing.T) {
	// this is a totally ad-hoc limit. test harnesses may add fds.
	// but if this is really much higher than 20, there's obviously leaks.
	limit := 20
	start := time.Now()
	for countOpenFiles(t) > limit {
		<-time.After(time.Second)
		t.Log("open fds:", countOpenFiles(t))
		if time.Now().Sub(start) > (time.Second * 15) {
			t.Error("fd leak!")
		}
	}
}

func countOpenFiles(t *testing.T) int {
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("lsof -p %v", os.Getpid())).Output()
	if err != nil {
		t.Fatal(err)
	}
	return bytes.Count(out, []byte("\n"))
}

func getPort(a net.Addr) string {
	if a == nil {
		return ""
	}
	s := strings.Split(a.String(), ":")
	if len(s) > 1 {
		return s[1]
	}
	return ""
}

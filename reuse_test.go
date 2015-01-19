package reuseport

import (
	"bytes"
	"io"
	"net"
	"os"
	"testing"
)

func echo(c net.Conn) {
	io.Copy(c, c)
	c.Close()
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

func TestListenSamePort(t *testing.T) {

	// any ports
	any := [][]string{
		[]string{"tcp", "127.0.0.1:0"},
		[]string{"tcp", "[::1]:0"},
		[]string{"tcp4", "127.0.0.1:0"},
		[]string{"tcp6", "[::1]:0"},
		[]string{"udp", "127.0.0.1:0"},
		[]string{"udp", "[::1]:0"},
		[]string{"udp4", "127.0.0.1:0"},
		[]string{"udp6", "[::1]:0"},
	}

	// specific ports. off in CI
	specific := [][]string{
		[]string{"tcp", "127.0.0.1:5556"},
		[]string{"tcp", "[::1]:5557"},
		[]string{"tcp4", "127.0.0.1:5558"},
		[]string{"tcp6", "[::1]:5559"},
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

		l1, err := Listen(network, addr)
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l1.Close()
		t.Log("listening", l1.Addr())

		l2, err := Listen(l1.Addr().Network(), l1.Addr().String())
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l2.Close()
		t.Log("listening", l2.Addr())

		l3, err := Listen(l2.Addr().Network(), l2.Addr().String())
		if err != nil {
			t.Fatal(err)
			continue
		}
		defer l3.Close()
		t.Log("listening", l3.Addr())

		if l1.Addr().String() != l2.Addr().String() {
			t.Fatal("addrs should match", l1.Addr(), l2.Addr())
		}

		if l1.Addr().String() != l3.Addr().String() {
			t.Fatal("addrs should match", l1.Addr(), l3.Addr())
		}
	}
}

func TestListenDialSamePort(t *testing.T) {

	any := [][]string{
		[]string{"tcp", "127.0.0.1:0", "127.0.0.1:0"},
		[]string{"tcp4", "127.0.0.1:0", "127.0.0.1:0"},
		[]string{"tcp6", "[::1]:0", "[::1]:0"},
		[]string{"udp", "127.0.0.1:0", "127.0.0.1:0"},
		[]string{"udp4", "127.0.0.1:0", "127.0.0.1:0"},
		[]string{"udp6", "[::1]:0", "[::1]:0"},
	}

	specific := [][]string{
		[]string{"tcp", "127.0.0.1:5570", "127.0.0.1:5571"},
		[]string{"tcp4", "127.0.0.1:5572", "127.0.0.1:5573"},
		[]string{"tcp6", "[::1]:5573", "[::1]:5574"},
		[]string{"udp", "127.0.0.1:5670", "127.0.0.1:5671"},
		[]string{"udp4", "127.0.0.1:5672", "127.0.0.1:5673"},
		[]string{"udp6", "[::1]:5673", "[::1]:5674"},
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
			t.Fatal(err)
			continue
		}
		defer c1.Close()
		t.Log("dialed", c1.LocalAddr(), c1.RemoteAddr())

		if l1.Addr().String() != c1.LocalAddr().String() {
			t.Fatal("addrs should match", l1.Addr(), c1.LocalAddr())
		}

		if l2.Addr().String() != c1.RemoteAddr().String() {
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
	}
}

func TestUnixNotSupported(t *testing.T) {

	testCases := [][]string{
		[]string{"unix", "/tmp/foo"},
	}

	for _, tcase := range testCases {
		network := tcase[0]
		addr := tcase[1]
		t.Log("testing", network, addr)

		_, err := Listen(network, addr)
		if err == nil {
			t.Fatal("unix supported")
			continue
		}
	}
}

// +build darwin freebsd dragonfly netbsd openbsd linux

package reuseport

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"

	resolve "github.com/jbenet/go-net-resolve-addr"
	sockaddrnet "github.com/jbenet/go-sockaddr/net"
)

const (
	tcp4       = 52 // "4"
	tcp6       = 54 // "6"
	filePrefix = "port."
)

func dial(dialer net.Dialer, netw, addr string) (c net.Conn, err error) {
	var (
		family, fd     int
		file           *os.File
		remoteSockaddr syscall.Sockaddr
		localSockaddr  syscall.Sockaddr
	)

	netAddr, err := resolve.ResolveAddr("dial", netw, addr)
	if err != nil {
		fmt.Println("resolve addr failed")
		return nil, err
	}
	fmt.Println("resolve addr ok")

	switch netAddr.(type) {
	case *net.TCPAddr, *net.UDPAddr:
	default:
		return nil, ErrUnsupportedProtocol
	}

	switch dialer.LocalAddr.(type) {
	case *net.TCPAddr, *net.UDPAddr:
	default:
		return nil, ErrUnsupportedProtocol
	}

	family = sockaddrnet.NetAddrAF(netAddr)
	localSockaddr = sockaddrnet.NetAddrToSockaddr(dialer.LocalAddr)
	remoteSockaddr = sockaddrnet.NetAddrToSockaddr(netAddr)

	if fd, err = syscall.Socket(family, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); err != nil {
		fmt.Println("tcp socket failed")
		return nil, err
	}

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, soReuseAddr, 1); err != nil {
		fmt.Println("reuse addr failed")
		return nil, err
	}

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, soReusePort, 1); err != nil {
		fmt.Println("reuse port failed")
		return nil, err
	}

	if localSockaddr != nil {
		if err = syscall.Bind(fd, localSockaddr); err != nil {
			fmt.Println("bind failed")
			return nil, err
		}
	}

	// Set backlog size to the maximum
	if err = syscall.Connect(fd, remoteSockaddr); err != nil {
		fmt.Println("connect failed")
		return nil, err
	}

	// File Name get be nil
	file = os.NewFile(uintptr(fd), filePrefix+strconv.Itoa(os.Getpid()))
	if c, err = net.FileConn(file); err != nil {
		return nil, err
	}

	if err = file.Close(); err != nil {
		return nil, err
	}

	return c, err
}

func listen(netw, addr string) (l net.Listener, err error) {
	var (
		family, fd int
		file       *os.File
		sockaddr   syscall.Sockaddr
	)

	netAddr, err := resolve.ResolveAddr("listen", netw, addr)
	if err != nil {
		fmt.Println("resolve addr failed")
		return nil, err
	}
	fmt.Println("resolve addr ok")

	switch netAddr.(type) {
	case *net.TCPAddr, *net.UDPAddr:
	default:
		return nil, ErrUnsupportedProtocol
	}

	family = sockaddrnet.NetAddrAF(netAddr)
	sockaddr = sockaddrnet.NetAddrToSockaddr(netAddr)

	if fd, err = syscall.Socket(family, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); err != nil {
		fmt.Println("socket failed")
		return nil, err
	}

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, soReusePort, 1); err != nil {
		fmt.Println("setsockopt reuseport failed")
		return nil, err
	}

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, soReuseAddr, 1); err != nil {
		fmt.Println("setsockopt reuseaddr failed")
		return nil, err
	}

	if err = syscall.Bind(fd, sockaddr); err != nil {
		fmt.Println("bind failed")
		return nil, err
	}

	// Set backlog size to the maximum
	if err = syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		fmt.Println("listen failed")
		return nil, err
	}

	// File Name get be nil
	file = os.NewFile(uintptr(fd), filePrefix+strconv.Itoa(os.Getpid()))
	if l, err = net.FileListener(file); err != nil {
		return nil, err
	}

	if err = file.Close(); err != nil {
		return nil, err
	}

	return l, err
}

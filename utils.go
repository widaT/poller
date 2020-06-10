package poller

import (
	"crypto/tls"
	"errors"
	"net"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func Listener2Fd(ln net.Listener, nonblock bool) (fd int, err error) {
	var file *os.File
	switch l := ln.(type) {
	case *net.TCPListener:
		file, err = l.File()
	case *net.UnixListener:
		file, err = l.File()
	}
	if err != nil {
		return
	}
	fd = int(file.Fd())
	if nonblock {
		err = Nonblock(fd)
	}
	return
}

func PacketConn2Fd(pconn net.PacketConn, nonblock bool) (fd int, err error) {
	var file *os.File
	switch pf := pconn.(type) {
	case *net.UDPConn:
		file, err = pf.File()
	}
	if err != nil {
		return
	}
	fd = int(file.Fd())
	if nonblock {
		err = Nonblock(fd)
	}
	return
}

type fakeTLSConn struct {
	Conn net.Conn
}

func NetConn2Fd(conn net.Conn, nonblock bool) (fd int, err error) {
	var f *os.File
	if c, ok := conn.(*tls.Conn); ok {
		p := unsafe.Pointer(c)
		fakeTLSConn := (*fakeTLSConn)(p)
		conn = fakeTLSConn.Conn
	}
	switch typ := conn.(type) {
	case *net.TCPConn:
		f, err = typ.File()
		if err != nil {
			return 0, err
		}
	case *net.UDPConn:
		f, err = typ.File()
		if err != nil {
			return 0, err
		}
	default:
		return 0, errors.New("unsupported net conn")
	}
	fd = int(f.Fd())
	if nonblock {
		err = Nonblock(fd)
	}
	return
}

func Nonblock(fd int) error {
	return unix.SetNonblock(fd, true)
}

func Block(fd int) error {
	return unix.SetNonblock(fd, false)
}

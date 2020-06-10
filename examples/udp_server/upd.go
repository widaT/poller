package main

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"
	"golang.org/x/sys/unix"
)

const UDP_SOCKET poller.Token = poller.Token(0)

// run it
// then nc  -u localhost 8888
func main() {
	poll, err := poller.New()
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.ListenPacket("udp", ":8888")
	if err != nil {
		log.Fatal(err)
	}

	fd, _ := poller.PacketConn2Fd(ln, true)
	//also
	//addr, _ := net.ResolveUDPAddr("udp", ":8888")
	//conn, _ := net.ListenUDP("udp", addr)
	//fd, _ := poller.NetConn2Fd(conn, true)

	err = poll.Register(fd, UDP_SOCKET, interest.READABLE, pollopt.Edge)
	if err != nil {
		log.Fatal(err)
	}
	events := poller.MakeEvents(128)
	for {
		n, err := poll.Select(events, -1)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			log.Fatal(err)
		}
		for i := 0; i < n; i++ {
			ev := events[i]
			switch ev.Token() {
			case UDP_SOCKET:
				loopReadUDP(fd)
			default:
				log.Fatal(errors.New("unreachable"))
			}
		}
	}
}

func loopReadUDP(fd int) error {
	receivedData := make([]byte, 4096)
	n, sa, err := unix.Recvfrom(fd, receivedData, 0)
	if err != nil || n == 0 {
		if err != nil && err != unix.EAGAIN {
			log.Fatal(err)
		}
		return nil
	}
	fmt.Println(string(receivedData[:n]))
	unix.Sendto(fd, receivedData[:n], 0, sa)
	return nil
}

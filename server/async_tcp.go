package server

import (
	"log"
	"net"
	"strconv"
	"syscall"

	"github.com/knightfall22/wumi/config"
	"github.com/knightfall22/wumi/core"
)

var conn_clients int

func RunASyncTCPServer() error {
	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	log.Println("Starting asynchronous server on", address)

	max_clients := 20000

	//Create EPOLL even object to hold events
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)

	//create a socket(IPV4, non blocking socket that does not drop the socket)
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}

	defer syscall.Close(serverFD)

	// Set the Socket operate in blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// Bind the IP and the port
	ip4 := net.ParseIP(config.Host)
	if err := syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	//start listening
	if err = syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	//Async.io start here!!

	//creating EPOLL instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}

	defer syscall.Close(epollFD)

	//specify the event we want to get hints about
	//and set the socket on which to monitor
	socketServerEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	//listen and read events on the server itself
	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent); err != nil {
		return err
	}

	for {
		//see if any FD is ready for an IO
		nevents, err := syscall.EpollWait(epollFD, events[:], -1)
		if err != nil {
			continue
		}

		for i := range nevents {
			//check if the socket itself is ready for an IO
			if int(events[i].Fd) == serverFD {
				//accept incoming request from a server
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err", err)
					continue
				}

				//increase the number of concurrent clients
				conn_clients++
				syscall.SetNonblock(fd, true)

				// add this new tcp client to be monitored
				socketClientEvent := syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}

				if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &socketClientEvent); err != nil {
					log.Fatalln(err)
				}
			} else {
				comm := core.FDComm{
					Fd: int(events[i].Fd),
				}

				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					conn_clients--
					continue
				}

				response(cmd, comm)
			}
		}
	}
}

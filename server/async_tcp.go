package server

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/knightfall22/wumi/config"
	"github.com/knightfall22/wumi/core"
)

var conn_clients int

var cronFrequency time.Duration = 1 * time.Second
var lastCronExecTime time.Time = time.Now()

const (
	_ int32 = 1 << iota
	EngineStatus_Waiting
	EngineStatus_BUSY
	EngineStatus_SHUTTING_DOWN
)

var eStatus = EngineStatus_Waiting

func WaitForSignal(wg *sync.WaitGroup, sigs <-chan os.Signal) {
	defer wg.Done()

	<-sigs

	//if server is busy continue to wait
	for atomic.LoadInt32(&eStatus) == EngineStatus_BUSY {
	}

	atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)

	core.ShutDown()
	os.Exit(0)

}

func RunASyncTCPServer(wg *sync.WaitGroup) error {
	defer wg.Done()
	defer func() {
		atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)
	}()

	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	log.Println("Starting asynchronous server on", address)

	max_clients := 1000

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

	for atomic.LoadInt32(&eStatus) != EngineStatus_SHUTTING_DOWN {
		if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys()
			lastCronExecTime = time.Now()
		}

		//see if any FD is ready for an IO
		nevents, err := syscall.EpollWait(epollFD, events[:], -1)
		if err != nil {
			continue
		}

		if !atomic.CompareAndSwapInt32(&eStatus, EngineStatus_Waiting, EngineStatus_BUSY) {
			switch eStatus {
			case EngineStatus_SHUTTING_DOWN:
				return nil
			}
		}

		for i := 0; i < nevents; i++ {
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

				cmds, err := readCommands(comm)
				if err != nil {
					if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_DEL, int(events[i].Fd), nil); err != nil {
						log.Println("Error removing fd from epoll:", err)
					}
					syscall.Close(int(events[i].Fd))
					conn_clients--
					continue
				}

				response(cmds, comm)
			}
		}

		atomic.StoreInt32(&eStatus, EngineStatus_Waiting)
	}

	return nil
}

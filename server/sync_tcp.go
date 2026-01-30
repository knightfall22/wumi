package server

import (
	"io"
	"log"
	"net"
	"strconv"

	"github.com/knightfall22/wumi/config"
)

func RunSyncTCPServer() {
	log.Println("Starting synchronous server on", config.Host, config.Port)

	var con_client = 0

	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))

	//listening to configed host:port
	lnr, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	for {
		//blocking call waiting for new client to connect
		c, err := lnr.Accept()
		if err != nil {
			panic(err)
		}

		//increment the number of concurrent clients
		con_client++
		log.Printf("client connected with remote address: %s, concurrent clients: %d\n", c.RemoteAddr().String(), con_client)

		for {
			cmd, err := readCommand(c)
			if err != nil {
				c.Close()
				con_client--
				log.Printf("client disconnected remote address: %s, concurrent clients: %d\n", c.RemoteAddr().String(), con_client)

				if err == io.EOF {
					break
				}
				log.Println("err", err)
			}

			log.Println("command", cmd)
			if err := response(cmd, c); err != nil {
				log.Println("err write", err)
			}
		}
	}
}

func readCommand(c net.Conn) (string, error) {
	buf := make([]byte, 521)
	n, err := c.Read(buf[:])
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

func response(cmd string, c net.Conn) error {
	if _, err := c.Write([]byte(cmd)); err != nil {
		return err
	}
	return nil
}

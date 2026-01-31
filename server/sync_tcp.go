package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/knightfall22/wumi/config"
	"github.com/knightfall22/wumi/core"
)

func RunSyncTCPServer() {
	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	log.Println("Starting synchronous server on", address)

	var con_client = 0

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
			response(cmd, c)
		}
	}
}

func readCommand(c io.ReadWriter) (*core.RedisCmd, error) {
	buf := make([]byte, 521)
	n, err := c.Read(buf[:])
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil
}

func response(cmd *core.RedisCmd, c io.ReadWriter) {
	if err := core.EvalAndRespond(cmd, c); err != nil {
		respondError(err, c)
	}
}

func respondError(err error, c io.ReadWriter) {
	log.Println("error", err)
	fmt.Fprintf(c, "-%s\r\n", err)
}

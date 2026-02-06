package server

import (
	"io"
	"strings"

	"github.com/knightfall22/wumi/core"
)

// func RunSyncTCPServer() {
// 	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
// 	log.Println("Starting synchronous server on", address)

// 	var con_client = 0

// 	//listening to configed host:port
// 	lnr, err := net.Listen("tcp", address)
// 	if err != nil {
// 		panic(err)
// 	}

// 	for {
// 		//blocking call waiting for new client to connect
// 		c, err := lnr.Accept()
// 		if err != nil {
// 			panic(err)
// 		}

// 		//increment the number of concurrent clients
// 		con_client++
// 		log.Printf("client connected with remote address: %s, concurrent clients: %d\n", c.RemoteAddr().String(), con_client)

// 		for {
// 			cmds, err := readCommands(c)
// 			if err != nil {
// 				c.Close()
// 				con_client--
// 				log.Printf("client disconnected remote address: %s, concurrent clients: %d\n", c.RemoteAddr().String(), con_client)

// 				if err == io.EOF {
// 					break
// 				}
// 				log.Println("err", err)
// 			}

// 			log.Println("command", cmds)
// 			response(cmds, c)
// 		}
// 	}
// }

func readCommands(c io.ReadWriter) (core.RedisCmds, error) {
	buf := make([]byte, 521)
	n, err := c.Read(buf[:])
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}

	values, err := core.Decode(buf[:n])
	if err != nil {
		return nil, err
	}

	cmds := make(core.RedisCmds, 0)

	for _, v := range values {
		tokens, err := toArrayString(v.([]any))
		if err != nil {
			return nil, err
		}

		cmds = append(cmds, &core.RedisCmd{
			Cmd:  strings.ToUpper(tokens[0]),
			Args: tokens[1:],
		})
	}

	return cmds, nil
}

func toArrayString(value []any) ([]string, error) {
	stringValue := make([]string, len(value))
	for i := range value {
		stringValue[i] = value[i].(string)
	}

	return stringValue, nil
}

func response(cmds core.RedisCmds, c *core.Client) {
	core.EvalAndRespond(cmds, c)
}

// func respondError(err error, c io.ReadWriter) {
// 	log.Println("error", err)
// 	fmt.Fprintf(c, "-%s\r\n", err)
// }

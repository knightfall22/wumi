package core

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/knightfall22/wumi/config"
)

func dumpkey(file *os.File, key string, value *Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, value.Value)
	tokens := strings.Split(cmd, " ")

	_, err := file.Write(Encode(tokens, false))
	if err != nil {
		log.Println("error writing to log file: ", err)
	}
}

func DumpAllAof() {
	fp, err := os.OpenFile(config.AOFFILE, os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Printf("error with aof file: %v\n", err)
		return
	}

	log.Println("rewriting AOF file at", config.AOFFILE)
	for k, obj := range store {
		dumpkey(fp, k, obj)
	}
	log.Println("AOF rewrite complete")
}

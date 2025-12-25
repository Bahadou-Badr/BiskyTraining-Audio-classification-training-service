package logger

import (
	"log"
	"os"
)

var L *log.Logger

func InitLogger(env string) {
	L = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}

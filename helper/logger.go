package helper

import (
	"log"
	"os"
)

var Mylog = log.New(os.Stderr, "app: ", log.LstdFlags|log.Lshortfile)

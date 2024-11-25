package main

import (
	"os"

	//"os"

	"github.com/KinMod-ui/thelaLocator/helper"
	"github.com/KinMod-ui/thelaLocator/websockets"
)

type user struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

func main() {

	port := os.Args[1]
	helper.Mylog.Println("Entering app")

	websockets.SetupWSServer(port)

	helper.Mylog.Println("Exiting app")
}

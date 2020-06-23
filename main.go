package main

import (
	"visitw_backend/misc"
	"visitw_backend/router"
)

func init() {

}
func main() {

	router.Start(misc.Config.Port)

}

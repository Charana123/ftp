package main

import (
	"github.com/Charana123/ftp/server"
	"github.com/Charana123/ftp/server/driver"
)

func main() {
	driver := &driver.ExampleDriver{
		MaxConnections: 50,
	}
	server.StartServer(driver)
}

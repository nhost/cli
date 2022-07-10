package util

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func GetPort(low, hi int) uint32 {

	//
	//  Initialize the seed
	//
	//  This is done to prevent Go from choosing pseudo-random numbers
	rand.Seed(time.Now().UnixNano())

	//  generate a random port value
	port := uint32(low + rand.Intn(hi-low))

	//  validate whether the port is available
	if !PortAvailable(port) {
		return GetPort(low, hi)
	}

	//  return the value, if it's available
	return port
}

func PortAvailable(port uint32) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}

	ln.Close()

	ln, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}

	ln.Close()
	return true
}

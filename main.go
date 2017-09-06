package main

import (
	"flag"
	"net"
	"strconv"
	"sync"

	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/golib"
)

func handleTcpConn(_ *sync.WaitGroup, conn *net.TCPConn) {
	remote := conn.RemoteAddr()
	log.Debugln("Accepted TCP connection from", remote)
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Errorf("Error receiving TCP data from %v: %v\n", remote, err)
	} else {
		printData(data, remote.String(), "TCP")
	}
}

func handleUdpPacket(_ *sync.WaitGroup, remoteAddr *net.UDPAddr, packet []byte) {
	printData(packet, remoteAddr.String(), "UDP")
}

func printData(data []byte, remoteAddr string, proto string) {
	dataStr := string(data)
	if len(dataStr) > 0 && dataStr[len(dataStr)-1] == '\n' {
		dataStr = dataStr[:len(dataStr)-1] // Newline will be added by logger
	}
	log.Printf("%v from %v: %v", proto, remoteAddr, dataStr)
}

func main() {
	var tcpEndpoints, udpEndpoints golib.StringSlice
	flag.Var(&tcpEndpoints, "t", "TCP endpoints to listen on")
	flag.Var(&udpEndpoints, "u", "UDP endpoints to listen on")
	golib.RegisterFlags(golib.FlagsAll)
	flag.Parse()
	defer golib.ProfileCpu()()

	if len(tcpEndpoints)+len(udpEndpoints) == 0 {
		log.Fatalln("Please specify at least one UDP or TCP endpoint to listen on.")
	}

	var tasks golib.TaskGroup
	for _, endpoint := range tcpEndpoints {
		tasks.Add(&golib.TCPListenerTask{
			ListenEndpoint: endpoint,
			Handler:        handleTcpConn,
		})
	}
	for _, endpoint := range udpEndpoints {
		tasks.Add(&golib.UDPListenerTask{
			ListenEndpoint: endpoint,
			Handler:        handleUdpPacket,
		})
	}
	tasks.PrintWaitAndStop()
}

func parsePorts(strings golib.StringSlice) ([]int, error) {
	res := make([]int, len(strings))
	for i, str := range strings {
		port, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		res[i] = port
	}
	return res, nil
}

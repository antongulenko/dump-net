package main

import (
	"flag"
	"net"
	"strconv"
	"sync"

	"bufio"

	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

func handleTcpConn(_ *sync.WaitGroup, conn *net.TCPConn) {
	remote := conn.RemoteAddr()
	log.Debugln("Accepted TCP connection from", remote)
	b := bufio.NewReader(conn)
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			log.Errorf("Error receiving TCP data from %v: %v\n", remote, err)
			break
		} else {
			printData(line, conn.LocalAddr(), remote.String(), "TCP")
		}
	}
}

func handleUdpPacket(_ *sync.WaitGroup, localAddr net.Addr, remoteAddr *net.UDPAddr, packet []byte) {
	printData(packet, localAddr, remoteAddr.String(), "UDP")
}

func printData(data []byte, localAddr net.Addr, remoteAddr string, proto string) {
	dataStr := string(data)
	if len(dataStr) > 0 && dataStr[len(dataStr)-1] == '\n' {
		dataStr = dataStr[:len(dataStr)-1] // Newline will be added by logger
	}
	log.Printf("[%v] Received on %v from %v: %v", proto, localAddr, remoteAddr, dataStr)
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

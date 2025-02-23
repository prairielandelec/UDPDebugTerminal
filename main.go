package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const MulticastAddr = "239.239.239.239"

const BuffSize = 16384

type UserConfig struct {
	HdwInterface *net.Interface
	IPAddr       string
	Port         string
}

var Config UserConfig

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Need an Interface Index, Host IP, and Port Number")

		interfaces, err := net.Interfaces()
		if err != nil {
			fmt.Println("Could Not Find Any Interfaces")
			fmt.Println(err)
			os.Exit(1)
		}

		for _, iface := range interfaces {
			fmt.Printf("Index: %d, Name: %s, HardwareAddr: %s, Flags: %s\n",
				iface.Index, iface.Name, iface.HardwareAddr, iface.Flags)
		}
		os.Exit(1)
	}

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	ArgHdwIdx, _ := strconv.Atoi(os.Args[1])
	Config.Port = os.Args[3]
	Config.IPAddr = os.Args[2]

	var ifaceerr error
	Config.HdwInterface, ifaceerr = net.InterfaceByIndex(ArgHdwIdx)
	if ifaceerr != nil {
		fmt.Println(ifaceerr)
		os.Exit(1)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", MulticastAddr+":"+Config.Port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Printf("Using interface %s", Config.HdwInterface.Name)
	conn, err := net.ListenMulticastUDP("udp", Config.HdwInterface, udpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Listening to:\t%s:%s on %s\n", Config.IPAddr, Config.Port, udpAddr.String())

	packetChan := make(chan *net.UDPAddr)

	// Start the listener in a goroutine
	go listenForPackets(conn, packetChan)

	for {
		select {
		case <-interruptChan:
			fmt.Println("\nReceived Ctrl+C, closing connection...")
			return
		case <-packetChan:
			continue
		}
	}
}

func listenForPackets(conn *net.UDPConn, packetChan chan<- *net.UDPAddr) {
	buf := make([]byte, BuffSize)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				return
			}
			fmt.Println(err)
			continue
		}

		if addr.IP.String() == Config.IPAddr {
			now := time.Now()
			formattedTime := now.Format("15:04:05")
			fmt.Printf("%s: %s\n", formattedTime, string(buf[:n]))
			packetChan <- addr
		}
	}
}

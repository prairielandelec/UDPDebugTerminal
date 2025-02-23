package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

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

	ifaceIdx, _ := strconv.Atoi(os.Args[1])

	// Resolve the UDP address for the multicast group
	udpAddr, err := net.ResolveUDPAddr("udp", "239.239.239.239:"+os.Args[3])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get the network interface by index
	iface, err := net.InterfaceByIndex(ifaceIdx) // Replace 6 with the actual index of your interface
	fmt.Printf("Using Interface: %s\n", iface.Name)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create a UDP connection to listen on the specified port
	conn, err := net.ListenMulticastUDP("udp", iface, udpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Listening to:%s:%s on %s\n", os.Args[2], os.Args[3], udpAddr.String())

	// Buffer to hold incoming data
	buf := make([]byte, 1024)

	for {
		// Read from the connection
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Check if the packet is from the desired source address
		if addr.IP.String() == os.Args[2] {
			// Print the data read from the connection to the terminal
			now := time.Now()
			formattedTime := now.Format("15:04:05")
			fmt.Printf("%s: %s\n", formattedTime, string(buf[:n]))
		}
	}
}

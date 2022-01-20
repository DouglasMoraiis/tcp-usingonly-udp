package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

func checkError(err error, msg string){
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro em " + msg + "\n", err.Error())
		os.Exit(1)
	}
}

func handleClient(conn *net.UDPConn)  {
	var netBuffer [60000]byte
	var fileBuffer [60000]byte

	_, err := conn.Read(netBuffer[0:])
	checkError(err, "Read")

	base64.StdEncoding.Decode(fileBuffer[:], netBuffer[:])

	ioutil.WriteFile("meucru.png", fileBuffer[:], 0444)
}

func main() {
	service := ":3200"

	udpAddr, _ := net.ResolveUDPAddr("udp", service)

	conn, _ := net.ListenUDP("udp", udpAddr)

	for {
		handleClient(conn)
	}
}

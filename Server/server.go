package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

func checkError(err error, msg string){
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro em " + msg + "\n", err.Error())
		os.Exit(1)
	}
}

func checkParams(args []string) (string, string) {
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Error: Argumentos esperados: <porta> <diretório>")
		os.Exit(1)
	}
	port := os.Args[1]
	dir := os.Args[2]
	return port, dir
}

func handleClient(conn *net.UDPConn, dir string)  {
	var netBuffer [524]byte
	var fileBuffer [524]byte

	size, err := conn.Read(netBuffer[0:])
	fmt.Println(size)
	checkError(err, "Read")

	_, err = base64.StdEncoding.Decode(fileBuffer[0:size], netBuffer[0:size])
	checkError(err, "Decode")

	dirFile := strings.TrimSpace(dir) + "file.png"
	err = ioutil.WriteFile(dirFile, fileBuffer[0:size], 0666)
	checkError(err, "WriteFile")
	os.Exit(0)
}

func main() {
	port, dir := checkParams(os.Args)

	udpAddr, _ := net.ResolveUDPAddr("udp", port)

	conn, _ := net.ListenUDP("udp", udpAddr)

	for {
		handleClient(conn, dir)
	}
}

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"os"
)

func checkError(err error, msg string){
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro em " + msg + "\n", err.Error())
		os.Exit(1)
	}
}

func checkParams(args []string) (net.IP, string, *os.File) {
	if len(args) != 4 {
		fmt.Fprintf(os.Stderr, "Argumentos esperados: <hostname/ip> <porta> <arquivo> em %s\n", args[0])
		os.Exit(1)
	}
	addr := net.ParseIP(args[1])
	if addr == nil {
		fmt.Fprintf(os.Stderr, "%s não é um hostname/ip válido.\n", args[1])
		os.Exit(1)
	}
	ip := net.ParseIP(os.Args[1])
	port := os.Args[2]
	file, err := os.Open(os.Args[3])
	checkError(err, "<arquivo> inválido")
	return ip, port, file
}

func readFile() []byte {
	file, err := os.Open("./go.png")
	checkError(err, "file")
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	bytes := make([]byte, fileSize)

	bufferReader := bufio.NewReader(file)
	bufferReader.Read(bytes)

	return bytes
}

func main() {
	_, _, file := checkParams(os.Args)
	file.Close()

	binary := readFile()
	encode := base64.StdEncoding.EncodeToString(binary)

	udpAddr, err := net.ResolveUDPAddr("udp", ":3200")
	checkError(err, "ResolveUDPAddr")

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err, "ListenUDP")

	conn.Write([]byte(encode))

	os.Exit(0)
}
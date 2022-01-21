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
		fmt.Fprintf(os.Stderr, "Error em " + msg + "\n", err.Error())
		os.Exit(1)
	}
}

func checkParams(args []string) (string, string) {
	if len(args) != 4 {
		fmt.Fprintf(os.Stderr, "Error: Argumentos esperados: <hostname/ip> <porta> <arquivo>")
		os.Exit(1)
	}

	addr, err := net.ResolveIPAddr("ip", args[1]+args[2])
	if err == nil {
		fmt.Fprintf(os.Stderr, "Error: %s não é um hostname/ip válido.\n", args[1])
		os.Exit(1)
	}

	ip := addr.String()
	port := args[2]
	return ip, port
}

func readFile() []byte {
	file, err := os.Open(os.Args[3])
	checkError(err, "Open file")
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	fmt.Println(fileSize)

	bytes := make([]byte, fileSize)

	bufferReader := bufio.NewReader(file)
	bufferReader.Read(bytes)

	return bytes
}

func main() {
	_, port := checkParams(os.Args)

	binary := readFile()
	encode := base64.StdEncoding.EncodeToString(binary)

	udpAddr, err := net.ResolveUDPAddr("udp", port)
	checkError(err, "ResolveUDPAddr")

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err, "ListenUDP")

	conn.Write([]byte(encode))

	os.Exit(0)
}
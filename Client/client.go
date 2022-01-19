package main

import (
	"fmt"
	//"github.com/google/gopacket"
	"net"
	"os"
)

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

func checkError(err error, msg string){
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro em " + msg + "\n", err.Error())
		os.Exit(1)
	}
}

func main() {
	ip, port, file := checkParams(os.Args)
	defer file.Close()
	fmt.Printf("# IP: %s \n", string(ip))
	fmt.Printf("# PORT: %s \n", port)
	fileState, _ := file.Stat()
	fileName := fileState.Name()
	fmt.Printf(fileName)

	udpAddr, err := net.ResolveUDPAddr("udp", port)
	checkError(err, "ResolveUDPAddr")

	conn, err := net.ListenUDP("udp", udpAddr)
	checkError(err, "ListenUDP")

	conn.Write([]byte("Oi"))
	os.Exit(0)
}
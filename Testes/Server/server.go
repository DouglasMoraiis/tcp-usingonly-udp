package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var cont = uint64(0)

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

func DeleteAllFilesParts(){
	for i := uint64(0); i < cont; i++{
		e := os.Remove("file"+strconv.FormatUint(i,10)+".png")
		fmt.Println("conferindo remoção do arquivo ", e)
		if e != nil{
			log.Fatal(e)
		}
	}
}

func handleClient(conn *net.UDPConn, dir string)  {
	var netBuffer [524]byte
	var fileBuffer [524]byte

	size, err := conn.Read(netBuffer[0:])
	fmt.Println(size)
	checkError(err, "Read")

	_, err = base64.StdEncoding.Decode(fileBuffer[0:size], netBuffer[0:size])
	checkError(err, "Decode")

	dirFile := strings.TrimSpace(dir) + ("file" + strconv.FormatUint(cont, 10) + ".png")
	cont++
	err = ioutil.WriteFile(dirFile, fileBuffer[0:size], 0666)
	checkError(err, "WriteFile")

	if size < 200 {
		unionFiles, err := os.Create("unionFiles.png")
		if err != nil{
			log.Fatal(err)
		}
		defer unionFiles.Close()

		for i := uint64(0); i < cont; i++{
			func(){
				file, err := os.Open("file" + strconv.FormatUint(i,10) +".png")
				if err != nil{
					log.Fatal(err)
				}
				defer file.Close()

				ArquivoFinalEuAcho, err := io.Copy(unionFiles, file)
				if err != nil{
					log.Fatal(err)
				}
				fmt.Println("arquivo final eu acho bytes: %d\n", ArquivoFinalEuAcho)
			}()
		}
		fmt.Println("UnionFiles.txt bytes: %d\n", unionFiles)

		DeleteAllFilesParts()
		os.Exit(0)
	}
	//os.Exit(0)
}

func main() {
	port, dir := checkParams(os.Args)

	udpAddr, _ := net.ResolveUDPAddr("udp", port)

	conn, _ := net.ListenUDP("udp", udpAddr)

	for {
		handleClient(conn, dir)
	}
}

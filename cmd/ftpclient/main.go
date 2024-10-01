package main

import (
	"fmt"
	"os"

	"github.com/bigbr41n/GFTP-client/pkg/ftp"
)



func main() {
	//parse arguments passed
	if len(os.Args) != 3 {
		fmt.Println("Usage: gftp <host> <port>")
        return
	}

	host := os.Args[1]
	port := os.Args[2]

	//create a new FTP client
	client := ftp.NewClient(host, port)

    //connect to the FTP server
    conn, err := client.Dial()
    if err != nil {
        fmt.Printf("Error connecting to FTP server: %v\n", err)
        return
    }

    client.HandleConnection(conn)
}
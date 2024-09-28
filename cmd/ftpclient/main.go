package main

import (
	"fmt"
	"net"
	"os"
)



func main() {
	//parse arguments passed
	if len(os.Args) != 3 {
		fmt.Println("Usage: gftp <host> <port>")
        return
	}

	fmt.Println(os.Args[1], os.Args[2])
	host := fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]);

	//connect to FTP server
	conn , err := net.Dial("tcp", host);
	if err!= nil {
        fmt.Printf("Error connecting to FTP server: %v\n", err)
        return
    }

	//read server response
	buf := make([]byte, 1024)
    conn.Read(buf)
    fmt.Println(string(buf))

    //close connection
    conn.Close()
}
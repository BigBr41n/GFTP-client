package ftp

import (
	"fmt"
	"net"

	"github.com/bigbr41n/GFTP-client/pkg/command"
)


type FTPClient struct {
	host string
	port string
}

type FTPClientInterface interface {
	Dial() (net.Conn, error)
	HandleConnection(conn net.Conn)
}

// NewClient returns a new instance of FTPClient implementing the FTPClientInterface
func NewClient(host, port string) FTPClientInterface {
	return &FTPClient{
		host: host,
		port: port,
	}
}

// Dial connects to the FTP server and returns the connection
func (c *FTPClient) Dial() (net.Conn, error) {
	// initiate connection
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", c.host, c.port))

	if err != nil {
		return nil, fmt.Errorf("error connecting to FTP server: %v", err)

	}

	// return the connection
	return conn, nil
}

// HandleConnection reads responses from the FTP server
func (c *FTPClient) HandleConnection(conn net.Conn) {
	handler := command.NewCommandsHandler(conn)
	handler.HandleCommands()
}

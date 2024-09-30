package command

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)




type FTPClient struct {
	conn     net.Conn
} 


type FTPhandler interface {
	HandleCommands() 
}

//Factory Function
func NewCommandsHandler(conn net.Conn) FTPhandler {
	return &FTPClient{
			conn:     conn,
	}
}

//handle Authentication
func (ftpc * FTPClient) auth() {

	//buffer to communicate with server for auth
	buffer := make([]byte,128)

	for {
		n, err := ftpc.conn.Read(buffer)
		if err != nil {
			fmt.Print("GFTP> Authentication Error: ", err)
			return
		}
		msg := string(buffer[:n])
		fmt.Printf("GFTP> %s", msg)
		if msg == "230 User logged in, proceed.\r\n" {
			return 
		}
		
		//read from the Stdin
		val , err := ftpc.readInput()
		if err != nil {
			fmt.Print("GPTP> Error reading input : ", err)
		}

		//send the value to the serve
		ftpc.conn.Write([]byte(val))
	}

}

func (ftpc * FTPClient) HandleCommands() {

	//close the connection 
	defer ftpc.conn.Close()

	//handle Auth
	ftpc.auth()

	// Handle commands in a loop
	buffer := make([]byte, 2048)
	for {
			// Read the command from the client (Stdin)
			fmt.Print("GFTP> ")
			command, err := ftpc.readInput()
			if err != nil {
				fmt.Print("GFTP> 500 Internal server error.\r\n")
				return
			}

			// Process the command
			switch {
				case strings.HasPrefix(command, "PUT"):
						ftpc.handlePUT(strings.TrimSpace(command[3:]))
				case strings.HasPrefix(command, "GET"):
						ftpc.handleGET(strings.TrimSpace(command[3:]))
				case strings.HasPrefix(command, "QUIT"):
						fmt.Print("GFTP> 221 Goodbye.\r\n")
						ftpc.conn.Close()
						return
				default:
						// "\r\n" used to flush the buffer
						ftpc.conn.Write([]byte(command + "\r\n"))
						//read the response from the server and show it to the user
            	        n, err := ftpc.conn.Read(buffer)
            	        if err!= nil {
            	            fmt.Printf("GFTP> Error reading server response: %v\r\n", err)
            	            return
            	        }
            	        fmt.Printf("GFTP>\n%s\r\n", string(buffer[:n]))
			}
	}
}





// PUT uploads a file to the FTP server
func (ftp *FTPClient) handlePUT(localFilePath string) error {
    // Open the local file for reading
    file, err := os.Open(localFilePath)
    if err != nil {
        return fmt.Errorf("error opening local file: %v", err)
    }
    defer file.Close()

    // Get the filename from the local file path
    filename := filepath.Base(localFilePath)

    // Send the PUT command to the server
    _, err = ftp.conn.Write([]byte(fmt.Sprintf("PUT %s\r\n", filename)))
    if err != nil {
        return fmt.Errorf("error sending PUT command: %v", err)
    }

    // Wait for the server's response (optional)
    buf := make([]byte, 1024)
    n, err := ftp.conn.Read(buf)
    if err != nil {
        return fmt.Errorf("error reading server response: %v", err)
    }
    response := string(buf[:n])
    if response != "Ready to receive file "+filename+"...\r\n" {
        return fmt.Errorf("unexpected server response: %s", response)
    }

    // Send the file to the server
    _, err = io.Copy(ftp.conn, file)
    if err != nil {
        return fmt.Errorf("error sending file: %v", err)
    }

    // Wait for the final response from the server
    n, err = ftp.conn.Read(buf)
    if err != nil {
        return fmt.Errorf("error reading server response: %v", err)
    }
    response = string(buf[:n])
    fmt.Println("Server response:", response)

    return nil
}


// handleGET handles the download of a file from the FTP server
func (ftpc *FTPClient) handleGET(filename string) {
	// Get the value of the USER environment variable
	user := os.Getenv("USER")

	// Fallback to "GFTPC" if the USER environment variable is empty
	if user == "" {
		user = "GFTPC"
	}

	// Define the directory path with the fallback
	dirPath := fmt.Sprintf("/home/%s/GFTP", user)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		fmt.Printf("GFTP> Error creating directory: %v\n", err)
		return
	}

	// Change to the newly created or existing directory
	if err := os.Chdir(dirPath); err != nil {
		fmt.Printf("GFTP> Error changing directory: %v\n", err)
		return
	}

	// Ensure the file is created with a unique name if it already exists
	finalFilename := createUniqueFile(filename)

	// Create the file
	file, err := os.Create(finalFilename)
	if err != nil {
		fmt.Print("GFTP> 451 Requested action aborted: local error in processing.\r\n")
		return
	}
	defer file.Close()

	fmt.Print("GFTP> 150 Opening data connection.\r\n")

	//send the GET command request to the server
	_, err = ftpc.conn.Write([]byte(fmt.Sprintf("GET %s\r\n", filename)))
	if err!= nil {
        fmt.Print("GFTP> 451 Requested action aborted: local error in processing.\r\n")
        return
    }
	//file size 
	var size int64
	// read the size of the file sent from the server 
	err = binary.Read(ftpc.conn, binary.LittleEndian, &size)
	if err!= nil {
        fmt.Print("GFTP> 426 Transfer failed.\r\n")
        return
    }

	fmt.Printf("GFTP> File size: %d bytes.\r\n", size)
	// Copy data from the FTP connection to the file with the size sent from the server
	var totalRead int64
	_ , err = io.CopyN(file, ftpc.conn, size-totalRead)
	if err != nil {
		fmt.Print("GFTP> 426 Transfer failed.\r\n")
		return
	}


	fmt.Printf("GFTP> 226 Transfer complete. file path : %s/%s \r\n", dirPath,finalFilename)
}


// createUniqueFile adds a slug to the filename if it already exists
func createUniqueFile(filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	newFilename := filename
	counter := 1

	for {
		// Check if the file already exists
		if _, err := os.Stat(newFilename); os.IsNotExist(err) {
			// File does not exist, we can use this filename
			break
		}

		// File exists, generate a new filename with a counter (e.g., file-1.txt, file-2.txt)
		newFilename = fmt.Sprintf("%s-%d%s", base, counter, ext)
		counter++
	}

	return newFilename
}


// readInput reads a line of input from standard input
func (ftpc *FTPClient) readInput() (string, error) {
	reader := bufio.NewReader(os.Stdin) 
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

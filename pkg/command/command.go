package command

import (
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




func NewCommandsHandler(conn net.Conn) FTPhandler {
	return &FTPClient{
			conn:     conn,
	}
}



func (ftpc * FTPClient) HandleCommands() {

	buffer := make([]byte, 1024)
	for {
			// Read the command from the client
			n, err := ftpc.conn.Read(buffer)
			if err != nil {
				fmt.Print("ftp > Error reading command.\r\n")
			}

			command, err := ftpc.readInput(fmt.Sprintf("%s ", buffer[:n]))
			if err != nil {
				fmt.Print("ftp > 500 Internal server error.\r\n")
				return
			}

			// Process the command
			switch {
			case strings.HasPrefix(command, "PUT"):
					ftpc.handlePUT(strings.TrimSpace(command[3:]))
			case strings.HasPrefix(command, "GET"):
					ftpc.handleGET(strings.TrimSpace(command[3:]))
			case strings.HasPrefix(command, "QUIT"):
					fmt.Print("ftp <> 221 Goodbye.\r\n")
					ftpc.conn.Close()
		return
			default:
					fmt.Print("500 Unknown command.\r\n")
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
		fmt.Printf("ftp > Error creating directory: %v\n", err)
		return
	}

	// Change to the newly created or existing directory
	if err := os.Chdir(dirPath); err != nil {
		fmt.Printf("ftp > Error changing directory: %v\n", err)
		return
	}

	// Ensure the file is created with a unique name if it already exists
	finalFilename := createUniqueFile(filename)

	// Create the file
	file, err := os.Create(finalFilename)
	if err != nil {
		fmt.Print("ftp > 451 Requested action aborted: local error in processing.\r\n")
		return
	}
	defer file.Close()

	fmt.Print("ftp > 150 Opening data connection.\r\n")

	// Copy data from the FTP connection to the file
	if _, err := io.Copy(file, ftpc.conn); err != nil {
		fmt.Print("ftp > 426 Transfer failed.\r\n")
		return
	}

	fmt.Print("ftp > 226 Transfer complete.\r\n")
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


func (ftpc * FTPClient) readInput(prompt string) (string, error) {
	if err := ftpc.writeResponse(prompt); err != nil {
			return "", err
	}

	buffer := make([]byte, 1024)
	n, err := ftpc.conn.Read(buffer)
	if err != nil {
			return "", err
	}

	return strings.TrimSpace(string(buffer[:n])), nil
}

func (ftpc * FTPClient) writeResponse(res string) error {
	_, err := ftpc.conn.Write([]byte(res))
	return err
}
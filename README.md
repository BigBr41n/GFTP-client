# GFTP client

a custom FTP client based on TCP connection.
This FTP client used only to connect with a custom GFTP server (repo in my account) and don't follow any FTP standers (a fully custom FTP client).

## Usage

0. enter the username a password (should be perviously added to the server DB manually or by an API)
1. LS : list all files in the current directory
2. CD : change dir to sub directory
3. PWD: print current directory
4. RM : remove a file remotely
5. DRM: delete a directory & its content remotely
6. MKDIR : create new subdirectory in the current working directory
7. GET : to get a file from the server
8. PUT : to put a file in the server
9. QUIT : quit the server

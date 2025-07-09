package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"strings"
)

const port = ":8080"

const proto = "HTTP/1.1"
const method = "GET"

type status string

const (
	resOK   status = "200 OK"
	resBad  status = "400 Bad Request"
	resNF   status = "404 Not Found"
	resFail status = "500 Server Error"
)

func main() {

	root, err := os.OpenRoot(".")
	if err != nil {
		log.Fatalf("Could not open directory: %v. ERR: %v", ".", err.Error())
	}
	defer root.Close()

	// we have the wd as root
	rootFS := root.FS()

	In, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Could not start listner on: 8080")
	}
	defer In.Close()

	fmt.Printf("Listening on 0.0.0.0%s", port)
	for {
		conn, err := In.Accept()
		if err != nil {
			log.Fatal(err.Error())
		}
		go handle(conn, rootFS) // Request Handler
	}
}

// Request Handler
func handle(conn net.Conn, fsys fs.FS) {

	defer conn.Close()

	req := make([]byte, 1024)
	_, err := conn.Read(req)
	if err != nil {
		conn.Write(getResponse(resBad))
		fmt.Fprint(conn, "Bad Request")
		return
	}

	if len(req) > 1024 {
		conn.Write(getResponse(resBad))
		fmt.Fprint(conn, "Request too large")
		return
	}

	// req to string
	reqParts := strings.Split(string(req), "\r\n")

	if len(reqParts) == 0 {
		conn.Write(getResponse(resBad))
		fmt.Fprint(conn, "Bad Protocol")
		return
	}

	firstLine := reqParts[0]

	// Parse First line
	flParts := strings.Split(firstLine, " ")

	// Check for len
	if len(flParts) != 3 {
		conn.Write(getResponse(resBad))
		return
	}

	// Validate First Line
	if flParts[0] != method || flParts[2] != proto {
		conn.Write(getResponse(resBad))
		return
	}

	// Incase the request arrives from a browser
	// check for favicon.ico
	if flParts[1] == "/favicon.ico" {
		return
	}

	// return the asked directory / file
	filteredPath := strings.Join(strings.Split(flParts[1], "")[1:], "")

	if filteredPath == "" {
		filteredPath = "."
	}

	// Load File
	f, err := fsys.Open(filteredPath)
	if err != nil {
		conn.Write(getResponse(resNF))
		fmt.Fprint(conn, "File Not Fount")
		return
	}
	defer f.Close()

	// Check file type
	ft, err := f.Stat()
	if err != nil {
		conn.Write(getResponse(resNF))
		fmt.Fprint(conn, "File Not Fount")
		return
	}

	// if file is of type file - stream to client
	if !ft.IsDir() {
		// Stream File - Downloadable
		// conn.Write(getResponse(resOK))
		fmt.Fprintf(conn,
			"%s 200 OK\r\nContent-Disposition: attachment;filename=\"%s\"\r\nContent-Length: %d\r\nConnection: close\r\n\r\n",
			proto, ft.Name(), ft.Size(),
		)
		_, err = io.Copy(conn, f)
		if err != nil {
			conn.Write(getResponse(resFail))
			fmt.Fprint(conn, "Server Error")
			return
		}

		return
	}

	// Send file tree for path
	dirInfo, err := fs.ReadDir(fsys, filteredPath)
	if err != nil {
		conn.Write(getResponse(resFail))
		fmt.Fprint(conn, "Server Error")
		return
	}

	// we have the query
	conn.Write(getResponse(resOK))
	fmt.Fprintf(conn, "you asked for %v\n", flParts[1])
	for _, v := range dirInfo {
		fmt.Fprintf(conn, "%v\n", v)
	}

}

// first line builder
func getResponse(status status) []byte {
	return fmt.Appendf(nil, "%s %s\r\n\r\n", proto, status)
}

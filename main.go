package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

var port = ":8080"

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

	if len(os.Args) != 1 {
		a := os.Args[1]
		intA, err := strconv.Atoi(a)
		if err != nil {
			log.Fatal("Port argument must be a number")
		}
		if intA <= 0 || intA > 65535 {
			log.Fatal("Port out or range: 0 - 65535")
		}

		if intA > 0 && intA <= 1023 {
			if os.Getuid() != 0 {
				log.Fatal("user not SUDO, port not allowed")
			}
		}
		port = fmt.Sprintf(":"+"%s", a)
	}

	root, err := os.OpenRoot(".")
	if err != nil {
		log.Fatalf("Could not open directory: %v. ERR: %v", ".", err.Error())
	}
	defer root.Close()

	// we have the wd as root
	rootFS := root.FS()

	In, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Could not start listner on: %s", port)
	}

	// SERVER
	printInfo()
	swg := new(sync.WaitGroup)
	swg.Add(1)
	go server(In, rootFS, swg)

	// Cracefull Exist
	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)
	<-intChan // Interrupt Signal Capture
	err = In.Close()
	if err != nil {
		fmt.Printf("Failed to close listner. ERR: %s", err.Error())
	}
	swg.Wait()
	fmt.Printf("\nServer Shutdown\n")
}

func server(In net.Listener, rootFS fs.FS, servWG *sync.WaitGroup) {
	defer servWG.Done()
	wg := new(sync.WaitGroup)

	for {
		conn, err := In.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			fmt.Printf("Error accepting request: %v", err.Error())
			break
		}
		wg.Add(1)
		go handle(conn, rootFS, wg) // Request Handler
	}
	wg.Wait()
}

// Request Handler
func handle(conn net.Conn, fsys fs.FS, wg *sync.WaitGroup) {

	defer wg.Done()
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
		conn.Write([]byte("HTTP/1.1 204 No Content\r\n\r\n"))
		return
	}

	// return the asked directory / file
	filteredPath := flParts[1]
	filteredPath = strings.TrimPrefix(filteredPath, "/")
	filteredPath = strings.TrimSuffix(filteredPath, "/")

	if filteredPath == "" {
		filteredPath = "."
	}

	// Load File
	f, err := fsys.Open(filteredPath)
	if err != nil {
		conn.Write(getResponse(resNF))
		fmt.Fprint(conn, "Not Fount")
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
		fmt.Printf("[%s] - Sent file: \"%s\", to remote: \"%s\"\n", time.Now().Format(time.DateTime), ft.Name(), conn.RemoteAddr())

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
	fmt.Fprintf(conn, "you asked for %v\n\n", flParts[1])
	fmt.Printf("[%s] - Sent Dir: \"%s\", to remote: \"%s\"\n", time.Now().Format(time.DateTime), ft.Name(), conn.RemoteAddr())
	for _, v := range dirInfo {
		fmt.Fprintf(conn, "%v\n", v)
	}

}

// first line builder
func getResponse(status status) []byte {
	return fmt.Appendf(nil, "%s %s\r\n\r\n", proto, status)
}

// Print Starting info
func printInfo() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// Get Interfaces
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("serving dir: \"%s\" - via HTTP \n\n", wd)
	fmt.Printf("-- http://localhost%s\n", port)
	for _, v := range addrs {
		if ipNet, ok := v.(*net.IPNet); ok {
			if ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
				fmt.Printf("-- http://%s%s\n", ipNet.IP, port)
			}
		}
	}
	fmt.Println()
}

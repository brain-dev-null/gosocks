package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
    listener, err := net.Listen("tcp", "127.0.0.1:8000")

    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }

    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Fatal(err)    
            os.Exit(1)
        }

        reader := bufio.NewReader(conn)
        req := request{"", "", make(map[string]string)}

        for {
            message, _ := reader.ReadString('\n')
            if len(message) == 0 {
                err := conn.Close()
                if err != nil {
                    log.Fatal(err)
                    os.Exit(1)
                }
                os.Exit(0)
            }

            trimmedMessage := strings.TrimSuffix(message, "\r\n")

            if req.method == "" {
                parseFirstLine(&req, &trimmedMessage)
            } else if len(trimmedMessage) > 0 {
                parseHeaderLine(&req, &trimmedMessage)
            } else {
                log.Println("Read complete, answering")
                log.Printf("Request: %+v\n", req)
                writeAnswer(&conn)
                conn.Close()
                os.Exit(0)
            }
        }
    }
}

func parseFirstLine(req *request, line *string) {
    elems := strings.Split(*line, " ")
    method := elems[0]
    path := elems[1]

    req.method = method
    req.path = path
}

func parseHeaderLine(req *request, line *string) {
    elems := strings.Split(*line, ": ")
    headerName := elems[0]
    headerValue := elems[1]
    req.headers[headerName] = headerValue
}

func writeAnswer(conn *net.Conn) {
    writer := bufio.NewWriter(*conn)
    content := "<html><h1>Hello Web</h1></html>"
    _, err := writer.WriteString("HTTP/1.1 200 OK\r\n")
    writer.WriteString("Content-type: text/html\r\n")
    writer.WriteString(fmt.Sprintf("Content-length: %d\r\n", len(content)))
    writer.WriteString("\r\n")
    writer.WriteString(content)
    writer.Flush()

    if err != nil {
        log.Fatal(err)
    }
}

type request struct {
    method string
    path string
    headers map[string]string
}

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
		req := request{"", "", make(map[string]string), make(map[string]string)}

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
				break
			}
		}
		printRequest(&req)
		writeAnswer(&conn)
		conn.Close()
	}
}

func parseFirstLine(req *request, line *string) {
	elems := strings.Split(*line, " ")
	method := elems[0]
	uri := elems[1]

	path, parameters := parseUri(uri)

	req.method = method
	req.path = path
	req.parameters = parameters
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
	method     string
	path       string
	parameters map[string]string
	headers    map[string]string
}

func printRequest(req *request) {
	fmt.Println("=== Request Path ====")
	fmt.Printf("%s %s\n", req.method, req.path)

	fmt.Println("==== Headers ====")
	for headerName, headerValue := range req.headers {
		fmt.Printf("%s: %s\n", headerName, headerValue)
	}

	if len(req.parameters) > 0 {
		fmt.Println("=== Parameters ===")
		for paramName, paramValue := range req.parameters {
			fmt.Printf("%s: %s\n", paramName, paramValue)
		}
	}
}

func parseUri(url string) (string, map[string]string) {
	elements := strings.SplitN(url, "?", 2)
	path := elements[0]

	if len(elements) != 2 {
		return path, make(map[string]string)
	}

	parameterString := elements[1]

	parameters := make(map[string]string)

	for _, param := range strings.Split(parameterString, "&") {
		paramElems := strings.SplitN(param, "=", 2)

		if len(paramElems) != 2 {
			return path, make(map[string]string)
		}

		parameters[paramElems[0]] = paramElems[1]

	}

	return path, parameters
}

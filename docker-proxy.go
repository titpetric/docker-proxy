package main

import "io"
import "fmt"
import "net"
import "time"
import "flag"

func check_strict(e error) {
	if e != nil {
		panic(e)
	}
}

func check_warn(e error) bool {
	if e != nil {
		fmt.Println(e)
		return true
	}
	return false
}

func log(s string, min int, verbose int) {
	if verbose >= min {
		fmt.Println(s);
	}
}
	

func server(port string, verbose int) {
	log(fmt.Sprintf("Started server on port %s", port), 0, verbose);
	conn, err := net.Listen("tcp", ":"+port)
	defer conn.Close()
	check_strict(err)
	for {
		so, err := conn.Accept()
		if check_warn(err) {
			continue
		}
		log("Accepted new connection", 1, verbose);
		go proxyHandler(so, verbose)
	}
}

func proxyHandler(so net.Conn, verbose int) {
	defer log("Closed connection", 1, verbose);
	defer so.Close();
	so_buf := make([]byte, 1024*1024)
	doc_buf := make([]byte, 1024*1024)

	so.SetReadDeadline(time.Now().Add(30 * time.Second));
	so_len, err := so.Read(so_buf)
	if so_len == 0 || err == io.EOF {
		log("Closed connection (timeout)", 1, verbose);
		return;
	}
	check_strict(err)

	log("Opened docker.sock connection for client", 1, verbose);

	doc_socket, err := net.Dial("unix", "/var/run/docker.sock")
	defer doc_socket.Close()

	check_strict(err)
	_, err = doc_socket.Write(so_buf[:so_len])

	log(fmt.Sprintf("Incoming (len=%d): %s", so_len, string(so_buf[:so_len])), 2, verbose);

	total := 0;
	for {
		// Only set a read timeout after we received the first byte. Some requests
		// might be long running and need some time to put together a response.
		if (total > 0) {
			doc_socket.SetReadDeadline(time.Now().Add(1 * time.Second));
		}

		doc_len, err := doc_socket.Read(doc_buf)
		total = total + doc_len;

		if (doc_len > 0) {
			log(fmt.Sprintf("Outgoing (len=%d): %s", doc_len, string(doc_buf[:doc_len])), 2, verbose);
			so.Write(doc_buf[:doc_len])
		}

		if doc_len == 0 || err != nil {
			log(fmt.Sprintf("Closed docker.sock connection for client (total bytes=%d)", total), 1, verbose);
			return;
		}
	}
}

func main() {
	var (
		portPtr = flag.String("p", "9999", "HTTP port to listen to")
		verbose = flag.Int("v", 0, "Verbosity (0=off, 1=log connections, 2=all")
	)
	flag.Parse()
	server(*portPtr, *verbose)
}

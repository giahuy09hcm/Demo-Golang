package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	serverAddr := "192.168.110.23" // Địa chỉ IP của server
	serverPort := "8080"           // Cổng của server

	conn, err := net.Dial("tcp", serverAddr+":"+serverPort)
	if err != nil {
		fmt.Println("Lỗi kết nối:", err)
		return
	}
	defer conn.Close()

	// Gửi yêu cầu đọc file log
	conn.Write([]byte("READ_LOG\n"))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line) // In ra toàn bộ dòng log
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Lỗi đọc từ server:", err)
	}
}

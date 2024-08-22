package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	serverAddr := "192.168.110.23"
	serverPort := "8080"

	conn, err := net.Dial("tcp", serverAddr+":"+serverPort)
	if err != nil {
		fmt.Println("Lỗi kết nối:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Nhập lệnh: ")
		command, _ := reader.ReadString('\n')
		_, err = conn.Write([]byte(command))
		if err != nil {
			fmt.Println("Lỗi gửi lệnh:", err)
			continue
		}

		// Đọc và xử lý kết quả từ server
		messageBuffer := make([]byte, 0)
		for {
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Lỗi đọc từ server:", err)
				break
			}
			messageBuffer = append(messageBuffer, buffer[:n]...)

			// Kiểm tra dấu hiệu kết thúc lệnh (2 ký tự NULL liên tiếp)
			if len(messageBuffer) >= 2 && messageBuffer[len(messageBuffer)-2] == 0 && messageBuffer[len(messageBuffer)-1] == 0 {
				message := string(messageBuffer[:len(messageBuffer)-2])
				// Kiểm tra xem có phải là thông báo lỗi hay không
				if strings.HasPrefix(message, "Lỗi thực thi lệnh") {
					fmt.Println("Lỗi:", message) // In ra thông báo lỗi
				} else {
					fmt.Print(message) // In ra kết quả bình thường
				}
				break
			}
		}
	}
}

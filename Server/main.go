package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Tạo file log
	logFile, err := os.OpenFile("command_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Lỗi mở file log:", err)
		return
	}
	defer logFile.Close()

	// Tạo bufio.Writer để flush dữ liệu ngay lập tức
	logWriter := bufio.NewWriter(logFile)
	defer logWriter.Flush()

	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client đã ngắt kết nối")
			} else {
				fmt.Println("Lỗi đọc từ kết nối:", err)
			}
			return
		}

		fmt.Println("Lệnh nhận được:", message)

		// Kiểm tra nếu lệnh là yêu cầu đọc log
		if message == "READ_LOG\n" {
			// Mở file log ở chế độ theo dõi (tail -f)
			cmd := exec.Command("cmd", "/C", "type command_log.txt & powershell Get-Content command_log.txt -Wait")

			stdout, _ := cmd.StdoutPipe()
			stderr, _ := cmd.StderrPipe()

			if err := cmd.Start(); err != nil {
				conn.Write([]byte("Lỗi mở file log\n"))
				return
			}

			// Sao chép stdout vào conn
			go io.Copy(conn, stdout)

			// Sao chép stderr vào conn (nếu cần)
			go io.Copy(conn, stderr)

			// Đợi lệnh kết thúc (hoặc client ngắt kết nối)
			cmd.Wait()

			return // Thoát khỏi vòng lặp sau khi kết thúc đọc log
		}

		// Kiểm tra nếu lệnh là ping
		isPing := strings.HasPrefix(message, "ping")
		if isPing {
			message = message + " -t" // Thêm tùy chọn -t để ping liên tục
		}

		cmd := exec.Command("cmd", "/C", message) // Windows

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		// Chạy lệnh
		if err := cmd.Start(); err != nil {
			errorMessage := fmt.Sprintf("Lỗi bắt đầu lệnh: %s\n", err)
			conn.Write([]byte(errorMessage))
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)

		// Đọc và log stdout
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				logWriter.WriteString(fmt.Sprintf("[%s] STDOUT: %s\n", time.Now().Format("2006-01-02 15:04:05"), line))
				logWriter.Flush()
				conn.Write([]byte(line + "\n"))
			}
			stdout.Close()
		}()

		// Đọc và log stderr
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				logWriter.WriteString(fmt.Sprintf("[%s] STDERR: %s\n", time.Now().Format("2006-01-02 15:04:05"), line))
				logWriter.Flush()
				conn.Write([]byte(line + "\n"))
			}
			stderr.Close()
		}()

		// Đặt timeout cho lệnh ping để tránh chạy mãi mãi
		if isPing {
			go func() {
				time.Sleep(5 * time.Second) // Timeout sau 5 giây
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			}()
		}

		wg.Wait() // Đợi cho cả stdout và stderr được sao chép xong

		// Đợi lệnh hoàn thành
		err = cmd.Wait()

		// Kiểm tra lỗi và gửi kết quả về cho client
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() != 0 {
					errorMessage := fmt.Sprintf("Lỗi thực thi lệnh (mã %d): %s\n", exitErr.ExitCode(), exitErr)
					conn.Write([]byte(errorMessage))
				} else {
					conn.Write([]byte("Lệnh thực thi thành công (có thể có cảnh báo)\n\n")) // Thêm 2 dòng trống để client tiếp tục nhập
				}
			} else {
				errorMessage := fmt.Sprintf("Lỗi thực thi lệnh: %s\n\n", err)
				conn.Write([]byte(errorMessage))
			}
		} else {
			conn.Write([]byte("Lệnh thực thi thành công\n\n")) // Thêm 2 dòng trống để client tiếp tục nhập
		}
	}
}

func main() {
	port := "8080" // Cổng server

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Lỗi lắng nghe:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Server đang lắng nghe trên cổng", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Lỗi chấp nhận kết nối:", err)
			continue
		}
		go handleConnection(conn)
	}
}

package main

import (
	"bufio"
	"fmt"
	"log"

	"github.com/Microsoft/go-winio"
)

const pipeName = `\\.\pipe\mypipe`

func main() {
	// 名前付きパイプのリスナーを作成
	listener, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Go server is listening...")

	// クライアントからの接続を待機
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Client connected.")

	// 受信
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("Received from C#: " + scanner.Text())
		break
	}

	// 送信
	fmt.Fprintln(conn, "Hello from Go")

	// 受信
	scanner = bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("Received from C#: " + scanner.Text())
		break
	}
}

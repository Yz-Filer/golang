package main

import (
	"bufio"
	"fmt"
	"log"

	"github.com/Microsoft/go-winio"
)

const pipeName = `\\.\pipe\mypipe`

func main() {
	// 名前付きパイプに接続
	conn, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 送信
	fmt.Fprintln(conn, "Hello from Go")

	// 受信
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("Received from C#: " + scanner.Text())
		break
	}

	// 送信
	fmt.Fprintln(conn, "Go is done")

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

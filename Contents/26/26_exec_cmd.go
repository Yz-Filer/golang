package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

func main() {
	//-----------------------------------------------------------------------------
	// コマンドを実行し終了を待つ（stdout/stderr取得なし）
	//-----------------------------------------------------------------------------
	cmd := exec.Command("cmd", "/c", "echo stdout出力 & echo stderr出力 >&2")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("◆Run終了")
	
	
	//-----------------------------------------------------------------------------
	// コマンドを実行し終了を待つ（stderr取得なし）
	//-----------------------------------------------------------------------------
	cmd = exec.Command("cmd", "/c", "echo stdout出力 & echo stderr出力 >&2")
	ret, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("◆Output終了\n%s\n", string(ret))
	
	
	//-----------------------------------------------------------------------------
	// コマンドを実行し終了を待つ
	//-----------------------------------------------------------------------------
	cmd = exec.Command("cmd", "/c", "echo stdout出力 & echo stderr出力 >&2")
	ret, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("◆CombinedOutput終了\n%s\n", string(ret))
	
	
	//-----------------------------------------------------------------------------
	// stdinで渡した文字列を入力としたコマンドを実行し終了を待つ
	//-----------------------------------------------------------------------------
	cmd = exec.Command("cmd", "/c", "more & echo stderr出力 >&2")
	cmd.Stdin = bytes.NewBufferString("stdin入力からstdoutへ出力\n")
	
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("◆Run（標準入出力、エラー出力）終了")
	fmt.Print(strings.Trim(stdoutBuf.String(), "\n"))
	fmt.Println(stderrBuf.String())


	//-----------------------------------------------------------------------------
	// pipeを使ってstdinで渡した文字列を入力としたコマンドを実行し終了を待つ
	//-----------------------------------------------------------------------------
	cmd = exec.Command("cmd", "/c", "more & echo stderr出力 >&2")

	// StdinPipe を取得
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	// StdoutPipe を取得
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// StderrPipe を取得
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	// コマンドの実行を開始 (非同期)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	// StdinPipe にデータを書き込む (goルーチンで実行してデッドロックを防ぐ)
	go func() {
		defer stdinPipe.Close()
		writer := bufio.NewWriter(stdinPipe)
		_, err := writer.WriteString("stdin入力からstdoutへ出力\n")
		if err != nil {
			log.Fatal(err)
		}
		err = writer.Flush()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// StdoutPipe からの出力を読み取る
	var stdoutBuf2 bytes.Buffer
	_, err = io.Copy(&stdoutBuf2, stdoutPipe)
	if err != nil {
		log.Fatal(err)
	}

	// StderrPipe からのエラー出力を読み取る
	var stderrBuf2 bytes.Buffer
	_, err = io.Copy(&stderrBuf2, stderrPipe)
	if err != nil {
		log.Fatal(err)
	}

	// コマンドの終了を待つ
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("◆Start（標準入出力、エラー出力）終了")
	fmt.Print(strings.Trim(stdoutBuf2.String(), "\n"))
	fmt.Println(stderrBuf2.String())

}

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	WithCancelCause()
	WithDeadlineCause()
	WithTimeoutCause()
	WithTimeoutCause2()
	AfterFunc()
}

// WithCancelCauseを使ったサンプル
func WithCancelCause() {
	// cancel関数実行時に中断するコンテキスト
	ctx, cancel := context.WithCancelCause(context.Background())

	go func() {
		// 外部コマンドを実行
		cmd := exec.CommandContext(ctx, "cmd", "/c", "start", "/WAIT", "timeout", "/T", "3", "/NOBREAK")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("コマンド実行エラー: %v\n出力: %s", err, string(output))
		}
	}()
	
	// 1秒後に中断
	time.Sleep(1 * time.Second) 
	cancel(errors.New("canceled by CancelCauseFunc"))
	
	// 3秒後に中断理由を表示
	time.Sleep(3 * time.Second)
	fmt.Println(context.Cause(ctx))
}

// WithDeadlineCauseを使ったサンプル
func WithDeadlineCause() {
	// 「現在時間 + 1秒」の時刻に中断するコンテキスト
	// ※第3引数に指定してるerrorが中断理由になる
	ctx, cancel := context.WithDeadlineCause(context.Background(), time.Now().Add(1 * time.Second), errors.New("canceled by CancelCauseFunc"))
	defer cancel()

	go func() {
		// 外部コマンドを実行
		cmd := exec.CommandContext(ctx, "cmd", "/c", "start", "/WAIT", "timeout", "/T", "3", "/NOBREAK")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("コマンド実行エラー: %v\n出力: %s", err, string(output))
		}
	}()
	
	// 3秒後に中断理由を表示
	time.Sleep(3 * time.Second)
	fmt.Println(context.Cause(ctx))
}

// WithTimeoutCauseを使ったサンプル
func WithTimeoutCause() {
	// 1秒後に中断するコンテキスト
	// ※第3引数に指定してるerrorが中断理由になる
	ctx, cancel := context.WithTimeoutCause(context.Background(), 1 * time.Second, errors.New("canceled by CancelCauseFunc"))
	defer cancel()

	go func() {
		// 外部コマンドを実行
		cmd := exec.CommandContext(ctx, "cmd", "/c", "start", "/WAIT", "timeout", "/T", "3", "/NOBREAK")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("コマンド実行エラー: %v\n出力: %s", err, string(output))
		}
	}()
	
	// 3秒後に中断理由を表示
	time.Sleep(3 * time.Second)
	fmt.Println(context.Cause(ctx))
}

// AfterFuncを使ったサンプル
func AfterFunc() {
	// 1秒後に中断するコンテキスト
	ctx, cancel := context.WithTimeoutCause(context.Background(), 1 * time.Second, errors.New("canceled by CancelCauseFunc"))
	defer cancel()
	
	// 1秒後に中断した後、"run after func"を表示
	// ※中断前にstopf()を実行した場合、第2引数の関数は実行されない
	// ※中断せずに外部コマンド実行が完了した後、stopf()を実行した場合、第2引数の関数は実行されない
	stopf := context.AfterFunc(ctx, func() {
		fmt.Println("run after func")
	})
	defer stopf()

	go func() {
		// 外部コマンドを実行
		cmd := exec.CommandContext(ctx, "cmd", "/c", "start", "/WAIT", "timeout", "/T", "3", "/NOBREAK")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("コマンド実行エラー: %v\n出力: %s", err, string(output))
		}
	}()
	
	// 3秒後に中断理由を表示
	time.Sleep(3 * time.Second)
	fmt.Println(context.Cause(ctx))
}

// WithTimeoutCause時に外部コマンドを強制終了させるサンプル
func WithTimeoutCause2() {
	// 1秒後に中断するコンテキスト
	// ※第3引数に指定してるerrorが中断理由になる
	ctx, cancel := context.WithTimeoutCause(context.Background(), 1 * time.Second, errors.New("canceled by CancelCauseFunc"))
	defer cancel()

	go func() {
		// 外部コマンドを実行
		cmd := exec.CommandContext(ctx, "cmd", "/c", "start", "/WAIT", "timeout", "/T", "3", "/NOBREAK")
		
		// 中断時に自動的にコールされるcmd.Cancel関数で子プロセスも含めてkillする
		cmd.Cancel = func() error {
			pidStr := strconv.Itoa(cmd.Process.Pid)
			killCmd := exec.Command("taskkill", "/PID", pidStr, "/F", "/T")
			return killCmd.Run()
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("コマンド実行エラー: %v\n出力: %s", err, string(output))
		}
	}()
	
	// 3秒後に中断理由を表示
	time.Sleep(3 * time.Second)
	fmt.Println(context.Cause(ctx))
}


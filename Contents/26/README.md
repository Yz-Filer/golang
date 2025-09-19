[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 26. 外部コマンドの実行方法メモ  

外部コマンドの実行方法を自分用の備忘として記載しておきます。  
尚、`exec.CommandContext`については前の章で使ってるので説明は省略します。  

## 26.1 Cmd構造体の作成  

実行ファイルと引数を指定してCmd構造体を取得するコードは以下のようになります。  

```go
cmd := exec.Command("cmd", "/c", "echo stdout出力 & echo stderr出力 >&2")
```

上記コードは標準出力と標準エラー出力に文字列を表示するコマンドを設定した例になります。  

## 26.2 コマンドを実行し終了を待つ（stdout/stderr取得なし）  

コマンドを実行して終了を待つコードは以下のようになります。  

```go
cmd := exec.Command("cmd", "/c", "echo stdout出力 & echo stderr出力 >&2")
err := cmd.Run()
if err != nil {
	log.Fatal(err)
}
fmt.Println("◆Run終了")
```

`Run`を実行していますが、標準出力も標準エラー出力も取得することはできません。  

## 26.3 コマンドを実行し終了を待つ（stdout取得あり/stderr取得なし）  

`Run`と比較して標準出力の取得が出来るコードは以下のようになります。  

```go
cmd = exec.Command("cmd", "/c", "echo stdout出力 & echo stderr出力 >&2")
ret, err := cmd.Output()
if err != nil {
	log.Fatal(err)
}
fmt.Printf("◆Output終了\n%s\n", string(ret))
```

`ret`に標準出力が格納されてます。  

## 26.4 コマンドを実行し終了を待つ（stdout/stderr取得あり）  

標準出力と標準エラー出力を統合して取得が出来るコードは以下のようになります。  

```go
cmd = exec.Command("cmd", "/c", "echo stdout出力 & echo stderr出力 >&2")
ret, err = cmd.CombinedOutput()
if err != nil {
	log.Fatal(err)
}
fmt.Printf("◆CombinedOutput終了\n%s\n", string(ret))
```

`ret`に標準出力と標準エラー出力が足された物が格納されてます。  

## 26.5 stdinで渡した文字列を入力としたコマンドを実行し終了を待つ（stdout/stderr取得あり）

標準入力からデータを渡し、標準出力と標準エラー出力を別々に取得することが出来るコードは以下のようになります。  

```go
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
```

このコードは全てのデータを標準入力に一括で渡し、全てのデータを標準出力と標準エラー出力から一括で取得するコードになります。  

## 26.6 pipeを使ってstdinで渡した文字列を入力としたコマンドを実行し終了を待つ（stdout/stderr取得あり）  

標準入力からpipeへデータを渡し、標準出力と標準エラー出力をpipeから別々に取得することが出来るコードは以下のようになります。

```go
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
```

`Start`でcmdを実行していますが、`Run`と違い`Start`は終了を待ちません。cmd実行後に標準入力にデータを流していきますが、pipeで渡しながらPipeから受け取るため、goルーチンを使っています。  
受信側も非同期処理が必要な場合は、標準出力や標準エラー出力もgoルーチンにするかどうかの検討が必要となります。  

`stdinPipe.Close()`が行われるまで、pipeからの入力を待ち続けるため、忘れずにCloseして下さい。  
`Wait`で処理が終わるのを待ち、処理終了後に受信したデータを表示しています。  

## 26.7 おわりに  

外部コマンドの実行方法についての説明は以上となります。  
作成したファイルは、
[ここ](26_exec_cmd.go)
に置いてます。  

> [!TIP]  
> 外部コマンド起動時に一瞬コンソール画面が表示される場合は、cmdに以下の設定をすれば解消します。
> ```go
> cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
> ```
> 
> 新しいDOS窓を開いてコマンドを実行し、DOS窓を閉じたくない場合は、以下のようになります。  
> ```go
> exec.Command("cmd", "/C", "start", "cmd", "/K", "dir")
> ```

<br>

「[27. プロセス間通信（名前付きパイプ）のメモ](../27/README.md)」へ

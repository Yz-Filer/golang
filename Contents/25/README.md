[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 25. コンテキストの使い方メモ  

外部コマンドの実行方法を備忘として追加しようとしていたのですが、コンテキストを使う関数があったので、先にコンテキストをテーマにしようと思います。  
コンテキストは、WEB検索すると色々難しい事が書いてある場合が多いので、使い方に絞った紹介をしていきます。  

詳しい方からするとズレてるかもしれませんが、コンテキストは、サーバー接続のタイムアウトなど、コンテキストに対応した処理を中断するために使えます。但し、子プロセスがある外部コマンドなどのように強制中断が出来ない物は、最後まで実行されるけど異常終了（中断）のような扱いになるようです。  
子プロセスがある外部コマンドの中断方法も「25.6」で紹介します。  

## 25.1 空のコンテキストを作成  

空のコンテキスト作成は、`ContextNew()`とかではなく、以下のようなコードになるようです。  

```go
ctx := context.Background()
```

> [!TIP]  
> 今回は紹介しませんが、key/value形式でコンテキスト内に値を保持する  
> `func WithValue(parent Context, key, val any) Context`  
> や、親が中断されても中断されない  
> `func WithoutCancel(parent Context) Context`  
> などの関数もあります。  

## 25.2 コマンドで中断可能にする：WithCancel/WithCancelCause  

`WithCancelCause`を使ったコードは以下のようなコードになります。  
`WithCancel`と`WithCancelCause`の違いは、中断時にerrorを渡せるかどうかとなります。errorを渡すことで、どこで中断されたかを把握出来るようになります。  

```go
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
```

goルーチンで3秒間待機するコマンドを実行した1秒後に中断しています。但し、実行中の処理は「start」を使ってるため3秒経過するまで継続し、強制終了はしないようでした。  
最後に`context.Cause(ctx)`にて、中断時に指定したerror内容を取得しています。  

`WithCancel`と`WithCancelCause`は、サーバー接続時などにボタン押下にて接続を中断するような場合に使えそうです。  

> [!TIP]  
> - 最後に3秒待ってますが、実際にはチャネルを使って待ち受けることになるかと思います。  
> - 親コンテキストが中断された場合、子のコンテキストも中断されます。  

## 25.3 時刻指定で中断する：WithDeadline/WithDeadlineCause  

`WithDeadlineCause`を使ったコードは以下のようなコードになります。  
`WithDeadline`と`WithDeadlineCause`の違いは「25.2」と同様です。  

```go
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
```

`time.Now().Add(1 * time.Second)`にて現在時刻の1秒後に中断するように設定しています。中断時に渡すerrorも`WithDeadlineCause`の引数で指定しています。  
`cancel()`は実行する必要があるようなので`defer`で最後に実行するようにしています。もちろん、`cancel()`を使って中断することもできます。  

## 25.4 指定時間経過後に中断する：WithTimeout/WithTimeoutCause  

`WithTimeoutCause`を使ったコードは以下のようなコードになります。  
`WithTimeout`と`WithTimeoutCause`の違いは「25.2」と同様です。  

```go
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
```

`WithDeadlineCause`とほぼ同じで、違いは中断時刻を指定するか、経過時間を指定するかの部分のみとなります。  

## 25.5 中断後の処理を指定する：AfterFunc  

中断後の処理を予め指定するコードは以下のようなコードになります。  

```go
// 1秒後に中断するコンテキスト
ctx, cancel := context.WithTimeoutCause(context.Background(), 1 * time.Second, errors.New("canceled by CancelCauseFunc"))
defer cancel()

// 1秒後に中断した後、"run after func"を表示
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
```

`WithTimeoutCause`を使って1秒後に中断させています。`AfterFunc`関数にて中断後に実行する関数を定義しています。  

`AfterFunc`の戻り値`stopf`は、`AfterFunc`で登録した関数を解除する時に実行する関数となります。中断前/中断後/中断なしのそれぞれの挙動は以下のようになります。  
- 中断前：stopf()を実行した場合、第2引数の関数は実行されなくなる  
- 中断後：第2引数の関数は実行されているため、stopf()を実行しても何もされない  
- 中断なし：第2引数の関数は実行されないため、stopf()を実行しても何もされない  

`stopf`の戻り値は`bool`型で、`AfterFunc`の第2引数で指定した関数を解除した時（中断前）のみtrueを返します。  

`AfterFunc`の第2引数で指定した関数が実行中かどうかの検知方法は用意されてません。  
`AfterFunc`実行時に中断済みであった場合、第2引数の関数は即時実行されます。  
1つのコンテキストに`AfterFunc`を複数実行した場合、第2引数の関数はそれぞれ独立して動きます。  

> [!CAUTION]  
> `AfterFunc`の第2引数で指定した関数は、goルーチンで実行されます。gotk3を使ったUI操作（中断のダイアログ表示）などを行う場合は、`glib.IdleAdd()`を使うなどの検討が必要となります。  

## 25.6 子プロセスがある外部コマンド中断時に強制終了させる方法  

`exec.CommandContext`限定になりますが、コンテキストによる中断時に強制終了させる方法を紹介します。  
子プロセスがある外部コマンドを中断する場合のコードは以下のようになります。  

```go
cmd := exec.CommandContext(ctx, "cmd", "/c", "start", "/WAIT", "timeout", "/T", "3", "/NOBREAK")
cmd.Cancel = func() error {
	pidStr := strconv.Itoa(cmd.Process.Pid)
	killCmd := exec.Command("taskkill", "/PID", pidStr, "/F", "/T")
	return killCmd.Run()
}
```

goルーチン内の外部コマンドを定義した後に、`cmd.Cancel`関数にプロセスをkillする関数を設定します。コンテキストにより中断された時に、`cmd.Cancel`が自動的に実行されます。  
プロセスのkillは「taskkill /T」コマンドを使って、子プロセスも含めてkillするように設定します。  

## 25.7 おわりに  

コンテキストの使い方の説明は以上となります。  
作成したファイルは、
[ここ](25_context.go)
に置いてます。  

</br>

「[26. 外部コマンドの実行方法メモ](../26/README.md)」へ

# 14. カスタムシグナル

メインウィンドウと編集ウィンドウ・付箋ウィンドウ間で非同期に連携する必要があるため、カスタムシグナルを使いたいと思います。  

gotk3では、ボタンクリックなどのシグナルを`Connect()`で待ち受けたりしてるので、馴染みがある機能だと思います。gotk3の古いVerでは引数を渡すことが出来なかったため、使い道が限定されていましたが、現在のVerでは引数を渡すことが出来るようになったため、使い勝手が良くなりました。  

## 14.1 引数なし、戻り値なしのシグナル  

まずは、引数なし、戻り値なしのサンプルです。  

```go
// シグナルの作成
_, err = glib.SignalNew("test1")
if err != nil {
	log.Fatal(err)
}

// シグナル受信側
window.Connect("test1", func() {
})
	：
	：
// シグナル送信側
ret, err := window.Emit("test1", glib.TYPE_POINTER)
```

"test1"という名前のシグナルを作成し、`Emit()`で送信してます。  
受信側は`Connect()`で待ち受けます。  
gotk3のドキュメントには、戻り値なしの場合、Emit()の第2引数は`glib.TYPE_NONE`にするよう記載されてる物もありますが、戻り値なしの場合`nil`が返ってきますのでエラーになります。そのため、第2引数に`glib.TYPE_POINTER`を指定してます。  

## 14.2 引数なし、戻り値ありのシグナル  

次は、引数なし、戻り値ありのサンプルです。  

```go
// シグナルの作成
_, err = glib.SignalNew("test2")
if err != nil {
	log.Fatal(err)
}

// シグナル受信側
window.Connect("test2", func() int {
	return 200
})
	：
	：
// シグナル送信側
ret, err := window.Emit("test2", glib.TYPE_INT)
```

シグナル作成部分は同じです。シグナル受信部分は、コールバック関数の戻り値の型を`int`に設定して、`return 200`で値を戻してます。  
シグナル送信側は、第2引数に`glib.TYPE_INT`を指定してシグナルを送信してます。`ret`には200が代入されます。  

## 14.3 引数あり、戻り値あり  

最後に、引数あり、戻り値ありのサンプルです。  

```go
// シグナルの作成
_, err = glib.SignalNewV("test3", glib.TYPE_BOOLEAN, 2, glib.TYPE_INT, glib.TYPE_STRING)
if err != nil {
	log.Fatal(err)
}

// シグナル受信側
window.Connect("test3", func(win *gtk.ApplicationWindow, arg1 int, arg2 string) bool {
	fmt.Println(arg1, arg2)
	return true
})
	：
	：
// シグナル送信側
ret, err := window.Emit("test3", glib.TYPE_BOOLEAN, 1000, "--arg2--")
```

シグナルの作成では、引数を渡すため`glib.SignalNew()`ではなく`glib.SignalNewV()`を使ってます。第2引数が戻り値の型、第3引数が引数の数、第4引数以降が引数の型となります。  
シグナル受信側は、第1引数にオブジェクトの型（サンプルではwindowの型）、第2引数以降はシグナルから送信されてくる引数になります。コールバック関数の戻り値の型は、シグナル作成時に指定した`bool`型を設定してます。  
シグナル送信側は、第2引数に戻り値の型、第3引数以降にシグナルで渡す引数を設定します。上記サンプルの場合`ret`には`true`が代入されます。  

今回作成する付箋アプリは、シグナルの引数に付箋のKeyを渡すことで、メインウィンドウ上の表の更新を行います。  

## 14.4 「[4.4 マルチスレッドで気をつけるところ](../04/README.md#44-%E3%83%9E%E3%83%AB%E3%83%81%E3%82%B9%E3%83%AC%E3%83%83%E3%83%89%E3%81%A7%E6%B0%97%E3%82%92%E3%81%A4%E3%81%91%E3%82%8B%E3%81%A8%E3%81%93%E3%82%8D)」の問題は改善する？  

goルーチン内でUI操作を行うとgotk3アプリはクラッシュすると説明しましたが、シグナルを使う事で改善するかどうか考えてみました。  
goルーチン内でUI操作を行わなければ良いので、`Emit()`がgoルーチン内で安全に実行できればシグナルと組み合わせることで問題がなくなるのではないかと思いGeminiに聞いてみました。  

> Emit()はUI操作ではないため、glib.IdleAdd()は不要という認識で正しいです。  
> - Emit()は、単にシグナルを発行し、登録されたシグナルハンドラを呼び出すだけです。  
> - Emit()自体は、UIウィジェットの更新や変更などのUI操作を行いません。  
> 
> したがって、GoルーチンからEmit()しても、スレッドセーフの問題は発生しません。  

```go
// シグナルの作成
_, err = glib.SignalNew("test4")
if err != nil {
	log.Fatal(err)
}

// シグナル受信側
window.Connect("test4", func() bool {
	// ここでモーダルダイアログを表示
	return true
})
	：
	：
// シグナル送信側
go func() {
    ret, err := window.Emit("test4", glib.TYPE_BOOLEAN)
	fmt.Println(ret, err)
}()
```

試してみた所、モーダルダイアログを表示した瞬間フリーズしました。試しに、`Emit()`を`glib.IdeleAdd()`の中に入れてみた所問題なく動作しました。  
再度Geminiに聞いてみました。  

> シグナルハンドラ内でモーダルダイアログを表示する場合、ゴルーチン内で Emit() を呼び出す際には glib.IdleAdd() を使用する必要があります。  

ということで、`Emit()`もUI操作と同様に考える必要がありそうです。結論としては、goルーチン内でUI操作を行うためには`glib.IdeleAdd()`が必須であり、「4.4章」で記載した内容はシグナルを使っても改善しそうにないことが分かりました。

</br>

「[15. CSSを使った書式設定](../15/README.md)」へ

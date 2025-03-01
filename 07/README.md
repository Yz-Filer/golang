# 7. メッセージダイアログを表示したい
ここまでで、ほぼほぼ「Hello World!」から進化してないので、ダイアログを表示するアプリを作成してみます。  
今回は、以下の３つに分類して進めていきます。  
- gtk標準のメッセージダイアログ（モーダルダイアログ）
- 自作したカスタムメッセージダイアログ（モーダルダイアログ）
- カスタムメッセージダイアログを使ったモードレスダイアログ

![](image/window.jpg)  

あと、ステータスバーも追加してます。  

![](image/glade_window.jpg)  

作成したファイルは、
[ここ](glade/07_MainWindow.glade)
に置いてます。  

## 7.1 標準メッセージダイアログ
特徴についてGeminiに聞いてみました。  

> - 作成時の戻り値でエラーを返さない。  
> - OSやデスクトップ環境によって、ボタンの配置（YES/NOなど）が逆になることや、テキストが英語から日本語になることがある。  
> - Windows環境では、アイコンが表示されない。  
> - カスタムボタンの追加も可能。（`dialog.AddButton("Custom Button", gtk.RESPONSE_YES)`）  
> - 通常はモーダルダイアログとして表示され、親ウィンドウの操作をブロックする。

メッセージダイアログはエラー表示などに使うので、作成が失敗することがない前提なのでしょうか？  

> メッセージダイアログはエラー表示など、ユーザーに何らかの情報を伝えるために使用されるものであり、作成が失敗することは稀であるという前提があります。そのため、GTK3のメッセージダイアログは、作成時の戻り値でエラーを返す仕組みになっていません。

> [!NOTE]
> あくまでもGeminiの回答となります。開発者の考えがどうなのかは分かりません。

アイコンが表示されないことと、OSによってボタン配置やボタンラベルが変わるというのは残念な感じがしました。  
ボタンについては、OS標準の操作性を維持するためということで仕方がないようです。  
作成したコードは以下になります。  

```go
dialog := gtk.MessageDialogNew(window1, gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_INFO, gtk.BUTTONS_OK_CANCEL, "標準メッセージダイアログ")
defer dialog.Destroy()

dialog.SetTitle ("タイトル")
dialog.FormatSecondaryText("このダイアログはgtk3標準のメッセージダイアログです")

ret := dialog.Run()

// 標準メッセージダイアログの応答処理
switch ret {
	case gtk.RESPONSE_OK:
		log.Println("標準メッセージダイアログで、OKが押されました")
	case gtk.RESPONSE_CANCEL:
		log.Println("標準メッセージダイアログで、CANCELが押されました")
	case gtk.RESPONSE_DELETE_EVENT:
		log.Println("標準メッセージダイアログが閉じられました")
}
```

`MessageDialogNew()`の引数に指定している項目の意味は以下になります。  

| 引数の場所 | 引数値 | 説明 |
| --- | --- | --- |
| 2 | DIALOG_MODAL | モーダルダイアログを指定 |
|  | DIALOG_DESTROY_WITH_PARENT | 親と一緒に破棄（モーダルダイアログでは親だけ閉じれないので不要な筈） |
| 3 | MESSAGE_INFO | 電球型アイコン |
|  | MESSAGE_WARNING | 「！」型アイコン |
|  | MESSAGE_QUESTION | 「？」型アイコン |
|  | MESSAGE_ERROR | 通行止め型アイコン |
|  | MESSAGE_OTHER | アイコンなし |
| 4 | BUTTONS_NONE | ボタンなし |
|  | BUTTONS_OK | OKボタン |
|  | BUTTONS_CLOSE | CLOSEボタン |
|  | BUTTONS_CANCEL | CANCELボタン |
|  | BUTTONS_YES_NO | YESとNOのボタン |
|  | BUTTONS_OK_CANCEL | OKとCANCELのボタン |

`dialog.Run()`の戻り値は以下になります。  

| 戻り値 | 説明 |
| --- | --- |
| RESPONSE_NONE | ボタン押下以外で終了した時用 |
| RESPONSE_REJECT | `AddButton()`で追加したボタン押下 |
| RESPONSE_ACCEPT | `AddButton()`で追加したボタン押下 |
| RESPONSE_DELETE_EVENT | 右上の「×」ボタン押下 |
| RESPONSE_OK | OKボタン押下 |
| RESPONSE_CANCEL | CANCELボタン押下 |
| RESPONSE_CLOSE | CLOSEボタン押下 |
| RESPONSE_YES | YESボタン押下 |
| RESPONSE_NO | NOボタン押下 |
| RESPONSE_APPLY | `AddButton()`で追加したボタン押下 |
| RESPONSE_HELP | `AddButton()`で追加したボタン押下 |

## 7.2 カスタムメッセージダイアログ
やっぱりアイコンを表示したいので、メッセージダイアログを自作してみました。  
![](image/glade_dialog.jpg)  
レスポンスを返す全てのボタンを並べてます。あと、アイコン表示用のImageと処理中表示用のSpinnerも追加しました。  
作成したファイルは、
[ここ](glade/07_DIALOG.glade)
に置いてます。  

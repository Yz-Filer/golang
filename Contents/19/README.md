[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 19. ディレクトリ配下の更新を監視したい  

指定したディレクトリ配下のファイルやサブディレクトリに更新があったかどうかを監視する方法を説明していきます。  
ディレクトリ更新の監視は以下のような流れとなります。  

- ディレクトリのファイルハンドルを取得  
- `ReadDirectoryChangesW()`で監視登録と更新検知時に実行するコールバック関数を登録  
- 一度検知すると監視が終了してしまうので、再度監視を登録  
- ファイルハンドルをクローズして監視を終了  

> [!NOTE]  
> 指定したディレクトリ自体の更新は検知出来ないため、ディレクトリ自体の更新を検知するためには、親ディレクトリを監視対象にする必要があります。  

## 19.1 ディレクトリのファイルハンドルを取得  

ファイルハンドルの取得は以下のようなコードになります。  

```go
targetDir := `D:\test`
targetDirUTF16Ptr := windows.StringToUTF16Ptr(targetDir)
fHandle, w32err := win32.CreateFile(
	targetDirUTF16Ptr,
	win32.FILE_LIST_DIRECTORY,
	win32.FILE_SHARE_READ | win32.FILE_SHARE_WRITE | win32.FILE_SHARE_DELETE,
	nil,
	win32.OPEN_EXISTING,
	win32.FILE_FLAG_BACKUP_SEMANTICS | win32.FILE_FLAG_OVERLAPPED,
	0)
if fHandle == 0 || w32err != win32.NO_ERROR {
	return fmt.Errorf("CreateFileの失敗")
}
```

> [!CAUTION]  
> 実際にはディレクトリの存在確認や、Create済かどうかなどの確認も必要です。  

## 19.2 監視登録と更新検知時に実行するコールバック関数を登録  

```go
bufferSize := 4096
buffer := make([]byte, bufferSize)
overlapped = win32.OVERLAPPED {}
var bytesReturned uint32
ret, w32err := win32.ReadDirectoryChangesW(
	fHandle,
	unsafe.Pointer(&buffer[0]),
	uint32(bufferSize),
	win32.BoolToBOOL(false),
	win32.FILE_NOTIFY_CHANGE_FILE_NAME |
		win32.FILE_NOTIFY_CHANGE_DIR_NAME |
		win32.FILE_NOTIFY_CHANGE_ATTRIBUTES |
		win32.FILE_NOTIFY_CHANGE_SIZE |
		win32.FILE_NOTIFY_CHANGE_LAST_WRITE |
		win32.FILE_NOTIFY_CHANGE_CREATION |
		win32.FILE_NOTIFY_CHANGE_SECURITY,
	&bytesReturned,
	&overlapped,
	// 更新検知時のコールバック関数
	syscall.NewCallback(func(dwErrorCode uint32, dwNumberOfBytesTransfered uint32, lpOverlapped *win32.OVERLAPPED) uintptr {
		
		// 変更されたファイルの一覧を表示
		offset := 0
		for {
			fileInfo := (*win32.FILE_NOTIFY_INFORMATION)(unsafe.Pointer(&buffer[offset]))
			fileName := syscall.UTF16ToString((*[win32.MAX_PATH]uint16)(unsafe.Pointer(&fileInfo.FileName[0]))[:fileInfo.FileNameLength / 2])
			fmt.Printf("%s\n", fileName)
			if fileInfo.NextEntryOffset == 0 {
				break
			}
			offset += int(fileInfo.NextEntryOffset)
		}
		
		// シグナル送信
		window1.Emit("directory_changed", glib.TYPE_POINTER, targetDir, uint(dwErrorCode))
		return 0
	}),
)
if ret == win32.FALSE || w32err != win32.NO_ERROR {
	return fmt.Errorf("ReadDirectoryChangesの失敗")
}
```

変更があったファイルの情報は`buffer`に格納されます。  
`BoolToBOOL()`で指定している引数は、サブディレクトリ配下も監視対象にするかどうかの指定となります。`true`が監視対象、`false`が監視対象外になります。  
`FILE_NOTIFY_・・・`で指定している引数は、何が更新されたら検知するのかを指定しています。  
インラインでコールバック関数を定義してます。外部で定義する場合は`buffer`にアクセスしているので変数のスコープに気をつけてください。  



> [!NOTE]  
> `Emit()`を`glib.IdlaAdd()`を使って実行すると凄く遅くなったので`glib.IdlaAdd()`は使ってません。フリーズやクラッシュなどはしてないようなので、問題ないと思いますが不具合が出たら修正してください。  

> [!TIP]  
> 単体のファイルを監視をする場合は、「変更されたファイルの一覧を表示」でファイル名を取得しているので、そこへ検知する処理を追加することになります。その場合、大量のファイルが更新されても問題ないか、bufferサイズが適切かどうかの検証が必要となります。  
> ネットワークで接続されてるディレクトリの更新を監視する場合、bufferサイズは64KBが上限の可能性がありますので注意して下さい。  

> [!CAUTION]  
> 実際にはファイルハンドルが取得済みかどうか、監視中かどうかなどの確認も必要です。  

## 19.3 一度検知すると監視が終了してしまうので、再度監視を登録  

ディレクトリ更新検知時のシグナル受信側の処理は以下のようなコードになります。  

```go
window1.Connect("directory_changed", func(parent *gtk.ApplicationWindow, directory string, dwErrorCode uint) {
	// ディレクトリ更新検知でエラーが発生していた場合
	if dwErrorCode != uint(win32.ERROR_SUCCESS) {
		ShowErrorDialog(parent, fmt.Errorf("ReadDirectoryChangesW コールバックエラー: %v", dwErrorCode))
		return
	}
	
	fmt.Printf("ディレクトリの更新を検知しました：%s\n", directory)
	
	// ディレクトリ監視を再開
	// ※停止中のイベントもある程度キャッシュに溜まってる
	err = DirWatchStart(directory)
	if err != nil {
		ShowErrorDialog(parent, err)
	}
})
```

「19.2」で説明した監視登録のコードを関数`DirWatchStart()`としてコールしてます。  

> [!NOTE]  
> `DirWatchStart()`実行までは非監視状態であることをステータス用変数などを使って管理した方が良いかもしれません。  

## 19.4 ファイルハンドルをクローズして監視を終了  

ファイルハンドルをクローズすると、ディレクトリ更新監視は終了します。  

```go
ret, w32err := win32.CloseHandle(fHandle)
if ret == win32.FALSE || w32err != win32.NO_ERROR {
	return fmt.Errorf("CloseHandleの失敗")
}
```

> [!CAUTION]  
> 実際にはファイルハンドルが取得済みかどうかなどの確認も必要です。  

## 19.5 おわりに  

複数ファイル更新時は、複数ファイルまとめて1イベントではなく、ファイル単位かそれに近いレベルでイベントが発生します。そのため、イベント到着時に、画面更新やファイル取り込みなどの処理を即開始してしまうと、処理中に次のイベントが発生してしまいます。  
（10ファイルコピーされた時に、最後に1回だけ処理をしたくても、イベントが10回届くので1ファイル毎に10回処理をすることになる）  

監視再開の実行をせずに処理を開始すれば、処理中にイベントは発生しませんが、イベントがキャッシュに溜まってるため、ファイル単位レベルで順々にイベントが到着する状況は変わりません。  

自アプリ内でファイルコピーなどを実装する場合は、監視を止めた後ファイルコピーを行い、コピー後に監視を再開するようにプログラムする必要があります。  

他アプリからファイルが更新される場合、更新検知後に、監視を止めて数秒待ってから画面更新やデータ取り込みなどの処理を開始し、処理が終わったら監視を再開するようなプログラムにする必要があります。  

作成したファイルは、
[ここ](19_SimpleWindow_directory.go)と[ここ](19_library.go)
に置いてます。   
「Caution」で記載した確認処理なども盛り込んで関数化しています。  

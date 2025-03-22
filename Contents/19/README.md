[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 19. ディレクトリ配下の更新を監視したい  

指定したディレクトリ配下のファイルやサブディレクトリに更新があったかどうかを監視する方法を説明していきます。  
ディレクトリ更新監視は以下のような流れとなります。  

- ディレクトリのファイルハンドルを取得  
- `ReadDirectoryChangesW()`で監視登録と更新検知時に実行するコールバック関数を登録  
- 一度検知すると監視が終了してしまうので、再度監視を登録  
- ファイルハンドルをクローズして監視を終了  

> [!NOTE]  
> 指定したディレクトリ自体の更新は検知出来ないため、ディレクトリ自体の更新を検知するには、親ディレクトリを監視対象にする必要があります。  

## 19.2 ディレクトリのファイルハンドルを取得  

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

## 19.3 監視登録と更新検知時に実行するコールバック関数を登録  

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

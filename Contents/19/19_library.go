package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"github.com/zzl/go-win32api/win32"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// ディレクトリ変更監視用構造体
type DirWatchInfo struct {
	handle		win32.HANDLE
	watching	bool
	subdir		bool
	overlapped	win32.OVERLAPPED
}

// ディレクトリ変更監視用マップ
var DirWatchMap = make(map[string]DirWatchInfo)

//-----------------------------------------------------------------------------
// ディレクトリを監視リストに追加
//-----------------------------------------------------------------------------
func DirWatchOpen(targetDir string, isSubdirWatch bool) error {
	// ディレクトリが存在しない場合
	ok, err := DirectoryExists(targetDir)
	if err != nil {
		return fmt.Errorf("ディレクトリの存在確認でエラーが発生しました")
	}
	if !ok {
		return fmt.Errorf("存在しないディレクトリが指定されました")
	}
	
	// 登録済みの場合
	dwMap, ok := DirWatchMap[targetDir]
	if ok {
		return nil
	}
	
	targetDirUTF16Ptr := windows.StringToUTF16Ptr(targetDir)

	// ディレクトリのハンドルを取得
	fHandle, w32err := win32.CreateFile(targetDirUTF16Ptr, win32.FILE_LIST_DIRECTORY, win32.FILE_SHARE_READ | win32.FILE_SHARE_WRITE | win32.FILE_SHARE_DELETE, nil, win32.OPEN_EXISTING, win32.FILE_FLAG_BACKUP_SEMANTICS | win32.FILE_FLAG_OVERLAPPED, 0)
	if fHandle == 0 || w32err != win32.NO_ERROR {
		return fmt.Errorf("CreateFileの失敗")
	}
	dwMap.handle = fHandle
	dwMap.watching = false
	dwMap.subdir = isSubdirWatch
	DirWatchMap[targetDir] = dwMap
	
	return nil
}

//-----------------------------------------------------------------------------
// ディレクトリ監視を開始
//-----------------------------------------------------------------------------
func DirWatchStart(targetDir string) error {
	// 未登録の場合
	dwMap, ok := DirWatchMap[targetDir]
	if !ok {
		return fmt.Errorf("指定されたディレクトリは未登録です")
	}
	
	// 監視中の場合は何もしない
	if dwMap.watching {
		return nil
	}
	
	bufferSize := 4096
	buffer := make([]byte, bufferSize)
	
	dwMap.overlapped = win32.OVERLAPPED {}

	// 変更を監視
	var bytesReturned uint32
	ret, w32err := win32.ReadDirectoryChangesW(
		dwMap.handle,
		unsafe.Pointer(&buffer[0]),
		uint32(bufferSize),
		win32.BoolToBOOL(dwMap.subdir),
		win32.FILE_NOTIFY_CHANGE_FILE_NAME |
			win32.FILE_NOTIFY_CHANGE_DIR_NAME |
			win32.FILE_NOTIFY_CHANGE_ATTRIBUTES |
			win32.FILE_NOTIFY_CHANGE_SIZE |
			win32.FILE_NOTIFY_CHANGE_LAST_WRITE |
			win32.FILE_NOTIFY_CHANGE_CREATION |
			win32.FILE_NOTIFY_CHANGE_SECURITY,
		&bytesReturned,
		&dwMap.overlapped,
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
			// ※glib.IdlaAddを使うと凄く遅くなる。とりあえず不具合は起きてない
			window1.Emit("directory_changed", glib.TYPE_POINTER, targetDir, uint(dwErrorCode))
			return 0
		}),
	)
	if ret == win32.FALSE || w32err != win32.NO_ERROR {
		return fmt.Errorf("ReadDirectoryChangesの失敗")
	}
	
	dwMap.watching = true

	DirWatchMap[targetDir] = dwMap
	
	return nil
}

//-----------------------------------------------------------------------------
// ディレクトリ監視を終了しmapから削除
//-----------------------------------------------------------------------------
func DirWatchClose(targetDir string) error {
	// 未登録の場合
	dwMap, ok := DirWatchMap[targetDir]
	if !ok {
		return fmt.Errorf("指定されたディレクトリは未登録です")
	}
	
	// ディレクトリハンドルのクローズ
	// ※クローズにより監視が終了
	ret, w32err := win32.CloseHandle(dwMap.handle)
	if ret == win32.FALSE || w32err != win32.NO_ERROR {
		return fmt.Errorf("CloseHandleの失敗")
	}
	delete(DirWatchMap, targetDir)
	
	return nil
}

//-----------------------------------------------------------------------------
// ディレクトリの存在確認
//-----------------------------------------------------------------------------
func DirectoryExists(dirname string) (bool, error) {
	f, err := os.Stat(dirname)
	if err != nil {
		// ファイルが存在しない場合
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		// その他のエラーの場合
		return false, err
	}
	
	// ファイルが存在した場合、ディレクトリ判定
	return f.IsDir(), nil
}

//-----------------------------------------------------------------------------
// グレードからgtkオブジェクトを取得
// [input]
//   builder	gladeファイルから構築したbuilder（初回はnilを指定）
//				同一ファイル内に含まれてるオブジェクトを取得する場合は前回取得したbuilderを指定
//   glade		gladeファイルから取得したxml文字列（embedから代入した変数を指定）
//   id			gladeでオブジェクトに設定したID
// [return]
//   オブジェクト
//   Builder
//   error
//-----------------------------------------------------------------------------
func GetObjFromGlade[T any](builder *gtk.Builder, glade string, id string) (T, *gtk.Builder, error) {
	var err error
	
	// builderがNULLの場合、gladeから読み込む
	if builder == nil {
		if len(glade) == 0 {
			return *new(T), nil, fmt.Errorf("Could not get glade string")
		}
		builder, err = gtk.BuilderNewFromString(glade)
		if err != nil {
			return *new(T), nil, fmt.Errorf("Could not create builder: %w", err)
		}
	}
	
	// Builderの中からオブジェクトを取得
	obj, err := builder.GetObject(id)
	if err != nil {
		return *new(T), nil, fmt.Errorf("Could not get object: %w", err)
	}
	gtkObject, ok := obj.(T)
	if !ok {
		return *new(T), nil, fmt.Errorf("Could not convert object to gtk object")
	}
	
	return gtkObject, builder, nil
}

//-----------------------------------------------------------------------------
// キューに溜まった処理をすべて実行させる
// フリーズすることがあるので、100回まで
//-----------------------------------------------------------------------------
func DoEvents() {
	for i := 0; i < 100; i++ {
		if !gtk.EventsPending() {
			break
		}
	    gtk.MainIteration()
	}
}

//-----------------------------------------------------------------------------
// エラーメッセージを表示する
//-----------------------------------------------------------------------------
func ShowErrorDialog(parent gtk.IWindow, err error) {
	dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "エラーが発生しました")
	dialog.FormatSecondaryText("%s", err.Error())
	dialog.SetTitle ("error")
	dialog.Run()
	dialog.Destroy()
}


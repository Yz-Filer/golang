// ファイルエクスプローラーのコンテキストメニューを表示
package main
/*
#cgo pkg-config: gdk-3.0
#include <gdk/gdk.h>
#include <gdk/gdkwin32.h>
#cgo LDFLAGS: -lsetupapi -lshlwapi -lole32 -loleaut32 -luuid -Wl,--allow-multiple-definition
#include "context_menu.c"
*/
import "C"

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
	"github.com/zzl/go-win32api/win32"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/01_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application
var window1 *gtk.ApplicationWindow
var Hwnd uintptr

//-----------------------------------------------------------------------------
// ファイルパスをCGO用メモリに格納
//-----------------------------------------------------------------------------
func prepareFilePaths(filePaths []string) (unsafe.Pointer, C.int, error) {
	fileCount := len(filePaths)
	if fileCount == 0 {
		return nil, C.int(0), fmt.Errorf("no file")
	}

	// CのWCHAR** を確保
	// sizeof(WCHAR*) はC.int(unsafe.Sizeof((*C.WCHAR)(nil)))で取得できる
	cArray := C.malloc(C.size_t(C.int(unsafe.Sizeof((*C.WCHAR)(nil))) * C.int(fileCount)))
	if cArray == nil {
		return nil, C.int(0), fmt.Errorf("failed to allocate memory for C array")
	}
	
	for i, path := range filePaths {
		// UTF-16でNULL終端されたワイド文字列を生成
		wPath, err := windows.UTF16PtrFromString(path)
		if err != nil {
			return nil, C.int(0), err
		}

		// C.mallocでメモリを確保し、wPathをコピー
		// C.wcslen(wPath) * sizeof(wchar_t) + sizeof(wchar_t)
		wlen := C.wcslen((*C.WCHAR)(unsafe.Pointer(wPath)))
		cWcharPtr := C.malloc(C.size_t(wlen+1) * C.sizeof_wchar_t)
		if cWcharPtr == nil {
			return nil, C.int(0), fmt.Errorf("failed to allocate memory for WCHAR string")
		}

		// メモリをコピー
		C.wcscpy((*C.WCHAR)(cWcharPtr), (*C.WCHAR)(unsafe.Pointer(wPath)))

		// ポインタ配列にコピーしたワイド文字列のポインタを格納
		ptr := (**C.WCHAR)(unsafe.Pointer(uintptr(cArray) + uintptr(i)*unsafe.Sizeof((*C.WCHAR)(nil))))
		*ptr = (*C.WCHAR)(cWcharPtr)
	}
	
	return cArray, C.int(fileCount), nil
}

//-----------------------------------------------------------------------------
// gtk.WindowからWindowsのWindowハンドルを取得する
//-----------------------------------------------------------------------------
func GetWindowHandle(window *gtk.ApplicationWindow) (uintptr, error) {
	gdkWin, err := window.GetWindow()
	if err != nil {
		return uintptr(0), err
	}
	return uintptr(C.gdk_win32_window_get_handle((*C.GdkWindow)(unsafe.Pointer(gdkWin.Native())))), nil
}

//-----------------------------------------------------------------------------
// メイン
//-----------------------------------------------------------------------------
func main() {
	const appID = "org.example.myapp"
	var err error
	
	// 「コピー」「切り取り」のためにOleを初期化
	ret := win32.OleInitialize(nil) 
	if ret != win32.S_OK {
		log.Fatal("OLEの初期化に失敗しました")
	}
	
	///////////////////////////////////////////////////////////////////////////
	// 新しいアプリケーションの作成
	///////////////////////////////////////////////////////////////////////////
	application, err = gtk.ApplicationNew(appID, glib.APPLICATION_NON_UNIQUE)
	if err != nil {
		log.Fatal("Could not create application:", err)
	}

	///////////////////////////////////////////////////////////////////////////
	// アプリケーション起動時のイベント（必須ではない）
	///////////////////////////////////////////////////////////////////////////
	application.Connect("startup", func() {
		log.Println("application startup")	
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーション アクティブ時のイベント
	///////////////////////////////////////////////////////////////////////////
	application.Connect("activate", func() {
		// gladeからウィンドウを取得
		window1, _, err = GetObjFromGlade[*gtk.ApplicationWindow](nil, window1Glade, "MAIN_WINDOW")
		if err != nil {
			log.Fatal(err)
		}
		
		
		// リソースからアプリケーションのアイコンを設定
		iconPixbuf, err := gdk.PixbufNewFromDataOnly(icon)
		if err != nil {
			log.Fatal("Could not create pixbuf from bytes:", err)
		}
		defer iconPixbuf.Unref()
		
		// ウィンドウにアイコンを設定
		window1.SetIcon(iconPixbuf)
		
		// ウィンドウのプロパティを設定（必須ではない）
		window1.SetPosition(gtk.WIN_POS_MOUSE)
		
		
		//-----------------------------------------------------------
		// マウスボタンを押したときのイベントハンドラ
		//-----------------------------------------------------------
		window1.Connect("button-press-event", func(win *gtk.ApplicationWindow, event *gdk.Event) bool {
			e := gdk.EventButtonNewFromEvent(event)
			switch e.Button() {
				case gdk.BUTTON_SECONDARY:							// 右クリック時
					// ファイルパスをCGO用メモリに格納
					filePaths, fileCount, err := prepareFilePaths([]string{"D:\\test\\ううう"})
					if err != nil {
						log.Println("ファイルパスのメモリ格納に失敗しました")
						return false
					}
					
					// 確保したメモリを解放する
					defer C.free_wchars_array((**C.WCHAR)(filePaths), fileCount)
					
					// コンテキストメニューを表示する
					x, y := e.MotionValRoot()
					ok := C.ShowContextMenu(C.HWND(unsafe.Pointer(Hwnd)), C.int(x), C.int(y), (**C.WCHAR)(filePaths), fileCount)
					if !ok {
						log.Println("コンテキストメニューの表示に失敗しました。")
					}
					// イベントの伝播を停止
					return true
			}
			// イベントを伝播
			return false
		})
	
		//-----------------------------------------------------------
		// ウィンドウ最小化、最大化時の処理（必須ではない）
		// Linuxは挙動が異なるかも
		//-----------------------------------------------------------
		window1.Connect("window-state-event", func(parent *gtk.ApplicationWindow, event *gdk.Event) bool {
			// gdk.EventWindowState を取得
			windowStateEvent := gdk.EventWindowStateNewFromEvent(event)
			
			if windowStateEvent != nil {
				// 最小化された場合
				if windowStateEvent.ChangedMask() == (gdk.WINDOW_STATE_ICONIFIED | gdk.WINDOW_STATE_FOCUSED) {
					log.Println("ウィンドウが最小化されました")
				}
				
				// 最大化された場合
				if windowStateEvent.NewWindowState() == (gdk.WINDOW_STATE_MAXIMIZED | gdk.WINDOW_STATE_FOCUSED) {
					log.Println("ウィンドウが最大化されました")
				}
			}
			
			// イベントの伝播を停止
			return true
		})
		
		//-----------------------------------------------------------
		// 閉じるボタンが押された時の処理（必須ではない）
		// まだ、閉じる前のため、キャンセルが可能
		//-----------------------------------------------------------
		window1.Connect("delete-event", func(parent *gtk.ApplicationWindow, event *gdk.Event) bool {
			log.Println("ウィンドウのクローズが試みられました")
			
			// ウィンドウクローズ処理を中断
			//return true
			
			// ウィンドウクローズ処理を継続
			return false
		})
		
		//-----------------------------------------------------------
		// メインウィンドウを閉じた後の処理（必須ではない）
		// この後、アプリケーションの"shutdown"イベントも呼ばれる
		//-----------------------------------------------------------
		window1.Connect("destroy", func() {
			log.Println("ウィンドウが閉じられました")
		})
		
		// アプリケーションを設定
		window1.SetApplication(application)

		// ウィンドウの表示
		window1.ShowAll()
		


		//-----------------------------------------------------------
		// ウィンドウハンドルの取得
		//-----------------------------------------------------------
		Hwnd, err = GetWindowHandle(window1)
		if err != nil {
			log.Fatal("ウィンドウハンドルの取得に失敗: ", err)
		}
	})
		
	///////////////////////////////////////////////////////////////////////////
	// アプリケーション終了時のイベント（必須ではない）
	///////////////////////////////////////////////////////////////////////////
	application.Connect("shutdown", func() {
		log.Println("application shutdown")
		win32.OleUninitialize()
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	// Runに引数を渡してるけど、application側で取りだすより
	// go側でグローバル変数にでも格納した方が楽
	os.Exit(application.Run(os.Args))
}


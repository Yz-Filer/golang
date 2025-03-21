// クリップボードの監視/USBドライブ抜き差し監視/USBドライブイジェクト
// メッセージ用ウィンドウ作成版
package main

/*
#cgo LDFLAGS: -lsetupapi -Wl,--allow-multiple-definition
#include "usb-ejecter.c"
*/
import "C"

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"github.com/zzl/go-win32api/win32"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/18_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application
var window1 *gtk.ApplicationWindow


//-----------------------------------------------------------------------------
// メッセージ受信用ウィンドウのウィンドウプロシージャ
//-----------------------------------------------------------------------------
func WndProc(hwnd win32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch (msg) {
		case win32.WM_CLIPBOARDUPDATE:
			// シグナルを送信
			glib.IdleAdd(func() {
				window1.Emit("clipboard_update", glib.TYPE_POINTER)
			})
		case win32.WM_DEVICECHANGE:
			hdr := (*win32.DEV_BROADCAST_HDR)(unsafe.Pointer(lParam))
			if hdr == nil {
				break
			}

			if hdr.Dbch_devicetype == win32.DBT_DEVTYP_VOLUME {
				// ドライブレターの取得
				vol := (*win32.DEV_BROADCAST_VOLUME)(unsafe.Pointer(lParam))
				drvLetter := ""
				for i := 0; i < 26; i++ {
					if (vol.Dbcv_unitmask >> i) & 1 == 1 {
						drvLetter = string('A' + i) + ":"
						break
					}
				}
				
				// シグナルを送信
				switch uint32(wParam) {
					case win32.DBT_DEVICEARRIVAL:			// ドライブが追加された場合
						glib.IdleAdd(func() {
							window1.Emit("device_add", glib.TYPE_POINTER, drvLetter)
						})
					case win32.DBT_DEVICEREMOVECOMPLETE:	// ドライブが取り外された場合
						glib.IdleAdd(func() {
							window1.Emit("device_remove", glib.TYPE_POINTER, drvLetter)
						})
				}
			}
		case win32.WM_DESTROY:
			win32.PostQuitMessage(0)
	}
	return win32.DefWindowProc(hwnd, msg, wParam, lParam)
}

//-----------------------------------------------------------------------------
// メイン
//-----------------------------------------------------------------------------
func main() {
	const appID = "org.example.myapp"
	var err error
	var builder *gtk.Builder

	//-----------------------------------------------------------
	// メッセージ用ウィンドウ作成
	//-----------------------------------------------------------
	
	var w32err win32.WIN32_ERROR
	
	// ウィンドウクラスの登録
	className := windows.StringToUTF16Ptr("window class")
	wndClass := win32.WNDCLASSEX{
		CbSize	: uint32(unsafe.Sizeof(win32.WNDCLASSEX{})),
		LpfnWndProc   : syscall.NewCallback(WndProc),
		LpszClassName : className,
	}
	_, w32err = win32.RegisterClassEx(&wndClass)
	if w32err != win32.NO_ERROR {
		log.Fatal("RegisterClassExの失敗")
	}

	// モジュールハンドルを取得
	// ※hInstance=0でもCreateWindowExは動くがdll化する場合などは動かなくなる可能性がある
	hInstance, w32err := win32.GetModuleHandleW(nil)

	// メッセージ受信用ウィンドウの作成
	// ※DBT_DEVTYP_VOLUMEメッセージを受信するためにトップレベルウィンドウが必要
	hwnd, w32err := win32.CreateWindowEx(win32.WS_EX_APPWINDOW, className, nil, win32.WS_OVERLAPPEDWINDOW, 0, 0, 0, 0, 0, 0, hInstance, nil)
	if hwnd == 0 || w32err != win32.NO_ERROR {
		log.Fatal("CreateWindowExの失敗")
	}

	//-----------------------------------------------------------
	// クリップボード監視の登録
	//-----------------------------------------------------------
	
	// クリップボードの更新メッセージをhwndへ送信
	ret, w32err := win32.AddClipboardFormatListener(hwnd)
	if ret == win32.FALSE || w32err != win32.NO_ERROR {
		log.Fatal("AddClipboardFormatListenerの失敗")
	}

	//-----------------------------------------------------------
	// デバイス監視（ドライブレターの増減）の登録
	//-----------------------------------------------------------
	
	// デバイス変更通知メッセージをhwndへ送信
	notificationFilter := win32.DEV_BROADCAST_DEVICEINTERFACE_{
		Dbcc_size:       uint32(unsafe.Sizeof(win32.DEV_BROADCAST_DEVICEINTERFACE_{})),
		Dbcc_devicetype: uint32(win32.DBT_DEVTYP_DEVICEINTERFACE),
		Dbcc_reserved:   0,
		Dbcc_classguid:  win32.GUID_IO_VOLUME_DEVICE_INTERFACE,
	}
	hDevNotify, w32err := win32.RegisterDeviceNotification(hwnd, unsafe.Pointer(&notificationFilter), win32.DEVICE_NOTIFY_WINDOW_HANDLE)
	if hDevNotify == nil || w32err != win32.NO_ERROR {
		log.Fatalf("RegisterDeviceNotificationの失敗")
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
		window1, builder, err = GetObjFromGlade[*gtk.ApplicationWindow](nil, window1Glade, "MAIN_WINDOW")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからコンボボックステキストを取得
		comboBoxText, _, err := GetObjFromGlade[*gtk.ComboBoxText](builder, "", "COMBOBOXTEXT")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからボタンを取得
		button, _, err := GetObjFromGlade[*gtk.Button](builder, "", "BUTTON")
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
		// カスタムシグナル
		//-----------------------------------------------------------

		// カスタムシグナル（WM_CLIPBOARDUPDATE）を作成
		_, err = glib.SignalNew("clipboard_update")
		if err != nil {
			log.Fatal("Could not create signal: ", err)
		}

		// カスタムシグナル（DBT_DEVICEARRIVAL）を作成
		_, err = glib.SignalNewV("device_add", glib.TYPE_POINTER, 1, glib.TYPE_STRING)
		if err != nil {
			log.Fatal("Could not create signal: ", err)
		}

		// カスタムシグナル（DBT_DEVICEREMOVECOMPLETE）を作成
		_, err = glib.SignalNewV("device_remove", glib.TYPE_POINTER, 1, glib.TYPE_STRING)
		if err != nil {
			log.Fatal("Could not create signal: ", err)
		}
		
		
		//-----------------------------------------------------------
		// クリップボード更新検知のシグナル処理
		//-----------------------------------------------------------
		window1.Connect("clipboard_update", func() {
			fmt.Println("Clipboard content changed!")
		})

		//-----------------------------------------------------------
		// ドライブ追加検知のシグナル処理
		//-----------------------------------------------------------
		window1.Connect("device_add", func(parent *gtk.ApplicationWindow, drvLetter string) {
			if drvLetter != "" {
				fmt.Println(drvLetter, "was added.")
			} else {
				fmt.Println("A new unknown device was detected.")
			}
		})

		//-----------------------------------------------------------
		// ドライブ取り外し検知のシグナル処理
		//-----------------------------------------------------------
		window1.Connect("device_remove", func(parent *gtk.ApplicationWindow, drvLetter string) {
			if drvLetter != "" {
				fmt.Println(drvLetter, "was removed.")
			} else {
				fmt.Println("An unknown device has been disconnected.")
			}
		})

		//-----------------------------------------------------------
		// ejectボタン押下時の処理
		//-----------------------------------------------------------
		button.Connect("clicked", func() {
			str := comboBoxText.GetActiveText()
			if C.EjectDriveByLetter(C.wchar_t(str[0])) {
				log.Printf("USBドライブ[%s]が取り外しを行いました。\n", str)
			} else {
				log.Printf("USBドライブ[%s]の取り外しに失敗しました。\n", str)
			}
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
	})
		
	///////////////////////////////////////////////////////////////////////////
	// アプリケーション終了時のイベント（必須ではない）
	///////////////////////////////////////////////////////////////////////////
	application.Connect("shutdown", func() {
		log.Println("application shutdown")

		// デバイス変更通知のメッセージ送信を停止
		win32.UnregisterDeviceNotification(hDevNotify)
		
		// クリップボードの更新メッセージ送信を停止
		win32.RemoveClipboardFormatListener(hwnd)
		
		// メッセージ受信用ウィンドウの破棄
		win32.DestroyWindow(hwnd)
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	// Runに引数を渡してるけど、application側で取りだすより
	// go側でグローバル変数にでも格納した方が楽
	os.Exit(application.Run(os.Args))
}


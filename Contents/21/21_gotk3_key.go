// キー入力
package main

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

var (
	Imm32				= windows.NewLazyDLL("imm32.dll")
	ImmGetContext		= Imm32.NewProc("ImmGetContext")
	ImmSetOpenStatus	= Imm32.NewProc("ImmSetOpenStatus")
	ImmReleaseContext	= Imm32.NewProc("ImmReleaseContext")
)

//go:embed glade/20_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application

// ウィンドウハンドル
var hwnd uintptr

// メッセージフック用ハンドル
var HookHandleM win32.HHOOK

//-----------------------------------------------------------------------------
// WH_GETMESSAGE用のコールバック関数
//-----------------------------------------------------------------------------
func hookProcM(nCode int, wParam, lParam uintptr) uintptr {
	if nCode >= 0 {
		msg := (*win32.MSG)(unsafe.Pointer(lParam))
		// 自ウィンドウの時だけ処理
		if hwnd == msg.Hwnd {
			switch (msg.Message) {
				case win32.WM_INPUT:
					// RAWINPUTのサイズを取得
					var size uint32
					win32.GetRawInputData(msg.LParam, win32.RID_INPUT, nil, &size, uint32(unsafe.Sizeof(win32.RAWINPUTHEADER{})))

					if size > 0 {
						// RawInputデータを取得
						buf := make([]byte, size)
						if win32.GetRawInputData(msg.LParam, win32.RID_INPUT, unsafe.Pointer(&buf[0]), &size, uint32(unsafe.Sizeof(win32.RAWINPUTHEADER{}))) == size {
							raw := (*win32.RAWINPUT)(unsafe.Pointer(&buf[0]))
							
							// キーボード入力を処理
							if raw.Header.DwType == uint32(win32.RIM_TYPEKEYBOARD) {
								keyboard := raw.Data.KeyboardVal()
								switch (uint32(keyboard.Flags) & 0x01) {
									case win32.RI_KEY_MAKE:				// キーを押した時
										// 2つあるキーはE0プレフィックスでどちらのキーか判定
										// ※「shiftキー + テンキー0」のInsertとInsertキーの判別にも使える
										if (uint32(keyboard.Flags) & 0x02) == win32.RI_KEY_E0 {
											fmt.Print("1:右の")
										} else {
											fmt.Print("1:左の")
										}
										
										// 例としてInsertキーのみ判別。他はキーコード表示。
										if keyboard.VKey == uint16(win32.VK_INSERT) {
											fmt.Println("Insertキーが押されました")
										} else {
											fmt.Println(keyboard.VKey, "が押されました")
										}
									case win32.RI_KEY_BREAK:			// キーを離した時
								}
							}
						} else {
							log.Println("GetRawInputData (data) failed")
						}
					} else {
						log.Println("GetRawInputData (size) failed")
					}
			}
		}
	}
	return uintptr(win32.CallNextHookEx(HookHandleM, int32(nCode), wParam, lParam))
}

//-----------------------------------------------------------------------------
// メイン
//-----------------------------------------------------------------------------
func main() {
	const appID = "org.example.myapp"
	var window1 *gtk.ApplicationWindow
	var builder *gtk.Builder
	var err error
	var w32err win32.WIN32_ERROR
	
	
	//-----------------------------------------------------------
	// Windowsメッセージのhook
	//-----------------------------------------------------------

	// 物理キーの判別をするためのフック
	HookHandleM, w32err = win32.SetWindowsHookEx(win32.WH_GETMESSAGE, uintptr(syscall.NewCallback(hookProcM)), 0, win32.GetCurrentThreadId())
	if HookHandleM == 0 || w32err != win32.NO_ERROR {
		log.Fatalf("SetWindowsHookEx failed: %v", win32.GetLastError())
	}

	devices := []win32.RAWINPUTDEVICE{
		{
			UsUsagePage: 0x01, // Generic Desktop
			UsUsage:     0x06, // Keyboard
			DwFlags:     0,
			HwndTarget:  hwnd,
		},
	}

	pRawInputDevices := (*win32.RAWINPUTDEVICE)(unsafe.Pointer(&devices[0]))
	uiNumDevices := uint32(len(devices))
	cbSize := uint32(unsafe.Sizeof(devices[0]))

	ret, w32err := win32.RegisterRawInputDevices(pRawInputDevices, uiNumDevices, cbSize)
	if ret == win32.FALSE || w32err != win32.NO_ERROR {
		log.Fatal("RegisterRawInputDevicesの失敗")
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
		
		// gladeからテキストをクリップボードへ格納するボタンを取得
		btnTxtTo, _, err := GetObjFromGlade[*gtk.Button](builder, "", "BTN_TXT_TO")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからテキストをクリップボードから取得するボタンを取得
		btnTxtFrom, _, err := GetObjFromGlade[*gtk.Button](builder, "", "BTN_TXT_FROM")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeから画像をクリップボードへ格納するボタンを取得
		btnImgTo, _, err := GetObjFromGlade[*gtk.Button](builder, "", "BTN_IMG_TO")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeから画像をクリップボードから取得するボタンを取得
		btnImgFrom, _, err := GetObjFromGlade[*gtk.Button](builder, "", "BTN_IMG_FROM")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからText用Entryを取得
		entryText, _, err := GetObjFromGlade[*gtk.Entry](builder, "", "ENTRY1")
		if err != nil {
			log.Fatal(err)
		}
		
		
		// gladeからImageを取得
		image1, _, err := GetObjFromGlade[*gtk.Image](builder, "", "IMAGE1")
		if err != nil {
			log.Fatal(err)
		}
		
		// クリップボードを取得
		clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
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
		// キーが押された時のイベントハンドラ
		// ※「SHIFT + 0」が「Insert」になるなどキーボードの割り当てが優先される
		//-----------------------------------------------------------
		window1.Connect("key-press-event", func(win *gtk.ApplicationWindow, event *gdk.Event) bool {
			keyEvent := gdk.EventKeyNewFromEvent(event)
			keyVal := keyEvent.KeyVal()
			keyState := gdk.ModifierType(keyEvent.State() & 0x0F)
			
			switch keyState {
				case gdk.SHIFT_MASK:	// SHIFTキー
					switch keyVal {
						case gdk.KEY_a, gdk.KEY_A:				fmt.Println(" [shift + a] が押されました")
						case gdk.KEY_Shift_L, gdk.KEY_Shift_R:	fmt.Println(" [shift] が押されました")
						default:
							// 押されたキーを表示
							keyName := gdk.KeyValName(keyVal)
							fmt.Printf(" [shift + %s] が押されました\n", keyName)
					}
				case gdk.CONTROL_MASK:	// CTRLキー
					switch keyVal {
						case gdk.KEY_a, gdk.KEY_A:					fmt.Println(" [ctrl + a] が押されました")
						case gdk.KEY_Control_L, gdk.KEY_Control_R:	fmt.Println(" [ctrl] が押されました")
						default:
							// 押されたキーを表示
							keyName := gdk.KeyValName(keyVal)
							fmt.Printf(" [ctrl + %s] が押されました\n", keyName)
					}
				case gdk.MOD1_MASK:		// ALTキー
					switch keyVal {
						case gdk.KEY_a, gdk.KEY_A:				fmt.Println(" [alt + a] が押されました")
						default:
							// 押されたキーを表示
							keyName := gdk.KeyValName(keyVal)
							fmt.Printf(" [alt + %s] が押されました\n", keyName)
					}
				default:
					switch keyVal {
						case gdk.KEY_a, gdk.KEY_A:		fmt.Println(" [a] が押されました")
						case gdk.KEY_Alt_L, gdk.KEY_Alt_R:		fmt.Println(" [alt] が押されました")
						default:
							// 押されたキーを表示
							keyName := gdk.KeyValName(keyVal)
							fmt.Printf(" [%s] が押されました\n", keyName)
					}
			}
			
			// イベントを伝播
			return false
		})
		
		//-----------------------------------------------------------
		// キーが離された時のイベントハンドラ
		//-----------------------------------------------------------
		window1.Connect("key-release-event", func(win *gtk.ApplicationWindow, event *gdk.Event) bool {
			fmt.Println("キーが離されました")
			
			// イベントを伝播
			return false
		})

		//-----------------------------------------------------------
		// マウスを移動したときのイベントハンドラ
		//-----------------------------------------------------------
		window1.Connect("motion-notify-event", func(win *gtk.ApplicationWindow, event *gdk.Event) bool {
			// SHIFTキーの状態を検知
			if (uint16(win32.GetKeyState(int32(win32.VK_SHIFT))) & 0x8000) != 0 {
				fmt.Println("SHIFTキーが押されている")
			}
			// CONTROLキーの状態を検知
			if (uint16(win32.GetKeyState(int32(win32.VK_CONTROL))) & 0x8000) != 0 {
				fmt.Println("CONTROLキーが押されている")
			}
			// ALTキーの状態を検知
			if (uint16(win32.GetKeyState(int32(win32.VK_MENU))) & 0x8000) != 0 {
				fmt.Println("ALTキーが押されている")
			}
			
			// イベントを伝播
			return false
		})
		
		//-----------------------------------------------------------
		// テキストをクリップボードへ送る
		// ★送った後、IMEをOFFにする
		//-----------------------------------------------------------
		btnTxtTo.Connect("clicked", func() {
			text, err := entryText.GetText()
			if err != nil {
				ShowErrorDialog(window1, err)
				return
			}
			clipboard.SetText(text)
			
			// IMEコンテキストのハンドルを取得
			himc, _, err := ImmGetContext.Call(uintptr(hwnd))
			if himc == 0 || err.Error() != "The operation completed successfully." {
				ShowErrorDialog(window1, err)
				return
			}
			
			// IMEをOFFにする
			ret, _, err := ImmSetOpenStatus.Call(uintptr(himc), uintptr(win32.FALSE))
			if win32.BOOL(ret) == win32.FALSE || err.Error() != "The operation completed successfully." {
				ShowErrorDialog(window1, err)
				return
			}
			
			// IMEコンテキストの解放
			ret, _, err = ImmReleaseContext.Call(uintptr(hwnd), uintptr(himc))
			if win32.BOOL(ret) == win32.FALSE || err.Error() != "The operation completed successfully." {
				ShowErrorDialog(window1, err)
				return
			}
		})
		
		//-----------------------------------------------------------
		// テキストをクリップボードから取得する
		// ★送った後、IMEをONにする
		//-----------------------------------------------------------
		btnTxtFrom.Connect("clicked", func() {
			// テキスト形式のデータが存在した場合、Entryに表示
			if clipboard.WaitIsTextAvailable() {
				text, err := clipboard.WaitForText()
				if err != nil {
					ShowErrorDialog(window1, err)
					return
				}
				entryText.SetText(text)
			}
			
			// IMEコンテキストのハンドルを取得
			himc, _, err := ImmGetContext.Call(uintptr(hwnd))
			if himc == 0 || err.Error() != "The operation completed successfully." {
				ShowErrorDialog(window1, err)
				return
			}
			
			// IMEをONにする
			ret, _, err := ImmSetOpenStatus.Call(uintptr(himc), uintptr(win32.TRUE))
			if win32.BOOL(ret) == win32.FALSE || err.Error() != "The operation completed successfully." {
				ShowErrorDialog(window1, err)
				return
			}
			
			// IMEコンテキストの解放
			ret, _, err = ImmReleaseContext.Call(uintptr(hwnd), uintptr(himc))
			if win32.BOOL(ret) == win32.FALSE || err.Error() != "The operation completed successfully." {
				ShowErrorDialog(window1, err)
				return
			}
		})
		
		//-----------------------------------------------------------
		// 画像をクリップボードへ送る
		//-----------------------------------------------------------
		btnImgTo.Connect("clicked", func() {
			// gtk.Imageからgdk.Pixbufを取得
			pixbuf := image1.GetPixbuf()
			if pixbuf == nil {
				ShowErrorDialog(window1, fmt.Errorf("画像が空か、取得に失敗"))
				return
			}

			// クリップボードに画像データを送信
			clipboard.SetImage(pixbuf)
		})
		
		//-----------------------------------------------------------
		// 画像をクリップボードから取得する
		//-----------------------------------------------------------
		btnImgFrom.Connect("clicked", func() {
			// 画像形式のデータが存在した場合、Imageに表示
			if clipboard.WaitIsImageAvailable() {
				pixbuf, err := clipboard.WaitForImage()
				if err != nil {
					ShowErrorDialog(window1, err)
					return
				}
				image1.SetFromPixbuf(pixbuf)
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


		// ウィンドウハンドルを取得
		hwnd, err = GetWindowHandle(window1)
		if err != nil {
			log.Fatal(err)
		}

	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーション終了時のイベント（必須ではない）
	///////////////////////////////////////////////////////////////////////////
	application.Connect("shutdown", func() {
		log.Println("application shutdown")

		// WindowsメッセージのUnhook
		win32.UnhookWindowsHookEx(HookHandleM)
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	// Runに引数を渡してるけど、application側で取りだすより
	// go側でグローバル変数にでも格納した方が楽
	os.Exit(application.Run(os.Args))
}


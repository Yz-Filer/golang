package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"unsafe"
	
	"golang.org/x/sys/windows"
	"github.com/zzl/go-win32api/win32"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/35_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application


var (
	OleDragDrop				= syscall.NewLazyDLL("OleDragDrop.dll")
	OLE_IDropSource_Start	= OleDragDrop.NewProc("OLE_IDropSource_Start")
)

const (
	WM_DROPGETDATA	= win32.WM_USER + 101
	WM_DROPDRAGEND	= win32.WM_USER + 102
)


//-----------------------------------------------------------------------------
// メッセージ受信用ウィンドウのウィンドウプロシージャ
//-----------------------------------------------------------------------------
func WndProc(hwnd win32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch (msg) {
		case WM_DROPGETDATA:
			files := []string{`D:\test\bb.txt`}

			// UTF-16 文字列のスライスを作成
			u16str := []uint16{}
			for _, file := range files {
				u16, err := windows.UTF16FromString(file)
				if err != nil {
					return uintptr(1)
				}
				u16str = append(u16str, u16...)
			}
			u16str = append(u16str, 0) // 終端のヌル文字を追加

			// メモリの割り当て
			size := unsafe.Sizeof(win32.DROPFILES{}) + uintptr(len(u16str) * 2)
			hMem, ret := win32.GlobalAlloc(win32.GMEM_MOVEABLE, uintptr(size))
			if ret != win32.NO_ERROR {
				return uintptr(1)
			}

			// メモリのロック
			dataPtr, ret := win32.GlobalLock(hMem)
			if ret != win32.NO_ERROR {
				return uintptr(1)
			}
			defer win32.GlobalUnlock(hMem)

			// DROPFILES 構造体の作成とメモリへのコピー
			dropFiles := (*win32.DROPFILES)(dataPtr)
			dropFiles.PFiles = win32.DWORD(unsafe.Sizeof(win32.DROPFILES{})) // オフセットを設定
			dropFiles.Pt = win32.POINT{X: 0, Y: 0}
			dropFiles.FNC = win32.FALSE
			dropFiles.FWide = win32.TRUE

			// データをメモリにコピー
			utf16Ptr := (*uint16)(unsafe.Pointer(uintptr(dataPtr) + unsafe.Sizeof(win32.DROPFILES{})))
			copy(unsafe.Slice(utf16Ptr, len(u16str) * 2), u16str)

			// メモリをlParamが指すポインタへ代入
			handlePtr := (*windows.Handle)(unsafe.Pointer(lParam))
			*handlePtr = windows.Handle(hMem)
			
			return uintptr(0)

		case WM_DROPDRAGEND:
			switch win32.HRESULT(wParam) {
				case win32.DRAGDROP_S_DROP:
					log.Println("Drop完了")
				case win32.DRAGDROP_S_CANCEL:
					log.Println("Dropキャンセル")
			}
			return uintptr(0)

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
	var window1 *gtk.ApplicationWindow
	var err error
	var w32err win32.WIN32_ERROR

	// oleのInitialize
	win32.OleInitialize(nil)

	//-----------------------------------------------------------
	// メッセージ用ウィンドウ作成
	//-----------------------------------------------------------
	
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
	if w32err != win32.NO_ERROR {
		log.Fatal("GetModuleHandleWの失敗")
	}

	// メッセージ受信用ウィンドウの作成
	hwnd, w32err := win32.CreateWindowEx(0, className, nil, 0, 0, 0, 0, 0, win32.HWND_MESSAGE, 0, hInstance, nil)
	if hwnd == 0 || w32err != win32.NO_ERROR {
		log.Fatal("CreateWindowExの失敗")
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


		// TargetEntryの設定
		targetEntryURI, err := gtk.TargetEntryNew("text/uri-list", gtk.TARGET_OTHER_APP, 0)
		if err != nil {
			log.Fatal(err)
		}
		targetEntries := []gtk.TargetEntry{*targetEntryURI}

		// Drop先の設定
		window1.DragDestSet(gtk.DEST_DEFAULT_ALL, targetEntries, gdk.ACTION_COPY | gdk.ACTION_MOVE)

		//-----------------------------------------------------------
		// Drop時のデータ受信
		//-----------------------------------------------------------
		window1.Connect("drag-data-received", func(entry *gtk.ApplicationWindow, context *gdk.DragContext, x, y int, data *gtk.SelectionData, info, time uint) {
			// 受信したデータを表示
			switch info {
				case 0:
					fmt.Printf("text/uri-list:\n  %s\n", strings.ReplaceAll(string(data.GetData()), "\r\n",", "))
			}
		})


		var dragging bool
		var offsetX, offsetY int32

		//-----------------------------------------------------------
		// マウスボタンを押したときのイベントハンドラ
		//-----------------------------------------------------------
		window1.Connect("button-press-event", func(win *gtk.ApplicationWindow, event *gdk.Event) {
			e := gdk.EventButtonNewFromEvent(event)
			if e.Button() == gdk.BUTTON_PRIMARY {
				dragging = true
				x, y := e.MotionVal()
				offsetX = int32(x)
				offsetY = int32(y)
			}
		})

		//-----------------------------------------------------------
		// マウスを移動したときのイベントハンドラ
		//-----------------------------------------------------------
		window1.Connect("motion-notify-event", func(win *gtk.ApplicationWindow, event *gdk.Event) {
			if dragging {
				e := gdk.EventMotionNewFromEvent(event)
				
				// マウスの移動量を取得
				x, y := e.MotionVal()
				dx := int32(x) - offsetX
				dy := int32(y) - offsetY
				if dx < 0 { dx *= -1 }
				if dy < 0 { dy *= -1 }
				
				// Dragと判定される閾値を取得
				dragcx, w32err := win32.GetSystemMetrics(win32.SM_CXDRAG)
				if w32err != win32.NO_ERROR {
					ShowErrorDialog(window1, fmt.Errorf("GetSystemMetricsの失敗"))
				}
				dragcy, w32err := win32.GetSystemMetrics(win32.SM_CYDRAG)
				if w32err != win32.NO_ERROR {
					ShowErrorDialog(window1, fmt.Errorf("GetSystemMetricsの失敗"))
				}
				
				// 移動量が閾値を超えていたらDrag開始
				if dragcx < dx || dragcy < dy {
					cf := []uint{uint(win32.CF_HDROP)}
					_, _, err := OLE_IDropSource_Start.Call(uintptr(hwnd), uintptr(WM_DROPGETDATA), uintptr(WM_DROPDRAGEND), uintptr(unsafe.Pointer(&cf[0])), uintptr(1), uintptr(win32.DROPEFFECT_COPY | win32.DROPEFFECT_MOVE | win32.DROPEFFECT_LINK | win32.DROPEFFECT_SCROLL))
					if err.Error() != "The operation completed successfully." {
						ShowErrorDialog(window1, fmt.Errorf("OLE_IDropSource_Startの失敗"))
					}
					dragging = false
				}
			}
		})

		//-----------------------------------------------------------
		// マウスボタンを離したときのイベントハンドラ
		//-----------------------------------------------------------
		window1.Connect("button-release-event", func(win *gtk.ApplicationWindow, event *gdk.Event) {
			e := gdk.EventButtonNewFromEvent(event)
			if e.Button() == gdk.BUTTON_PRIMARY {
				dragging = false
			}
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
	// アプリケーション終了時のイベント
	///////////////////////////////////////////////////////////////////////////
	application.Connect("shutdown", func() {
		log.Println("application shutdown")

		// メッセージ受信用ウィンドウの破棄
		win32.DestroyWindow(hwnd)

		// oleのUninitialize
		win32.OleUninitialize()

	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	os.Exit(application.Run(os.Args))
}


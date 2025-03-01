// シンプルなウィンドウ表示
// -- gladeファイルを読み込んでウィンドウを表示する --
package main

import (
	_ "embed"
	"log"
	"os"

	"golang.org/x/sys/windows"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/01_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application

//-----------------------------------------------------------------------------
// メイン
//-----------------------------------------------------------------------------
func main() {
	const appID = "org.example.myapp"
	var window1 *gtk.ApplicationWindow
	var err error
	
	///////////////////////////////////////////////////////////////////////////
	// Mutexを作成または開く（必須ではない）
	// 多重起動を防止する場合、Mutexを使う
	// ⇒Windowsの場合、ApplicationNew()の第2引数では制御できなさそう
	//  「glib.APPLICATION_FLAGS_NONE」の場合、多重起動防止として機能せず且つ不具合が出る
	//  そのため、「glib.APPLICATION_NON_UNIQUE」を指定して多重起動許可にしておく
	///////////////////////////////////////////////////////////////////////////
	mutex, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(appID))
	if err != nil {
		log.Fatal("アプリケーションは既に起動しています")
	}
	defer windows.CloseHandle(mutex)
	
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
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	// Runに引数を渡してるけど、application側で取りだすより
	// go側でグローバル変数にでも格納した方が楽
	os.Exit(application.Run(os.Args))
}


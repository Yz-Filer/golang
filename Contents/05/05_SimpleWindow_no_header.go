// シンプルなウィンドウ表示
// -- 半透明にしてヘッダを消してマウスで移動させる --
package main

import (
	_ "embed"
	"log"
	"os"

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

		// ウィンドウを半透明にする
		// window1.SetOpacity(0.0)では透明にならなかった。
		// 色はRGBAの順に0～1で指定（A=0が透明）
		color := gdk.NewRGBA(1.0, 0.97, 0.82, 0.8)
		window1.OverrideBackgroundColor(gtk.STATE_FLAG_NORMAL, color)


		// ウィンドウのヘッダーバーを消す
		window1.SetDecorated(false)

		var dragging bool
		var offsetX, offsetY int

		//-----------------------------------------------------------
		// マウスボタンを押したときのイベントハンドラ
		//-----------------------------------------------------------
		window1.Connect("button-press-event", func(win *gtk.ApplicationWindow, event *gdk.Event) {
			e := gdk.EventButtonNewFromEvent(event)
			if e.Button() == gdk.BUTTON_PRIMARY {
				dragging = true
				x, y := e.MotionVal()
				offsetX = int(x)
				offsetY = int(y)
			}
		})

		//-----------------------------------------------------------
		// マウスを移動したときのイベントハンドラ
		//-----------------------------------------------------------
		window1.Connect("motion-notify-event", func(win *gtk.ApplicationWindow, event *gdk.Event) {
			if dragging {
				e := gdk.EventMotionNewFromEvent(event)
				x, y := e.MotionVal()
				dx := int(x) - offsetX
				dy := int(y) - offsetY
				winX, winY := win.GetPosition()
				win.Move(winX+dx, winY+dy)
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
	// アプリケーション終了時のイベント（必須ではない）
	///////////////////////////////////////////////////////////////////////////
	application.Connect("shutdown", func() {
		log.Println("application shutdown")
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	os.Exit(application.Run(os.Args))
}


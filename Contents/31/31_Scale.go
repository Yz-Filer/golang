package main

import (
	_ "embed"
	"log"
	"os"
	"strconv"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/31_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application

var FontSize = 12.0 * 1024.0

//-----------------------------------------------------------------------------
// 書式の設定
//-----------------------------------------------------------------------------
func ApplyStyle(widget gtk.IWidget, scale float64) error {
	// プロバイダーを作成
	cssProvider, err := gtk.CssProviderNew()
	if err != nil {
		return err
	}
	
	// コンテキストを取得
	var context *gtk.StyleContext
	context, err = widget.(*gtk.TextView).GetStyleContext()
	if err != nil {
		return err
	}
	
	// CSS文字列を作成
	cssStr := "text, .view {\n"
	cssStr+= "  font-family: MS Gothic;\n"
	cssStr+= "  font-size: " + strconv.Itoa(int(FontSize * scale / 100.0 / 1024.0)) + "pt;\n"
	cssStr+= "}"
	
	// CSSをロード
	err = cssProvider.LoadFromData(cssStr)
	if err != nil {
		return err
	}
	
	// 書式を反映
	context.AddProvider(cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	
	return nil
}

//-----------------------------------------------------------------------------
// メイン
//-----------------------------------------------------------------------------
func main() {
	const appID = "org.example.myapp"
	var window1 *gtk.ApplicationWindow
	var builder *gtk.Builder
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
		window1, builder, err = GetObjFromGlade[*gtk.ApplicationWindow](nil, window1Glade, "MAIN_WINDOW")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからtextviewを取得
		textView1, _, err := GetObjFromGlade[*gtk.TextView](builder, "", "TEXTVIEW")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからadjustmentを取得
		adjustment1, _, err := GetObjFromGlade[*gtk.Adjustment](builder, "", "ADJUSTMENT")
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
		// マウスホイール（垂直）が回転した時、拡大/縮小
		//-----------------------------------------------------------
		textView1.Connect("scroll_event", func(self *gtk.TextView, e *gdk.Event) bool {
			event := gdk.EventScrollNewFromEvent(e)
			
			// ctrlキー + 垂直方向の場合
			if event.State() == gdk.CONTROL_MASK && event.DeltaY() != 0 {
				// マウスホイールの変動量を10刻みにする
				scale := adjustment1.GetValue() - float64(event.DeltaY()) * 10.0
				
				// スライダーを移動
				// ※書式設定は"value_changed"シグナル内で実行
				adjustment1.SetValue(scale)
				
				// シグナルを伝播しない
				return true
			}
			
			return false
		})
		
		//-----------------------------------------------------------
		// スケールの値が変更された時の処理
		//-----------------------------------------------------------
		adjustment1.Connect("value_changed", func() {
			// 変動量を5刻みにする
			scale := float64(int(adjustment1.GetValue() / 5.0)) * 5.0
			adjustment1.SetValue(scale)
			
			// 書式を設定
			err := ApplyStyle(textView1, scale)
			if err != nil {
				ShowErrorDialog(window1, err)
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
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	// Runに引数を渡してるけど、application側で取りだすより
	// go側でグローバル変数にでも格納した方が楽
	os.Exit(application.Run(os.Args))
}


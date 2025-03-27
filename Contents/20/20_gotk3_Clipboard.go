// クリップボード
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/20_MainWindow.glade
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
		// テキストをクリップボードへ送る
		//-----------------------------------------------------------
		btnTxtTo.Connect("clicked", func() {
			text, err := entryText.GetText()
			if err != nil {
				ShowErrorDialog(window1, err)
				return
			}
			clipboard.SetText(text)
		})
		
		//-----------------------------------------------------------
		// テキストをクリップボードから取得する
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


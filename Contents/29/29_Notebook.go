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

//go:embed glade/29_MainWindow.glade
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

		// gladeからNotebookを取得
		note, _, err := GetObjFromGlade[*gtk.Notebook](builder, "", "NOTEBOOK")
		if err != nil {
			log.Fatal(err)
		}

		// gladeからページを追加するボタンを取得
		btnAdd, _, err := GetObjFromGlade[*gtk.Button](builder, "", "BUTTON_ADD")
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
		
		
		
		count := 1
		
		//-----------------------------------------------------------
		// ページの追加ボタン押下時の処理
		//-----------------------------------------------------------
		btnAdd.Connect("clicked", func() {
			// gladeからBoxを取得
			box1, builder, err := GetObjFromGlade[*gtk.Box](nil, window1Glade, "BOX")
			if err != nil {
				ShowErrorDialog(window1, err)
			}
			
			// gladeからBox内の表示ボタンを取得
			btnShow, _, err := GetObjFromGlade[*gtk.Button](builder, "", "BUTTON")
			if err != nil {
				ShowErrorDialog(window1, err)
			}
			
			// gladeからEventBoxを取得
			eventBox1, _, err := GetObjFromGlade[*gtk.EventBox](builder, "", "EVENTBOX")
			if err != nil {
				ShowErrorDialog(window1, err)
			}
			
			// gladeからLabelを取得
			label1, _, err := GetObjFromGlade[*gtk.Label](builder, "", "LABEL")
			if err != nil {
				ShowErrorDialog(window1, err)
			}
			
			// gladeからページCloseボタンを取得
			btnClose, _, err := GetObjFromGlade[*gtk.Button](builder, "", "BUTTON_CLOSE")
			if err != nil {
				ShowErrorDialog(window1, err)
			}
			
			// タブのラベル文字列を設定
			id := count
			count++
			label1.SetText(fmt.Sprintf("%02d", id))
			
			// タブの右端に、box1の内容のページとlabel1・btnCloseのタブを持ったページを追加
			// ※label1・btnCloseはeventBox1の中
			note.AppendPage(box1, eventBox1)
			
			// タブをマウスで並べ替え可
			note.SetTabReorderable(box1, true)
			
			//-----------------------------------------------------------
			// 表示ボタン押下時の処理
			//-----------------------------------------------------------
			btnShow.Connect("clicked", func() {
				log.Printf("このページのIDは「%d」です。\n", id)
			})
			
			//-----------------------------------------------------------
			// ページCloseボタン押下時の処理
			//-----------------------------------------------------------
			btnClose.Connect("clicked", func() {
				// ページを削除
				note.RemovePage(note.PageNum(box1))
			})
			
			//-----------------------------------------------------------
			// label1のシグナルハンドラはEventBoxが代替して処理
			// ※右クリックで閉じたい時とかに使う
			//-----------------------------------------------------------
			eventBox1.Connect("button-press-event", func(eventBox *gtk.EventBox, ev *gdk.Event) bool {
				log.Println("EventBoxがクリックされました。")
				
				// イベントを伝播
				return false
			})
		})
		
		//-----------------------------------------------------------
		// ページ選択が変更になった時の処理
		//-----------------------------------------------------------
		note.Connect("switch-page", func(notebook *gtk.Notebook, page *gtk.Widget, page_num int) {
			log.Printf("%d番目のページに切り替わりました。\n", page_num + 1)
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


// DnD
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
	"unsafe"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/22_MainWindow.glade
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
		
		// gladeからEntryを取得
		entryText1, _, err := GetObjFromGlade[*gtk.Entry](builder, "", "ENTRY1")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからEntryを取得
		entryText2, _, err := GetObjFromGlade[*gtk.Entry](builder, "", "ENTRY2")
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
		// DnDターゲットの設定
		//-----------------------------------------------------------

		targetEntryText, err := gtk.TargetEntryNew("text/plain", gtk.TARGET_SAME_APP, 0)
		if err != nil {
			log.Fatal(err)
		}
		
		targetEntryURI, err := gtk.TargetEntryNew("text/uri-list", gtk.TARGET_OTHER_APP, 1)
		if err != nil {
			log.Fatal(err)
		}
		targetEntriesText := []gtk.TargetEntry{*targetEntryText, *targetEntryURI}
		
		//-----------------------------------------------------------
		// Drag側の設定
		//-----------------------------------------------------------
		
		// Dragソースの設定(text)
		entryText1.DragSourceSet(gdk.BUTTON1_MASK, targetEntriesText, gdk.ACTION_COPY | gdk.ACTION_MOVE)
		

		// DnD開始時のシグナルハンドラ
		entryText1.Connect("drag-begin", func(entry *gtk.Entry, context *gdk.DragContext) {
			fmt.Println("begin")

			// Dragされたターゲットリストを表示
			targetList := context.ListTargets()
			n := targetList.Length()
			for i := uint(0); i < n; i++ {
				switch uintptr(targetList.NthData(i).(unsafe.Pointer)) {
					case uintptr(gdk.GdkAtomIntern("text/plain", false)):
						fmt.Println("start target: text/plain")
					case uintptr(gdk.GdkAtomIntern("text/uri-list", false)):
						fmt.Println("start target: text/uri-list")
				}
			}
		})
		
		// MOVE発生時のシグナルハンドラ
		entryText1.Connect("drag-data-delete", func(entry *gtk.Entry, context *gdk.DragContext) {
			fmt.Println("delete")
		})
		
		// DnD失敗時のシグナルハンドラ
		entryText1.Connect("drag-failed", func(entry *gtk.Entry, context *gdk.DragContext, result int) bool {
			fmt.Println("failed", result)
			
			// 失敗が処理済みならtrue
			return true
		})
		
		// DnD終了時のシグナルハンドラ
		entryText1.Connect("drag-end", func(entry *gtk.Entry, context *gdk.DragContext) {
			fmt.Println("end")
		})

		// データ受信要求への応答
		entryText1.Connect("drag-data-get", func(entry *gtk.Entry, context *gdk.DragContext, data *gtk.SelectionData, info, time uint) {
/*
			// 要求されたデータをSelectionData引数へ格納
			// ※Entryの場合は、選択領域が自動的に設定される。変更は出来なさそう
			switch info {
				case 0:
					text, err := entryText1.GetText()
					if err != nil {
						ShowErrorDialog(window1, err)
						return
					}
					ok := data.SetText(text)
					if !ok {
						ShowErrorDialog(window1, fmt.Errorf("DnDテキストデータ送信に失敗"))
						return
					}
			}
*/
		})
		
		
		//-----------------------------------------------------------
		// Drop側の設定
		//-----------------------------------------------------------
		
		// Drop先の設定
		entryText2.DragDestSet(gtk.DEST_DEFAULT_ALL, targetEntriesText, gdk.ACTION_COPY | gdk.ACTION_MOVE)
		
		// データ受信
		entryText2.Connect("drag-data-received", func(entry *gtk.Entry, context *gdk.DragContext, x, y int, data *gtk.SelectionData, info, time uint) {
			// 受信したデータを表示
			switch info {
				case 0:
					// 受信したデータをEntryに設定
					// ※Entryの場合は、自動的に設定される。変更は出来なさそう
//					entryText2.SetText(data.GetText())
					fmt.Printf("text/plain:\n  %s\n", data.GetText())
				case 1:
					entryText2.SetText(strings.ReplaceAll(string(data.GetData()), "\r\n",", "))
					fmt.Printf("text/uri-list:\n  %s\n", strings.ReplaceAll(string(data.GetData()), "\r\n",", "))
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


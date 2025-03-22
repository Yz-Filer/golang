// ディレクトリ配下の更新を監視
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/zzl/go-win32api/win32"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/19_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application
var window1 *gtk.ApplicationWindow




//-----------------------------------------------------------------------------
// メイン
//-----------------------------------------------------------------------------
func main() {
	const appID = "org.example.myapp"
	var err error

	//-----------------------------------------------------------
	// ディレクトリ配下の更新監視
	// ※ディレクトリ単位にオープン/スタート/クローズが必要
	//-----------------------------------------------------------
	
	targetDir1 := `D:\test`
	
	// ディレクトリ監視をオープン
	err = DirWatchOpen(targetDir1, false)
	if err != nil {
		log.Fatal(err)
	}
	
	// ディレクトリ監視を開始
	err = DirWatchStart(targetDir1)
	if err != nil {
		log.Fatal(err)
	}
	
	// 監視登録されてるディレクトリ一覧を表示
	fmt.Println("ディレクトリ監視に登録されてるディレクトリの一覧（trueが監視中）")
	for k, v := range DirWatchMap {
		fmt.Println("  ", k, v.watching)
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
		// カスタムシグナル
		//-----------------------------------------------------------

		// カスタムシグナル（ディレクトリ監視）を作成
		_, err = glib.SignalNewV("directory_changed", glib.TYPE_POINTER, 2, glib.TYPE_STRING, glib.TYPE_UINT)
		if err != nil {
			log.Fatal("Could not create signal: ", err)
		}

		//-----------------------------------------------------------
		// ディレクトリ更新検知のシグナル処理
		//-----------------------------------------------------------
		window1.Connect("directory_changed", func(parent *gtk.ApplicationWindow, directory string, dwErrorCode uint) {
			// ディレクトリ更新検知でエラーが発生していた場合
			if dwErrorCode != uint(win32.ERROR_SUCCESS) {
				ShowErrorDialog(parent, fmt.Errorf("ReadDirectoryChangesW コールバックエラー: %v", dwErrorCode))
				return
			}
			
			fmt.Printf("ディレクトリの更新を検知しました：%s\n", directory)
			
			// 検知後は監視が外れてるためwatching項目をfalseに変更
			dwMap, ok := DirWatchMap[directory]
			if !ok {
				ShowErrorDialog(parent, fmt.Errorf("ディレクトリ監視の管理情報が見つかりませんでした"))
				return
			}
			dwMap.watching = false
			DirWatchMap[directory] = dwMap
			
			// ディレクトリ監視を再開
			// ※停止中のイベントもある程度バッファに溜まってる
			err = DirWatchStart(directory)
			if err != nil {
				ShowErrorDialog(parent, err)
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

		// ディレクトリ更新の監視を停止
		DirWatchClose(targetDir1)
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	// Runに引数を渡してるけど、application側で取りだすより
	// go側でグローバル変数にでも格納した方が楽
	os.Exit(application.Run(os.Args))
}


// 閉じるボタン/最小化ボタン押下時にタスクトレイへ格納
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

//go:embed glade/06_Menu.glade
var menu1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application

// メニューを作成
func buildMenu(parent *gtk.ApplicationWindow, stIcon *gtk.StatusIcon) (*gtk.Menu, error) {
	// gladeからメニューを取得
	menu1, builder, err := GetObjFromGlade[*gtk.Menu](nil, menu1Glade, "STATUSICON_MENU")
	if err != nil {
		return nil, err
	}
	
	// gladeからmenuItem1を取得
	menuItem1, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_SHOW")
	if err != nil {
		return nil, err
	}
	
	// gladeからmenuItem2を取得
	menuItem2, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_QUIT")
	if err != nil {
		return nil, err
	}
	
	// menuItem1選択時に親ウィンドウをタスクトレイから出す
	menuItem1.Connect("activate", func(){
		if parent != nil {
			stIcon.SetVisible(false)
			parent.SetPosition(gtk.WIN_POS_NONE)
			parent.Present()
			parent.Deiconify()
		}
	})
	
	// menuItem2選択時にアプリを終了
	menuItem2.Connect("activate", func(){
		stIcon.SetVisible(false)
		menu1.Destroy()
		application.Quit()
	})
	
	return menu1, nil
}

func main() {
	const appID = "org.example.myapp"

	var window1 *gtk.ApplicationWindow
	var statusIcon *gtk.StatusIcon
	var err error
	
	///////////////////////////////////////////////////////////////////////////
	// Mutexを作成または開く
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
		
		// ウィンドウのプロパティを設定
		window1.SetPosition(gtk.WIN_POS_MOUSE)


		// リソースからタスクトレイアイコンを設定
		statusIcon, err = gtk.StatusIconNewFromPixbuf(iconPixbuf)
		if err != nil {
			log.Fatal("Could not create status icon: ", err)
		}
		defer statusIcon.Unref()
		
		// タスクトレイアイコン用のメニューを作成
		menu, err := buildMenu(window1, statusIcon)
		if err != nil {
			log.Fatal("Could not create status menu: ", err)
		}
		
		//-----------------------------------------------------------
		// ウィンドウ最小化時の処理
		// Linuxは挙動が異なるかも
		//-----------------------------------------------------------
		window1.Connect("window-state-event", func(parent *gtk.ApplicationWindow, event *gdk.Event) bool {
			// gdk.EventWindowState を取得
			windowStateEvent := gdk.EventWindowStateNewFromEvent(event)
			
			if windowStateEvent != nil {
				// 最小化された場合、タスクトレイに格納
				if windowStateEvent.ChangedMask() == (gdk.WINDOW_STATE_ICONIFIED | gdk.WINDOW_STATE_FOCUSED) {
					log.Println("ウィンドウを最小化し、タスクトレイに格納しました")
					window1.Hide()
					statusIcon.SetVisible(true)
				}
			}
			
			// イベントの伝播を停止
			return true
		})

		//-----------------------------------------------------------
		// 閉じるボタンが押されたら、タスクトレイに格納
		//-----------------------------------------------------------
		window1.Connect("delete-event", func(parent *gtk.ApplicationWindow, event *gdk.Event) bool {
			log.Println("ウィンドウをタスクトレイに格納しました")
			
			window1.Hide()
			statusIcon.SetVisible(true)
			
			// クローズ処理をキャンセル
			return true
		})
		
		//*********************************************************************
		// タスクトレイの処理
		//*********************************************************************
		
		//-----------------------------------------------------------
		// タスクトレイアイコン右クリック時にメニューを表示
		// ※メニューは項目を選択しないと消えない
		// ※アイコンの位置にメニューが表示されない
		//-----------------------------------------------------------
		statusIcon.Connect("popup-menu", func(statusIcon *gtk.StatusIcon, button uint, activateTime uint) {
			menu.PopupAtStatusIcon(statusIcon, gdk.Button(button), uint32(activateTime))
		})

		//-----------------------------------------------------------
		// タスクトレイアイコンをクリックしたらウィンドウを再表示
		// Connect activate (double click or single click, depends on the desktop environment)
		//-----------------------------------------------------------
		statusIcon.Connect("activate", func() {
			statusIcon.SetVisible(false)
			window1.SetPosition(gtk.WIN_POS_NONE)		// これを指定しておくと元の位置に表示されるみたい
			window1.Present()
			window1.Deiconify()							// 最小化されたウィンドウを元に戻す
		})
		
		// 起動時はタスクトレイアイコンを非表示にする
		statusIcon.SetVisible(false)
		
		// アプリケーションを設定
		window1.SetApplication(application)

		// ウィンドウの表示
		window1.ShowAll()
	})

	os.Exit(application.Run(os.Args))
}

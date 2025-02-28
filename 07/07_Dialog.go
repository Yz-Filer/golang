// メッセージダイアログの表示
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/07_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application


// エラーメッセージを表示する
func ShowErrorDialog(parent *gtk.ApplicationWindow, err error) {
	dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "エラーが発生しました")
	dialog.FormatSecondaryText("%s", err.Error())
	dialog.SetTitle ("error")
	dialog.Run()
	dialog.Destroy()
}


func main() {
	const appID = "org.example.myapp"
	var window1 *gtk.ApplicationWindow
	var builder *gtk.Builder
	var statusbar *gtk.Statusbar
	var button1, button2, button3 *gtk.Button
	var err error
	
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
		window1, builder, err = GetObjFromGlade[*gtk.ApplicationWindow](nil, window1Glade, "MAIN_WINDOW")
		if err != nil {
			log.Fatal(err)
		}
		
		// リソースからアプリケーションのアイコンを設定
		iconPixbuf, err := gdk.PixbufNewFromDataOnly(icon)
		if err != nil {
			log.Fatal("Could not create pixbuf from bytes: ", err)
		}
		defer iconPixbuf.Unref()
		
		// ウィンドウにアイコンを設定
		window1.SetIcon(iconPixbuf)
		
		// ウィンドウのプロパティを設定
		window1.SetPosition(gtk.WIN_POS_MOUSE)

		// gladeからステータスバーを取得
		statusbar, _, err = GetObjFromGlade[*gtk.Statusbar](builder, "", "STATUSBAR")
		if err != nil {
			log.Fatal(err)
		}
		// ステータスバーへ文字列を表示（第一引数は、文字列の識別番号）
		statusbar.Push(0, "ステータスはここに表示")
		
		// gladeから標準メッセージダイアログボタンを取得
		button1, _, err = GetObjFromGlade[*gtk.Button](builder, "", "STANDARD_MESSAGE_DIALOG")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからカスタムメッセージダイアログボタンを取得
		button2, _, err = GetObjFromGlade[*gtk.Button](builder, "", "CUSTOM_MESSAGE_DIALOG")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからモードレスダイアログボタンを取得
		button3, _, err = GetObjFromGlade[*gtk.Button](builder, "", "MODELESS_DIALOG")
		if err != nil {
			log.Fatal(err)
		}
		
		
		//-----------------------------------------------------------
		// 「標準メッセージダイアログ」ボタン押下時
		//-----------------------------------------------------------
		button1.Connect("clicked", func() {

			// ダイアログの表示
			dialog := gtk.MessageDialogNew(window1, gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_INFO, gtk.BUTTONS_OK_CANCEL, "標準メッセージダイアログ")
			defer dialog.Destroy()

			dialog.SetTitle ("タイトル")
			dialog.FormatSecondaryText("このダイアログはgtk3標準のメッセージダイアログです")
			
			ret := dialog.Run()

			// 標準メッセージダイアログの応答処理
			switch ret {
				case gtk.RESPONSE_OK:
					log.Println("標準メッセージダイアログで、OKが押されました")
				case gtk.RESPONSE_CANCEL:
					log.Println("標準メッセージダイアログで、CANCELが押されました")
				case gtk.RESPONSE_DELETE_EVENT:
					log.Println("標準メッセージダイアログが閉じられました")
			}
		})
		
		
		//-----------------------------------------------------------
		// 「カスタムメッセージダイアログ」ボタン押下時
		//-----------------------------------------------------------
		button2.Connect("clicked", func() {
			// REJECT, ACCEPT, OK, CANCEL, CLOSE, YES, NO, APPLY, HELP
			buttonFlg := [9]string{"", "", "OK", "CANCEL", "", "YES", "NO", "", ""}
			
			// ダイアログの作成
			dialog1, _, err := CustomMessageDialogNew(window1, "タイトル", gtk.MESSAGE_INFO, buttonFlg, "メッセージダイアログ", "このダイアログは自作のメッセージダイアログです")
			if err != nil {
				log.Println("Could not create dialog: ", err)
				ShowErrorDialog(window1, err)
				return
			}
			defer dialog1.Destroy()
			
			// ダイアログをモーダルに設定して表示
			dialog1.SetModal(true)
			ret := dialog1.Run()
			
			// カスタムメッセージダイアログの応答処理
			switch ret {
				case gtk.RESPONSE_REJECT:
					log.Println("カスタムメッセージダイアログで、REJECTが押されました")
				case gtk.RESPONSE_ACCEPT:
					log.Println("カスタムメッセージダイアログで、ACCEPTが押されました")
				case gtk.RESPONSE_OK:
					log.Println("カスタムメッセージダイアログで、OKが押されました")
				case gtk.RESPONSE_CANCEL:
					log.Println("カスタムメッセージダイアログで、CANCELが押されました")
				case gtk.RESPONSE_CLOSE:
					log.Println("カスタムメッセージダイアログで、CLOSEが押されました")
				case gtk.RESPONSE_YES:
					log.Println("カスタムメッセージダイアログで、YESが押されました")
				case gtk.RESPONSE_NO:
					log.Println("カスタムメッセージダイアログで、NOが押されました")
				case gtk.RESPONSE_APPLY:
					log.Println("カスタムメッセージダイアログで、APPLYが押されました")
				case gtk.RESPONSE_HELP:
					log.Println("カスタムメッセージダイアログで、HELPが押されました")
				case gtk.RESPONSE_DELETE_EVENT:
					log.Println("カスタムメッセージダイアログが閉じられました")
				default:
					log.Println("カスタムメッセージダイアログで、想定してないレスポンスを受信しました")
			}
		})
		
		
		//-----------------------------------------------------------
		// 「モードレスダイアログ」ボタン押下時
		//-----------------------------------------------------------
		button3.Connect("clicked", func() {
			// REJECT, ACCEPT, OK, CANCEL, CLOSE, YES, NO, APPLY, HELP
			buttonFlg := [9]string{"", "", "", "", "", "", "", "", ""}
			
			// ダイアログの作成
			// 「gtk.MESSAGE_OTHER」を指定して、アイコンを非表示にする（スピナー表示のため）
			dialog1, spinner1, err := CustomMessageDialogNew(window1, "タイトル", gtk.MESSAGE_OTHER, buttonFlg, "モードレスダイアログ", "3秒経ったらクローズします")
			if err != nil {
				log.Println("Could not create dialog: ", err)
				ShowErrorDialog(window1, err)
				return
			}
			
			// ダイアログが表示された時のシグナル処理
			// 3秒待ってクローズ
			dialog1.Connect("show", func(dlg *gtk.Dialog) {
				// 複数ダイアログ開いた時に並列に動かすためにgoルーチンを使用
				// （後から開いたダイアログの方がUIのメインループを優先して使うっぽい）
				// 画像ビューアのようにマウス操作するようなアプリではアクティブにした方が優先されると思うからgoルーチン不要かも？
				go func() {
					for i := 0; i <= 300; i++ {
						DoEvents()
						time.Sleep(10 * time.Millisecond)
						
						// goルーチン内のUI操作はglib.IdleAddを使って安全に実行
						glib.IdleAdd(func() {
							dlg.SetTitle(fmt.Sprintf("%d", 300-i))
						})
					}
					
					// goルーチン内のUI操作はglib.IdleAddを使って安全に実行
					glib.IdleAdd(func() {
						dlg.Destroy()
					})
					log.Println("モードレスダイアログが閉じられました")
				}()
			})
			
			// スピナーをShow・Start
			spinner1.Show()
			spinner1.Start()
			
			// ダイアログをモードレスに設定して表示
			dialog1.SetModal(false)
			dialog1.Show()
			
			// モードレスの場合隠れることがあるから、最前面に表示
			dialog1.SetKeepAbove(true)
			dialog1.SetKeepAbove(false)
		})



		// アプリケーションを設定
		window1.SetApplication(application)

		// ウィンドウの表示
		window1.ShowAll()
	})

	os.Exit(application.Run(os.Args))
}

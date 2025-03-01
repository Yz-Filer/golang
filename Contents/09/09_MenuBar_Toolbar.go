// メニューと色々なダイアログの表示
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

//go:embed glade/09_MainWindow.glade
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

// メニュー「開く」を選択した場合、ファイル選択ダイアログを表示
func menuOpen(parent *gtk.ApplicationWindow) (string, error) {
	log.Println("Openを選択")
	if parent == nil {
		return "", fmt.Errorf("parent is null")
	}
	
	// ファイル選択ダイアログを作成
	fcd, err := gtk.FileChooserDialogNewWith2Buttons("Open", parent, gtk.FILE_CHOOSER_ACTION_OPEN, "ok (_O)", gtk.RESPONSE_OK, "cancel (_C)", gtk.RESPONSE_CANCEL)
	if err != nil {
		return "", err
	}
	defer fcd.Destroy()
	
	// 「OK」で終わった場合は、ファイル名を返却
	if fcd.Run() == gtk.RESPONSE_OK {
		return fcd.GetFilename(), nil
	}
	return "", nil
}

// メニュー「フォント」を選択した場合、フォント選択ダイアログを表示
func menuFont(parent *gtk.ApplicationWindow) (string, error) {
	log.Println("Fontを選択")
	if parent == nil {
		return "", fmt.Errorf("parent is null")
	}
	
	// フォント選択ダイアログを作成
	fcd, err := gtk.FontChooserDialogNew("Font", parent)
	if err != nil {
		return "", err
	}
	defer fcd.Destroy()
	
	// 「OK」で終わった場合は、フォント名を返却
	if fcd.Run() == gtk.RESPONSE_OK {
		return fcd.GetFont(), nil
	}
	return "", nil
}

// メニュー「色」を選択した場合、色選択ダイアログを表示
func menuColor(parent *gtk.ApplicationWindow) (string, error) {
	log.Println("Colorを選択")
	if parent == nil {
		return "", fmt.Errorf("parent is null")
	}
	
	// 色選択ダイアログを作成
	ccd, err := gtk.ColorChooserDialogNew("Color", parent)
	if err != nil {
		return "", err
	}
	defer ccd.Destroy()
	
	// 「OK」で終わった場合は、RGBAを返却
	if ccd.Run() == gtk.RESPONSE_OK {
		return ccd.GetRGBA().String(), nil
	}
	return "", nil
}

// メニュー「カレンダー」を選択した場合、日付選択ダイアログを表示
func menuCalendar(parent *gtk.ApplicationWindow) (string, error) {
	log.Println("Calendarを選択")
	if parent == nil {
		return "", fmt.Errorf("parent is null")
	}
	
	// ダイアログを作成 
	dialog, err := gtk.DialogNew()
	if err != nil {
		return "", err
	}
	defer dialog.Destroy()
	
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)

	// カレンダーウィジェットを作成して追加
	calendar, err := gtk.CalendarNew()
	if err != nil {
		return "", err
	}
	gca, err := dialog.GetContentArea()
	if err != nil {
		return "", err
	}
	gca.Add(calendar)

	// 実行日を設定
	year, month, day := time.Now().Date()
	calendar.SelectMonth(uint(month) - 1, uint(year))
	calendar.SelectDay(uint(day))

	dialog.ShowAll()

	// ダブルクリックされた場合、レスポンスコード「OK」を送信
	calendar.Connect("day_selected_double_click", func() {
		dialog.Response(gtk.RESPONSE_OK)
	})

	// 「OK」で終わった場合は、日付を返却
	if dialog.Run() == gtk.RESPONSE_OK {
		year, month, day := calendar.GetDate()
		return fmt.Sprintf("%04d/%02d/%02d", year, int(month) + 1, day), nil
	}

	return "", nil
}

// メニュー「ABOUT」を選択した場合、ABOUTダイアログを表示
func menuAbout(parent *gtk.ApplicationWindow) error {
	log.Println("Aboutを選択")
	if parent == nil {
		return fmt.Errorf("parent is null")
	}
	
	// ABOUTダイアログを作成
	abd, err := gtk.AboutDialogNew()
	if err != nil {
		return err
	}
	defer abd.Destroy()
	
	abd.SetTransientFor(parent)
	
	// ロゴに親アイコンを設定
	parentIcon, err := parent.GetIcon()
	if err == nil {
		abd.SetLogo(parentIcon)
	}
	
	abd.SetProgramName("プログラムの名前")
	abd.SetVersion("バージョン x.xx")
	abd.SetComments("コメントをここに記載")
	abd.SetWebsiteLabel("ウェブサイト")
	abd.SetWebsite("ウェブサイトのURL")
	
	abd.SetCopyright("Copyright (c) 20xx Firstname Lastname")
	abd.SetLicense("LicenseTypeを指定しない場合、ここにライセンスを記載")
	abd.SetLicenseType(gtk.LICENSE_MIT_X11)
	abd.SetWrapLicense(true)	// ライセンス表示を改行する

	abd.SetAuthors([]string{"開発した人"})
	abd.SetDocumenters([]string{"ドキュメントを作成した人"})
	abd.SetTranslatorCredits("翻訳した人")
	abd.SetArtists([]string{"グラフィックデザイン、UIデザインなどに貢献した人"})
	
	abd.Run()
	return nil
}

// ツールボタンのシグナル処理を設定
func buildToolButton(parent *gtk.ApplicationWindow, builder *gtk.Builder) error {
	// gladeからtoolButtonOpenを取得
	toolButtonOpen, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "TOOLBUTTON_OPEN")
	if err != nil {
		return err
	}
	
	// gladeからtoolButtonFontを取得
	toolButtonFont, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "TOOLBUTTON_FONT")
	if err != nil {
		return err
	}
	
	// gladeからtoolButtonColorを取得
	toolButtonColor, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "TOOLBUTTON_COLOR")
	if err != nil {
		return err
	}
	
	// gladeからtoolButtonCalendarを取得
	toolButtonCalendar, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "TOOLBUTTON_CALENDAR")
	if err != nil {
		return err
	}
	
	// gladeからtoolButtonQuitを取得
	toolButtonQuit, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "TOOLBUTTON_QUIT")
	if err != nil {
		return err
	}
	
	// gladeからtoolButtonAboutを取得
	toolButtonAbout, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "TOOLBUTTON_ABOUT")
	if err != nil {
		return err
	}
	
	// toolButtonOpen選択時
	toolButtonOpen.Connect("clicked", func(){
		ret, err := menuOpen(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		if len(ret) != 0 {
			log.Println(ret)
		}
	})
	
	// toolButtonFont選択時
	toolButtonFont.Connect("clicked", func(){
		ret, err := menuFont(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		if len(ret) != 0 {
			log.Println(ret)
		}
	})
	
	// toolButtonColor選択時
	toolButtonColor.Connect("clicked", func(){
		ret, err := menuColor(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		if len(ret) != 0 {
			log.Println(ret)
		}
	})
	
	// toolButtonCalendar選択時
	toolButtonCalendar.Connect("clicked", func(){
		ret, err := menuCalendar(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		if len(ret) != 0 {
			log.Println(ret)
		}
	})
	
	// toolButtonQuit選択時にアプリを終了
	toolButtonQuit.Connect("clicked", func(){
		log.Println("toolButtonQuit")
		application.Quit()
	})
	
	// toolButtonAbout選択時
	toolButtonAbout.Connect("clicked", func(){
		err := menuAbout(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	return nil
}

// メニューアイテムのシグナル処理を設定
func buildMenuItem(parent *gtk.ApplicationWindow, builder *gtk.Builder) error {
	// gladeからmenuItemOpenを取得
	menuItemOpen, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_OPEN")
	if err != nil {
		return err
	}
	
	// gladeからmenuItemFontを取得
	menuItemFont, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_FONT")
	if err != nil {
		return err
	}
	
	// gladeからmenuItemColorを取得
	menuItemColor, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_COLOR")
	if err != nil {
		return err
	}
	
	// gladeからmenuItemCalendarを取得
	menuItemCalendar, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_CALENDAR")
	if err != nil {
		return err
	}
	
	// gladeからmenuItemQuitを取得
	menuItemQuit, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_QUIT")
	if err != nil {
		return err
	}
	
	// gladeからmenuItemAboutを取得
	menuItemAbout, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_ABOUT")
	if err != nil {
		return err
	}
	
	// menuItemOpen選択時
	menuItemOpen.Connect("activate", func(){
		ret, err := menuOpen(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		if len(ret) != 0 {
			log.Println(ret)
		}
	})
	
	// menuItemFont選択時
	menuItemFont.Connect("activate", func(){
		ret, err := menuFont(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		if len(ret) != 0 {
			log.Println(ret)
		}
	})
	
	// menuItemColor選択時
	menuItemColor.Connect("activate", func(){
		ret, err := menuColor(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		if len(ret) != 0 {
			log.Println(ret)
		}
	})
	
	// menuItemCalendar選択時
	menuItemCalendar.Connect("activate", func(){
		ret, err := menuCalendar(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		if len(ret) != 0 {
			log.Println(ret)
		}
	})
	
	// menuItemQuit選択時にアプリを終了
	menuItemQuit.Connect("activate", func(){
		log.Println("menuItemQuit")
		application.Quit()
	})
	
	// menuItemAbout選択時
	menuItemAbout.Connect("activate", func(){
		err := menuAbout(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	return nil
}

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
	// アプリケーション アクティブ時のイベント
	///////////////////////////////////////////////////////////////////////////
	application.Connect("activate", func() {
		// gladeからウィンドウを取得
		window1, builder, err = GetObjFromGlade[*gtk.ApplicationWindow](nil, window1Glade, "MENU_WINDOW")
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


		//*********************************************************************
		// メニュー/ツールボタンの処理
		//*********************************************************************

		// gladeからメニューアイテムを取得し、シグナル処理を設定
		err = buildMenuItem(window1, builder)
		if err != nil {
			log.Fatal(err)
		}

		// gladeからツールボタンを取得し、シグナル処理を設定
		err = buildToolButton(window1, builder)
		if err != nil {
			log.Fatal(err)
		}
		
		
		// アプリケーションを設定
		window1.SetApplication(application)

		// ウィンドウの表示
		window1.ShowAll()
	})

	os.Exit(application.Run(os.Args))
}

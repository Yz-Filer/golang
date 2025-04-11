package main

import (
	"container/ring"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/33_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

// undo/redo用構造体
type UndoRedoStr struct {
	isAvailable		bool
	isInsert		bool
	offset			int
	text			string
}

// ファイル用構造体
type FileStr struct {
	textView		*gtk.TextView
	textBuffer		*gtk.TextBuffer
	label			*gtk.Label
	directory		string
	fileName		string
	charset			string
	isCrLf			bool
	isWrap			bool
	isEdit			bool
	isNew			bool
	isUndoEnable	bool
	isRedoEnable	bool
	doingUndoRedo	bool
	editCount		int
	scale			float64
	ubuf			*ring.Ring
	rbuf			*ring.Ring
}

var (
	application *gtk.Application
	WorkDir string
	
	FileMap = make(map[string]FileStr)
	FontSize = 12.0 * 1024.0
	NewFileCount = 1
	FileId = 1
)

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
	
	// 実行ファイルのフルパスを取得
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to retrieve executable path: ", err)
	}
	
	// 実行ファイルのディレクトリを作業ディレクトリに設定
	WorkDir = filepath.Dir(exePath)
	
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
		
		// gladeからadjustmentを取得
		adjustment1, _, err := GetObjFromGlade[*gtk.Adjustment](builder, "", "ADJUSTMENT")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからNotebookを取得
		note, _, err := GetObjFromGlade[*gtk.Notebook](builder, "", "NOTEBOOK")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからnewボタンを取得
		btnNew, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "BUTTON_NEW")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからopenボタンを取得
		btnOpen, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "BUTTON_OPEN")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeから上書き保存ボタンを取得
		btnSave, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "BUTTON_SAVE")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeから名前を付けて保存ボタンを取得
		btnSaveAs, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "BUTTON_SAVEAS")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからundoボタンを取得
		btnUndo, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "BUTTON_UNDO")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからredoボタンを取得
		btnRedo, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "BUTTON_REDO")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからwrapボタンを取得
		btnWrap, _, err := GetObjFromGlade[*gtk.ToggleToolButton](builder, "", "BUTTON_WRAP")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからステータスバーのLabel(カーソル位置)を取得
		labelPoint, _, err := GetObjFromGlade[*gtk.Label](builder, "", "LABEL_POINT")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからステータスバーのLabel(改行コード)を取得
		labelCrLf, _, err := GetObjFromGlade[*gtk.Label](builder, "", "LABEL_CRLF")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからステータスバーのLabel(文字コード)を取得
		labelCharSet, _, err := GetObjFromGlade[*gtk.Label](builder, "", "LABEL_CHARSET")
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
		window1.SetDefaultSize(640, 480)
		
		
		nextPageScale := -1.0
		nextWrap := false
		
		//-----------------------------------------------------------
		// 選択ページ変更時の処理
		//-----------------------------------------------------------
		note.Connect("switch-page", func(self *gtk.Notebook, page *gtk.Widget, page_num int) {
			// 選択後のページのFileStrを取得
			fs, _, err := GetFileStrFromPageNum(note, page_num)
			if err != nil {
				ShowErrorDialog(window1, err)
				return
			}
			
			// 各ボタンの活性状態を設定
			btnSave.SetSensitive(!fs.isNew)
			btnRedo.SetSensitive(fs.isRedoEnable)
			btnUndo.SetSensitive(fs.isUndoEnable)
			
			// ステータスバーにカーソル位置を表示
			iter := fs.textBuffer.GetIterAtMark(fs.textBuffer.GetInsert())
			line := iter.GetLine() + 1
			offset := iter.GetLineOffset() + 1
			labelPoint.SetText(fmt.Sprintf("%d行、%d列", line, offset))
			
			// ステータスバーに改行コードを表示
			if fs.isCrLf {
				labelCrLf.SetText("CRLF")
			} else {
				labelCrLf.SetText("LF")
			}
			
			// ステータスバーに文字コードを表示
			labelCharSet.SetText(fs.charset)
			
			// 移動後ページのWrapボタンを設定
			if btnWrap.GetActive() != fs.isWrap {
				nextWrap = true
				btnWrap.SetActive(fs.isWrap)
			}
			
			// 移動後ページのscaleを設定
			nextPageScale = fs.scale
			adjustment1.SetValue(fs.scale)
		})
		
		//-----------------------------------------------------------
		// スケールの値が変更された時の処理
		//-----------------------------------------------------------
		adjustment1.Connect("value_changed", func() {
			// ページ切替時は、スライダーの設定のみ実施して終了
			if nextPageScale > 0 {
				adjustment1.SetValue(nextPageScale)
				nextPageScale = -1
				return
			}
			
			// 選択中のページのFileStrを取得
			fs, id, err := GetFileStrFromPageNum(note)
			if err != nil {
				ShowErrorDialog(window1, err)
				return
			}
			
			// 変動量を5刻みにする
			fs.scale = float64(int(adjustment1.GetValue() / 5.0)) * 5.0
			adjustment1.SetValue(fs.scale)
			FileMap[id] = fs
			
			// 書式を設定
			err = ApplyStyle(fs.textView, fs.scale)
			if err != nil {
				ShowErrorDialog(window1, err)
			}
		})
		
		//-----------------------------------------------------------
		// 新規ボタン押下時の処理
		//-----------------------------------------------------------
		btnNew.Connect("clicked", func() {
			err := AddPage(window1, btnUndo, btnRedo, note, adjustment1, labelPoint)
			if err != nil {
				ShowErrorDialog(window1, err)
			}
		})
		
		//-----------------------------------------------------------
		// 開くボタン押下時の処理
		//-----------------------------------------------------------
		btnOpen.Connect("clicked", func() {
			err := OpenFile(window1, btnUndo, btnRedo, note, adjustment1, labelPoint)
			if err != nil {
				ShowErrorDialog(window1, err)
			}
		})
		
		//-----------------------------------------------------------
		// 上書き保存ボタン押下時の処理
		//-----------------------------------------------------------
		btnSave.Connect("clicked", func() {
			err := SaveFile(note)
			if err != nil {
				ShowErrorDialog(window1, err)
			}
			
			// 各ボタンの活性状態を設定
			btnRedo.SetSensitive(false)
			btnUndo.SetSensitive(false)
		})
		
		//-----------------------------------------------------------
		// 名前を付けて保存ボタン押下時の処理
		//-----------------------------------------------------------
		btnSaveAs.Connect("clicked", func() {
			err := SaveAsFile(window1, note)
			if err != nil {
				if err.Error() == "cancel" {
					return
				}
				ShowErrorDialog(window1, err)
			}
			
			// 各ボタンの活性状態を設定
			btnSave.SetSensitive(true)
			btnRedo.SetSensitive(false)
			btnUndo.SetSensitive(false)
		})
		
		//-----------------------------------------------------------
		// undoボタン押下時の処理
		//-----------------------------------------------------------
		btnUndo.Connect("clicked", func() {
			// 選択中のページのFileStrを取得
			fs, id, err := GetFileStrFromPageNum(note)
			if err != nil {
				ShowErrorDialog(window1, err)
				return
			}
			
			// undoバッファから最新の値を取得
			value := fs.ubuf.Value.(UndoRedoStr)
			if !value.isAvailable {
				return
			}
			
			// undo/redoによる編集を開始
			fs.doingUndoRedo = true
			FileMap[id] = fs
			
			// 挿入データの場合は、削除
			// 削除データの場合は、挿入
			iterStart := fs.textBuffer.GetIterAtOffset(value.offset)
			if value.isInsert {
				iterEnd := fs.textBuffer.GetIterAtOffset(value.offset + len(value.text))
				fs.textBuffer.Delete(iterStart, iterEnd)
			} else {
				fs.textBuffer.Insert(iterStart, value.text)
			}
			
			// 次のredoバッファに保存し、enable
			fs.rbuf = fs.rbuf.Next()
			fs.rbuf.Value = value
			fs.isRedoEnable = true
			
			// 取りだし済みのデータを利用不可にし、1つ前のバッファへ
			fs.ubuf.Value = UndoRedoStr {isAvailable: false}
			fs.ubuf = fs.ubuf.Prev()
			
			// 利用不可データになった場合、disable
			if !fs.ubuf.Value.(UndoRedoStr).isAvailable {
				fs.isUndoEnable = false
			}
			
			// editCountが「1」→「0」で編集前まで戻されたデータ
			fs.editCount--
			if fs.editCount == 0 {
				fs.isEdit = false
				fs.label.SetText(fs.fileName)
			}
			
			// undo/redoによる編集を終了
			fs.doingUndoRedo = false
			
			btnRedo.SetSensitive(fs.isRedoEnable)
			btnUndo.SetSensitive(fs.isUndoEnable)
			FileMap[id] = fs
		})
		
		//-----------------------------------------------------------
		// redoボタン押下時の処理
		//-----------------------------------------------------------
		btnRedo.Connect("clicked", func() {
			// 選択中のページのFileStrを取得
			fs, id, err := GetFileStrFromPageNum(note)
			if err != nil {
				ShowErrorDialog(window1, err)
				return
			}
			
			// redoバッファから最新の値を取得
			value := fs.rbuf.Value.(UndoRedoStr)
			if !value.isAvailable {
				return
			}
			
			// undo/redoによる編集を開始
			fs.doingUndoRedo = true
			FileMap[id] = fs
			
			// 挿入データの場合は、挿入
			// 削除データの場合は、削除
			iterStart := fs.textBuffer.GetIterAtOffset(value.offset)
			if value.isInsert {
				fs.textBuffer.Insert(iterStart, value.text)
			} else {
				iterEnd := fs.textBuffer.GetIterAtOffset(value.offset + len(value.text))
				fs.textBuffer.Delete(iterStart, iterEnd)
			}
			
			// 次のundoバッファに保存し、enable
			fs.ubuf = fs.ubuf.Next()
			fs.ubuf.Value = value
			fs.isUndoEnable = true
			
			// 取りだし済みのデータを利用不可にし、1つ前のバッファへ
			fs.rbuf.Value = UndoRedoStr {isAvailable: false}
			fs.rbuf = fs.rbuf.Prev()
			
			// 利用不可データになった場合、disable
			if !fs.rbuf.Value.(UndoRedoStr).isAvailable {
				fs.isRedoEnable = false
			}
			
			// editCountが「0」→「1」で編集されたデータ
			fs.editCount++
			if fs.editCount == 1 {
				fs.isEdit = true
				fs.label.SetText("*" + fs.fileName)
			}
			
			// undo/redoによる編集を終了
			fs.doingUndoRedo = false
			
			btnRedo.SetSensitive(fs.isRedoEnable)
			btnUndo.SetSensitive(fs.isUndoEnable)
			FileMap[id] = fs
		})
		
		//-----------------------------------------------------------
		// wrapボタン押下時の処理
		//-----------------------------------------------------------
		btnWrap.Connect("toggled", func() {
			// ページ切替時は、フラグをoffにして終了
			if nextWrap {
				nextWrap = false
				return
			}
			
			// 選択中のページのFileStrを取得
			fs, id, err := GetFileStrFromPageNum(note)
			if err != nil {
				ShowErrorDialog(window1, err)
				return
			}
			
			fs.isWrap = !fs.isWrap
			FileMap[id] = fs
			
			if fs.isWrap {
				fs.textView.SetWrapMode(gtk.WRAP_CHAR)
			} else {
				fs.textView.SetWrapMode(gtk.WRAP_NONE)
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
			
			// 編集中のファイルがあるか確認
			isExistEditFile := false
			for _, v := range FileMap {
				if v.isEdit {
					isExistEditFile = true
					break
				}
			}
			
			// 編集中のファイルがある場合
			if isExistEditFile {
				dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "未保存のファイルがあります。\n終了を中止しますか？")
				defer dialog.Destroy()
				
				ret := dialog.Run()
				if ret != gtk.RESPONSE_NO {
					// ウィンドウクローズ処理を中断
					return true
				}
			}
			
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


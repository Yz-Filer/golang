package main

import (
	"container/ring"
	_ "embed"
	"log"
	"os"
	"strconv"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/32_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application

var FontSize = 12.0 * 1024.0

// undo/redo用構造体
type UndoRedoStr struct {
	isAvailable		bool
	isInsert		bool
	offset			int
	text			string
}

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
		
		// textBufferを取得
		textBuffer1, err := textView1.GetBuffer()
		
		// gladeからadjustmentを取得
		adjustment1, _, err := GetObjFromGlade[*gtk.Adjustment](builder, "", "ADJUSTMENT")
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
		
		// undo/redoリングバッファの初期化
		ubuf := ring.New(10)
		rbuf := ring.New(10)
		for i := 0; i < 10; i++ {
			ubuf.Value = UndoRedoStr {isAvailable: false}
			ubuf = ubuf.Move(i)
			rbuf.Value = UndoRedoStr {isAvailable: false}
			rbuf = rbuf.Move(i)
		}
		ubuf = ubuf.Move(0)
		rbuf = rbuf.Move(0)
		doingUndoRedo := false
		editCount := 0
		
		// undo/redoボタンの初期状態はdisable
		btnUndo.SetSensitive(false)
		btnRedo.SetSensitive(false)
		
		//-----------------------------------------------------------
		// テキスト挿入時の処理
		//-----------------------------------------------------------
		textBuffer1.Connect("insert_text", func(textBuffer *gtk.TextBuffer, pos *gtk.TextIter, text string) {
			// undo/redoによる編集時は何もしない
			if doingUndoRedo {
				doingUndoRedo = false
				return
			}
			
			// ユーザによる手入力時、redoバッファをクリアし、disable
			// ※undo後しかredoは実行させない
			for i := 0; i < 10; i++ {
				rbuf.Value = UndoRedoStr {isAvailable: false}
				rbuf = rbuf.Move(i)
			}
			btnRedo.SetSensitive(false)
			
			// 次のundoバッファに編集内容を保存し、enable
			ubuf = ubuf.Next()
			ubuf.Value = UndoRedoStr {
				isAvailable:	true,
				isInsert:		true,
				offset:			pos.GetOffset(),
				text:			text,
			}
			btnUndo.SetSensitive(true)
			
			// editCountが「0」→「1」で編集されたデータ
			editCount++
			if editCount == 1 {
				log.Println("変更発生")
			}
		})
		
		//-----------------------------------------------------------
		// テキスト削除時の処理
		//-----------------------------------------------------------
		textBuffer1.Connect("delete_range", func(textBuffer *gtk.TextBuffer, start *gtk.TextIter, end *gtk.TextIter) {
			// undo/redoによる編集時は何もしない
			if doingUndoRedo {
				doingUndoRedo = false
				return
			}
			
			// ユーザによる手入力時、redoバッファをクリアし、disable
			// ※undo後しかredoは実行させない
			for i := 0; i < 10; i++ {
				rbuf.Value = UndoRedoStr {isAvailable: false}
				rbuf = rbuf.Move(i)
			}
			btnRedo.SetSensitive(false)
			
			// 次のundoバッファに編集内容を保存し、enable
			ubuf = ubuf.Next()
			ubuf.Value = UndoRedoStr {
				isAvailable:	true,
				isInsert:		false,
				offset:			start.GetOffset(),
				text:			start.GetText(end),
			}
			btnUndo.SetSensitive(true)
			
			// editCountが「0」→「1」で編集されたデータ
			editCount++
			if editCount == 1 {
				log.Println("変更発生")
			}
		})
		
		//-----------------------------------------------------------
		// undoボタン押下時の処理
		//-----------------------------------------------------------
		btnUndo.Connect("clicked", func() {
			// undoバッファから最新の値を取得
			value := ubuf.Value.(UndoRedoStr)
			if !value.isAvailable {
				return
			}
			
			// undo/redoによる編集を開始
			doingUndoRedo = true
			
			// 挿入データの場合は、削除
			// 削除データの場合は、挿入
			iterStart := textBuffer1.GetIterAtOffset(value.offset)
			if value.isInsert {
				iterEnd := textBuffer1.GetIterAtOffset(value.offset + len(value.text))
				textBuffer1.Delete(iterStart, iterEnd)
			} else {
				textBuffer1.Insert(iterStart, value.text)
			}
			
			// 次のredoバッファに保存し、enable
			rbuf = rbuf.Next()
			rbuf.Value = value
			btnRedo.SetSensitive(true)
			
			// 取りだし済みのデータを利用不可にし、1つ前のバッファへ
			ubuf.Value = UndoRedoStr {isAvailable: false}
			ubuf = ubuf.Prev()
			
			// 利用不可データになった場合、disable
			if !ubuf.Value.(UndoRedoStr).isAvailable {
				btnUndo.SetSensitive(false)
			}
			
			// editCountが「1」→「0」で編集前まで戻されたデータ
			editCount--
			if editCount == 0 {
				log.Println("変更前に戻りました")
			}
		})
		
		//-----------------------------------------------------------
		// redoボタン押下時の処理
		//-----------------------------------------------------------
		btnRedo.Connect("clicked", func() {
			// redoバッファから最新の値を取得
			value := rbuf.Value.(UndoRedoStr)
			if !value.isAvailable {
				return
			}
			
			// undo/redoによる編集を開始
			doingUndoRedo = true
			
			// 挿入データの場合は、挿入
			// 削除データの場合は、削除
			iterStart := textBuffer1.GetIterAtOffset(value.offset)
			if value.isInsert {
				textBuffer1.Insert(iterStart, value.text)
			} else {
				iterEnd := textBuffer1.GetIterAtOffset(value.offset + len(value.text))
				textBuffer1.Delete(iterStart, iterEnd)
			}
			
			// 次のundoバッファに保存し、enable
			ubuf = ubuf.Next()
			ubuf.Value = value
			btnUndo.SetSensitive(true)
			
			// 取りだし済みのデータを利用不可にし、1つ前のバッファへ
			rbuf.Value = UndoRedoStr {isAvailable: false}
			rbuf = rbuf.Prev()
			
			// 利用不可データになった場合、disable
			if !rbuf.Value.(UndoRedoStr).isAvailable {
				btnRedo.SetSensitive(false)
			}
			
			// editCountが「0」→「1」で編集されたデータ
			editCount++
			if editCount == 1 {
				log.Println("変更発生")
			}
		})
		
		//-----------------------------------------------------------
		// ctrl + z, ctrl + y押下時の処理
		//-----------------------------------------------------------
		textView1.Connect("key-press-event", func(textView *gtk.TextView, event *gdk.Event) bool {
			keyEvent := gdk.EventKeyNewFromEvent(event)
			keyVal := keyEvent.KeyVal()
			keyState := gdk.ModifierType(keyEvent.State() & 0x0F)
			
			switch keyState {
				case gdk.CONTROL_MASK:	  // CTRLキー
					switch keyVal {
						case gdk.KEY_y, gdk.KEY_Y:
							_, err := btnRedo.Emit("clicked", glib.TYPE_POINTER)
							if err != nil {
								ShowErrorDialog(window1, err)
							}
						case gdk.KEY_z, gdk.KEY_Z:
							_, err := btnUndo.Emit("clicked", glib.TYPE_POINTER)
							if err != nil {
								ShowErrorDialog(window1, err)
							}
					}
			}
			
			// イベントを伝播
			return false
		})
		
		//-----------------------------------------------------------
		// マウスホイール（垂直）が回転した時、拡大/縮小
		//-----------------------------------------------------------
		textView1.Connect("scroll_event", func(self *gtk.TextView, e *gdk.Event) bool {
			event := gdk.EventScrollNewFromEvent(e)
			
			// ctrlキー + 垂直方向の場合
			if event.State() == gdk.CONTROL_MASK && event.DeltaY() != 0 {
				scale := adjustment1.GetValue() - float64(event.DeltaY()) * 10.0
				adjustment1.SetValue(scale)
				
				// 書式を設定
				err := ApplyStyle(textView1, scale)
				if err != nil {
					ShowErrorDialog(window1, err)
				}
				
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


package main

/*
#cgo pkg-config: gdk-3.0
#include <gdk/gdk.h>
#include <gdk/gdkwin32.h>
*/
import "C"

import (
	_ "embed"
	"fmt"
	"strconv"
	"unsafe"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)


//go:embed glade/07_DIALOG.glade
var dialog1Glade string

//-----------------------------------------------------------------------------
// グレードからgtkオブジェクトを取得
// [input]
//   builder	gladeファイルから構築したbuilder（初回はnilを指定）
//				同一ファイル内に含まれてるオブジェクトを取得する場合は前回取得したbuilderを指定
//   glade		gladeファイルから取得したxml文字列（embedから代入した変数を指定）
//   id			gladeでオブジェクトに設定したID
// [return]
//   オブジェクト
//   Builder
//   error
//-----------------------------------------------------------------------------
func GetObjFromGlade[T any](builder *gtk.Builder, glade string, id string) (T, *gtk.Builder, error) {
	var err error
	
	// builderがNULLの場合、gladeから読み込む
	if builder == nil {
		if len(glade) == 0 {
			return *new(T), nil, fmt.Errorf("Could not get glade string.")
		}
		builder, err = gtk.BuilderNewFromString(glade)
		if err != nil {
			return *new(T), nil, fmt.Errorf("Could not create builder: %w", err)
		}
	}
	
	// Builderの中からオブジェクトを取得
	obj, err := builder.GetObject(id)
	if err != nil {
		return *new(T), nil, fmt.Errorf("Could not get object: %w", err)
	}
	gtkObject, ok := obj.(T)
	if !ok {
		return *new(T), nil, fmt.Errorf("Could not convert object to gtk object.")
	}
	
	return gtkObject, builder, nil
}


//-----------------------------------------------------------------------------
// カスタムメッセージダイアログの作成
// [input]
//   parent			親ウィンドウ
//   title			ダイアログのタイトル
//   messageType	gtk.MessageTypeによって、「！」とか「？」とかを表示
//   buttonLabel	9種類のボタンに表示する文字列を設定（空文字列の場合、ボタン非表示）
//   label1			メッセージ1行目（文字が少し大きく、太字）
//   label2			メッセージ2行目
// [return]
//   Dialog
//   Spinner
//   error
//-----------------------------------------------------------------------------
func CustomMessageDialogNew(parent gtk.IWindow, title string, messageType gtk.MessageType, buttonLabel [9]string, label1 string, label2 string) (*gtk.Dialog, *gtk.Spinner, error) {
	// gladeからダイアログを取得
	dialog, builder2, err := GetObjFromGlade[*gtk.Dialog](nil, dialog1Glade, "DIALOG")
	if err != nil {
		return nil, nil, err
	}
	
	// ダイアログのタイトルを設定する
	dialog.SetTitle(title)
	
	// 親が指定された場合、親と紐づけ、アイコンを継承
	if parent != nil {
		dialog.SetTransientFor(parent)
		parentIcon, err := parent.(*gtk.ApplicationWindow).GetIcon()
		if err == nil {
			dialog.SetIcon(parentIcon)
		}
	}
	
	// ダイアログ内のボタンを取得
	dialogButtons := [9]*gtk.Button{}
	dialogButtonIDs := [9]string{"BUTTON_REJECT", "BUTTON_ACCEPT", "BUTTON_OK", "BUTTON_CANCEL", "BUTTON_CLOSE", "BUTTON_YES", "BUTTON_NO", "BUTTON_APPLY", "BUTTON_HELP"}
	dialogButtonResponse := [9]gtk.ResponseType{gtk.RESPONSE_REJECT, gtk.RESPONSE_ACCEPT, gtk.RESPONSE_OK, gtk.RESPONSE_CANCEL, gtk.RESPONSE_CLOSE, gtk.RESPONSE_YES, gtk.RESPONSE_NO, gtk.RESPONSE_APPLY, gtk.RESPONSE_HELP}
	for i := 0; i < 9; i++ {

		// gladeからdialog buttonを取得
		dialogButtons[i], _, err = GetObjFromGlade[*gtk.Button](builder2, "", dialogButtonIDs[i])
		if err != nil {
			return nil, nil, err
		}
		
		// buttonLabelが空文字の場合、表示しない
		// 空文字じゃない場合、指定文字列を表示し、クリック時にレスポンスタイプを返す
		if len(buttonLabel[i]) == 0 {
			dialogButtons[i].Hide()
		} else {
			id := i		// iは変動するため、ローカル変数に代入
			dialogButtons[id].SetLabel(buttonLabel[id])
			dialogButtons[id].Connect("clicked", func(button *gtk.Button) {
				dialog.Response(dialogButtonResponse[id])
			})
		}
	}
	
	// ダイアログ内のアイコンを設定する（「？」とか「！」とか）
	dialogImage, _, err := GetObjFromGlade[*gtk.Image](builder2, "", "ICON")
	if err != nil {
		return nil, nil, err
	}
	
	// messageTypeによって表示するアイコンを設定(gtk.MESSAGE_OTHERの場合アイコン非表示)
	iconType := "dialog-information"
	switch messageType {
		case gtk.MESSAGE_INFO : iconType = "dialog-information"
		case gtk.MESSAGE_WARNING : iconType = "dialog-warning"
		case gtk.MESSAGE_QUESTION : iconType = "dialog-question"
		case gtk.MESSAGE_ERROR : iconType = "dialog-error"
		default : iconType = "dialog-password"
	}
	if messageType != gtk.MESSAGE_OTHER {
		dialogImage.SetFromIconName(iconType, gtk.ICON_SIZE_DND)
	} else {
		dialogImage.Hide()
	}
	
	// ダイアログ内のスピナーを取得し、非表示にする
	dialogSpinner, _, err := GetObjFromGlade[*gtk.Spinner](builder2, "", "SPINNER")
	if err != nil {
		return nil, nil, err
	}
	dialogSpinner.Hide()
	
	// gladeからラベルを取得（1行目）
	dialogLabel1, _, err := GetObjFromGlade[*gtk.Label](builder2, "", "LABEL1")
	if err != nil {
		return nil, nil, err
	}
	
	// ラベルに文字列を設定。文字列が未指定の場合非表示にする。
	dialogLabel1.SetText(label1)
	if len(label1) == 0 {
		dialogLabel1.Hide()
	}
	
	// メッセージが長い場合、折り返す
	dialogLabel1.SetLineWrap(true)
	dialogLabel1.SetMaxWidthChars(120)
	
	// gladeからラベルを取得（2行目）
	dialogLabel2, _, err := GetObjFromGlade[*gtk.Label](builder2, "", "LABEL2")
	if err != nil {
		return nil, nil, err
	}
	
	// ラベルに文字列を設定。文字列が未指定の場合非表示にする。
	dialogLabel2.SetText(label2)
	if len(label2) == 0 {
		dialogLabel2.Hide()
	}
	
	// メッセージが長い場合、折り返す
	dialogLabel2.SetLineWrap(true)
	dialogLabel2.SetMaxWidthChars(120)
	
	// ダイアログを最前面に表示
//	dialog.ShowAll()
//	dialog.SetKeepAbove(true)
//	dialog.SetKeepAbove(false)
	
	return dialog, dialogSpinner, nil
}

//-----------------------------------------------------------------------------
// キューに溜まった処理をすべて実行させる
// フリーズすることがあるので、100回まで
//-----------------------------------------------------------------------------
func DoEvents() {
	for i := 0; i < 100; i++ {
		if !gtk.EventsPending() {
			break
		}
	    gtk.MainIteration()
	}
}

//-----------------------------------------------------------------------------
// エラーメッセージを表示する
//-----------------------------------------------------------------------------
func ShowErrorDialog(parent gtk.IWindow, err error) {
	dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "エラーが発生しました")
	dialog.FormatSecondaryText("%s", err.Error())
	dialog.SetTitle ("error")
	dialog.Run()
	dialog.Destroy()
}

//-----------------------------------------------------------------------------
// ListStoreに格納されてる値を取得
//-----------------------------------------------------------------------------
func GetListStoreValue[T any] (iModel gtk.ITreeModel, iter *gtk.TreeIter, id int) (T, error) {

	model := iModel.ToTreeModel()

	// 値を取得
	colVal, err := model.GetValue(iter, id)
	if err != nil {
		return *new(T), err
	}
	
	// 値をgolang形式に変換
	col, err := colVal.GoValue()
	if err != nil {
		return *new(T), err
	}
	
	// interfaceをT型に変換
	ret, ok := col.(T)
	if !ok {
		return *new(T), fmt.Errorf("type assertion failed")
	}
	return ret, nil
}

//-----------------------------------------------------------------------------
// keyからListStoreのIterを取得
//-----------------------------------------------------------------------------
func GetIterFromKey(listStore *gtk.ListStore, key string) (*gtk.TreeIter, error) {
	var err1 error = nil
	var iter1 *gtk.TreeIter = nil
	
	listStore.ForEach(func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
		// 値を取得
		col1, err := GetListStoreValue[string] (model, iter, 0)
		if err != nil {
			err1 = err
			
			// 検索を終了する
			return true
		}
		
		if col1 == key {
			iter1 = iter
			
			// 検索を終了する
			return true
		}
		
		// 検索を続ける
		return false
	})
	
	if iter1 == nil && err1 == nil {
		err1 = fmt.Errorf("Not found.")
	}
	
	return iter1, err1
}

//-----------------------------------------------------------------------------
// gtk.WindowからWindowsのWindowハンドルを取得する
//-----------------------------------------------------------------------------
func GetWindowHandle(window *gtk.Window) (uintptr, error) {
	gdkWin, err := window.GetWindow()
	if err != nil {
		return uintptr(0), err
	}
	return uintptr(C.gdk_win32_window_get_handle((*C.GdkWindow)(unsafe.Pointer(gdkWin.Native())))), nil
}

//-----------------------------------------------------------------------------
// タスクバーへの表示を抑止する
//-----------------------------------------------------------------------------
func SetSkipTaskbarHint(window *gtk.Window) error {
	// Windowsのウィンドウハンドルを取得
	hwnd, err := GetWindowHandle(window)
	if err != nil {
		return err
	}

	// ウィンドウスタイルの取得
	var gwl_exstyle int = GWL_EXSTYLE
	exStyle, _, err := GetWindowLongPtr.Call(uintptr(hwnd), uintptr(gwl_exstyle))
	if err.Error() != "The operation completed successfully." {
		return fmt.Errorf("Failed to retrieve window style.")
	}
	
	// ツールウィンドウ(タスクバーに表示されないウィンドウ)へウィンドウスタイルを変更する
	newExStyle := exStyle | WS_EX_TOOLWINDOW 
	_, _, err = SetWindowLongPtr.Call(uintptr(hwnd), uintptr(gwl_exstyle), uintptr(newExStyle))
	if err.Error() != "The operation completed successfully." {
		return fmt.Errorf("Failed to set window style.")
	}
	
	return nil
}

//-----------------------------------------------------------------------------
// pango.Styleを文字列に変換
//-----------------------------------------------------------------------------
func FontStyleToString(style pango.Style) string {
	switch style {
		case pango.STYLE_OBLIQUE:	return "oblique"
		case pango.STYLE_ITALIC:	return "italic"
		default:					return "normal"
	}
}

//-----------------------------------------------------------------------------
// pango.Stretchを文字列に変換
//-----------------------------------------------------------------------------
func FontStretchToString(stretch pango.Stretch) string {
	switch stretch {
		case pango.STRETCH_ULTRA_CONDENSED:			return "ultra-condensed"
		case pango.STRETCH_EXTRA_CONDENSEDStretch:	return "extra-condensed"
		case pango.STRETCH_CONDENSEDStretch:		return "condensed"
		case pango.STRETCH_SEMI_CONDENSEDStretch:	return "semi-condensed"
		case pango.STRETCH_SEMI_EXPANDEDStretch:	return "semi-expanded"
		case pango.STRETCH_EXPANDEDStretch:			return "expanded"
		case pango.STRETCH_EXTRA_EXPANDEDStretch:	return "extra-expanded"
		case pango.STRETCH_ULTRA_EXPANDEDStretch:	return "ultra-expanded"
		default:									return "normal"
	}
}

//-----------------------------------------------------------------------------
// CSS文字列を作成
//-----------------------------------------------------------------------------
func BuildCSS(st StickyStr, isTextView bool) string {
	css := "* {\n"
	
	if isTextView {
		// 選択領域のカーソルの色まで同じになるので、TextViewは範囲を限定
		css = "text, .view {\n"
	}
	
	css+= "  font-family: " + st.FontFamily + ";\n"
	css+= "  font-size: " + strconv.Itoa(st.FontSize / 1024) + "pt;\n"
	css+= "  font-style: " + FontStyleToString(st.FontStyle) + ";\n"
	css+= "  font-weight: " + strconv.Itoa(int(st.FontWeight)) + ";\n"
	css+= "  font-stretch: " + FontStretchToString(st.FontStretch) + ";\n"
	css+= "  color: rgba("
	css+= strconv.Itoa(st.FgColor[0]) + ", "
	css+= strconv.Itoa(st.FgColor[1]) + ", "
	css+= strconv.Itoa(st.FgColor[2]) + ", "
	css+= strconv.FormatFloat(float64(st.FgColor[3]) / 255.0, 'f', 2, 64) + ");\n"
	css+= "  background-color: rgba("
	css+= strconv.Itoa(st.BgColor[0]) + ", "
	css+= strconv.Itoa(st.BgColor[1]) + ", "
	css+= strconv.Itoa(st.BgColor[2]) + ", "
	css+= strconv.FormatFloat(float64(st.BgColor[3]) / 255.0, 'f', 2, 64) + ");\n"
	css+= "}"
	
	return css
}

//-----------------------------------------------------------------------------
// 書式を設定
//-----------------------------------------------------------------------------
func ApplyStyle(widget gtk.IWidget, st StickyStr) error {
	// プロバイダーを作成
	cssProvider, err := gtk.CssProviderNew()
	if err != nil {
		return err
	}
	
	cssStr := ""
	
	// コンテキストを取得し、CSS文字列を作成
	var context *gtk.StyleContext
	switch widget.(type) {
		case *gtk.TextView:
			context, err = widget.(*gtk.TextView).GetStyleContext()
			cssStr = BuildCSS(st, true)
		default:
			context, err = widget.(*gtk.Window).GetStyleContext()
			cssStr = BuildCSS(st, false)
	}
	if err != nil {
		return err
	}
	
	// CSSをロード
	err = cssProvider.LoadFromData(cssStr)
	if err != nil {
		return err
	}
	
	// 書式を反映
	context.AddProvider(cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	
	return nil
}

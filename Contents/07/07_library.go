package main

import (
	_ "embed"
	"fmt"

	"github.com/gotk3/gotk3/gtk"
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
			return *new(T), nil, fmt.Errorf("Could not get glade string")
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
		return *new(T), nil, fmt.Errorf("Could not convert object to gtk object")
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

package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)


//-----------------------------------------------------------------------------
// 第2引数で指定したページからFileStrを取得
// ※第2引数が未指定の時、現在ページからFileStrを取得
//-----------------------------------------------------------------------------
func GetFileStrFromPageNum(note *gtk.Notebook, page_nums ...int) (FileStr, string, error) {
	var page_num int
	if len(page_nums) == 0 {
		page_num = note.GetCurrentPage()
	} else {
		page_num = page_nums[0]
	}

	// 選択されてるページを取得
	scrolledWindow1, err := note.GetNthPage(page_num)
	if err != nil {
		return FileStr{}, "", err
	}
	
	// 選択ページからIDを取得
	id, err := scrolledWindow1.(*gtk.ScrolledWindow).GetName()
	if err != nil {
		return FileStr{}, "", err
	}
	
	// 選択ページからFileStrを取得
	fs, ok := FileMap[id]
	if !ok {
		return FileStr{}, "", fmt.Errorf("選択中のページの取得に失敗しました。")
	}
	
	return fs, id, nil
}

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
// エラーメッセージを表示する
//-----------------------------------------------------------------------------
func ShowErrorDialog(parent gtk.IWindow, err error) {
	dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "エラーが発生しました")
	dialog.FormatSecondaryText("%s", err.Error())
	dialog.SetTitle ("error")
	dialog.Run()
	dialog.Destroy()
}

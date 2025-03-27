package main

/*
#cgo pkg-config: gdk-3.0
#include <gdk/gdk.h>
#include <gdk/gdkwin32.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/gotk3/gotk3/gtk"
)


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

//-----------------------------------------------------------------------------
// gtk.WindowからWindowsのWindowハンドルを取得する
//-----------------------------------------------------------------------------
func GetWindowHandle(window *gtk.ApplicationWindow) (uintptr, error) {
	gdkWin, err := window.GetWindow()
	if err != nil {
		return uintptr(0), err
	}
	return uintptr(C.gdk_win32_window_get_handle((*C.GdkWindow)(unsafe.Pointer(gdkWin.Native())))), nil
}

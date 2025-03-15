// 付箋の表示
package main

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//-----------------------------------------------------------------------------
// 付箋用メニューを作成
//-----------------------------------------------------------------------------
func buildStickyMenu(grandParent *gtk.ApplicationWindow, builder *gtk.Builder, key string) (*gtk.Menu, error) {
	if grandParent == nil {
		return nil, fmt.Errorf("GrandParent window not found.")
	}

	// gladeからメニューを取得
	menu1, builder, err := GetObjFromGlade[*gtk.Menu](builder, WindowGlade, "STICKY_MENU")
	if err != nil {
		return nil, err
	}
	
	// gladeからmenuItem1を取得
	menuItem1, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_NEW_STICKY")
	if err != nil {
		return nil, err
	}
	// gladeからmenuItem2を取得
	menuItem2, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_REMOVE_STICKY")
	if err != nil {
		return nil, err
	}
	
	// menuItem1選択時に編集画面表示(新規)シグナル発行
	menuItem1.Connect("activate", func(){
		_, err := grandParent.Emit("edit_window_new", glib.TYPE_POINTER, EDIT_NEW, "")
		if err != nil {
			ShowErrorDialog(grandParent, err)
		}
	})
	
	// menuItem2選択時に付箋破棄シグナル発行
	menuItem2.Connect("activate", func(){
		_, err := grandParent.Emit("sticky_note_destroy", glib.TYPE_POINTER, key)
		if err != nil {
			ShowErrorDialog(grandParent, err)
		}
	})
	
	return menu1, nil
}

//-----------------------------------------------------------------------------
// 付箋用ウィンドウを表示
//-----------------------------------------------------------------------------
func StickyNoteWindowShow(parent *gtk.ApplicationWindow, key string) error {
	// 付箋ウィンドウを作成
	sWindow, err := StickyNoteWindowNew(parent, key)
	if err != nil {
		return err
	}
	
	// 付箋ウィンドウを表示
	sWindow.Show()

	// 付箋がタスクバーに表示されないようにする（Showの後にやる必要あり）
	err = SetSkipTaskbarHint(sWindow)
	if err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// 付箋用ウィンドウを作成
//-----------------------------------------------------------------------------
func StickyNoteWindowNew(parent *gtk.ApplicationWindow, key string) (*gtk.Window, error) {
	// 構造体から付箋情報を取得
	st, ok := StickyMap[key]
	if !ok {
		return nil, fmt.Errorf("The sticky note does not exist.")
	}

	// 付箋用ウィンドウをgladeから取得
	sWindow, builder, err := GetObjFromGlade[*gtk.Window](nil, WindowGlade, "STICKY_WINDOW")
	if err != nil {
		return nil, err
	}

	// Labelをgladeから取得
	label, _, err := GetObjFromGlade[*gtk.Label](builder, "", "STICKY_NOTE")
	if err != nil {
		return nil, err
	}
	
	// 付箋用メニューを作成
	menu, err := buildStickyMenu(parent, builder, key)
	if err != nil {
		return nil, err
	}

	sWindow.SetModal(false)
	
	// 書式を反映
	err = ApplyStyle(sWindow, st)
	if err != nil {
		return nil, err
	}
	
	// 構造体のNoteをラベルへ反映
	label.SetText(st.Note)
	
	// ウィンドウサイズを変更（幅にTextViewとずれがあるので1引いた）
	label.SetSizeRequest(st.Width - 1, st.Height)
	sWindow.Resize(st.Width + 7, st.Height + 8)
	
	// 新規ではない場合、座標を指定
	if st.X >= 0 && st.Y >= 0 {
		sWindow.Move(st.X, st.Y)
	}
	
	// Alignの設定
	label.SetYAlign(0.0)
	label.SetXAlign(st.Align)
	switch st.Align {
		case 0.5:	label.SetJustify(gtk.JUSTIFY_CENTER)
		case 1.0:	label.SetJustify(gtk.JUSTIFY_RIGHT)
		default:	label.SetJustify(gtk.JUSTIFY_LEFT)
	}
	
	// 破棄用に構造体へ付箋ウィンドウを保存
	st.stickyWindow = sWindow
	StickyMap[key] = st


	var dragging bool
	var offsetX, offsetY int
	var winX, winY int

	//-----------------------------------------------------------
	// マウスボタンを押したときのイベントハンドラ
	//-----------------------------------------------------------
	sWindow.Connect("button-press-event", func(win *gtk.Window, event *gdk.Event) {
		e := gdk.EventButtonNewFromEvent(event)
		switch e.Button() {
			case gdk.BUTTON_PRIMARY:							// 左クリック時
				if e.Type() == gdk.EVENT_DOUBLE_BUTTON_PRESS {	// ダブルクリック時
					// 現在位置を保存
					st.X, st.Y = sWindow.GetPosition()
					StickyMap[key] = st
					
					// 更新モードで編集ウィンドウを表示するシグナルを送信
					_, err := parent.Emit("edit_window_new", glib.TYPE_POINTER, EDIT_UPDATED, key)
					if err != nil {
						ShowErrorDialog(parent, err)
						return
					}
					
					// 付箋を破棄
					sWindow.Destroy()
				} else {										// シングルクリック時
					if !dragging{
						dragging = true
						winX, winY = win.GetPosition()
						x, y := e.MotionValRoot()
						offsetX = int(x)
						offsetY = int(y)
					}
				}
			case gdk.BUTTON_SECONDARY:							// 右クリック時
				// メニューを表示
				menu.PopupAtPointer(event)
		}
	})

	//-----------------------------------------------------------
	// マウスを移動したときのイベントハンドラ
	//-----------------------------------------------------------
	sWindow.Connect("motion-notify-event", func(win *gtk.Window, event *gdk.Event) {
		if dragging {
			e := gdk.EventMotionNewFromEvent(event)
			x, y := e.MotionValRoot()
			dx := int(x) - offsetX
			dy := int(y) - offsetY
			win.Move(winX + dx, winY + dy)
		}
	})

	//-----------------------------------------------------------
	// マウスボタンを離したときのイベントハンドラ
	//-----------------------------------------------------------
	sWindow.Connect("button-release-event", func(win *gtk.Window, event *gdk.Event) {
		e := gdk.EventButtonNewFromEvent(event)
		if e.Button() == gdk.BUTTON_PRIMARY {
			if dragging{
				dragging = false
				
				// 現在位置を保存
				st.X, st.Y = sWindow.GetPosition()
				StickyMap[key] = st
			}
		}
	})
	
	return sWindow, nil
}

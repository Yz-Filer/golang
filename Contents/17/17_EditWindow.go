// 編集ウィンドウの表示
package main

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/pango"
)


//-----------------------------------------------------------------------------
// RGBA配列（0-255）からgdk.RGBAを返す
//-----------------------------------------------------------------------------
func ColorAryToRGBA(c [4]int) *gdk.RGBA {
	return gdk.NewRGBA(float64(c[0]) / 255.0, float64(c[1]) / 255.0, float64(c[2]) / 255.0, float64(c[3]) / 255.0)
}

//-----------------------------------------------------------------------------
// 色の選択
//-----------------------------------------------------------------------------
func ColorChoose(parent *gtk.Window, st *StickyStr, isForeground bool) error{
	var rgba *gdk.RGBA
	
	if parent == nil {
		return fmt.Errorf("parent is null.")
	}
	
	// 色選択ダイアログを作成
	ccd, err := gtk.ColorChooserDialogNew("色を選択", parent)
	if err != nil {
		return err
	}
	defer ccd.Destroy()
	
	// 現在色の設定
	if isForeground {
		rgba = ColorAryToRGBA(st.FgColor)
	} else {
		rgba = ColorAryToRGBA(st.BgColor)
	}
	ccd.SetRGBA(rgba)
	
	// 「OK」で終わった場合は、RGBAを構造体へ設定
	if ccd.Run() == gtk.RESPONSE_OK {
		rgba = ccd.GetRGBA()
		if isForeground {
			st.FgColor[0] = int(rgba.GetRed() * 255)
			st.FgColor[1] = int(rgba.GetGreen() * 255)
			st.FgColor[2] = int(rgba.GetBlue() * 255)
			st.FgColor[3] = int(rgba.GetAlpha() * 255)
		} else {
			st.BgColor[0] = int(rgba.GetRed() * 255)
			st.BgColor[1] = int(rgba.GetGreen() * 255)
			st.BgColor[2] = int(rgba.GetBlue() * 255)
			st.BgColor[3] = int(rgba.GetAlpha() * 255)
		}
	}
	return nil
}

//-----------------------------------------------------------------------------
// フォントの選択
//-----------------------------------------------------------------------------
func FontChoose(parent *gtk.Window, st *StickyStr) error {
	if parent == nil {
		return fmt.Errorf("parent is null.")
	}
	
	// フォント選択ダイアログを作成
	fcd, err := gtk.FontChooserDialogNew("Fontの選択", parent)
	if err != nil {
		return err
	}
	defer fcd.Destroy()
	
	// 現在フォントの設定
	fd := pango.FontDescriptionNew()
	fd.SetFamily(st.FontFamily)
	fd.SetSize(st.FontSize)
	fd.SetStyle(st.FontStyle)
	fd.SetWeight(pango.Weight(st.FontWeight))
	fd.SetStretch(st.FontStretch)

	fcd.SetFont(fd.ToString())
	
	// 「OK」で終わった場合は、フォントを構造体へ設定
	if fcd.Run() == gtk.RESPONSE_OK {
		desc := pango.FontDescriptionFromString(fcd.GetFont())
		st.FontFamily = desc.GetFamily()
		st.FontSize = desc.GetSize()
		st.FontStyle = desc.GetStyle()
		st.FontWeight = int(desc.GetWeight())
		st.FontStretch = desc.GetStretch()
	}
	return nil
}

//-----------------------------------------------------------------------------
// 編集ウィンドウの作成
//-----------------------------------------------------------------------------
func EditWindowNew(parent *gtk.ApplicationWindow, key string) (*gtk.Window, error) {
	isNew := false
	
	// 新規の場合は、付箋用構造体を初期化
	st, ok := StickyMap[key]
	if !ok {
		isNew = true
		st = StickyStr {
			X: -1,
			Y: -1,
			Width: -1,
			Height: -1,
			Note: "",
			FontFamily: "Yu Gothic",
			FontSize: 10240,
			FontStyle: pango.STYLE_NORMAL,
			FontWeight: 400,
			FontStretch: pango.STRETCH_NORMALStretch,
			FgColor: [4]int{0, 0, 0, 255},
			BgColor: [4]int{255, 255, 255, 255},
			Align: 0,
		}
	}
	
	// 編集用ウィンドウをgladeから取得
	eWindow, builder, err := GetObjFromGlade[*gtk.Window](nil, WindowGlade, "EDIT_WINDOW")
	if err != nil {
		return nil, err
	}
	eWindow.SetModal(false)

	parentIcon, err := parent.GetIcon()
	if err == nil {
		eWindow.SetIcon(parentIcon)
	}
	
	// TextViewをgladeから取得
	textView, _, err := GetObjFromGlade[*gtk.TextView](builder, "", "EW_TEXTVIEW")
	if err != nil {
		return nil, err
	}
	
	// HeaderBarをgladeから取得
	headerBar, _, err := GetObjFromGlade[*gtk.HeaderBar](builder, "", "EW_HEADERBAR")
	if err != nil {
		return nil, err
	}
	
	// Toolbarをgladeから取得
	toolbar, _, err := GetObjFromGlade[*gtk.Toolbar](builder, "", "EW_TOOLBAR")
	if err != nil {
		return nil, err
	}
	
	// OKボタンをgladeから取得
	ewButtonOk, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "EW_BUTTON_OK")
	if err != nil {
		return nil, err
	}
	
	// Cancelボタンをgladeから取得
	ewButtonCancel, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "EW_BUTTON_CANCEL")
	if err != nil {
		return nil, err
	}
	
	// Colorボタンをgladeから取得
	ewButtonColor, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "BG_COLOR")
	if err != nil {
		return nil, err
	}
	
	// Fontボタンをgladeから取得
	ewButtonFont, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "FONT")
	if err != nil {
		return nil, err
	}
	
	// Font colorボタンをgladeから取得
	ewButtonFontColor, _, err := GetObjFromGlade[*gtk.ToolButton](builder, "", "FG_COLOR")
	if err != nil {
		return nil, err
	}
	
	// AlignLeftボタンをgladeから取得
	ewButtonAlignLeft, _, err := GetObjFromGlade[*gtk.ToggleToolButton](builder, "", "ALIGN_LEFT")
	if err != nil {
		return nil, err
	}
	
	// AlignCenterボタンをgladeから取得
	ewButtonAlignCenter, _, err := GetObjFromGlade[*gtk.ToggleToolButton](builder, "", "ALIGN_CENTER")
	if err != nil {
		return nil, err
	}
	
	// AlignRightボタンをgladeから取得
	ewButtonAlignRight, _, err := GetObjFromGlade[*gtk.ToggleToolButton](builder, "", "ALIGN_RIGHT")
	if err != nil {
		return nil, err
	}
	
	// 付箋新規作成ボタンをgladeから取得
	ewButtonNew, _, err := GetObjFromGlade[*gtk.Button](builder, "", "EW_NEW_BUTTON")
	if err != nil {
		return nil, err
	}
	
	// サイズ・座標を設定
	headerBarHeight, _ := headerBar.GetPreferredHeight()
	toolbarHeight, _ := toolbar.GetPreferredHeight()
	
	if st.Width >= 0 && st.Height >= 0 {
		eWindow.Resize(st.Width + 8, st.Height + toolbarHeight)
		eWindow.Move(st.X, st.Y - headerBarHeight)
	} else {
		eWindow.Resize(380, 128 + toolbarHeight)
	}
	
	// 書式を反映
	err = ApplyStyle(textView, st)
	if err != nil {
		return nil, err
	}
	
	// 文字列を反映
	err = SetTextToTextView(textView, st.Note)
	if err != nil {
		return nil, err
	}
	
	// AlignによりToggleToolButtonの状態を設定
	switch st.Align {
		case 0.5:
			ewButtonAlignLeft.SetActive(false)
			ewButtonAlignCenter.SetActive(true)
			ewButtonAlignRight.SetActive(false)
		case 1.0:
			ewButtonAlignLeft.SetActive(false)
			ewButtonAlignCenter.SetActive(false)
			ewButtonAlignRight.SetActive(true)
		default:
			ewButtonAlignLeft.SetActive(true)
			ewButtonAlignCenter.SetActive(false)
			ewButtonAlignRight.SetActive(false)
	}

	// OKボタン押下時の処理
	ewButtonOk.Connect("clicked", func() {
		// TextViewからテキストを取得
		note, err := GetTextFromTextView(textView)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		
		// 構造体へテキスト/サイズ/座標を代入
		st.Note = note
		st.Width = textView.GetAllocatedWidth()
		st.Height = textView.GetAllocatedHeight()
		st.X, st.Y = eWindow.GetPosition()
		st.Y += headerBarHeight
		
		// 構造体の更新結果をマップに格納
		StickyMap[key] = st
		
		
		// 付箋ウィンドウを表示するようシグナルを発行
		mode := STICKY_UPDATED
		if isNew {
			mode = STICKY_NEW
		}
		_, err = parent.Emit("sticky_note_show", glib.TYPE_POINTER, mode, key)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		
		// 編集ウィンドウ破棄
		eWindow.Destroy()
	})
	
	// Cancelボタン押下時の処理
	ewButtonCancel.Connect("clicked", func() {
		// 付箋が存在したら編集前の付箋を表示
		_, ok := StickyMap[key]
		if ok {
			// 付箋ウィンドウを表示するようシグナルを発行
			_, err = parent.Emit("sticky_note_show", glib.TYPE_POINTER, STICKY_NOT_UPDATED, key)
			if err != nil {
				ShowErrorDialog(parent, err)
				return
			}
		}
		
		eWindow.Destroy()
	})
	
	// クローズボタン押下時の処理
	eWindow.Connect("delete-event", func(win *gtk.Window, event *gdk.Event) bool {
		// 付箋が存在したら編集前の付箋を表示
		_, ok := StickyMap[key]
		if ok {
			// 付箋ウィンドウを表示するようシグナルを発行
			_, err = parent.Emit("sticky_note_show", glib.TYPE_POINTER, STICKY_NOT_UPDATED, key)
			if err != nil {
				ShowErrorDialog(parent, err)
				return true
			}
		}
		
		// 編集ウィンドウのクローズを継続
		return false
	})
		
	// Colorボタン押下時の処理
	ewButtonColor.Connect("clicked", func() {
		// 背景色を選択
		err := ColorChoose(eWindow, &st, false)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		
		// 書式へ反映
		err = ApplyStyle(textView, st)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	// Fontボタン押下時の処理
	ewButtonFont.Connect("clicked", func() {
		// フォントを選択
		err := FontChoose(eWindow, &st)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		
		// 書式へ反映
		err = ApplyStyle(textView, st)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	// Font colorボタン押下時の処理
	ewButtonFontColor.Connect("clicked", func() {
		// フォント色を選択
		err := ColorChoose(eWindow, &st, true)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		
		// 書式へ反映
		err = ApplyStyle(textView, st)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	// AlignLeftボタン押下時の処理
	ewButtonAlignLeft.Connect("toggled", func() {
		if ewButtonAlignLeft.GetActive() {
			// 左寄せにする
			st.Align = 0.0
			
			ewButtonAlignCenter.SetActive(false)
			ewButtonAlignRight.SetActive(false)
		}
	})
	
	// AlignCenterボタン押下時の処理
	ewButtonAlignCenter.Connect("toggled", func() {
		if ewButtonAlignCenter.GetActive() {
			// 中央寄せにする
			st.Align = 0.5
			
			ewButtonAlignLeft.SetActive(false)
			ewButtonAlignRight.SetActive(false)
		}
	})
	
	// AlignRightボタン押下時の処理
	ewButtonAlignRight.Connect("toggled", func() {
		if ewButtonAlignRight.GetActive() {
			// 右寄せにする
			st.Align = 1.0
			
			ewButtonAlignLeft.SetActive(false)
			ewButtonAlignCenter.SetActive(false)
		}
	})
	
	// 付箋新規作成ボタン押下時にシグナル発行
	ewButtonNew.Connect("clicked", func() {
		_, err := parent.Emit("edit_window_new", glib.TYPE_POINTER, EDIT_NEW, "")
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	return eWindow, nil
}

//-----------------------------------------------------------------------------
// TextViewから文字列を取得
//-----------------------------------------------------------------------------
func GetTextFromTextView(tv *gtk.TextView) (string, error) {
	buffer, err := tv.GetBuffer()
	if err != nil {
		return "", err
	}

	start, end := buffer.GetStartIter(), buffer.GetEndIter()
	text, err := buffer.GetText(start, end, false)
	if err != nil {
		return "", err
	}

	return text, nil
}

//-----------------------------------------------------------------------------
// TextViewに文字列を設定
//-----------------------------------------------------------------------------
func SetTextToTextView(tv *gtk.TextView, txt string) error {
	buffer, err := tv.GetBuffer()
	if err != nil {
		return err
	}

	buffer.SetText(txt)
	return nil
}

package main

import (
	_ "embed"
	"fmt"
	"log"
	"math"
	"os"
	"unsafe"
	
	"github.com/zzl/go-win32api/win32"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/34_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application

type SelectionStr struct {
	startX float64
	startY float64
	endX   float64
	endY   float64
	active bool
}

type WorkAreaStr struct {
	x	int32
	y	int32
	w	int32
	h	int32
}

const (
	OtherWidth = 2
	OtherHeight = 81
	HeaderHeight = 23
)

var (
	WorkArea	WorkAreaStr
	SrcPixbuf	*gdk.Pixbuf
	Width		int
	Height		int
	Scale		float64
	ScaleMode	int
)


//-----------------------------------------------------------------------------
// 渡した座標から、画像の範囲内の座標を取得
//-----------------------------------------------------------------------------
func GetPointWithinImageArea(imgX, imgY, x, y float64) (float64, float64) {
	x = math.Max(x, imgX)
	x = math.Min(x, imgX + float64(Width) * Scale)
	y = math.Max(y, imgY)
	y = math.Min(y, imgY + float64(Height) * Scale)

	return x, y
}

//-----------------------------------------------------------------------------
// 画像ファイルのオープン
//-----------------------------------------------------------------------------
func menuOpen(parent *gtk.ApplicationWindow) ([]byte, error) {
	if parent == nil {
		return nil, fmt.Errorf("parent is null")
	}

	// OSのファイル選択ダイアログを作成
	fcd, err := gtk.FileChooserNativeDialogNew("開く", parent, gtk.FILE_CHOOSER_ACTION_OPEN, "開く(_O)", "キャンセル(_C)")
	if err != nil {
		return nil, err
	}
	defer fcd.Destroy()

	// 画像拡張子のフィルタを追加
	fileFilterTxt, err := gtk.FileFilterNew()
	if err != nil {
		return nil, err
	}
	fileFilterTxt.AddPattern("*.bmp;*.gif;*.icns;*.ico;*.cur;*.jpeg;*.jpe;*.jpg;*.png;*.pnm;*.pbm;*.pgm;*.ppm;*.qtif;*.qif;*.svg;*.svgz;*.svg.gz;*.tga;*.targa;*.tiff;*.tif;*.xbm;*.xpm;*.webp;")
	fileFilterTxt.SetName("画像ファイル")
	fcd.AddFilter(fileFilterTxt)

	// すべての拡張子のフィルタを追加
	fileFilterPng, err := gtk.FileFilterNew()
	if err != nil {
		return nil, err
	}
	fileFilterPng.AddPattern("*.*")
	fileFilterPng.SetName("すべてのファイル")
	fcd.AddFilter(fileFilterPng)

	// ダイアログを起動
	if fcd.Run() == int(gtk.RESPONSE_ACCEPT) {
		// 選択したファイルを読み込む
		data, err := os.ReadFile(fcd.GetFilename())
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, fmt.Errorf("canceled")
}

//-----------------------------------------------------------------------------
// 画像を拡大/縮小して表示
//-----------------------------------------------------------------------------
func ShowScaledImage(parent *gtk.ApplicationWindow, drawingArea *gtk.DrawingArea) error {
	if SrcPixbuf == nil {
		return nil
	}
	
	switch ScaleMode {
		case 0:
			Scale = 1.0
		case 50:
			Scale = 0.5
		case 100:
			Scale = 1.0
		case 200:
			Scale = 2.0
	}
	
	// 画面より大きい場合、画面内に収まるサイズを設定
	dw := math.Min(float64(Width) * Scale + 2, float64(WorkArea.w - OtherWidth))
	dh := math.Min(float64(Height) * Scale + 2, float64(WorkArea.h - OtherHeight))
	
	// 拡大率を設定し、幅と高さを拡大率から再計算
	if ScaleMode == 0 {
		wScale := float64(dw - 2) / float64(Width)
		hScale := float64(dh - 2) / float64(Height)
		Scale = math.Min(wScale, hScale)
		dw = math.Min(dw, float64(Width) * Scale + 2)
		dh = math.Min(dh, float64(Height) * Scale + 2)
	}
	
	// 画面からはみ出る場合、画面内に収まる座標を設定
	winX, winY := parent.GetPosition()
	dx := math.Min(float64(winX), float64(WorkArea.w - OtherWidth) - dw)
	dx = math.Max(dx, float64(WorkArea.x))
	dy := math.Min(float64(winY), float64(WorkArea.h - OtherHeight) - dh)
	dy = math.Max(dy, float64(WorkArea.y))
	
	// drawingAreaのサイズを変更
	drawingArea.SetSizeRequest(int(float64(Width) * Scale), int(float64(Height) * Scale))
	
	// ウィンドウを移動
	parent.Move(int(dx), int(dy))
	
	// ウィンドウのサイズを変更
	parent.Resize(int(dw) + OtherWidth, int(dh) + OtherHeight - HeaderHeight)
	DoEvents()
	
	drawingArea.QueueDraw()
	
	return nil
}

//-----------------------------------------------------------------------------
// メニューアイテムのシグナル処理を設定
//-----------------------------------------------------------------------------
func buildMenuItem(parent *gtk.ApplicationWindow, builder *gtk.Builder, drawingArea *gtk.DrawingArea) error {
	// gladeからmenuItemOpenを取得
	menuItemOpen, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_OPEN")
	if err != nil {
		return err
	}
	// gladeからmenuItemCloseを取得
	menuItemClose, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_CLOSE")
	if err != nil {
		return err
	}
	// gladeからmenuItemScale50を取得
	menuItemScale50, _, err := GetObjFromGlade[*gtk.RadioMenuItem](builder, "", "MENUITEM_SCALE_50")
	if err != nil {
		return err
	}
	// gladeからmenuItemScale100を取得
	menuItemScale100, _, err := GetObjFromGlade[*gtk.RadioMenuItem](builder, "", "MENUITEM_SCALE_100")
	if err != nil {
		return err
	}
	// gladeからmenuItemScale200を取得
	menuItemScale200, _, err := GetObjFromGlade[*gtk.RadioMenuItem](builder, "", "MENUITEM_SCALE_200")
	if err != nil {
		return err
	}
	// gladeからmenuItemScaleAutoを取得
	menuItemScaleAuto, _, err := GetObjFromGlade[*gtk.RadioMenuItem](builder, "", "MENUITEM_SCALE_AUTO")
	if err != nil {
		return err
	}
	menuItemScaleAuto.SetActive(true)
	
	// menuItemOpen選択時
	menuItemOpen.Connect("activate", func(){
		data, err := menuOpen(parent)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
		if data != nil {
			// 画像を初期化
			SrcPixbuf = nil
			parent.Resize(1, 1)
			DoEvents()
			
			// データからpixbufを取得
			SrcPixbuf, err = gdk.PixbufNewFromDataOnly(data)
			if err != nil {
				ShowErrorDialog(parent, err)
				return
			}
			
			// 画像サイズを取得
			Width = SrcPixbuf.GetWidth()
			Height = SrcPixbuf.GetHeight()
			
			// 画像を拡大/縮小して表示
			err = ShowScaledImage(parent, drawingArea)
			if err != nil {
				ShowErrorDialog(parent, err)
				return
			}
		}
	})
	
	menuItemScale50.Connect("toggled", func(){
		if !menuItemScale50.GetActive() {
			return
		}
		
		ScaleMode = 50
		
		// 画像を拡大/縮小して表示
		err = ShowScaledImage(parent, drawingArea)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	menuItemScale100.Connect("toggled", func(){
		if !menuItemScale100.GetActive() {
			return
		}
		
		ScaleMode = 100
		
		// 画像を拡大/縮小して表示
		err = ShowScaledImage(parent, drawingArea)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	menuItemScale200.Connect("toggled", func(){
		if !menuItemScale200.GetActive() {
			return
		}
		
		ScaleMode = 200
		
		// 画像を拡大/縮小して表示
		err = ShowScaledImage(parent, drawingArea)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	menuItemScaleAuto.Connect("toggled", func(){
		if !menuItemScaleAuto.GetActive() {
			return
		}
		
		ScaleMode = 0
		
		// 画像を拡大/縮小して表示
		err = ShowScaledImage(parent, drawingArea)
		if err != nil {
			ShowErrorDialog(parent, err)
			return
		}
	})
	
	// menuItemClose選択時にアプリを終了
	menuItemClose.Connect("activate", func(){
		application.Quit()
	})
	
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
	
	var clipboard *gtk.Clipboard
	var scrRect win32.RECT
	
	// タスクバーを除いた作業領域を取得
	_, w32err := win32.SystemParametersInfo(win32.SPI_GETWORKAREA , 0, unsafe.Pointer(&scrRect), 0)
	if w32err != win32.NO_ERROR {
		log.Fatalf("SystemParametersInfo failed: %v", win32.GetLastError())
	}
	
	WorkArea = WorkAreaStr {
		x:	scrRect.Left,
		y:	scrRect.Top,
		w:	scrRect.Right - scrRect.Left - 9,
		h:	scrRect.Bottom - scrRect.Top - 9,
	}
	
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
		
		// gladeからScrolledWindowを取得
		scrolledWindow, _, err := GetObjFromGlade[*gtk.ScrolledWindow](builder, "", "SCROLLEDWINDOW")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからViewPortを取得
		viewport, _, err := GetObjFromGlade[*gtk.Viewport](builder, "", "VIEWPORT")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからDrawingAreaを取得
		drawingArea, _, err := GetObjFromGlade[*gtk.DrawingArea](builder, "", "DRAWING_AREA")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからステータスバーのLabelを取得
		stLabel, _, err := GetObjFromGlade[*gtk.Label](builder, "", "STATUSBAR_LABEL")
		if err != nil {
			log.Fatal(err)
		}
		
		// メニューアイテムのシグナル処理を設定
		err = buildMenuItem(window1, builder, drawingArea)
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

		// scrolledWindowのプロパティを設定
		color := gdk.NewRGBA(0.5, 0.5, 0.5, 1.0)
		viewport.OverrideBackgroundColor(gtk.STATE_FLAG_NORMAL, color)
		
		
		// 選択範囲の状態を管理する構造体
		selection := SelectionStr {
			startX: 0,
			startY: 0,
			endX:	0,
			endY:	0,
			active:	false,
		}
		
		// 拡大率
		Scale = 1.0
		ScaleMode = 0
		
		// 画像の座標
		imgX, imgY := 0.0, 0.0
		
		SrcPixbuf = nil

		// イベントマスクを設定して、マウスの動きとボタンの状態を監視できるようにする
		drawingArea.SetEvents(int(gdk.POINTER_MOTION_MASK | gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK))

		//-----------------------------------------------------------
		// DrawingAreaの描画処理
		//-----------------------------------------------------------
		drawingArea.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
			if SrcPixbuf == nil {
				return
			}
			
			// 画像を拡大/縮小
			scaledW := int(float64(Width) * Scale)
			scaledH := int(float64(Height) * Scale)
			scaledPixbuf, err := SrcPixbuf.ScaleSimple(scaledW, scaledH, gdk.INTERP_BILINEAR)
			if err != nil {
				ShowErrorDialog(window1, err)
				return
			}
			
			// Pixbuf を描画
			surface, err := gdk.CairoSurfaceCreateFromPixbuf(scaledPixbuf, 1, nil)
			if err != nil {
				ShowErrorDialog(window1, err)
				return
			}
			defer surface.Close()
			
			// Cairo Surface を描画ソースとして設定
			daW := scrolledWindow.GetAllocatedWidth()
			daH := scrolledWindow.GetAllocatedHeight()
			imgX = math.Max(0.0, float64(daW - scaledW) / 2.0)
			imgY = math.Max(0.0, float64(daH - scaledH) / 2.0)
			cr.SetSourceSurface(surface, imgX, imgY)
			cr.Paint()
			
			// 選択範囲を描画
			cr.SetSourceRGBA(0, 0, 1, 0.5) // 青色の半透明
			cr.Rectangle(selection.startX, selection.startY, selection.endX - selection.startX, selection.endY - selection.startY)
			cr.Fill()
			
			text := fmt.Sprintf("スケール：%d%%", ScaleMode)
			if ScaleMode == 0 {
				text = fmt.Sprintf("スケール：自動(%d%%)", int(100.0 * Scale))
			}
			stLabel.SetText(text)
			return
		})

		//-----------------------------------------------------------
		// マウスボタンが押された時の処理
		//-----------------------------------------------------------
		drawingArea.Connect("button-press-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
			if SrcPixbuf == nil {
				return false		// イベントを伝播
			}
			
			mouseEvent := gdk.EventButtonNewFromEvent(ev)
			
			// マウスカーソルの座標を取得
			selection.startX, selection.startY = mouseEvent.MotionVal()
			
			// 画像の範囲内を開始/終了位置に設定
			selection.startX, selection.startY = GetPointWithinImageArea(imgX, imgY, selection.startX, selection.startY)
			selection.endX, selection.endY = selection.startX, selection.startY
			
			selection.active = true
			drawingArea.QueueDraw() // 再描画を要求
			
			return true
		})

		//-----------------------------------------------------------
		// マウスが移動した時の処理（ボタンが押されている間）
		//-----------------------------------------------------------
		drawingArea.Connect("motion-notify-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
			if SrcPixbuf == nil {
				return false		// イベントを伝播
			}
			
			if selection.active {
				mouseEvent := gdk.EventMotionNewFromEvent(ev)
				
				// マウスカーソルの座標を取得
				selection.endX, selection.endY = mouseEvent.MotionVal()
				
				// 画像の範囲内を終了位置に設定
				selection.endX, selection.endY = GetPointWithinImageArea(imgX, imgY, selection.endX, selection.endY)
				
				drawingArea.QueueDraw() // 再描画を要求
			}
			
			return true
		})

		//-----------------------------------------------------------
		// マウスボタンが離された時の処理
		//-----------------------------------------------------------
		drawingArea.Connect("button-release-event", func(da *gtk.DrawingArea, ev *gdk.Event) bool {
			if SrcPixbuf == nil {
				return false		// イベントを伝播
			}
			
			if selection.active {
				mouseEvent := gdk.EventButtonNewFromEvent(ev)
				
				// マウスカーソルの座標を取得
				selection.endX, selection.endY = mouseEvent.MotionVal()
				
				// 画像の範囲内を終了位置に設定
				selection.endX, selection.endY = GetPointWithinImageArea(imgX, imgY, selection.endX, selection.endY)
				
				selection.active = false
				drawingArea.QueueDraw() // 最終的な選択範囲を描画
			}
			
			return true
		})
		
		//-----------------------------------------------------------
		// サイズが変更された時の処理
		//-----------------------------------------------------------
		oldW, oldH := 0, 0
		window1.Connect("configure-event", func(win *gtk.ApplicationWindow, event *gdk.Event) {
			if SrcPixbuf == nil {
				return
			}
			
			// サイズを取得
			e := gdk.EventConfigureNewFromEvent(event)
			w, h := e.Width(), e.Height()
			
			// サイズが変わってない場合、何もしない
			if w == oldW && h == oldH {
				return
			}
			oldW, oldH = w, h
			
			// 選択範囲をクリア
			selection.startX, selection.startY = 0, 0
			selection.endX, selection.endY = 0, 0
			selection.active = false
			
			if ScaleMode != 0 {
				return
			}
			
			// 拡大率を設定
			wScale := float64(w - OtherWidth) / float64(Width)
			hScale := float64(h + HeaderHeight - OtherHeight) / float64(Height)
			Scale = math.Min(wScale, hScale)
			
			// drawingAreaのサイズを変更
			drawingArea.SetSizeRequest(int(float64(Width) * Scale), int(float64(Height) * Scale))
			
			drawingArea.QueueDraw()
		})
		
		//-----------------------------------------------------------
		// キーが押された時
		//-----------------------------------------------------------
		window1.Connect("key-press-event", func(win *gtk.ApplicationWindow, event *gdk.Event) bool {
			if SrcPixbuf == nil {
				return false		// イベントを伝播
			}
			
			keyEvent := gdk.EventKeyNewFromEvent(event)
			keyVal := keyEvent.KeyVal()
			keyState := gdk.ModifierType(keyEvent.State() & 0x0F)
			
			switch keyState {
				case gdk.CONTROL_MASK:	// CTRLキー
					switch keyVal {
						case gdk.KEY_a, gdk.KEY_A:		// 全て選択
							selection.startX = imgX
							selection.startY = imgY
							selection.endX = imgX + float64(Width) * Scale
							selection.endY = imgY + float64(Height) * Scale
							selection.active = false
							drawingArea.QueueDraw()
							
						case gdk.KEY_c, gdk.KEY_C:		// 選択範囲をクリップボードへコピー
							// 選択領域を元画像の選択領域へ変換
							stX := (math.Min(selection.startX, selection.endX) - imgX) / Scale
							stY := (math.Min(selection.startY, selection.endY) - imgY) / Scale
							endX := (math.Max(selection.startX, selection.endX) - imgX) / Scale
							endY := (math.Max(selection.startY, selection.endY) - imgY) / Scale
							destWidth := int(endX - stX)
							destHeight := int(endY - stY)
							
							// 選択領域を切り出したPixbufを取得
							// ※cropが見当たらないから、Scaleで切り出し
							newPixbuf, err := gdk.PixbufNew(SrcPixbuf.GetColorspace(), SrcPixbuf.GetHasAlpha(), SrcPixbuf.GetBitsPerSample(), destWidth, destHeight)
							if err != nil {
								ShowErrorDialog(window1, err)
								return false
							}
							SrcPixbuf.Scale(newPixbuf, 0, 0, destWidth, destHeight, -stX, -stY, 1.0, 1.0, gdk.INTERP_BILINEAR)
							clipboard.SetImage(newPixbuf)
							
						case gdk.KEY_v, gdk.KEY_V:		// クリップボードからペースト
							if clipboard.WaitIsImageAvailable() {
								// 画像を初期化
								SrcPixbuf = nil
								window1.Resize(1, 1)
								DoEvents()
								
								// クリップボードから画像を取得
								SrcPixbuf, err = clipboard.WaitForImage()
								if err != nil {
									ShowErrorDialog(window1, err)
									return false
								}
								
								// 画像サイズを取得
								Width = SrcPixbuf.GetWidth()
								Height = SrcPixbuf.GetHeight()
								
								// 画像を拡大/縮小して表示
								err = ShowScaledImage(window1, drawingArea)
								if err != nil {
									ShowErrorDialog(window1, err)
									return false
								}
							}
					}
			}
			
			// イベントを伝播
			return false
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

		// クリップボードを取得
		clipboard, err = gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
		if err != nil {
			log.Fatal(err)
		}
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


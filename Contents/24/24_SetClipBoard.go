package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"unsafe"
	
	"golang.org/x/sys/windows"
	"github.com/zzl/go-win32api/win32"
)

func main() {
	var ret win32.BOOL
	var w32err win32.WIN32_ERROR

	// データ格納用にクリップボードを開く
	ret, w32err = win32.OpenClipboard(0)
	if ret == win32.FALSE || w32err != win32.NO_ERROR {
		log.Fatal("OpenClipboardの失敗")
	}

	// クリップボードをクリア
	ret2, w32err := win32.EmptyClipboard()
	if ret2 == win32.FALSE || w32err != win32.NO_ERROR {
		log.Fatal("EmptyClipboardの失敗")
	}


	// テキストデータを格納
	fmt.Println(SetData(win32.CF_UNICODETEXT, "aaabbbあいいう"))
	
	// HDROP用データを格納(move)
	files := []string{
		"D:\\test\\日本語ファイル.txt",
	}
	fmt.Println(SetData(win32.CF_HDROP, files, win32.DROPEFFECT_MOVE))

	// byte配列を格納(「abcd」をWaveAudioとして格納)
	bytes := []byte{0x61, 0x62, 0x63, 0x64}
	fmt.Println(SetData(win32.CF_WAVE, bytes))
	
	// カスタムフォーマット名を指定してbyte配列を格納
	fmt.Println(SetData("Custom Format", bytes))

	// 新しいRGBA画像(緑色の正方形)を生成
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	green := color.RGBA{0, 255, 0, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{green}, image.Point{}, draw.Src)
	
	// 画像をCF_BITMAP形式で格納
	fmt.Println(SetData(win32.CF_BITMAP, img))


	// データ格納用クリップボードを閉じる
	win32.CloseClipboard()
}

//------------------------------------------------------------------------------------------------
// クリップボードにデータを格納する
//------------------------------------------------------------------------------------------------
func SetData(dataFormat, data any, optionalArgs ...uint32) error {
	var err error
	stdCode := map[string]win32.CLIPBOARD_FORMATS {
		"Text":win32.CF_TEXT,
		"Bitmap":win32.CF_BITMAP,
		"MetafilePict":win32.CF_METAFILEPICT,
		"SymbolicLink":win32.CF_SYLK,
		"Dif":win32.CF_DIF,
		"Tiff":win32.CF_TIFF,
		"OemText":win32.CF_OEMTEXT,
		"Dib":win32.CF_DIB,
		"Palette":win32.CF_PALETTE,
		"PenData":win32.CF_PENDATA,
		"Riff":win32.CF_RIFF,
		"WaveAudio":win32.CF_WAVE,
		"UnicodeText":win32.CF_UNICODETEXT,
		"EnhancedMetafile":win32.CF_ENHMETAFILE,
		"FileDrop":win32.CF_HDROP,
		"Locale":win32.CF_LOCALE,
		"DibV5":win32.CF_DIBV5,
		"Max":win32.CF_MAX,
		"OwnerDisplay":win32.CF_OWNERDISPLAY,
		"DspText":win32.CF_DSPTEXT,
		"DspBitmap":win32.CF_DSPBITMAP,
		"DspMetaFilePict":win32.CF_DSPMETAFILEPICT,
		"DspEnhancedMetafile":win32.CF_DSPENHMETAFILE,
		"PrivateFirst":win32.CF_PRIVATEFIRST,
		"PrivateLast":win32.CF_PRIVATELAST,
		"GdiObjFirst":win32.CF_GDIOBJFIRST,
		"GdiObjLast":win32.CF_GDIOBJLAST,
	}
	
	// フォーマットコードを取得
	var format win32.CLIPBOARD_FORMATS
	switch f := dataFormat.(type) {
		case win32.CLIPBOARD_FORMATS:
			format = f
		case string:
			sCode, ok := stdCode[f]
			if ok {
				// 標準フォーマットの場合、コードを設定
				format = sCode
			} else {
				// 標準フォーマット以外の場合、登録してみてコードを取得
				formatNameUTF16, err := windows.UTF16PtrFromString(f)
				if err != nil {
					return fmt.Errorf("UTF16PtrFromStringの失敗: %w", err)
				}
				ff, ret := win32.RegisterClipboardFormat(formatNameUTF16)
				if ret != win32.NO_ERROR {
					return fmt.Errorf("RegisterClipboardFormatの失敗")
				}
				format = win32.CLIPBOARD_FORMATS(ff)
			}
		default:
			return fmt.Errorf("指定したフォーマットが不正です")
	}

	// データを格納する
	switch format {
		case win32.CF_BITMAP:
			err = SetBitmap(data.(*image.RGBA))
		case win32.CF_TEXT, win32.CF_UNICODETEXT:
			err = SetText(format, data.(string))
		case win32.CF_HDROP:
			err = SetFileDropList(data.([]string), optionalArgs[0])
		default:
			v, ok := data.([]byte)
			if ok {
				err = SetDataFromByteArray(format, v)
			} else {
				err = fmt.Errorf("指定したフォーマットへはbyte配列のみ対応してます")
			}
	}
	if err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// クリップボードへbitmapを格納
//-----------------------------------------------------------------------------
func SetBitmap(img *image.RGBA) error {
	// HDC を取得
	hdcScreen := win32.GetDC(0)
	defer win32.ReleaseDC(0, hdcScreen)

	// メモリ DC を作成
	hdcMem := win32.CreateCompatibleDC(hdcScreen)
	defer win32.DeleteDC(hdcMem)

	// 画像情報を取得
	bounds := img.Bounds()

	// Bitmapを作成
	hBitmap := win32.CreateCompatibleBitmap(hdcScreen, int32(bounds.Dx()), int32(bounds.Dy()))
	if hBitmap == 0 {
		return fmt.Errorf("CreateCompatibleBitmapに失敗")
	}

	// Bitmapをメモリデバイスコンテキストに選択
	hOld := win32.SelectObject(hdcMem, win32.HGDIOBJ(hBitmap))
	defer win32.SelectObject(hdcMem, hOld)

	// Goのimage.RGBAのデータをメモリデバイスコンテキストに描画
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// 0～65535 → 0～255に変換
			colorRef := win32.RGB(uint8(r >> 8), uint8(g >> 8), uint8(b >> 8))
			win32.SetPixel(hdcMem, int32(x), int32(y), colorRef)
		}
	}

	// クリップボードへ格納
	h, w32err := win32.SetClipboardData(uint32(win32.CF_BITMAP), win32.HGLOBAL(hBitmap))
	if h == 0 || w32err != win32.NO_ERROR {
		return fmt.Errorf("クリップボードにデータを設定できませんでした")
	}

	return nil
}

//-----------------------------------------------------------------------------
// クリップボードへTextを格納
//-----------------------------------------------------------------------------
func SetText(format win32.CLIPBOARD_FORMATS, text string) error {
	if format != win32.CF_TEXT && format != win32.CF_UNICODETEXT {
		return fmt.Errorf("CF_TEXTかCF_UNICODETEXTを指定して下さい")
	}
	
	var hMem uintptr
	var ret win32.WIN32_ERROR
	switch format {
		case win32.CF_UNICODETEXT:
			bytes := windows.StringToUTF16(text)
			size := (len(bytes) + 1) * 2

			// メモリの割り当て
			hMem, ret = win32.GlobalAlloc(win32.GMEM_MOVEABLE, uintptr(size))
			if ret != win32.NO_ERROR {
				return fmt.Errorf("メモリ確保の失敗")
			}

			// メモリのロック
			dataPtr, ret := win32.GlobalLock(hMem)
			if ret != win32.NO_ERROR {
				win32.GlobalFree(hMem)
				return fmt.Errorf("メモリロックの失敗")
			}
			defer win32.GlobalUnlock(hMem)

			// データをメモリにコピー
			copy(unsafe.Slice((*uint16)(dataPtr), size), bytes)

		default:
			bytes := []byte(text)
			size := len(bytes) + 1

			// メモリの割り当て
			hMem, ret = win32.GlobalAlloc(win32.GMEM_MOVEABLE, uintptr(size))
			if ret != win32.NO_ERROR {
				return fmt.Errorf("メモリ確保の失敗")
			}

			// メモリのロック
			dataPtr, ret := win32.GlobalLock(hMem)
			if ret != win32.NO_ERROR {
				win32.GlobalFree(hMem)
				return fmt.Errorf("メモリロックの失敗")
			}
			defer win32.GlobalUnlock(hMem)

			// データをメモリにコピー
			copy(unsafe.Slice((*byte)(dataPtr), size), bytes)
	}
	
	// クリップボードへ格納
	h, w32err := win32.SetClipboardData(uint32(format), hMem)
	if h == 0 || w32err != win32.NO_ERROR {
		win32.GlobalFree(hMem)
		return fmt.Errorf("クリップボードにデータを設定できませんでした")
	}

	return nil
}

//-----------------------------------------------------------------------------
// クリップボードへファイルリストを格納
//-----------------------------------------------------------------------------
func SetFileDropList(files []string, preferredDropEffect uint32) error {

	// UTF-16 文字列のスライスを作成
	u16str := []uint16{}
	for _, file := range files {
		u16, err := windows.UTF16FromString(file)
		if err != nil {
			return fmt.Errorf("UTF16FromString error: %w", err)
		}
		u16str = append(u16str, u16...)
	}
	u16str = append(u16str, 0) // 終端のヌル文字を追加

	// メモリの割り当て
	size := unsafe.Sizeof(win32.DROPFILES{}) + uintptr(len(u16str) * 2)
	hMem, ret := win32.GlobalAlloc(win32.GMEM_MOVEABLE, uintptr(size))
	if ret != win32.NO_ERROR {
		return fmt.Errorf("GlobalAlloc error")
	}

	// メモリのロック
	dataPtr, ret := win32.GlobalLock(hMem)
	if ret != win32.NO_ERROR {
		return fmt.Errorf("GlobalLock error")
	}
	defer win32.GlobalUnlock(hMem)

	// DROPFILES 構造体の作成とメモリへのコピー
	dropFiles := (*win32.DROPFILES)(dataPtr)
	dropFiles.PFiles = win32.DWORD(unsafe.Sizeof(win32.DROPFILES{})) // オフセットを設定
	dropFiles.Pt = win32.POINT{X: 0, Y: 0}
	dropFiles.FNC = win32.FALSE
	dropFiles.FWide = win32.TRUE

	// データをメモリにコピー
	utf16Ptr := (*uint16)(unsafe.Pointer(uintptr(dataPtr) + unsafe.Sizeof(win32.DROPFILES{})))
	copy(unsafe.Slice(utf16Ptr, len(u16str) * 2), u16str)

	// クリップボードへ格納
	h, w32err := win32.SetClipboardData(uint32(win32.CF_HDROP), hMem)
	if h == 0 || w32err != win32.NO_ERROR {
		win32.GlobalFree(hMem)
		return fmt.Errorf("クリップボードにデータを設定できませんでした")
	}

	// フォーマットコードを取得
	formatNameUTF16, err := windows.UTF16PtrFromString("Preferred DropEffect")
	if err != nil {
		return fmt.Errorf("UTF16PtrFromStringの失敗: %w", err)
	}
	efFormat, ret := win32.RegisterClipboardFormat(formatNameUTF16)
	if ret != win32.NO_ERROR {
		return fmt.Errorf("RegisterClipboardFormatの失敗")
	}

	// メモリの割り当て
	hDropEffect, ret := win32.GlobalAlloc(win32.GMEM_MOVEABLE, unsafe.Sizeof(preferredDropEffect))
	if ret != win32.NO_ERROR {
		return fmt.Errorf("GlobalAlloc error")
	}

	// メモリのロック
	pDropEffect, ret := win32.GlobalLock(hDropEffect)
	if ret != win32.NO_ERROR {
		return fmt.Errorf("GlobalLock error")
	}
	defer win32.GlobalUnlock(hDropEffect)

	*(*win32.DWORD)(unsafe.Pointer(pDropEffect)) = preferredDropEffect

	h, w32err = win32.SetClipboardData(uint32(efFormat), hDropEffect)
	if h == 0 || w32err != win32.NO_ERROR {
		win32.GlobalFree(hMem)
		return fmt.Errorf("クリップボードにDropEffectを設定できませんでした")
	}

	return nil
}

//-----------------------------------------------------------------------------
// クリップボード(HGlobal)へ[]byteを格納
//-----------------------------------------------------------------------------
func SetDataFromByteArray(format win32.CLIPBOARD_FORMATS, bytes []byte) error {
	// メモリの割り当て
	hMem, ret := win32.GlobalAlloc(win32.GMEM_MOVEABLE, uintptr(len(bytes)))
	if ret != win32.NO_ERROR {
		return fmt.Errorf("メモリ確保の失敗")
	}

	// メモリのロック
	dataPtr, ret := win32.GlobalLock(hMem)
	if ret != win32.NO_ERROR {
		win32.GlobalFree(hMem)
		return fmt.Errorf("メモリロックの失敗")
	}
	defer win32.GlobalUnlock(hMem)

	// データをメモリにコピー
	copy(unsafe.Slice((*byte)(dataPtr), len(bytes)), bytes)
	
	// クリップボードへ格納
	h, w32err := win32.SetClipboardData(uint32(format), hMem)
	if h == 0 || w32err != win32.NO_ERROR {
		win32.GlobalFree(hMem)
		return fmt.Errorf("クリップボードにデータを設定できませんでした")
	}

	return nil
}

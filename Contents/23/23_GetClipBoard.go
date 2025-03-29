package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
	
	"golang.org/x/sys/windows"
	"github.com/zzl/go-win32api/win32"
)

func main() {
	var ret win32.BOOL
	var w32err win32.WIN32_ERROR

	// 実行ファイルのフルパス取得
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("実行ファイルのパス取得に失敗: ", err)
	}

	// データ抽出用にクリップボードを開く
	ret, w32err = win32.OpenClipboard(0)
	if ret == win32.FALSE || w32err != win32.NO_ERROR {
		log.Fatal("OpenClipboardの失敗")
	}
	defer win32.CloseClipboard()

	// 指定したフォーマットコードのデータがあるか確認
	ok, err := GetDataPresent(win32.CF_TEXT)
	if err != nil {
		log.Fatal("GetDataPresentの失敗")
	}
	if ok {
		fmt.Println("Text形式は格納されてます")
	} else {
		fmt.Println("Text形式は格納されてません")
	}

	// 指定したフォーマット名のデータがあるか確認
	ok, err = ContainsData("Dib")
	if err != nil {
		log.Fatal("ContainsDataの失敗")
	}
	if ok {
		fmt.Println("Dib形式は格納されてます")
	} else {
		fmt.Println("Dib形式は格納されてません")
	}

	// クリップボードに登録されている形式を列挙する
	formats := GetFormats()
	fmt.Println("格納されてるフォーマット名一覧")
	for _, format := range formats {
		formatName := GetFormatName(format)
		fmt.Printf("  %s(%#x)\n", formatName, format)
	}

	// クリップボードからデータを取得
	fmt.Println("データ取得")
	for _, format := range formats {
		formatName := GetFormatName(format)
		fmt.Printf("  %s(%#x):", formatName, format)
		
		// クリップボードからデータを取得
		data := GetData(format)
		switch v := data.(type) {
			case error:
				fmt.Println("---- クリップボードからのデータ取得に失敗:", v, "----")
			case string:
				fmt.Println("    "+ v)
			case []string:
				fmt.Printf("ファイル数: %d\n", len(v))
				for _, file := range v {
					fmt.Println("    " + file)
				}
			case []byte:
				// 「Preferred DropEffect」の場合、FileDropのエフェクトを表示
				if formatName == "Preferred DropEffect" {
					fmt.Println("")
					if win32.DROPEFFECT_COPY == uint32(v[0]) & win32.DROPEFFECT_COPY {
						fmt.Println("    FileDrop：copy")
					}
					if win32.DROPEFFECT_MOVE == uint32(v[0]) & win32.DROPEFFECT_MOVE {
						fmt.Println("    FileDrop：move")
					}
					if win32.DROPEFFECT_LINK == uint32(v[0]) & win32.DROPEFFECT_LINK {
						fmt.Println("    FileDrop：link")
					}
					break
				}
				
				// 保存ファイル名の作成
				fileName := SanitizeFilenameStrict("0-" + formatName, "_")
				filePath := filepath.Join(filepath.Dir(exePath), fileName)
				
				// ファイルへ保存
				err = SaveByteArrayData(filePath, v)
				if err != nil {
					fmt.Println("ファイルへの保存に失敗: ", err)
				}
				fmt.Println("  ファイル「"+ fileName + "」へ保存しました")
		}
	}
}

//------------------------------------------------------------------------------------------------
// クリップボードからデータを取得しanyで返す
//------------------------------------------------------------------------------------------------
func GetData(format win32.CLIPBOARD_FORMATS) any {
	switch format {
		case win32.CF_BITMAP:
			byteArray, err := GetBitmap(format)
			if err != nil {
				return err
			}
			return byteArray
		case win32.CF_TEXT, win32.CF_UNICODETEXT:
			text, err := GetText(format)
			if err != nil {
				return err
			}
			return text
		case win32.CF_HDROP:
			files, err := GetFileDropList(format)
			if err != nil {
				return err
			}
			return files
		default:
			byteArray, err := GetDataToByteArray(format)
			if err != nil {
				return err
			}
			return byteArray
	}
	
	return nil
}

//-----------------------------------------------------------------------------
// クリップボードからbitmapを[]byteで取得
//-----------------------------------------------------------------------------
func GetBitmap(format win32.CLIPBOARD_FORMATS) ([]byte, error) {
	if format != win32.CF_BITMAP {
		return nil, fmt.Errorf("CF_BITMAPを指定して下さい")
	}
	
	// クリップボードからデータを取得
	hBitmap, ret := win32.GetClipboardData(uint32(format))
	if ret != win32.NO_ERROR {
		return nil, fmt.Errorf("クリップボードのデータを取得できませんでした")
	}
	
	// ビットマップ情報を取得
	var bitmap win32.BITMAP
	if win32.GetObject(win32.HGDIOBJ(hBitmap), int32(unsafe.Sizeof(bitmap)), unsafe.Pointer(&bitmap)) == 0 {
		return nil, fmt.Errorf("GetObject Bitmapに失敗")
	}

	// HDC を取得
	hdcScreen := win32.GetDC(0)
	defer win32.ReleaseDC(0, hdcScreen)

	// コピー元メモリ DC を作成
	srcHdc := win32.CreateCompatibleDC(hdcScreen)
	defer win32.DeleteDC(srcHdc)
	
	// コピー先メモリ DC を作成
	dstHdc := win32.CreateCompatibleDC(hdcScreen)
	defer win32.DeleteDC(dstHdc)

	// ビットマップをコピー元メモリ DCに選択
	hBitmapOld := win32.SelectObject(srcHdc, win32.HGDIOBJ(hBitmap))
	if hBitmapOld == 0 {
		return nil, fmt.Errorf("SelectObjectの失敗")
	}
	defer win32.SelectObject(srcHdc, hBitmapOld)

	// BITMAPINFOHEADER を作成
	bmi := win32.BITMAPINFO{}
	bmi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bmi.BmiHeader))
	bmi.BmiHeader.BiWidth = bitmap.BmWidth
	bmi.BmiHeader.BiHeight = -bitmap.BmHeight // 正の値はボトムアップ、負の値はトップダウン
	bmi.BmiHeader.BiPlanes = 1
	bmi.BmiHeader.BiBitCount = bitmap.BmBitsPixel
	bmi.BmiHeader.BiCompression = uint32(win32.BI_RGB)
	bmi.BmiHeader.BiSizeImage = uint32(bitmap.BmWidth * bitmap.BmHeight * int32(bitmap.BmBitsPixel / 8))

	// CreateDIBSection を使用して空のビットマップを作成
	var bits unsafe.Pointer
	hBitmapDIB, ret := win32.CreateDIBSection(srcHdc, &bmi, win32.DIB_RGB_COLORS, unsafe.Pointer(&bits), 0, 0)
	if hBitmapDIB == 0 || ret != win32.NO_ERROR {
		return nil, fmt.Errorf("CreateDIBSectionに失敗")
	}
	defer win32.DeleteObject(win32.HGDIOBJ(hBitmapDIB))
	
	// ビットマップが作成されてなかったらエラー
	if bits == nil {
		return nil, fmt.Errorf("CreateDIBSectionに失敗")
	}

	// コピー先メモリ DC に作成した空のビットマップを選択
	hBitmapDIBOld := win32.SelectObject(dstHdc, win32.HGDIOBJ(hBitmapDIB))
	if hBitmapDIBOld == 0 {
		return nil, fmt.Errorf("SelectObjectに失敗")
	}
	defer win32.SelectObject(dstHdc, hBitmapDIBOld)

	// BitBlt でビットマップをコピー
	ret2, w32err := win32.BitBlt(dstHdc, 0, 0, bitmap.BmWidth, bitmap.BmHeight, srcHdc, 0, 0, win32.SRCCOPY )
	if ret2 == win32.FALSE || w32err != win32.NO_ERROR {
		return nil, fmt.Errorf("BitBltに失敗")
	}

	// BITMAPFILEHEADER を作成
	bfh := win32.BITMAPFILEHEADER{}
	bfh.BfType = 0x4D42					// "BM"
	bfh.BfSize = 14 + bmi.BmiHeader.BiSize + bmi.BmiHeader.BiSizeImage		// 14 = uint32(unsafe.Sizeof(bfh))のパディングなしサイズ
	bfh.BfOffBits = 14 + bmi.BmiHeader.BiSize

	// データを一つの []byte 配列にまとめる
	var buffer bytes.Buffer

	// BITMAPFILEHEADER を書き込み
	if err := binary.Write(&buffer, binary.LittleEndian, bfh); err != nil {
		return nil, fmt.Errorf("BITMAPFILEHEADER の書き込みに失敗しました: %w", err)
	}

	// BITMAPINFOHEADER を書き込み
	if err := binary.Write(&buffer, binary.LittleEndian, bmi.BmiHeader); err != nil {
		return nil, fmt.Errorf("BITMAPINFOHEADER の書き込みに失敗しました: %w", err)
	}

	// ビットマップデータを書き込み
	bitsSlice := unsafe.Slice((*byte)(bits), bitmap.BmWidth * bitmap.BmHeight * int32(bitmap.BmBitsPixel / 8))
	buffer.Write(bitsSlice)

	return buffer.Bytes(), nil
}

//-----------------------------------------------------------------------------
// クリップボードからファイルリストを取得
//-----------------------------------------------------------------------------
func GetFileDropList(format win32.CLIPBOARD_FORMATS) ([]string, error) {
	if format != win32.CF_HDROP {
		return nil, fmt.Errorf("CF_HDROPを指定して下さい")
	}
	
	hDrop, ret := win32.GetClipboardData(uint32(format))
	if ret != win32.NO_ERROR {
		return nil, fmt.Errorf("クリップボードのデータを取得できませんでした")
	}
	defer win32.GlobalUnlock(hDrop)
	
	// ファイル数を取得する
	fileCount := win32.DragQueryFile(hDrop, 0xFFFFFFFF, nil, 0)
	
	// ファイルリストを取得する
	files := make([]string, fileCount)
	for i := uint32(0); i < fileCount; i++ {
		// バッファサイズを取得
		bufferSize := win32.DragQueryFile(hDrop, i, nil, 0) + 1
		
		// ファイル名を取得
		buffer := make([]uint16, bufferSize)
		win32.DragQueryFile(hDrop, i, &buffer[0], uint32(bufferSize))
		fileName := windows.UTF16ToString(buffer)
		
		files[i] = fileName
	}
	
	return files, nil
}

//-----------------------------------------------------------------------------
// クリップボードからTextを取得
//-----------------------------------------------------------------------------
func GetText(format win32.CLIPBOARD_FORMATS) (string, error) {
	if format != win32.CF_TEXT && format != win32.CF_UNICODETEXT {
		return "", fmt.Errorf("CF_TEXTかCF_UNICODETEXTを指定して下さい")
	}

	data, err := GetDataToByteArray(format)
	if err != nil {
		return "", err
	}
	
	// ユニコードテキストの場合
	if format == win32.CF_UNICODETEXT {
		return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(&data[0]))), nil
	}

	// ANSIテキストの場合
	ansiBytes := unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), len(data))
	text := windows.BytePtrToString(&ansiBytes[0])

	return text, nil
}

//-----------------------------------------------------------------------------
// クリップボード(HGlobal)から[]byteを取得
//-----------------------------------------------------------------------------
func GetDataToByteArray(format win32.CLIPBOARD_FORMATS) ([]byte, error) {
	hGlobal, ret := win32.GetClipboardData(uint32(format))
	if ret != win32.NO_ERROR {
		return nil, fmt.Errorf("クリップボードのデータを取得できませんでした")
	}
	
	p, ret := win32.GlobalLock(hGlobal)
	if ret != win32.NO_ERROR {
		return nil, fmt.Errorf("GlobalLock error")
	}
	defer win32.GlobalUnlock(hGlobal)
	
	size, ret := win32.GlobalSize(hGlobal)
	if ret != win32.NO_ERROR {
		return nil, fmt.Errorf("GlobalSize error")
	}
	
	data := make([]byte, size)
	copy(data, unsafe.Slice((*byte)(unsafe.Pointer(p)), size))

	return data, nil
}

//------------------------------------------------------------------------------------------------
// クリップボードに指定したフォーマットコードのデータが含まれてるかを返す
//------------------------------------------------------------------------------------------------
func GetDataPresent(format win32.CLIPBOARD_FORMATS) (bool, error) {
	ret, w32err := win32.IsClipboardFormatAvailable(uint32(format))
	if w32err != win32.NO_ERROR {
		return false, fmt.Errorf("IsClipboardFormatAvailable error")
	}
	if ret == win32.TRUE {
		return true, nil
	} else {
		return false, nil
	}
}

//------------------------------------------------------------------------------------------------
// クリップボードに指定したフォーマット名のデータが含まれてるかを返す
//------------------------------------------------------------------------------------------------
func ContainsData(formatName string) (bool, error) {
	// フォーマット名からフォーマットコードに変換
	format := GetFormatCode(formatName)
	
	// コードが取得できなかった場合
	if format == 0 {
		return false, nil
	}
	
	return GetDataPresent(format)
}

//-----------------------------------------------------------------------------
// クリップボードに登録されている形式を列挙する
//-----------------------------------------------------------------------------
func GetFormats() []win32.CLIPBOARD_FORMATS {
	var w32err win32.WIN32_ERROR
	var formats []win32.CLIPBOARD_FORMATS
	var format uint32 = 0
	
	for {
		format, w32err = win32.EnumClipboardFormats(format)
		if format == 0 || w32err != win32.NO_ERROR {
			break // 列挙終了
		}
		formats = append(formats, win32.CLIPBOARD_FORMATS(format))
	}
	
	return formats
}

//-----------------------------------------------------------------------------
// フォーマット名からフォーマットコードを取得
//-----------------------------------------------------------------------------
func GetFormatCode(name string) win32.CLIPBOARD_FORMATS {
	// クリップボードのフォーマット一覧を取得
	formats := GetFormats()
	
	// フォーマット一覧内のフォーマット名と比較
	for _, format := range formats {
		formatName := GetFormatName(format)
		if formatName == name {
			return format
		}
	}
	return win32.CLIPBOARD_FORMATS(0)
}

//-----------------------------------------------------------------------------
// フォーマットコードからフォーマット名を取得
//-----------------------------------------------------------------------------
func GetFormatName(format win32.CLIPBOARD_FORMATS) string {
	stdName := map[win32.CLIPBOARD_FORMATS]string {
		win32.CF_TEXT:"Text",
		win32.CF_BITMAP:"Bitmap",
		win32.CF_METAFILEPICT:"MetafilePict",
		win32.CF_SYLK:"SymbolicLink",
		win32.CF_DIF:"Dif",
		win32.CF_TIFF:"Tiff",
		win32.CF_OEMTEXT:"OemText",
		win32.CF_DIB:"Dib",
		win32.CF_PALETTE:"Palette",
		win32.CF_PENDATA:"PenData",
		win32.CF_RIFF:"Riff",
		win32.CF_WAVE:"WaveAudio",
		win32.CF_UNICODETEXT:"UnicodeText",
		win32.CF_ENHMETAFILE:"EnhancedMetafile",
		win32.CF_HDROP:"FileDrop",
		win32.CF_LOCALE:"Locale",
		win32.CF_DIBV5:"DibV5",
		win32.CF_MAX:"Max",
		win32.CF_OWNERDISPLAY:"OwnerDisplay",
		win32.CF_DSPTEXT:"DspText",
		win32.CF_DSPBITMAP:"DspBitmap",
		win32.CF_DSPMETAFILEPICT:"DspMetaFilePict",
		win32.CF_DSPENHMETAFILE:"DspEnhancedMetafile",
		win32.CF_PRIVATEFIRST:"PrivateFirst",
		win32.CF_PRIVATELAST:"PrivateLast",
		win32.CF_GDIOBJFIRST:"GdiObjFirst",
		win32.CF_GDIOBJLAST:"GdiObjLast",
	}
	
	// 標準フォーマットの場合、フォーマット名を返却
	sName, ok := stdName[format]
	if ok {
		return sName
	}

	// 標準フォーマットではない場合、名前を取得
	name := make([]uint16, 256)
	len, ret := win32.GetClipboardFormatName(uint32(format), &name[0], int32(len(name)))
	
	if ret == win32.NO_ERROR && len > 0 {
		// 取得に成功したらフォーマット名を返却
		return windows.UTF16ToString(name)
	} else {
		// 取得に失敗したら判別不能
		return "Unknown format"
	}
}

//-----------------------------------------------------------------------------
// []byteをファイルに保存
//-----------------------------------------------------------------------------
func SaveByteArrayData(filePath string, data []byte) error {
	// ファイルを作成（上書き）
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// バイト配列をファイルに書き込み
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	
	return nil
}

//-----------------------------------------------------------------------------
// ファイル名に使用できない文字を置き換え
//-----------------------------------------------------------------------------
func SanitizeFilenameStrict(filename string, replacement string) string {
	var invalidChars = []string{
		"<", ">", ":", "\"", "/", "\\", "|", "?", "*", 
		".",                                     // 拡張子をつけたくないため追加
	}
	var result strings.Builder
	for _, r := range filename {
		char := string(r)
		isInvalid := false
		for _, invalid := range invalidChars {
			if strings.Contains(invalid, char) {
				isInvalid = true
				break
			}
		}
		if isInvalid {
			result.WriteString(replacement)
		} else {
			result.WriteString(char)
		}
	}
	return result.String()
}

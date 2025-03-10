[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 16. タスクバーにアイコンを表示させない方法  

付箋ウィンドウはウィンドウなので、タスクバーにアイコンが表示されます。メインウィンドウの子ウィンドウに設定すると、メインウィンドウ以外はタスクバーに表示されなくなるのですが、メインウィンドウをタスクトレイに格納してしまうと、付箋ウィンドウが全て非表示になってしまいます。  
そのため、付箋ウィンドウをメインウィンドウの子ウィンドウに設定せずにタスクバーに表示させないようにする必要があります。  

## 16.1 `SetSkipTaskbarHint()`を使った方法

どうやってやるのかをGeminiに聞いてみました。

> Windows環境のgotk3でウィンドウのアイコンをタスクバーに表示されないように設定する方法は、SetSkipTaskbarHint()を使用することで実現できます。  
> ```go
> // タスクバーに表示しない設定
> win.SetSkipTaskbarHint(true)
> ```
> SetSkipTaskbarHint()は、ウィンドウマネージャーに対するヒントであり、常に完全に機能するとは限りません。一部の環境では、設定が無視される場合があります。  

試してみた所、Windows環境では効果がありませんでした。

## 16.2 `user32.dll`を使った方法  

`SetSkipTaskbarHint()`以外の方法をGeminiに聞いてみました。  

> Windows APIのuser32.dllにあるSetWindowLong関数とGetWindowLong関数を使用することで、ウィンドウのスタイルを変更し、タスクバーにアイコンが表示されないようにできます。  
> ```go
> const (
>         GWL_EXSTYLE = -20
>         WS_EX_TOOLWINDOW = 0x00000080
> )
> 
> var (
>         user32 = windows.NewLazySystemDLL("user32.dll")
>         setWindowLongPtr = user32.NewProc("SetWindowLongW")
>         getWindowLongPtr = user32.NewProc("GetWindowLongW")
> )
> 
> func setWindowLong(hwnd uintptr, index int, newStyle uintptr) uintptr {
>         ret, _, _ := setWindowLongPtr.Call(hwnd, uintptr(index), newStyle)
>         return uintptr(ret)
> }
> 
> func getWindowLong(hwnd uintptr, index int) uintptr {
>         ret, _, _ := getWindowLongPtr.Call(hwnd, uintptr(index))
>         return uintptr(ret)
> }
>   ：
>   ：
> // ウィンドウのGDKハンドルを取得
> gdkWin := win.GetWindow()
> hwnd := gdkWin.GetHWND()
> 
> // 拡張ウィンドウスタイルを取得してツールウィンドウフラグを設定
> style := getWindowLong(hwnd, GWL_EXSTYLE)
> style |= WS_EX_TOOLWINDOW
> 
> // 新しいスタイルを設定
> setWindowLong(hwnd, GWL_EXSTYLE, style)
> ```

試してみた所、色々エラーが出たので修正していった結果、以下のようになりました。  

```go
/*
#cgo pkg-config: gdk-3.0
#include <gdk/gdk.h>
#include <gdk/gdkwin32.h>
*/
import "C"

const (
	WS_EX_TOOLWINDOW = 0x00000080
	GWL_EXSTYLE      = -20
)

var (
	User32				= syscall.NewLazyDLL("user32.dll")
	GetWindowLongPtr	= User32.NewProc("GetWindowLongPtrW")
	SetWindowLongPtr	= User32.NewProc("SetWindowLongPtrW")
)

// gtk.WindowからWindowsのWindowハンドルを取得する
func GetWindowHandle(window *gtk.Window) (uintptr, error) {
	gdkWin, err := window.GetWindow()
	if err != nil {
		return uintptr(0), err
	}
	return uintptr(C.gdk_win32_window_get_handle((*C.GdkWindow)(unsafe.Pointer(gdkWin.Native())))), nil
}

// タスクバーへの表示を抑止する
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
```

ウィンドウスタイルを取得して、`WS_EX_TOOLWINDOW`を設定する流れは変わってません。  
Windowsのウィンドウハンドルを取得するところが、Geminiに何度聞いても上手くいかなかったので、「WEB検索」して「Geminiに聞いて」を繰り返すことでなんとか動くようになりました。
結局CGOを使う方法になり、これが最適なのかどうかはよく分からなかったのですが動くようになったので良しとします。  


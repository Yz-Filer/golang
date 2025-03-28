[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 21. キー入力の検知、IMEのON/OFF制御をしたい  

キー入力の検知はgotk3のシグナルハンドラで実施できるのですが、「shiftキー + テンキー0」と「insertキー」が判別出来ないため、win32 apiを使った方法も紹介します。  

また、gotk3ではWindows環境でIMEを制御する関数は提供されてなさそうなので、win32 apiを使った方法を紹介します。  

## 21.1 gotk3のキー入力処理  

キーが押された時、キーが離された時のシグナルハンドラは以下のようになります。  

```go
//-----------------------------------------------------------
// キーが押された時
//-----------------------------------------------------------
window1.Connect("key-press-event", func(win *gtk.ApplicationWindow, event *gdk.Event) bool {
    keyEvent := &gdk.EventKey{event}
    keyVal := keyEvent.KeyVal()
    keyState := gdk.ModifierType(keyEvent.State() & 0x0F)
    
    switch keyState {
        case gdk.SHIFT_MASK:    // SHIFTキー
            switch keyVal {
                case gdk.KEY_a, gdk.KEY_A:                 fmt.Println(" [shift + a] が押されました")
                case gdk.KEY_Shift_L, gdk.KEY_Shift_R:     fmt.Println(" [shift] が押されました")
                default:
                    // 押されたキーを表示
                    keyName := gdk.KeyValName(keyVal)
                    fmt.Printf(" [shift + %s] が押されました\n", keyName)
            }
        case gdk.CONTROL_MASK:    // CTRLキー
            switch keyVal {
                case gdk.KEY_a, gdk.KEY_A:                 fmt.Println(" [ctrl + a] が押されました")
                case gdk.KEY_Control_L, gdk.KEY_Control_R: fmt.Println(" [ctrl] が押されました")
                default:
                    // 押されたキーを表示
                    keyName := gdk.KeyValName(keyVal)
                    fmt.Printf(" [ctrl + %s] が押されました\n", keyName)
            }
        case gdk.MOD1_MASK:        // ALTキー
            switch keyVal {
                case gdk.KEY_a, gdk.KEY_A:                 fmt.Println(" [alt + a] が押されました")
                default:
                    // 押されたキーを表示
                    keyName := gdk.KeyValName(keyVal)
                    fmt.Printf(" [alt + %s] が押されました\n", keyName)
            }
        default:
            switch keyVal {
                case gdk.KEY_a, gdk.KEY_A:                 fmt.Println(" [a] が押されました")
                case gdk.KEY_Alt_L, gdk.KEY_Alt_R:         fmt.Println(" [alt] が押されました")
                default:
                    // 押されたキーを表示
                    keyName := gdk.KeyValName(keyVal)
                    fmt.Printf(" [%s] が押されました\n", keyName)
            }
    }
    
    // イベントを伝播
    return false
})

//-----------------------------------------------------------
// キーが離された時
//-----------------------------------------------------------
window1.Connect("key-release-event", func(win *gtk.ApplicationWindow, event *gdk.Event) bool {
    fmt.Println("キーが離されました")
    
    // イベントを伝播
    return false
})
```

引数で渡されてくる`event`から、キーコードとModifierKey（shiftキー, ctrlキー, altキー）を取得しています。  
各ModifierKeyと同時押しがされてるかどうかにより`switch`で分岐させています。  
（「a」キー以外は端折ってます）  

キーが離された時も`event`から、キーコードとModifierKey（shiftキー, ctrlキー, altキー）を取得して`switch`で分岐処理するところは同じなので、端折ってます。  

`case gdk.KEY_Shift_L, gdk.KEY_Shift_R`と`case gdk.KEY_Control_L, gdk.KEY_Control_R`と`case gdk.KEY_Alt_L, gdk.KEY_Alt_R`はshiftキー, ctrlキー, altキーが単体で押された時の処理となります。
`case gdk.KEY_Alt_L, gdk.KEY_Alt_R`だけ、shiftキーやctrlキーと違う場所で判定していますので注意して下さい。  

他のシグナルハンドラでModifierKeyが押されてるかどうかを検知したい場合があります。例えば、「ctrlキーを押しながらボタンをクリックした時だけ特別な処理をしたい」というような場合、上記のshiftキー, ctrlキー, altキーが単体で押された時の処理の中で、グローバル変数などに値を設定することでModifierKeyの状態を知ることが出来ます。  

## 21.2 win32 apiでModifierKeyの状態を検知する方法  

「21.1」で記載したように、グローバル変数などでModifierKeyの状態を保持しておくのが、おそらくgtkの想定している使い方だと思いますが、win32 apiを使って検知する方法も紹介しておきます。  
`github.com/zzl/go-win32api/win32`を使ってますのでimportが必要です。  

```go
//-----------------------------------------------------------
// マウスを移動したとき
//-----------------------------------------------------------
window1.Connect("motion-notify-event", func(win *gtk.ApplicationWindow, event *gdk.Event) bool {
	// SHIFTキーの状態を検知
	if (uint16(win32.GetKeyState(int32(win32.VK_SHIFT))) & 0x8000) != 0 {
		fmt.Println("SHIFTキーが押されている")
	}
	// CONTROLキーの状態を検知
	if (uint16(win32.GetKeyState(int32(win32.VK_CONTROL))) & 0x8000) != 0 {
		fmt.Println("CONTROLキーが押されている")
	}
	// ALTキーの状態を検知
	if (uint16(win32.GetKeyState(int32(win32.VK_MENU))) & 0x8000) != 0 {
		fmt.Println("ALTキーが押されている")
	}
	
	// イベントを伝播
	return false
})
```

ウィンドウ（ヘッダ以外の部分）をクリックした状態でマウスを動かした時に発生するシグナルハンドラ内に処理を記載しました。  
`GetKeyState()`関数を使用して関数実行時のキーの押下状態を取得しています。  

## 21.3 「shiftキー + テンキー0」と「insertキー」を判別したい  

1. 「WH_KEYBOARD_LL」をフック  
   
   WEBで検索すると、「WH_KEYBOARD_LL」をフックするというHPが見つかるのですが、これは実行中のアプリがactiveかどうかに関係なくメッセージを受信してしまうので使い難いと思いました。参考までにコードは以下のようになります。  

   ```go
   var HookHandle win32.HHOOK
   
   //-----------------------------------------------------------------------------
   // WH_KEYBOARD_LL用のコールバック関数
   //-----------------------------------------------------------------------------
   func hookProc(nCode int, wParam, lParam uintptr) uintptr {
   	if nCode >= 0 {
   		kbdll := (*win32.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
   
   		switch uint32(wParam) {
   			case win32.WM_KEYDOWN:
   				if (kbdll.Flags & 0x01) == 0x01 {
   					fmt.Print("拡張キー：")
   				} else {
   					fmt.Print("キー：")
   				}
   				// 例としてInsertキーのみ判別。他はキーコード表示。
   				if kbdll.VkCode == uint32(win32.VK_INSERT) {
   					fmt.Println("Insertキーが押されました")
   				} else {
   					fmt.Println(kbdll.VkCode, "が押されました")
   				}
   			case win32.WM_KEYUP:
   		}
   	}
   	return uintptr(win32.CallNextHookEx(HookHandle, int32(nCode), wParam, lParam))
   }
   　：
   　：
   メイン関数の中
   	//-----------------------------------------------------------
   	// Windowsメッセージのhook
   	//-----------------------------------------------------------
   	HookHandle, w32err = win32.SetWindowsHookEx(win32.WH_KEYBOARD_LL, uintptr(syscall.NewCallback(hookProc)), 0, 0)
   	if HookHandle == 0 || w32err != win32.NO_ERROR {
   		log.Fatalf("SetWindowsHookEx failed: %v", win32.GetLastError())
   	}
   　：
   　：
   終了時
   	// WindowsメッセージのUnhook
   	win32.UnhookWindowsHookEx(HookHandle)
   ```

   メッセージフックの説明は「[18.4 メッセージの受信（メッセージフック）](../18#184-%E3%83%A1%E3%83%83%E3%82%BB%E3%83%BC%E3%82%B8%E3%81%AE%E5%8F%97%E4%BF%A1%E3%83%A1%E3%83%83%E3%82%BB%E3%83%BC%E3%82%B8%E3%83%95%E3%83%83%E3%82%AF)」で説明しているので端折ります。  

   `fmt.Print("拡張キー：")`の前行で拡張キーかどうかを判定してます。「拡張キー」とは「Insertキー」や「Homeキー」などです。この拡張キー判定により「shiftキー + テンキー0」か「insertキー」かが判別出来ます。  

   上記コードはどのキーが押されたかまで判定してますが、拡張キー判定の結果のみグローバル変数で保持して、あとの判定はgotk3のシグナルハンドラで行うような使い方が良いのではないかと思っています。  

> [!CAUTION]  
> WH_GETMESSAGEを同時にフックすると機能しなくなるようです。`CallNextHookEx`以外にもメッセージをディスパッチするコードが必要なのかと思ったのですがGeminiに聞いても解決しませんでした。  

2. 「WH_GETMESSAGE」をフック  
   
   「WH_GETMESSAGE」をフックして、「WM_KEYDOWN」か「WM_INPUT」メッセージを処理する方法です。「WM_KEYDOWN」メッセージの方が簡単ですが、altキーが拾えないので、「shiftキー + テンキー」の判定に限定して使うなど、目的を絞った使い方になります。  

     以下に「WM_KEYDOWN」と「WM_INPUT」のコードを表示しますが、「WM_INPUT」の方が複雑なので、「WM_INPUT」をメインに作成してます。「WM_KEYDOWN」はコードを参考にすれば対処出来ると思います。  
     （「WM_KEYUP」は端折ってます）

   ```go
   // ウィンドウハンドル
   var hwnd uintptr
   
   // メッセージフック用ハンドル
   var HookHandleM win32.HHOOK
   
   
   //-----------------------------------------------------------------------------
   // WH_GETMESSAGE用のコールバック関数
   //-----------------------------------------------------------------------------
   func hookProcM(nCode int, wParam, lParam uintptr) uintptr {
       if nCode >= 0 {
           msg := (*win32.MSG)(unsafe.Pointer(lParam))
           // 自ウィンドウの時だけ処理
           if hwnd == msg.Hwnd {
               switch (msg.Message) {
   
                   case win32.WM_KEYDOWN:
                       // 以下で拡張キー判別はできるが、altキーは拾えない
                       fmt.Println("--", msg.WParam, uint32(msg.LParam) & 0x1000000)
   
                   case win32.WM_INPUT:
                       // RAWINPUTのサイズを取得
                       var size uint32
                       win32.GetRawInputData(msg.LParam, win32.RID_INPUT, nil, &size, uint32(unsafe.Sizeof(win32.RAWINPUTHEADER{})))
   
                       if size > 0 {
                           // RawInputデータを取得
                           buf := make([]byte, size)
                           if win32.GetRawInputData(msg.LParam, win32.RID_INPUT, unsafe.Pointer(&buf[0]), &size, uint32(unsafe.Sizeof(win32.RAWINPUTHEADER{}))) == size {
                               raw := (*win32.RAWINPUT)(unsafe.Pointer(&buf[0]))
                               
                               // キーボード入力を処理
                               if raw.Header.DwType == uint32(win32.RIM_TYPEKEYBOARD) {
                                   keyboard := raw.Data.KeyboardVal()
                                   switch (uint32(keyboard.Flags) & 0x01) {
                                       case win32.RI_KEY_MAKE:                // キーを押した時
                                           // 2つあるキーはE0プレフィックスでどちらのキーか判定
                                           if (uint32(keyboard.Flags) & 0x02) == win32.RI_KEY_E0 {
                                               fmt.Print("1:右の")
                                           } else {
                                               fmt.Print("1:左の")
                                           }
                                           
                                           // 例としてInsertキーのみ判別。他はキーコード表示。
                                           if keyboard.VKey == uint16(win32.VK_INSERT) {
                                               fmt.Println("Insertキーが押されました")
                                           } else {
                                               fmt.Println(keyboard.VKey, "が押されました")
                                           }
                                       case win32.RI_KEY_BREAK:            // キーを離した時
                                   }
                               }
                           } else {
                               log.Println("GetRawInputData (data) failed")
                           }
                       } else {
                           log.Println("GetRawInputData (size) failed")
                       }
               }
           }
       }
       return uintptr(win32.CallNextHookEx(HookHandleM, int32(nCode), wParam, lParam))
   }
   　：
   　：
   メイン関数の中
       //-----------------------------------------------------------
       // Windowsメッセージのhook
       //-----------------------------------------------------------
   
       // 物理キーの判別をするためのフック
       HookHandleM, w32err = win32.SetWindowsHookEx(win32.WH_GETMESSAGE, uintptr(syscall.NewCallback(hookProcM)), 0, win32.GetCurrentThreadId())
       if HookHandleM == 0 || w32err != win32.NO_ERROR {
           log.Fatalf("SetWindowsHookEx failed: %v", win32.GetLastError())
       }
   
       devices := []win32.RAWINPUTDEVICE{
           {
               UsUsagePage: 0x01, // Generic Desktop
               UsUsage:     0x06, // Keyboard
               DwFlags:     0,
               HwndTarget:  hwnd,
           },
       }
   
       pRawInputDevices := (*win32.RAWINPUTDEVICE)(unsafe.Pointer(&devices[0]))
       uiNumDevices := uint32(len(devices))
       cbSize := uint32(unsafe.Sizeof(devices[0]))
   
       ret, w32err := win32.RegisterRawInputDevices(pRawInputDevices, uiNumDevices, cbSize)
       if ret == win32.FALSE || w32err != win32.NO_ERROR {
           log.Fatal("RegisterRawInputDevicesの失敗")
       }
   　：
   　：
   終了時
       // WindowsメッセージのUnhook
       win32.UnhookWindowsHookEx(HookHandleM)
   ```

   メイン関数の中の処理で`devices := []win32.RAWINPUTDEVICE{`の行以降は、「WM_INPUT」用なので、「WM_KEYDOWN」しか使わない場合は不要です。  
   「WM_INPUT」の場合、拡張キー判定ではなく、2つあるキー（「shiftキー + テンキー0」と「Insertキー」もそうですが、「左shiftキー」と「右shiftキー」なども該当）のどちらかを判定します。（「右の」「左の」と表示してるのは便宜上なので、実際の位置と異なることがあります）

   上記コードはどのキーが押されたかまで判定してますが、左右キー判定の結果のみグローバル変数で保持して、あとの判定はgotk3のシグナルハンドラで行うような使い方が良いのではないかと思っています。  

> [!CAUTION]  
> 「WH_GETMESSAGE用のコールバック関数」内で自ウィンドウハンドルに絞って処理してますが、ダイアログが複数ある場合、複数のウィンドウハンドルで絞るのか？、それとも受信したメッセージを全て処理するのか？などアプリ作成時には考慮しなければならないことがあると思います。  

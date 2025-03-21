[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 18. クリップボード監視/USBドライブ抜き差し監視/USBドライブイジェクトがしたい  

今回からは、gotk3に特化しない機能も説明していきます。  
- クリップボードの監視  
  クリップボードにデータが書き込まれたかどうかを検知する  

- USBドライブ抜き差し監視  
  USBメモリやUSB HDDなどがUSBに接続された/USBから取り外されたことを検知する  

- USBドライブイジェクト  
  USBメモリやUSB HDDなどをプログラムから取り外しする  

## 18.1 Windowsのイベント通知

Windowsのイベント通知に関してGeminiに聞いてみました。

> Windowsメッセージ  
> - Windowsメッセージとは、Windowsオペレーティングシステムがアプリケーションやウィンドウに対して送信する通知のことです。  
> - ユーザーの操作（マウスのクリック、キーボードの入力など）、システムの状態変化（ウィンドウのサイズ変更、タイマーの満了など）、他のアプリケーションからの通知など、さまざまなイベントによって生成されます。  
> - メッセージは、メッセージの種類（WM_XXX）、メッセージのパラメータ（wParam、lParam）、メッセージが送信されたウィンドウのハンドルなど、関連する情報を含んでいます。  

> WndProc（ウィンドウプロシージャ）
> - WndProcは、Windowsメッセージを処理するためにアプリケーションが提供するコールバック関数です。  
> - 特定のウィンドウに関連付けられており、そのウィンドウに送信されたメッセージを処理します。  
> - WndProcは、受信したメッセージの種類に応じて適切な処理を行い、必要に応じて応答を返します。  

> GetMessage
> - GetMessageは、アプリケーションのメッセージキューからメッセージを取得する関数です。  
> - メッセージキューは、アプリケーションに送信されたメッセージが一時的に格納される場所です。  
> - GetMessageは、メッセージキューにメッセージが存在するまで待機し、メッセージが到着するとそれを取得してWndProcに渡します。  
> - GetMessageは、アプリケーションのメッセージループ（メッセージを取得して処理する無限ループ）の中で使用されます。  

> Windowsメッセージ、WndProc、およびGetMessageの関係性  
> 1. Windowsオペレーティングシステムは、イベントが発生すると、対応するメッセージをアプリケーションのメッセージキューに送信します。  
> 1. アプリケーションは、メッセージループの中でGetMessageを呼び出して、メッセージキューからメッセージを取得します。  
> 1. GetMessageは、取得したメッセージをアプリケーションに渡し、アプリケーションはDispatchMessage関数を呼び出してメッセージをWndProcに送ります。  
> 1. WndProcは、受信したメッセージの種類に応じて適切な処理を行い、必要に応じて応答を返します。  

## 18.2 クリップボード更新/USBドライブの抜き差し通知の設定  

Windowsメッセージが対象のウィンドウへ通知されるように設定します。  

> [!NOTE]
> win32のシステムコールが多数収録されてる「[zzl/go-win32api/win32](https://pkg.go.dev/github.com/zzl/go-win32api/win32)」パッケージを使用しています。  
> 登録されてる関数が多いためHPの表示に時間がかかります  

- クリップボード更新通知  
  ```go
  ret, w32err := win32.AddClipboardFormatListener(hwnd)
  if ret == win32.FALSE || w32err != win32.NO_ERROR {
  	log.Fatal("AddClipboardFormatListenerの失敗")
  }
  ```
  
  win32の`AddClipboardFormatListener()`をコールするだけとなります。  
  `hwnd`は
  「[16.2 user32.dllを使った方法](../16#162-user32dll%E3%82%92%E4%BD%BF%E3%81%A3%E3%81%9F%E6%96%B9%E6%B3%95)」 
  で作成した`GetWindowHandle()`を使う事により取得できます。  

- USBドライブの抜き差し通知
  ```go
  notificationFilter := win32.DEV_BROADCAST_DEVICEINTERFACE_{
  	Dbcc_size:       uint32(unsafe.Sizeof(win32.DEV_BROADCAST_DEVICEINTERFACE_{})),
  	Dbcc_devicetype: uint32(win32.DBT_DEVTYP_DEVICEINTERFACE),
  	Dbcc_reserved:   0,
  	Dbcc_classguid:  win32.GUID_IO_VOLUME_DEVICE_INTERFACE,
  }
  hDevNotify, w32err = win32.RegisterDeviceNotification(hwnd, unsafe.Pointer(&notificationFilter), win32.DEVICE_NOTIFY_WINDOW_HANDLE)
  if hDevNotify == nil || w32err != win32.NO_ERROR {
  	log.Fatal("RegisterDeviceNotificationの失敗")
  }
  ```

  こちらもwin32の`RegisterDeviceNotification()`をコールするだけですが、引数に渡す`DEV_BROADCAST_DEVICEINTERFACE_`構造体の作成が必要となります。  
  
  > [!NOTE]  
  > `DEV_BROADCAST_DEVICEINTERFACE_`の末尾の「_」はパッケージ側の誤記だと思いますが、定義されてる通りに指定しないと認識しないので、そのまま使用してます  

## 18.3 クリップボード更新/USBドライブの抜き差し通知の解除  

以下のコマンドを使います。必要に応じてエラーハンドリングを行ってください。  

- クリップボード更新通知  

  ```go
  win32.RemoveClipboardFormatListener(hwnd)
  ```

- USBドライブの抜き差し通知
  ```go
  win32.UnregisterDeviceNotification(hDevNotify)
  ```

 


[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 10. メニューバー/ツールバー/標準ダイアログを使いたい(後編)
後編では、標準ダイアログを対象にします。  

## 10.1 ファイル選択ダイアログ
![](image/open.jpg)  

コードを以下に示します。  

```go
func menuOpen(parent *gtk.ApplicationWindow) (string, error) {
	log.Println("Openを選択")
	if parent == nil {
		return "", fmt.Errorf("parent is null")
	}
	
	// ファイル選択ダイアログを作成
	fcd, err := gtk.FileChooserDialogNewWith2Buttons("Open", parent, gtk.FILE_CHOOSER_ACTION_OPEN, "ok (_O)", gtk.RESPONSE_OK, "cancel (_C)", gtk.RESPONSE_CANCEL)
	if err != nil {
		return "", err
	}
	defer fcd.Destroy()
	
	// 「OK」で終わった場合は、ファイル名を返却
	if fcd.Run() == gtk.RESPONSE_OK {
		return fcd.GetFilename(), nil
	}
	return "", nil
}
```

`FileChooserDialogNewWith2Buttons()`の第3引数は、以下のようになります。  

| FileChooserAction |  
| --- |  
| FILE_CHOOSER_ACTION_OPEN |  
| FILE_CHOOSER_ACTION_SAVE |  
| FILE_CHOOSER_ACTION_SELECT_FOLDER |  
| FILE_CHOOSER_ACTION_CREATE_FOLDER |  

第4～第5引数および第6～第7引数は、ボタンラベルの文字列とレスポンスコードになります。  
レスポンスコードは、
[7.1 標準メッセージダイアログ](../07#71-%E6%A8%99%E6%BA%96%E3%83%A1%E3%83%83%E3%82%BB%E3%83%BC%E3%82%B8%E3%83%80%E3%82%A4%E3%82%A2%E3%83%AD%E3%82%B0)
に一覧を表示してます。  

> [!TIP]
> ファイル選択ダイアログは、OSの物も使えます。
> ```go
> // OSのファイル選択ダイアログを作成
> fcd, err := gtk.FileChooserNativeDialogNew("オープン", parent, gtk.FILE_CHOOSER_ACTION_OPEN, "OK (_O)", "Cancel (_C)")
> if err != nil {
> 	return "", err
> }
> defer fcd.Destroy()
> 
> // 追加のオプション選択項目を2つ設定
> // リストボックスのラベルに"Choice1/2"が表示され、
> // リストボックスのリストに"op_label1/2"が追加される
> fcd.AddChoice("id1", "Choice1", []string{"op1", "op2"}, []string{"op_label1", "op_label2"})
> fcd.AddChoice("id2", "Choice2", []string{"op1", "op2"}, []string{"op_label1", "op_label2"})
> 
> // オプション選択項目の初期値を設定
> fcd.SetChoice("id1", "op1")
> fcd.SetChoice("id2", "op1")
> 
> // txt拡張子のフィルタを追加
> fileFilterTxt, err := gtk.FileFilterNew()
> if err != nil {
> 	return "", err
> }
> fileFilterTxt.AddPattern("*.txt")
> fileFilterTxt.SetName("text")
> fcd.AddFilter(fileFilterTxt)
> 
> // png拡張子のフィルタを追加
> fileFilterPng, err := gtk.FileFilterNew()
> if err != nil {
> 	return "", err
> }
> fileFilterPng.AddPattern("*.png")
> fileFilterPng.SetName("png")
> fcd.AddFilter(fileFilterPng)
> 
> // ダイアログを起動し、選択結果を表示
> if fcd.Run() == int(gtk.RESPONSE_ACCEPT) {
> 	// オプション選択項目の選択結果"op1/2"を表示
> 	fmt.Printf("Choice1:%s\n", fcd.GetChoice("id1"))
> 	fmt.Printf("Choice2:%s\n", fcd.GetChoice("id2"))
> 	
> 	// 選択したファイルパスを表示
> 	fmt.Printf("File path:%s\n", fcd.GetFilename())
> }
> ```

> [!NOTE]
> `FileChooserDialog`は、表示するファイルのフィルタ設定や、複数選択可否、カレントフォルダの指定などの設定も出来ます。  
> 詳しくは、[gotk3/gtk/FileChooser](https://pkg.go.dev/github.com/gotk3/gotk3/gtk#FileChooser)で確認して下さい。  

## 10.2 フォント選択ダイアログ
![](image/font.jpg)  

コードを以下に示します。  

```go
func menuFont(parent *gtk.ApplicationWindow) (string, error) {
	log.Println("Fontを選択")
	if parent == nil {
		return "", fmt.Errorf("parent is null")
	}
	
	// フォント選択ダイアログを作成
	fcd, err := gtk.FontChooserDialogNew("Font", parent)
	if err != nil {
		return "", err
	}
	defer fcd.Destroy()
	
	// 「OK」で終わった場合は、フォント名を返却
	if fcd.Run() == gtk.RESPONSE_OK {
		return fcd.GetFont(), nil
	}
	return "", nil
}
```

`GetFont()`で取得した文字列には色々な物が含まれてるので利用する時は注意してください。  
以下の関数を使うことで、各項目の値が取得可能です。  

```go
desc := pango.FontDescriptionFromString(fcd.GetFont())
desc.GetFamily()
desc.GetSize()
desc.GetStyle()
desc.GetWeight()
desc.GetStretch()
```

> [!NOTE]
> `FontChooserDialog`は、カレントフォントの設定なども出来ます。  
> 詳しくは、[gotk3/gtk/FontChooser](https://pkg.go.dev/github.com/gotk3/gotk3/gtk#FontChooser)で確認して下さい。  


## 10.3 カラー選択ダイアログ
![](image/color.jpg)  

コードを以下に示します。  

```go
func menuColor(parent *gtk.ApplicationWindow) (string, error) {
	log.Println("Colorを選択")
	if parent == nil {
		return "", fmt.Errorf("parent is null")
	}
	
	// 色選択ダイアログを作成
	ccd, err := gtk.ColorChooserDialogNew("Color", parent)
	if err != nil {
		return "", err
	}
	defer ccd.Destroy()
	
	// 「OK」で終わった場合は、RGBAを返却
	if ccd.Run() == gtk.RESPONSE_OK {
		return ccd.GetRGBA().String(), nil
	}
	return "", nil
}
```

今回は`GetRGBA().String()`で文字列を取得してますが、`GetRGBA()`で取得した`*gdk.RGBA`から、`GetRed(), GetGreen(), GetBlue(), GetAlpha()`の各関数でそれぞれの値を取得することも可能です。  

> [!NOTE]
> `ColorChooserDialog`は、カレントカラーの設定なども出来ます。  
> 詳しくは、[gotk3/gtk/ColorChooser](https://pkg.go.dev/github.com/gotk3/gotk3/gtk#ColorChooser)で確認して下さい。  

## 10.4 日付選択ダイアログ
![](image/calendar.jpg)  

コードを以下に示します。  

```go
func menuCalendar(parent *gtk.ApplicationWindow) (string, error) {
	log.Println("Calendarを選択")
	if parent == nil {
		return "", fmt.Errorf("parent is null")
	}
	
	// ダイアログを作成 
	dialog, err := gtk.DialogNew()
	if err != nil {
		return "", err
	}
	defer dialog.Destroy()
	
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)

	// カレンダーウィジェットを作成して追加
	calendar, err := gtk.CalendarNew()
	if err != nil {
		return "", err
	}
	gca, err := dialog.GetContentArea()
	if err != nil {
		return "", err
	}
	gca.Add(calendar)

	// 実行日を設定
	year, month, day := time.Now().Date()
	calendar.SelectMonth(uint(month) - 1, uint(year))
	calendar.SelectDay(uint(day))

	dialog.ShowAll()

	// ダブルクリックされた場合、レスポンスコード「OK」を送信
	calendar.Connect("day_selected_double_click", func() {
		dialog.Response(gtk.RESPONSE_OK)
	})

	// 「OK」で終わった場合は、日付を返却
	if dialog.Run() == gtk.RESPONSE_OK {
		year, month, day := calendar.GetDate()
		return fmt.Sprintf("%04d/%02d/%02d", year, int(month) + 1, day), nil
	}

	return "", nil
}
```

`gtk.Calendar`は、標準ダイアログではなく、ウィジェットとなりますので、ダイアログを作成してそこに表示しています。  
`SetDisplayOptions(CalendarDisplayOptions)`関数を使う事で、以下のオプションが指定可能です。  

| CalendarDisplayOptions |  
| --- |  
| CALENDAR_SHOW_HEADING |  
| CALENDAR_SHOW_DAY_NAMES |  
| CALENDAR_NO_MONTH_CHANGE |  
| CALENDAR_SHOW_WEEK_NUMBERS |  
| CALENDAR_SHOW_DETAILS |  

> [!NOTE]
> コードに記載の通り、`Calendar`は、カレント日付の設定などが出来ます。  
> 詳しくは、[gotk3/gtk/Calendar](https://pkg.go.dev/github.com/gotk3/gotk3/gtk#Calendar)で確認して下さい。  

> [!CAUTION]
> Monthが0始まりで管理されてるので注意が必要です。  

## 10.5 ABOUTダイアログ
![](image/about1.jpg) ![](image/about2.jpg)  

コードを以下に示します。  

```go
func menuAbout(parent *gtk.ApplicationWindow) error {
	log.Println("Aboutを選択")
	if parent == nil {
		return fmt.Errorf("parent is null")
	}
	
	// ABOUTダイアログを作成
	abd, err := gtk.AboutDialogNew()
	if err != nil {
		return err
	}
	defer abd.Destroy()
	
	abd.SetTransientFor(parent)
	
	// ロゴに親アイコンを設定
	parentIcon, err := parent.GetIcon()
	if err == nil {
		abd.SetLogo(parentIcon)
	}
	
	abd.SetProgramName("プログラムの名前")
	abd.SetVersion("バージョン x.xx")
	abd.SetComments("コメントをここに記載")
	abd.SetWebsiteLabel("ウェブサイト")
	abd.SetWebsite("ウェブサイトのURL")
	
	abd.SetCopyright("Copyright (c) 20xx Firstname Lastname")
	abd.SetLicense("LicenseTypeを指定しない場合、ここにライセンスを記載")
	abd.SetLicenseType(gtk.LICENSE_MIT_X11)
	abd.SetWrapLicense(true)	// ライセンス表示を改行する

	abd.SetAuthors([]string{"開発した人"})
	abd.SetDocumenters([]string{"ドキュメントを作成した人"})
	abd.SetTranslatorCredits("翻訳した人")
	abd.SetArtists([]string{"グラフィックデザイン、UIデザインなどに貢献した人"})
	
	abd.Run()
	return nil
}
```

コードを読めば、どこに何を記載すれば良いのかが分かるかと思います。  

## 10.6 おわりに
メニューバー/ツールバー/標準ダイアログについての説明が終わりました。  
作成したコードは、
[ここ](../09/09_MenuBar_Toolbar.go)
に置いてます。 

</br>

「[11. 表形式にデータを表示したい](../11/README.md)」へ  

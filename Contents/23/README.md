[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 23. win32 apiのクリップボードを使いたい（受信）  

gotk3のクリップボードはテキストと画像のみにしか対応してないため、win32 apiのクリップボードを使ってみます。  
今回は、gotk3を使わないコンソールアプリとなります。  

## 23.1 対応している格納形式  

以下にGeminiの説明を記載してますが、対応しているのは「TYMED_HGLOBAL」のみとなります。「TYMED_MFPICT」は「TYMED_HGLOBAL」と同様の方法で対応出来てるみたいです。  
「TYMED_GDI」は基本的には対応してませんが、「CF_BITMAP」の場合のみ個別に対応してます。  
多くのデータが「TYMED_HGLOBAL」で格納されてますのでそれなりに利用できると思います。  

> Windows環境における TYMED 列挙型は、OLE (Object Linking and Embedding) やデータ転送のメカニズムにおいて、データの格納および転送に使用されるメディアの種類を指定するために用いられます。  
>   
> 1. TYMED_HGLOBAL  
>    データはグローバルメモリブロック (Global Memory Block) に格納されています。  
> 2. TYMED_FILE  
>    データはファイルシステム上のファイルに格納されています。  
> 3. TYMED_ISTREAM  
>    データはストリームオブジェクト (IStream インターフェースを実装したオブジェクト) に格納されています。  
> 4. TYMED_ISTORAGE  
>    データはストレージオブジェクト (IStorage インターフェースを実装したオブジェクト) に格納されています。  
> 5. TYMED_GDI  
>    データは GDI (Graphics Device Interface) オブジェクトのハンドルです。  
> 6. TYMED_MFPICT  
>    データはメタファイル (HMETAFILE) のハンドルです。  
> 7. TYMED_ENHMF  
>    データは拡張メタファイル (HENHMETAFILE) のハンドルです。  

> [!NOTE]  
> データ形式として「CF_HDROP」と「Preferred DropEffect」にも対応してますので、ファイルのcopyやmove（cut）にも対応できます。  

## 23.2 使い方  

win32 apiを一つ一つ説明していくと長くなってしまうので、作成した関数の使い方を紹介していきます。  

1. クリップボードの利用開始/終了  

   `OpenClipboard()`と`CloseClipboard()`を使います。 
   パッケージとして「github.com/zzl/go-win32api/win32」を使ってますのでimportが必要です。

   ```go
   ret, w32err = win32.OpenClipboard(0)
   if ret == win32.FALSE || w32err != win32.NO_ERROR {
   	log.Fatal("OpenClipboardの失敗")
   }
   defer win32.CloseClipboard()
   ```

1. 指定したフォーマットコードのデータがあるか確認  

   以下は、テキスト（ASCII）形式のデータがクリップボードに格納されてるかどうかを確認するコードとなります。  

   ```go
   ok, err := GetDataPresent(win32.CF_TEXT)
   if err != nil {
   	log.Fatal("GetDataPresentの失敗")
   }
   if ok {
   	fmt.Println("Text形式は格納されてます")
   } else {
   	fmt.Println("Text形式は格納されてません")
   }
   ```

1. 指定したフォーマット名のデータがあるか確認  

   以下は、クリップボード形式を名前で指定して、クリップボードに格納されてるかどうかを確認するコードとなります。  

   ```go
   ok, err = ContainsData("Dib")
   if err != nil {
   	log.Fatal("ContainsDataの失敗")
   }
   if ok {
   	fmt.Println("Dib形式は格納されてます")
   } else {
   	fmt.Println("Dib形式は格納されてません")
   }
   ```

1. クリップボードに登録されている形式を列挙する  

   以下は、現在クリップボードに格納されてる全ての形式を表示するコードとなります。  

   ```go
   formats := GetFormats()
   fmt.Println("格納されてるフォーマット名一覧")
   for _, format := range formats {
   	formatName := GetFormatName(format)
   	fmt.Printf("  %s(%#x)\n", formatName, format)
   }
   ```

1. クリップボードからデータを取得  

   以下は、クリップボードに格納されてるデータを取得するコードとなります。  
   テキストデータ以外は実行ファイルと同じディレクトリに保存します。  
   上で使った`formats := GetFormats()`を実行していることが前提となってます。  

   ```go
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
   ```

   `data := GetData(format)`の戻り値の型により場合分けを行ってます。  
   `[]string`型はファイルがクリップボードにコピーされてる形式のみを前提として処理してます。  
   `[]byte`型は`string`と`[]string`以外の物はこの型で返ってくる前提で処理してます。CF_BITMAPも`[]byte`に格納しています。  
   `Preferred DropEffect`はファイルがクリップボードにコピーされてる時に格納されてる筈です。このデータを検査することによりcopyなのかmoveなのかの判定が出来ます。  
   （必ず設定されるかどうかは不明ですが、ファイルエクスプローラーは設定されてました）  
   ファイルmoveの場合、moveさせるのは受信側の処理となるため、`Preferred DropEffect`が`win32.DROPEFFECT_MOVE`の場合、ファイルコピー後に元ファイルを削除する処理が必要となります。  

## 23.3 おわりに  

クリップボードに格納されてる形式の確認、格納されてる形式の一覧、格納されてるデータの取得について、自作関数の使い方を説明しました。  

作成したファイルは、
[ここ](23_GetClipBoard.go)
に置いてます。  

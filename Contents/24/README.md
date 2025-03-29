[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 24. win32 apiのクリップボードを使いたい（送信）  

今回はクリップボードへの格納の説明となります。  

## 24.1 対応している格納形式  

対応しているのは「TYMED_HGLOBAL」のみとなります。   
「TYMED_GDI」は基本的には対応してませんが、「CF_BITMAP」の場合のみ個別に対応してます。  

## 24.2 使い方  

23章同様、win32 apiを一つ一つ説明していくと長くなってしまうので、作成した関数の使い方を紹介していきます。  

1. クリップボードの利用開始/終了  

   基本的には23章と同様ですが、クリップボードへ格納する前に`EmptyClipboard()`で空にする必要があります。  

   ```go
   ret, w32err = win32.OpenClipboard(0)
   if ret == win32.FALSE || w32err != win32.NO_ERROR {
   	log.Fatal("OpenClipboardの失敗")
   }
   defer win32.CloseClipboard()

   // クリップボードをクリア
   ret2, w32err := win32.EmptyClipboard()
   if ret2 == win32.FALSE || w32err != win32.NO_ERROR {
   	log.Fatal("EmptyClipboardの失敗")
   }
   ```

1. テキストデータを格納  
   
   以下は、マルチバイト文字列をクリップボードへ格納するコードとなります。`SetData()`の戻り値の型はerror型です。  
   
   ```go
   fmt.Println(SetData(win32.CF_UNICODETEXT, "aaabbbあいいう"))
   ```

1. ファイルを格納  

   ファイルをクリップボードへ格納する場合は、ファイルパスを「CF_HDROP」形式で格納します。  
   以下は、「win32.DROPEFFECT_MOVE」を指定してファイルをクリップボードへ格納するコードとなります。

   ```go
   files := []string{
   	"D:\\test\\日本語ファイル.txt",
   }
   fmt.Println(SetData(win32.CF_HDROP, files, win32.DROPEFFECT_MOVE))
   ```

1. byte配列を格納  

   以下は、byte配列をクリップボードへ格納するコードとなります。「abcd」をbyte配列でWaveAudioとして格納しています。  

   ```go
   bytes := []byte{0x61, 0x62, 0x63, 0x64}
   fmt.Println(SetData(win32.CF_WAVE, bytes))
   ```

1. カスタムフォーマット名を指定してbyte配列を格納  

   以下は、カスタムフォーマット名を指定してbyte配列を格納するコードとなります。カスタムフォーマット名は存在しない場合、新規登録されます。  

   ```go
   bytes := []byte{0x61, 0x62, 0x63, 0x64}
   fmt.Println(SetData("Custom Format", bytes))
   ```

   23章で紹介してませんでしたが、`GetFormatCode(name string)`関数を23章のコード内に作成しています。この関数を使ってカスタムフォーマット名からフォーマットコードを取得すれば、クリップボードからカスタムフォーマット名で登録したデータを受信することが出来ます。  

1. CF_BITMAP形式で画像を格納  

   以下は、CF_BITMAP形式で画像を格納するコードとなります。"image", "image/color", "image/draw"のimportが必要となります。  

   ```go
   // 新しいRGBA画像(緑色の正方形)を生成
   img := image.NewRGBA(image.Rect(0, 0, 200, 200))
   green := color.RGBA{0, 255, 0, 255}
   draw.Draw(img, img.Bounds(), &image.Uniform{green}, image.Point{}, draw.Src)
   
   // 画像をCF_BITMAP形式で格納
   fmt.Println(SetData(win32.CF_BITMAP, img))
   ```

## 24.3 おわりに  

`SetData()`された形式の数だけクリップボードへ格納されていきます。（CF_TEXTとCF_BITMAPを`SetData()`した場合、2つの形式のデータがクリップボードへ格納されています）また、`EmptyClipboard()`を実行しないと、処理開始前に格納されているデータも残ったままとなりますので、注意して下さい。  


作成したファイルは、
[ここ](24_SetClipBoard.go)
に置いてます。  

</br>

「[25. ](../25/README.md)」へ

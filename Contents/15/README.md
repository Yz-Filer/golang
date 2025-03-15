[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 15. （まとめ）CSSを使った書式設定

「[8.2 ラベルの書式設定](../08#82-%E3%83%A9%E3%83%99%E3%83%AB%E3%81%AE%E6%9B%B8%E5%BC%8F%E8%A8%AD%E5%AE%9A)」で説明した「Pango Markup」を使う方法は、表示する文字列が決定した後に使う場合には良いのですが、編集中に使うには不便です。  
そのため、「[8.3 ウィンドウやダイアログ全体の書式設定](../08#83-%E3%82%A6%E3%82%A3%E3%83%B3%E3%83%89%E3%82%A6%E3%82%84%E3%83%80%E3%82%A4%E3%82%A2%E3%83%AD%E3%82%B0%E5%85%A8%E4%BD%93%E3%81%AE%E6%9B%B8%E5%BC%8F%E8%A8%AD%E5%AE%9A)」で紹介したCSSを使う方法を使っていきたいと思います。  

## 15.1 CSS文字列の作成  

今回は、編集ウィンドウのTextViewと付箋ウィンドウのWindow（およびWindowに含まれるLabel）が同様の見た目になるように同じ書式を設定していきます。  

```go
func BuildCSS(st StickyStr, isTextView bool) string {
	css := "* {\n"
	
	if isTextView {
		// 選択領域のカーソルの色まで同じになるので、TextViewは範囲を限定
		css = "text, .view {\n"
	}
	
	css+= "  font-family: " + st.FontFamily + ";\n"
	css+= "  font-size: " + strconv.Itoa(st.FontSize / 1024) + "pt;\n"
	css+= "  font-style: " + FontStyleToString(st.FontStyle) + ";\n"
	css+= "  font-weight: " + strconv.Itoa(int(st.FontWeight)) + ";\n"
	css+= "  font-stretch: " + FontStretchToString(st.FontStretch) + ";\n"
	css+= "  color: rgba("
	css+= strconv.Itoa(st.FgColor[0]) + ", "
	css+= strconv.Itoa(st.FgColor[1]) + ", "
	css+= strconv.Itoa(st.FgColor[2]) + ", "
	css+= strconv.FormatFloat(float64(st.FgColor[3]) / 255.0, 'f', 2, 64) + ");\n"
	css+= "  background-color: rgba("
	css+= strconv.Itoa(st.BgColor[0]) + ", "
	css+= strconv.Itoa(st.BgColor[1]) + ", "
	css+= strconv.Itoa(st.BgColor[2]) + ", "
	css+= strconv.FormatFloat(float64(st.BgColor[3]) / 255.0, 'f', 2, 64) + ");\n"
	css+= "}"
	
	return css
}
```

付箋ウィンドウ側は、編集などしなくて表示するだけなので、対象は「`*`」を指定しても問題なかったのですが、TextView側は、「`*`」だと範囲指定時の背景色も全て同じ色に設定されて選択範囲が見えなくなってしまいましたので、「`text, .view`」を指定しています。  

あとは、フォントの設定、フォント色の設定、背景色の設定となります。TextViewの背景色透明度も編集ウィンドウと同様に設定してますが、TextViewの裏にあるWindowの背景色を変更してないため、透明にはなりません。  

> [!NOTE]  
> FontChooserDialogで取得したフォントサイズは1024で割るとptに変換出来ます。  
> RGBの指定は0～255ですが、透明度の指定は0.0～1.0になります。  

> [!TIP]  
> gtk3でサポートしているCSSプロパティは、[ここ](https://docs.gtk.org/gtk3/css-properties.html)を参照して下さい。  

## 15.2 CSS文字列を反映  

編集ウィンドウのTextViewと付箋ウィンドウのWindowにそれぞれCSSを使って書式設定をしていきます。

```go
func ApplyStyle(widget gtk.IWidget, st StickyStr) error {
	// プロバイダーを作成
	cssProvider, err := gtk.CssProviderNew()
	if err != nil {
		return err
	}
	
	cssStr := ""
	
	// コンテキストを取得し、CSS文字列を作成
	var context *gtk.StyleContext
	switch widget.(type) {
		case *gtk.TextView:
			context, err = widget.(*gtk.TextView).GetStyleContext()
			cssStr = BuildCSS(st, true)
		default:
			context, err = widget.(*gtk.Window).GetStyleContext()
			cssStr = BuildCSS(st, false)
	}
	if err != nil {
		return err
	}
	
	// CSSをロード
	err = cssProvider.LoadFromData(cssStr)
	if err != nil {
		return err
	}
	
	// 書式を反映
	context.AddProvider(cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	
	return nil
}
```

「[8.3 ウィンドウやダイアログ全体の書式設定](../08#83-%E3%82%A6%E3%82%A3%E3%83%B3%E3%83%89%E3%82%A6%E3%82%84%E3%83%80%E3%82%A4%E3%82%A2%E3%83%AD%E3%82%B0%E5%85%A8%E4%BD%93%E3%81%AE%E6%9B%B8%E5%BC%8F%E8%A8%AD%E5%AE%9A)」で紹介した内容とほぼ同じですが、第1引数で渡されたオブジェクトの型を判別して対象を変更してます。  

</br>

「[16. （まとめ）タスクバーにアイコンを表示させない方法](../16/README.md)」へ

# 13. ファイルの存在確認とファイル入出力

## 13.1 ファイルの存在確認に使う関数

go言語では標準関数で直接ファイルの存在確認を行ってくれる関数は提供されてないようなので、自分で判定する必要があります。  
どのようなコードにすれば良いかをGeminiに聞いてみたところ、  

> ```go
> func fileExists(filename string) bool {
> 	_, err := os.Stat(filename)
> 	return err == nil || !os.IsNotExist(err)
> }
> ```

と回答がありました。  
少し前までは上記のようなコードをよく見たのですが、WEB検索すると`errors.Is()`を使ったコードも見かけるので、違いを聞いてみました。  

> Go 1.13以降では、エラーをラップする機能が導入され、errors.Is()はラップされたエラーに対しても正常に動作します。  

> 最近のGoのコーディングでは、errors.Is()を使用する方が推奨される傾向があります。os.ErrNotExistを使い、ファイルの不存在をチェックする方がより明確で、エラー処理も適切に行えるためです。  

ということなので、`errors.Is()`を使えばよさそうに思いました。  

## 13.2 ファイルの存在確認関数の戻り値  

「13.1」ではファイルが存在するか/存在しないかをtrue/falseで返却してます。  
IsNotExist以外のエラーを考慮しなくてよいのかをGeminiに聞いてみました。  

> ファイルの存在確認では、返り値がtrueまたはfalseだけで完全に判断できないケースがあります。  
> - 権限エラー  
>   ファイルが存在しても、読み取り権限がない場合はos.Stat()がエラーを返すことがあります。この場合、os.IsExist()はfalseを返しますが、ファイルは実際には存在します。  
> - シンボリックリンク  
>   シンボリックリンクの場合、リンク先のファイルが存在するかどうかを確認する必要があります。os.Stat()はシンボリックリンク自体の情報を返すため、リンク先の存在確認には追加の処理が必要です。  
> - ネットワークファイルシステム  
>   ネットワークファイルシステムでは、ファイルが存在するかどうかの確認に時間がかかる場合や、ネットワークの問題で誤った結果が返される場合があります。  

やはり、「ファイルが存在しない」以外のエラーも対応が必要そうです。シンボリックリンクの対応を考慮しない単純なファイルの存在確認だけなら、以下のような関数で良いみたいです。  

```go
func FileExists(filename string) (bool, error) {
	f, err := os.Stat(filename)
	if err != nil {
		// ファイルが存在しない場合
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		// その他のエラーの場合
		return false, err
	}
	
	// ファイルが存在した場合、ディレクトリ判定
	return !f.IsDir(), nil
}
```

`IsDir()`は呼び出し側で判定した方が良い場合もあるので状況に応じて変更して下さい。  

## 13.3 付箋アプリのファイル入出力

付箋アプリでは、付箋を以下のような構造体で管理しようと考えています。  

```go
type StickyStr struct {
	X				int
	Y				int
	Width			int
	Height			int
	Note			string
	FontFamily		string
	FontSize		int
	FontStyle		pango.Style
	FontWeight		int
	FontStretch		pango.Stretch
	FgColor			[4]int
	BgColor			[4]int
	Align			float64
	StickyWindow	*gtk.Window
}
```

jsonやxmlで保存することも考えたのですが、バイナリで保存したいと考えたため、Geminiに聞いてみました。  

> encoding/gobパッケージを使用すると、構造体をより簡単にバイナリファイルへ保存・読み込みできます。  

メンバに`*gtk.Window`が含まれており、保存対象から除外可能かどうかをGeminiに聞いてみました。  

> Go言語では、構造体のフィールドが公開されている（頭文字が大文字）場合にのみ、外部パッケージからアクセスできます。encoding/gob パッケージも外部パッケージであるため、公開されているフィールドのみをエンコード/デコードします。  
> したがって、*gtk.Window フィールドの頭文字を小文字にすると、gob エンコーダはそのフィールドを無視します。  

`StickyWindow`変数の頭文字を小文字に変更して`encoding/gob`を使う事にします。  
コードは以下のようになります。  

```go
// 既存のデータファイルをバックアップ
func BackupDataFile(filename, bakFilename string) error {
	// データファイルの存在確認
	ok, err := FileExists(filename)
	if err != nil {
		return err
	}

	// データファイルが存在しなかった場合、バックアップ不要
	if !ok {
		return nil
	}

	// バックアップファイルの存在確認
	ok, err = FileExists(bakFilename)
	if err != nil {
		return err
	}
	
	// バックアップファイルが存在していたら削除
	if ok {
		err := os.Remove(bakFilename)
		if err != nil {
			return err
		}
	}
	
	// データファイル名をバックアップファイル名にリネーム
	return os.Rename(filename, bakFilename)
}

// マップをファイルに保存する
func SaveStickyMap(filename string) error {
	// 既存のデータファイルをバックアップ
	err := BackupDataFile(filename, filename + ".bak")
	if err != nil {
		return err
	}

	// データファイルを作成し、データを保存
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(StickyMap)
	if err != nil {
		return err
	}

	return nil
}

// ファイルからマップを読み込む
func LoadStickyMap(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	StickyMap = make(map[string]StickyStr)
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&StickyMap)
	if err != nil {
		return err
	}

	return nil
}
```

`StickyMap`はグローバル変数となります。  
保存時は、既存のファイルを1面分までバックアップするようにしてます。  

</br>

「[14. カスタムシグナル](../14/README.md)」へ
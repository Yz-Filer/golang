[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 2. 他のパソコンでも実行したい
「[1. gotk3を使って、Simple windowを作成する](../01/README.md)」で作成したアプリは、他のパソコンにコピーしても、そのままでは動作しません。  
動作に必要なモジュールやアイコンなどもコピーする必要があります。  
インストーラー作成とかで自動的に洗い出す方法があれば良いのですが、情報を見つけられませんでした。  
Geminiに聞いても、具体的な解決方法は分かりませんでしたので、手動で対応する方法を載せてます

## 2.1 コマンドを使って依存モジュールを調査する
bashの依存モジュールを調査する方法をGeminiに聞いてみました。
> Linuxのlddコマンドと同様に、MSYS2でもlddコマンドが利用できます。
> これにより、指定した実行ファイルが依存している共有ライブラリ（.dllファイル）の一覧が表示されます。
> ```bash
> ldd /usr/bin/bash
> ```
今回は、MSYS2上で以下のようなコマンドを使ってgtkが使用してそうなファイルの依存モジュールを調査しました。  
```bash
ldd /mingw64/bin/libgtk-3-0.dll | grep "/mingw64"
```
> [!NOTE]  
> 環境によっては、指定パスが異なるかもしれません  
> 「libgtk-3-0.dll」部分は「libgdk_pixbuf-2.0-0.dll」などgtkに関連しそうな複数のライブラリを指定して調査しました。

## 2.2 依存モジュールを配置する
調査結果より、実行ファイルと同じディレクトリに以下のファイルが必要なことが分かりました。  

<pre>
D:\TEST
├─gdbus.exe
├─libLerc.dll
├─libatk-1.0-0.dll
├─libbrotlicommon.dll
├─libbrotlidec.dll
├─libbz2-1.dll
├─libcairo-2.dll
├─libcairo-gobject-2.dll
├─libdatrie-1.dll
├─libdeflate.dll
├─libepoxy-0.dll
├─libexpat-1.dll
├─libffi-8.dll
├─libfontconfig-1.dll
├─libfreetype-6.dll
├─libfribidi-0.dll
├─libgcc_s_seh-1.dll
├─libgdk-3-0.dll
├─libgdk_pixbuf-2.0-0.dll
├─libgio-2.0-0.dll
├─libglib-2.0-0.dll
├─libgmodule-2.0-0.dll
├─libgobject-2.0-0.dll
├─libgraphite2.dll
├─libgtk-3-0.dll
├─libharfbuzz-0.dll
├─libiconv-2.dll
├─libintl-8.dll
├─libjbig-0.dll
├─libjpeg-8.dll
├─liblzma-5.dll
├─libpango-1.0-0.dll
├─libpangocairo-1.0-0.dll
├─libpangoft2-1.0-0.dll
├─libpangowin32-1.0-0.dll
├─libpcre2-8-0.dll
├─libpixman-1-0.dll
├─libpng16-16.dll
├─librsvg-2-2.dll
├─libsharpyuv-0.dll
├─libstdc++-6.dll
├─libthai-0.dll
├─libtiff-6.dll
├─libwebp-7.dll
├─libwinpthread-1.dll
├─libxml2-2.dll
├─libzstd.dll
└─zlib1.dll
</pre>
glib-2.0は必須だそうです。  
`/mingw64/share/glib-2.0/schemas`を実行ファイルディレクトリ配下の「share/glib-2.0」ディレクトリ配下へコピー
<pre>
D:\TEST
└─share
    └─glib-2.0
        └schemas
          └全ファイル
</pre>

> [!CAUTION]
> - 作成するアプリによっては増える可能性があるので、コンソールにgtkからのエラーが出力されたら都度対処することになります。
> - 作成済みアプリがある場合、`ldd 作成済みアプリのパス | grep "/mingw64"`で足りないモジュールを探すことができます。  

## 2.3 マウスカーソルやアイコンを表示出来るようにする
マウスカーソルやアイコンの画像ファイルも必要です。  
`/mingw64/share/icons/Adwaita`を実行ファイルディレクトリ配下の「share/icons」ディレクトリ配下へコピー
<pre>
D:\TEST
└─share
    └─icons
        └─Adwaita
            └全ファイル(※)
</pre>
> [!TIP]
> ※結構ファイル数があるので、以下のファイルを最低限コピーして、コンソールにgtkのエラーが出たら対処するという方法もあります。
> - Adwaitaディレクトリ
>   - icon-theme.cache, index.theme
> - Adwaita/cursors
>   - alias.cur, copy.cur, default.cur, dnd-ask.cur, dnd-copy.cur, dnd-link.cur, dnd-move.cur, dnd-no-drop.cur, dnd-none.cur, move.cur, no-drop.cur, not-allowed.cur, wait.ani
> - Adwaita/scalable
>   - 全ディレクトリと全ファイル
> - Adwaita/symbolic
>   - actionsディレクトリの以下のファイル
>     - bookmark-new-symbolic.svg, folder-new-symbolic.svg. object-flip-horizontal-symbolic.svg, object-flip-vertical-symbolic.svg, object-rotate-left-symbolic.svg, object-rotate-right-symbolic.svg, object-select-symbolic.svg
> - Adwaita/symbolic-up-to-32
>   - 全ディレクトリと全ファイル
  
マウスカーソルやアイコン表示のために、以下も必要です。  
`/mingw64/lib/gdk-pixbuf-2.0/`を実行ファイルディレクトリ配下の「lib」ディレクトリ配下へコピー
<pre>
D:\TEST
└─lib
    └─gdk-pixbuf-2.0
        └─2.10.0
            ├─loaders.cache
            └─loaders
                └全ファイル
</pre>

## 2.4 おわりに
以上で、他のパソコンでも動かすことが出来るようになると思います。  
必要なファイルが多く、アプリ毎に準備するのは大変なので、一つのフォルダに複数の実行ファイルを置くようにするのが良いのかもしれません。  
環境変数でどうにかならないかと思ったのですが、解決方法は見つかりませんでした。  

</br>

「[3. テーマを使いたい](../03/README.md)」へ

# 3. テーマを使いたい
gtk3は実行ファイルが出来た後に、テーマを変更することでアプリの外観を変更することが可能です。  

## 3.1 設定ファイル
以下のファイルを作成し、テキストエディタで編集します。  
<pre>
D:\TEST
└─share
    └─gtk-3.0
        └─settings.ini
</pre>

1. 標準テーマ（Adwaita）  
   settings.iniを作成しないか、以下の内容を記載
   ```
   [Settings]
   ```
1. 標準テーマ（Adwaita）をダークテーマへ変更  
   以下の内容を記載
   ```
   [Settings]
   gtk-application-prefer-dark-theme = true
   ```

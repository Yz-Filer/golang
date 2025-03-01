# go言語＆gotk3をちょっとやり直してみたい
## はじめに
Go言語とgotk3を用いてGoogle検索で調べながらファイラーを作成してみたりしたのですが、動作が不安定な時があったため、ちゃんと調べ直したいと考えてました。  
以前は参考となる情報が少なく苦労しましたが、現在ではAI技術が発展しているため、Geminiに質問しながらコードを作成してみようと思います。  

元々gtk4への移行を検討していましたが、うちの非力なパソコンではコンパイルに時間がかかり過ぎる上、完成したアプリケーションも「もっさり」だったため、断念しました。  
gtk3へ割り切ることで、非推奨となった機能（タスクトレイへの格納など）も活用していきたいと考えています。  

> [!NOTE]
> 対象OSはWindowsとなります。  
> go言語やgotk3のプログラミング方法や環境構築などの導入部分の解説は端折ってます。

## コンテンツ
[1. gotk3を使って、Simple windowを作成する](01/README.md)  
<img src="01/image/window.jpg" height="89" />  

[2. 他のパソコンでも実行したい](02/README.md)  
<img src="02/image/computer_tokui_boy.png" height="89" />  

[3. テーマを使いたい](03/README.md)  
<img src="03/image/window3.jpg" height="89" />  

[4. これに気をつけないとアプリがクラッシュする](04/README.md)  
<img src="04/image/computer_note_bad.png" height="89" />  

[5. 半透明の付箋もどき](05/README.md)  
<img src="05/image/window_multi.jpg" height="89" />  

[6. タスクトレイに格納したい](06/README.md)  
<img src="06/image/taskbar_menu.jpg" height="89" />  

[7. 7. メッセージダイアログとステータスバーを表示したい](07/README.md)  
<img src="07/image/std_dialog.jpg" height="89" />  

[8. ヘッダーバー・ラベル書式・ウィンドウ書式のカスタマイズ](08/README.md)  
<img src="08/image/window.jpg" height="89" /> <img src="08/image/custom_dialog_markup.jpg" height="89" />  


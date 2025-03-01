[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 9. メニューバー/ツールバー/標準ダイアログを使いたい(前編)
前編では、メニューバー/ツールバーを対象にし、後編で、標準ダイアログを対象にします。  

> [!NOTE]  
> gtk3のウィジェットは「The Python GTK+ 3 Tutorial」の[5. Widget Gallery](https://python-gtk-3-tutorial.readthedocs.io/en/latest/gallery.html)が画像が多くて参考になりました。  

![](image/glade.jpg)  

gladeでApplicationWindow上に、gtkMenuBarとgtkToolBarを配置します。  
gtkMenuItemとgtkToolButtonは、左側のペインで親を右クリックし、「Edit...」-「+」で追加できます。  

作成したファイルは、
[ここ](glade/09_MainWindow.glade)
に置いてます。  

> [!NOTE]
> - MenuBarは、「gtkMenuBar」-「gtkMenuItem」-「gtkMenu」-「gtkMenuItem」のような親子関係になります。
> - MenuItemの「下線を使用する」プロパティは、「ダイアログ(_D)」のようにすると、下線の後ろの文字がニーモニックキーになり、「ALT+D」でメニューが選択出来るようになります。  
>   「ダイアログ(_D)」-「開く(_O)」のようにメニューをプルダウンした後のメニューアイテムを選択する場合、「ALT+D」-「O」になります。  
> - ToolButtonの画像は「画像」-「ストックID」から選択すると標準のアイコンが使用されます。「画像」-「アイコン名」から選択すると、おそらく標準のアイコンテーマから選択されるのだと思います。

## 9.1 メニューバー

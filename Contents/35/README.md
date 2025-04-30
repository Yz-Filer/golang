[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 35. dllを使ったDrag and Drop  

「[22. Drag and Dropが使い難い](../22/README.md)」で説明したように、Windows環境のgotk3だけではDrag and Dropが限定した使い方しか出来なかったため、dllを使ってDrag and Dropを実装してみます。  

使うdllは、「[OleDragDrop.dll](https://www.vector.co.jp/soft/win95/prog/se240117.html)」ですが、32bitでコンパイルされてるため、リコンパイルが必要です。  
幸い、ソースも公開してくれてますので、以下のコマンドでリコンパイルしてみました。  

```bat
cl /LD OleDragDrop.c uuid.lib ole32.lib user32.lib shell32.lib
```

> [!CAUTION]  
> `cl`はVisualStudioのコンパイラです。64bitでdllをコンパイルできれば、`cl`以外でコンパイルしても良いと思いますので自分の環境にあったコンパイラを調べてみて下さい。 


[go言語 & gotk3をちょっとやり直してみたい](../../README.md#go%E8%A8%80%E8%AA%9Egotk3%E3%82%92%E3%81%A1%E3%82%87%E3%81%A3%E3%81%A8%E3%82%84%E3%82%8A%E7%9B%B4%E3%81%97%E3%81%A6%E3%81%BF%E3%81%9F%E3%81%84)  

# 27. プロセス間通信（名前付きパイプ）のメモ  

"github.com/Microsoft/go-winio"パッケージを使って、C#アプリとプロセス間通信を実施するための自分用のメモとなります。  

「go-winio」はH.P.に関数の説明などがないため、関数の説明も軽くしておきますが、全ての関数は試してなく、Geminiの説明そのままの物もあります。実際に使用する場合は、想定通りの挙動をするかどうかを確認してから使用して下さい。  

## 27.1 サーバー側で使う関数  

- `func ListenPipe(path string, c *PipeConfig) (net.Listener, error)`  
  指定されたパス (名前) を持つ名前付きパイプのリスナーを作成します。  
  Windows における名前付きパイプのパスは通常`\\.\pipe\パイプ名`の形式です。  
  第2引数のPipeConfigは以下のメンバーを持ちます。(nilで指定なし)  
  - SecurityDescriptor: パイプのセキュリティ記述子  
  - MessageMode: メッセージモード(true)またはバイトストリームモード(false)(デフォルト)を指定します。通常はバイトストリームモードを使うようなので、メッセージモードの説明は省略します。  
  - InputBufferSize: 入力バッファのサイズ (バイト単位) を指定します。0 の場合はシステムデフォルトが使用されます。  
  - OutputBufferSize: 出力バッファのサイズ (バイト単位) を指定します。0 の場合はシステムデフォルトが使用されます。  

- `func (l *win32PipeListener) Accept() (net.Conn, error)`  
  クライアントからの接続を受け付けます。  

- `func (l *win32PipeListener) Close() error`  
  ListenPipe 関数によって作成された名前付きパイプのリスナーを閉じます。  

- `func (l *win32PipeListener) Addr() net.Addr`  
  リスナーがリッスンしているネットワークアドレスを返します。  

## 27.2 クライアント側で使う関数  

- `func DialPipe(path string, timeout *time.Duration) (net.Conn, error)`  
  指定されたパスを持つ名前付きパイプへの接続を試みます。  
  第2引数はタイムアウト期間を指定します。nilを指定すると無期限に待機します。  

- `func DialPipeContext(ctx context.Context, path string) (net.Conn, error)`  
  コンテキストを使用して名前付きパイプへの接続を試みます。  

- `func DialPipeAccess(ctx context.Context, path string, access uint32) (net.Conn, error)`  
  特定のアクセス権限を指定して名前付きパイプへの接続を試みます。コンテキストも使用できます。  
  accessは通常「golang.org/x/sys/windows」の以下の権限を指定します。  
  - windows.GENERIC_READ：読み取り権限  
  - windows.GENERIC_WRITE：書き込み権限  
  - windows.GENERIC_READ | windows.GENERIC_WRITE：読み取り/書き込み権限  

- `func DialPipeAccessImpLevel(ctx context.Context, path string, access uint32, impLevel PipeImpLevel) (net.Conn, error)`  
  特定のアクセス権限と偽装レベルを指定して名前付きパイプへの接続を試みます。コンテキストも使用できます。  
  クライアントAPから同一マシン上のサーバーAPへ、ファイルなどのアクセス権限を名前付きパイプと偽装の仕組みを通じて間接的に委譲するために使えます。  
  - winio.PipeImpLevelAnonymous  
    サーバーはクライアントに関するいかなる情報も得られず、クライアントになりすますこともできません。  
  - winio.PipeImpLevelIdentification  
    サーバーはクライアントの識別情報（セキュリティ ID やグループなど）を知ることができますが、クライアントになりすまして操作を行うことはできません。  
  - winio.PipeImpLevelImpersonation  
    サーバーがクライアントになりすましてローカルで操作を行うことができますが、ネットワーク越しに別のサーバーにアクセスする際には自分の資格情報を使用します。  
  - winio.PipeImpLevelDelegation  
    サーバーがクライアントになりすまして、他のサーバーを含むネットワーク全体で操作を行うことができる強力な権限です。  

## 27.3 共通の関数  

1. 取得したnet.Connで使える関数  
   - `func (f *win32Pipe) LocalAddr() net.Addr`  
     ローカルの名前付きパイプのパスを取得します。（LocalAddrもRemoteAddrも同じパイプ名）  

   - `func (f *win32Pipe) RemoteAddr() net.Addr`  
     接続先の名前付きパイプのパスを取得します。（LocalAddrもRemoteAddrも同じパイプ名）  

   - `func (f *win32Pipe) SetDeadline(t time.Time) error`  
     パイプ通信が一定時間内に完了しない場合に、処理を中断してリソースを解放したい場合に利用します。  
     例えば、クライアントが一定時間応答しない場合に、サーバー側で接続をタイムアウトさせるといった処理を実装できます。  

   - `func (f *win32Pipe) Disconnect() error`  
     サーバー側からクライアント側との間の接続を切断します。サーバー側は閉じられないためクライアント側から再度接続出来ます。  
     サーバー側を閉じるためにはClose()メソッドを呼び出す必要があります。  
     クライアント側からこのメソッドを実行しても期待する動作はしません。Close()メソッドを呼び出すべきです。  

1. 取得したnet.Connで使える関数(メッセージモード専用の関数）  
   通常はバイトストリームモードを使うようなので、メッセージモードの説明は省略します。  
   - `func (f *win32MessageBytePipe) CloseWrite() error`  
   - `func (f *win32MessageBytePipe) Write(b []byte) (int, error)`  
   - `func (f *win32MessageBytePipe) Read(b []byte) (int, error)`  

1. 取得したnet.Addrで使える関数  
   - `func (pipeAddress) Network() string`  
     "pipe"文字列を返します。  
   - `func (s pipeAddress) String() string`  
     名前付きパイプのパス文字列を返します。  

## 27.4 C#がサーバー側、goがクライアント側のコード  

サーバー側のコードは以下のようになります。  

```C#
using System;
using System.IO.Pipes;
using System.IO;

public class NamedPipe
{
    [STAThread]
    public static void Main(string[] args)
    {
        // 名前付きパイプを作成
        using (NamedPipeServerStream pipeServer = new NamedPipeServerStream("mypipe", PipeDirection.InOut))
        {
            Console.WriteLine("C# server is listening...");

            // クライアントの接続を待機
            pipeServer.WaitForConnection();

            Console.WriteLine("Client connected.");

            using (StreamReader reader = new StreamReader(pipeServer))
            using (StreamWriter writer = new StreamWriter(pipeServer) { AutoFlush = true })
            {
                // 受信
                string line = reader.ReadLine();
                Console.WriteLine("Received from Go: " + line);
                
                // 送信
                writer.WriteLine("Hello from C#");

                // 受信
                line = reader.ReadLine();
                Console.WriteLine("Received from Go: " + line);
            }
        }
    }
}
```

クライアント側のコードは以下のようになります。  

```go
package main

import (
    "bufio"
    "fmt"
    "log"

    "github.com/Microsoft/go-winio"
)

const pipeName = `\\.\pipe\mypipe`

func main() {
    // 名前付きパイプに接続
    conn, err := winio.DialPipe(pipeName, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // 送信
    fmt.Fprintln(conn, "Hello from Go")

    // 受信
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        fmt.Println("Received from C#: " + scanner.Text())
        break
    }

    // 送信
    fmt.Fprintln(conn, "Go is done")

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
}
```

> [!TIP]  
> C#のコンパイルは、
> `csc.exe /platform:x64 /target:exe *.cs`で実施してます。  
> `csc.exe`は  
> `C:\Windows\Microsoft.NET\Framework\v****\csc.exe`  
> (****は環境により異なります)  にあると思います。  

## 27.5 goがサーバー側、C#がクライアント側のコード  

サーバー側のコードは以下のようになります。  

```go
package main

import (
    "bufio"
    "fmt"
    "log"

    "github.com/Microsoft/go-winio"
)

const pipeName = `\\.\pipe\mypipe`

func main() {
    // 名前付きパイプのリスナーを作成
    listener, err := winio.ListenPipe(pipeName, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()

    fmt.Println("Go server is listening...")

    // クライアントからの接続を待機
    conn, err := listener.Accept()
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    fmt.Println("Client connected.")

    // 受信
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        fmt.Println("Received from C#: " + scanner.Text())
        break
    }

    // 送信
    fmt.Fprintln(conn, "Hello from Go")

    // 受信
    scanner = bufio.NewScanner(conn)
    for scanner.Scan() {
        fmt.Println("Received from C#: " + scanner.Text())
        break
    }
}
```

クライアント側のコードは以下のようになります。  

```C#
using System;
using System.IO.Pipes;
using System.IO;

public class NamedPipeClient
{
    public static void Main(string[] args)
    {
        // 名前付きパイプに接続
        using (NamedPipeClientStream pipeClient = new NamedPipeClientStream(".", "mypipe", PipeDirection.InOut))
        {
            try
            {
                pipeClient.Connect(5000); // タイムアウト設定 (ミリ秒)
            }
            catch (TimeoutException)
            {
                Console.WriteLine("Timeout waiting for server connection.");
                return;
            }
            catch (Exception e)
            {
                Console.WriteLine("Error connecting to server: {0}", e.Message);
                return;
            }

            using (StreamReader reader = new StreamReader(pipeClient))
            using (StreamWriter writer = new StreamWriter(pipeClient) { AutoFlush = true })
            {
                // 送信
                writer.WriteLine("Hello from C#");

                // 受信
                string line = reader.ReadLine();
                Console.WriteLine("Received from Go: " + line);

                // 送信
                writer.WriteLine("C# is done");
            }
        }
    }
}
```

## 27.6 おわりに  

プロセス間通信について紹介しました。サーバー/クライアントは、両方ともgoアプリにすることも出来ます。

作成したファイルは、  
[C#サーバー](27_winio_server.cs) / [goクライアント](27_winio_client.go)  
[goサーバー](27_winio_server.go) / [C#クライアント](27_winio_client.cs)  
に置いてます。  


</br>

「[28.](../28/README.md)」へ

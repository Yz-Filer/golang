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

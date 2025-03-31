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

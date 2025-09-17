// コンパイルコマンド: gcc your_code.c -o your_program.exe -lshlwapi -lole32 -loleaut32 -luuid

#ifndef UNICODE
#define UNICODE
#endif

#include <windows.h>
#include <shlobj.h>
#include <shlwapi.h>
#include <stdio.h>

#pragma comment(lib, "shlwapi.lib")

// COMインターフェイスを安全に解放するためのマクロ
#define SAFE_RELEASE(p) if (p) { (p)->lpVtbl->Release(p); (p) = NULL; }

// --- グローバル変数 ---
// IContextMenu2/3 は TrackPopupMenuEx がアクティブな間に発生するオーナー描画等のメッセージを
// ウィンドウプロシージャで処理するために必要。そのため、一時的にグローバル変数として保持する。
static IContextMenu2* g_pContextMenu2 = NULL;
static IContextMenu3* g_pContextMenu3 = NULL;

// --- 関数のプロトタイプ宣言 ---
LRESULT CALLBACK WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);
void FreePIDLArray(LPITEMIDLIST* pidls, int count);
HRESULT CreateContextMenuForFiles(HWND hwnd, WCHAR** paths, int fileCount, IContextMenu** ppContextMenu);
int ShowContextMenu(HWND hwnd, POINT pt, WCHAR** filePaths, int fileCount);

/**
 * @brief エントリーポイント
 */
int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE hPrevInstance, LPSTR lpCmdLine, int nCmdShow) {
	OleInitialize(NULL);

	const WCHAR CLASS_NAME[] = L"MyContextMenuWindowClass";

	WNDCLASSW wc = {0};
	wc.lpfnWndProc   = WindowProc;
	wc.hInstance	 = hInstance;
	wc.lpszClassName = CLASS_NAME;
	wc.hCursor	   = LoadCursor(NULL, IDC_ARROW);
	wc.hbrBackground = (HBRUSH)(COLOR_WINDOW + 1);

	RegisterClassW(&wc);

	HWND hwnd = CreateWindowExW(
		0, CLASS_NAME, L"Context Menu Refactored",
		WS_OVERLAPPEDWINDOW,
		CW_USEDEFAULT, CW_USEDEFAULT, 300, 200,
		NULL, NULL, hInstance, NULL
	);

	if (!hwnd) {
		OleUninitialize();
		return 0;
	}

	ShowWindow(hwnd, nCmdShow);
	UpdateWindow(hwnd);

	MSG msg = {0};
	while (GetMessage(&msg, NULL, 0, 0) > 0) {
		TranslateMessage(&msg);
		DispatchMessage(&msg);
	}

	OleUninitialize();
	return (int)msg.wParam;
}

/**
 * @brief ウィンドウプロシージャ
 */
LRESULT CALLBACK WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
	// オーナー描画などのメニュー関連メッセージを IContextMenu に転送する
	if (g_pContextMenu2 && (uMsg == WM_DRAWITEM || uMsg == WM_MEASUREITEM || uMsg == WM_INITMENUPOPUP)) {
		g_pContextMenu2->lpVtbl->HandleMenuMsg(g_pContextMenu2, uMsg, wParam, lParam);
		return 0;
	}

	if (g_pContextMenu3 && uMsg == WM_MENUCHAR) {
		LRESULT result;
		g_pContextMenu3->lpVtbl->HandleMenuMsg2(g_pContextMenu3, uMsg, wParam, lParam, &result);
		return result;
	}

	switch (uMsg) {
		case WM_RBUTTONUP: {
			POINT pt = { LOWORD(lParam), HIWORD(lParam) };
			ClientToScreen(hwnd, &pt);

			// --- デモ用のファイルパス (存在するファイル・フォルダのパスに変更してください) ---
			WCHAR* filePaths[] = {
				L"D:\\test\\ううう"
			};
			int fileCount = sizeof(filePaths) / sizeof(filePaths[0]);

			ShowContextMenu(hwnd, pt, filePaths, fileCount);
			return 0;
		}
		case WM_DESTROY:
			PostQuitMessage(0);
			return 0;

		case WM_PAINT: {
			PAINTSTRUCT ps;
			HDC hdc = BeginPaint(hwnd, &ps);
			FillRect(hdc, &ps.rcPaint, (HBRUSH)(COLOR_WINDOW + 1));
			EndPaint(hwnd, &ps);
			return 0;
		}
	}
	return DefWindowProcW(hwnd, uMsg, wParam, lParam);
}

/**
 * @brief コンテキストメニューの表示と選択されたコマンドの実行
 * @param hwnd オーナーウィンドウのハンドル
 * @param pt スクリーン座標でのメニュー表示位置
 * @param filePaths 対象となるファイルパスの配列
 * @param fileCount ファイル数
 * @return 成功した場合は0、失敗した場合は1
 */
int ShowContextMenu(HWND hwnd, POINT pt, WCHAR** filePaths, int fileCount) {
	IContextMenu* pContextMenu = NULL;
	HRESULT hr = CreateContextMenuForFiles(hwnd, filePaths, fileCount, &pContextMenu);
	if (FAILED(hr)) {
		wprintf(L"CreateContextMenuForFiles failed: 0x%08x\n", hr);
		return 1;
	}

	HMENU hMenu = CreatePopupMenu();
	if (!hMenu) {
		SAFE_RELEASE(pContextMenu);
		return 1;
	}

	// コンテキストメニューにアイテムを問い合わせて構築する
	hr = pContextMenu->lpVtbl->QueryContextMenu(pContextMenu, hMenu, 0, 1, 0x7FFF, CMF_NORMAL | CMF_EXPLORE);
	if (FAILED(hr)) {
		DestroyMenu(hMenu);
		SAFE_RELEASE(pContextMenu);
		return 1;
	}

	// オーナー描画のために IContextMenu2/3 インターフェイスを取得
	pContextMenu->lpVtbl->QueryInterface(pContextMenu, &IID_IContextMenu2, (void**)&g_pContextMenu2);
	pContextMenu->lpVtbl->QueryInterface(pContextMenu, &IID_IContextMenu3, (void**)&g_pContextMenu3);

	// メニューを表示し、ユーザーの選択を待つ
	UINT selectedCmd = TrackPopupMenuEx(hMenu, TPM_RETURNCMD | TPM_NONOTIFY, pt.x, pt.y, hwnd, NULL);

	if (selectedCmd > 0) {
		CMINVOKECOMMANDINFOEX cmi = { sizeof(cmi) };
		//cmi.cbSize = sizeof(cmi);
		cmi.lpVerb = MAKEINTRESOURCEA(selectedCmd - 1);
		cmi.fMask = CMIC_MASK_UNICODE;
		cmi.hwnd = hwnd;
		cmi.lpVerbW = MAKEINTRESOURCEW(selectedCmd - 1);
		cmi.nShow = SW_SHOWNORMAL;

		pContextMenu->lpVtbl->InvokeCommand(pContextMenu, (LPCMINVOKECOMMANDINFO)&cmi);
	}

	// --- 後始末 ---
	SAFE_RELEASE(g_pContextMenu2);
	SAFE_RELEASE(g_pContextMenu3);
	SAFE_RELEASE(pContextMenu);
	DestroyMenu(hMenu);

	return 0;
}

/**
 * @brief 指定されたファイル群のコンテキストメニューインターフェイスを取得する
 * @note  対象となるファイルはすべて同じ親フォルダに存在する必要があります
 * @param hwnd 親ウィンドウのハンドル
 * @param paths 対象となるファイルパスの配列
 * @param fileCount ファイル数
 * @param[out] ppContextMenu 取得した IContextMenu インターフェイスへのポインタ
 * @return 成功した場合は S_OK、それ以外はエラーコード
 */
HRESULT CreateContextMenuForFiles(HWND hwnd, WCHAR** paths, int fileCount, IContextMenu** ppContextMenu) {
	if (fileCount <= 0 || !paths || !ppContextMenu) {
		return E_INVALIDARG;
	}
	*ppContextMenu = NULL;

	// 1. 全てのパスを絶対PIDL (Pointer to an Item ID List) に変換
	LPITEMIDLIST* pidls = (LPITEMIDLIST*)calloc(fileCount, sizeof(LPITEMIDLIST));
	if (!pidls) {
		return E_OUTOFMEMORY;
	}

	HRESULT hr = S_OK;
	for (int i = 0; i < fileCount; ++i) {
		hr = SHParseDisplayName(paths[i], NULL, &pidls[i], 0, NULL);
		if (FAILED(hr)) {
			FreePIDLArray(pidls, i); // ここまでで確保した分を解放
			free(pidls);
			return hr;
		}
	}

	// 2. 最初のファイルから親フォルダの IShellFolder を取得
	IShellFolder* pParentFolder = NULL;
	hr = SHBindToParent(pidls[0], &IID_IShellFolder, (void**)&pParentFolder, NULL);
	if (FAILED(hr)) {
		FreePIDLArray(pidls, fileCount);
		free(pidls);
		return hr;
	}

	// 3. 全ての絶対PIDLから、親から見た子PIDLの配列を作成
	LPCITEMIDLIST* childPidls = (LPCITEMIDLIST*)calloc(fileCount, sizeof(LPCITEMIDLIST));
	if (!childPidls) {
		SAFE_RELEASE(pParentFolder);
		FreePIDLArray(pidls, fileCount);
		free(pidls);
		return E_OUTOFMEMORY;
	}

	for (int i = 0; i < fileCount; ++i) {
		childPidls[i] = ILFindLastID(pidls[i]);
	}

	// 4. IShellFolder から IContextMenu インターフェイスを取得
	hr = pParentFolder->lpVtbl->GetUIObjectOf(pParentFolder, hwnd, fileCount, childPidls,
											  &IID_IContextMenu, NULL, (void**)ppContextMenu);

	// --- 後始末 ---
	free(childPidls);
	SAFE_RELEASE(pParentFolder);
	FreePIDLArray(pidls, fileCount);
	free(pidls);

	return hr;
}

/**
 * @brief PIDLの配列を解放するヘルパー関数
 * @param pidls 解放するPIDLの配列
 * @param count 配列の要素数
 */
void FreePIDLArray(LPITEMIDLIST* pidls, int count) {
	if (!pidls) {
		return;
	}
	for (int i = 0; i < count; ++i) {
		if (pidls[i]) {
			ILFree(pidls[i]);
		}
	}
}

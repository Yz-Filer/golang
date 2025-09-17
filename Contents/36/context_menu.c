// コンパイルコマンド: gcc your_code.c -o your_program.exe -lshlwapi -lole32 -loleaut32 -luuid

#ifndef UNICODE
#define UNICODE
#endif

#include <shlobj.h>
#include <shlwapi.h>
#include <stdbool.h>
#include <stdio.h>
#include <windows.h>

#pragma comment(lib, "shlwapi.lib")

// COMインターフェイスを安全に解放するためのマクロ
#define SAFE_RELEASE(p) if (p) { (p)->lpVtbl->Release(p); (p) = NULL; }

// --- グローバル変数 ---
// IContextMenu2/3 は TrackPopupMenuEx がアクティブな間に発生するオーナー描画等のメッセージを
// ウィンドウプロシージャで処理するために必要。そのため、一時的にグローバル変数として保持する。
//static IContextMenu2* g_pContextMenu2 = NULL;
//static IContextMenu3* g_pContextMenu3 = NULL;

// --- 関数のプロトタイプ宣言 ---
void FreePIDLArray(LPITEMIDLIST* pidls, int count);
HRESULT CreateContextMenuForFiles(HWND hwnd, WCHAR** paths, int fileCount, IContextMenu** ppContextMenu);
bool ShowContextMenu(HWND hwnd, int x, int y, WCHAR** filePaths, int fileCount);
void free_wchars_array(WCHAR** ptr, int count);


/**
 * @brief コンテキストメニューの表示と選択されたコマンドの実行
 * @param hwnd オーナーウィンドウのハンドル
 * @param pt スクリーン座標でのメニュー表示位置
 * @param filePaths 対象となるファイルパスの配列
 * @param fileCount ファイル数
 * @return 成功した場合はtrue、失敗した場合はfalse
 */
bool ShowContextMenu(HWND hwnd, int x, int y, WCHAR** filePaths, int fileCount) {
	IContextMenu* pContextMenu = NULL;
	HRESULT hr = CreateContextMenuForFiles(hwnd, filePaths, fileCount, &pContextMenu);
	if (FAILED(hr)) {
		wprintf(L"CreateContextMenuForFiles failed: 0x%08x\n", hr);
		return false;
	}

	HMENU hMenu = CreatePopupMenu();
	if (!hMenu) {
		SAFE_RELEASE(pContextMenu);
		return false;
	}

	// コンテキストメニューにアイテムを問い合わせて構築する
	hr = pContextMenu->lpVtbl->QueryContextMenu(pContextMenu, hMenu, 0, 1, 0x7FFF, CMF_NORMAL | CMF_EXPLORE);
	if (FAILED(hr)) {
		DestroyMenu(hMenu);
		SAFE_RELEASE(pContextMenu);
		return false;
	}

	// オーナー描画のために IContextMenu2/3 インターフェイスを取得
//	pContextMenu->lpVtbl->QueryInterface(pContextMenu, &IID_IContextMenu2, (void**)&g_pContextMenu2);
//	pContextMenu->lpVtbl->QueryInterface(pContextMenu, &IID_IContextMenu3, (void**)&g_pContextMenu3);

	// メニューを表示し、ユーザーの選択を待つ
	UINT selectedCmd = TrackPopupMenuEx(hMenu, TPM_RETURNCMD | TPM_NONOTIFY, x, y, hwnd, NULL);

	if (selectedCmd > 0) {
		CMINVOKECOMMANDINFOEX cmi = { sizeof(cmi) };
		cmi.lpVerb = MAKEINTRESOURCEA(selectedCmd - 1);
		cmi.fMask = CMIC_MASK_UNICODE;
		cmi.hwnd = hwnd;
		cmi.lpVerbW = MAKEINTRESOURCEW(selectedCmd - 1);
		cmi.nShow = SW_SHOWNORMAL;

		pContextMenu->lpVtbl->InvokeCommand(pContextMenu, (LPCMINVOKECOMMANDINFO)&cmi);
	}

	// --- 後始末 ---
//	SAFE_RELEASE(g_pContextMenu2);
//	SAFE_RELEASE(g_pContextMenu3);
	SAFE_RELEASE(pContextMenu);
	DestroyMenu(hMenu);

	return true;
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


/**
 * @brief ファイルパス用に確保したメモリを解放するヘルパー関数
 * @param ptr 解放するPIDLの配列
 * @param count 配列の要素数
 */
void free_wchars_array(WCHAR** ptr, int count) {
    if (ptr == NULL) {
        return;
    }
    for (int i = 0; i < count; i++) {
        free(ptr[i]);
    }
    free(ptr);
}

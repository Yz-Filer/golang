// 付箋アプリ
package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//-----------------------------------------------------------------------------
// ファイル入出力
//-----------------------------------------------------------------------------

// ファイルの存在確認
func FileExists(filename string) (bool, error) {
	f, err := os.Stat(filename)
	if err != nil {
		// ファイルが存在しない場合
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		// その他のエラーの場合
		return false, err
	}
	
	// ファイルが存在した場合、ディレクトリ判定
	return !f.IsDir(), nil
}

// 既存のデータファイルをバックアップ
func BackupDataFile(filename, bakFilename string) error {
	// データファイルの存在確認
	ok, err := FileExists(filename)
	if err != nil {
		return err
	}

	// データファイルが存在しなかった場合
	if !ok {
		return nil
	}

	// バックアップファイルの存在確認
	ok, err = FileExists(bakFilename)
	if err != nil {
		return err
	}
	
	// バックアップファイルが存在していたら削除
	if ok {
		err := os.Remove(bakFilename)
		if err != nil {
			return err
		}
	}
	
	// データファイルをバックアップファイル名にリネーム
	return os.Rename(filename, bakFilename)
}

// マップをファイルに保存する
func SaveStickyMap(filename string) error {
	// 既存のデータファイルをバックアップ
	err := BackupDataFile(filename, filename + ".bak")
	if err != nil {
		return err
	}

	// データファイルを作成し、データを保存
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(StickyMap)
	if err != nil {
		return err
	}

	return nil
}

// ファイルからマップを読み込む
func LoadStickyMap(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	StickyMap = make(map[string]StickyStr)
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&StickyMap)
	if err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// ステータスアイコン用のメニューを作成
//-----------------------------------------------------------------------------
func buildMenu(parent *gtk.ApplicationWindow, stIcon *gtk.StatusIcon, builder *gtk.Builder) (*gtk.Menu, error) {
	if parent == nil {
		return nil, fmt.Errorf("Parent window not found.")
	}

	// gladeからメニューを取得
	menu1, _, err := GetObjFromGlade[*gtk.Menu](builder, "", "STATUSICON_MENU")
	if err != nil {
		return nil, err
	}
	
	// gladeからmenuItem1を取得
	menuItem1, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_SHOW")
	if err != nil {
		return nil, err
	}
	
	// gladeからmenuItem2を取得
	menuItem2, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_NEW")
	if err != nil {
		return nil, err
	}
	
	// gladeからmenuItem3を取得
	menuItem3, _, err := GetObjFromGlade[*gtk.MenuItem](builder, "", "MENUITEM_QUIT")
	if err != nil {
		return nil, err
	}
	
	// menuItem1選択時に親ウィンドウをタスクトレイから出す
	menuItem1.Connect("activate", func(){
		stIcon.SetVisible(false)
		parent.SetPosition(gtk.WIN_POS_NONE)
		parent.Present()
	})
	
	// menuItem2選択時に付箋新規作成ボタンシグナル発行
	menuItem2.Connect("activate", func(){
		_, err := parent.Emit("edit_window_new", glib.TYPE_POINTER, EDIT_NEW, "")
		if err != nil {
			ShowErrorDialog(parent, err)
		}
	})
	
	// menuItem3選択時にアプリを終了
	menuItem3.Connect("activate", func(){
		stIcon.SetVisible(false)
		menu1.Destroy()

		// ファイルへの書き込み
		err = SaveStickyMap(ConfFile)
		if err != nil {
			ShowErrorDialog(parent, err)
		}
		
		// アプリケーションの終了
		MyApplication.Quit()
	})
	
	return menu1, nil
}

//-----------------------------------------------------------------------------
// メイン
//-----------------------------------------------------------------------------
func main() {
	const appID = "org.example.myapp"
	var mWindow *gtk.ApplicationWindow
	var builder *gtk.Builder
	var err error
	
	// 実行ファイルのフルパス取得
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to retrieve executable path: ", err)
	}
	
	// 保存ファイル名の作成
	fileName := filepath.Base(exePath)
	extension := filepath.Ext(fileName)
	fileNameWithoutExt := strings.TrimSuffix(fileName, extension)
	ConfFile = filepath.Join(filepath.Dir(exePath), fileNameWithoutExt + ".dat")
	
	
	// 多重起動防止
	mutex, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(appID))
	if err != nil {
		log.Fatal("Application is already running.")
	}
	defer windows.CloseHandle(mutex)
	
	// 新しいアプリケーションの作成
	MyApplication, err = gtk.ApplicationNew(appID, glib.APPLICATION_NON_UNIQUE)
	if err != nil {
		log.Fatal("Could not create application: ", err)
	}

	///////////////////////////////////////////////////////////////////////////
	// アプリケーション起動時のイベント
	///////////////////////////////////////////////////////////////////////////
	MyApplication.Connect("startup", func() {
		log.Println("application startup")	
		
		// ファイルの読み込み
		err = LoadStickyMap(ConfFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				StickyMap = make(map[string]StickyStr)
			} else {
				log.Fatal("Failed to load file: ", err)
			}
		}
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーション アクティブ時のイベント
	///////////////////////////////////////////////////////////////////////////
	MyApplication.Connect("activate", func() {
		// gladeからウィンドウを取得
		mWindow, builder, err = GetObjFromGlade[*gtk.ApplicationWindow](nil, WindowGlade, "MAIN_WINDOW")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからTreeViewを取得
		treeView, _, err := GetObjFromGlade[*gtk.TreeView](builder, "", "TREEVIEW")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからListStoreを取得
		listStore, _, err := GetObjFromGlade[*gtk.ListStore](builder, "", "LISTSTORE")
		if err != nil {
			log.Fatal(err)
		}

		// gladeから付箋新規作成ボタンを取得
		mwNewButton, _, err := GetObjFromGlade[*gtk.Button](builder, "", "MW_NEW_BUTTON")
		if err != nil {
			log.Fatal(err)
		}

		// リソースからアプリケーションのアイコンを取得
		iconPixbuf, err := gdk.PixbufNewFromDataOnly(icon)
		if err != nil {
			log.Fatal("Could not create pixbuf from bytes: ", err)
		}
		defer iconPixbuf.Unref()
		
		// ウィンドウアイコンを設定
		mWindow.SetIcon(iconPixbuf)
		
		// タスクトレイアイコンを設定
		statusIcon, err := gtk.StatusIconNewFromPixbuf(iconPixbuf)
		if err != nil {
			log.Fatal("Could not create status icon: ", err)
		}
		defer statusIcon.Unref()
		
		// タスクトレイアイコン用のメニューを作成
		menu, err := buildMenu(mWindow, statusIcon, builder)
		if err != nil {
			log.Fatal("Could not create status menu: ", err)
		}
		
		// ソートを作成し、TreeViewに追加
		listSort, err := gtk.TreeModelSortNew(listStore)
		if err != nil {
			log.Fatal(err)
		}
		treeView.SetModel(listSort)
		
		// ウィンドウのプロパティを設定（必須ではない）
		mWindow.SetPosition(gtk.WIN_POS_MOUSE)
		mWindow.Resize(400, 300)


		// 起動時にファイルから読み込んだ付箋を表示
		for key, value := range StickyMap {
			// TreeViewに追加
			iter := listStore.Append()
			err = listStore.Set(iter, []int{0, 1}, []interface{}{key, strings.ReplaceAll(value.Note, "\n", " ")})
			if err != nil {
				log.Fatal("Could not add sticky note: ", err)
			}

			// 付箋ウィンドウを表示
			err = StickyNoteWindowShow(mWindow, key)
			if err != nil {
				log.Fatal("Could not create sticky note window: ", err)
			}
		}

		//-----------------------------------------------------------
		// カスタムシグナル作成
		//-----------------------------------------------------------

		// カスタムシグナル（編集画面表示）を作成
		_, err = glib.SignalNewV("edit_window_new", glib.TYPE_POINTER, 2, glib.TYPE_INT, glib.TYPE_STRING)
		if err != nil {
			log.Fatal("Could not create signal: ", err)
		}
		
		// カスタムシグナル（付箋表示）を作成
		_, err = glib.SignalNewV("sticky_note_show", glib.TYPE_POINTER, 2, glib.TYPE_INT, glib.TYPE_STRING)
		if err != nil {
			log.Fatal("Could not create signal: ", err)
		}
		
		// カスタムシグナル（付箋破棄）を作成
		_, err = glib.SignalNewV("sticky_note_destroy", glib.TYPE_POINTER, 1, glib.TYPE_STRING)
		if err != nil {
			log.Fatal("Could not create signal: ", err)
		}
		
		//*********************************************************************
		// メインウィンドウのシグナル処理
		//*********************************************************************

		//-----------------------------------------------------------
		// 付箋新規作成ボタン押下時の処理（シグナル発行）
		//-----------------------------------------------------------
		mwNewButton.Connect("clicked", func() {
			_, err := mWindow.Emit("edit_window_new", glib.TYPE_POINTER, EDIT_NEW, "")
			if err != nil {
				ShowErrorDialog(mWindow, err)
			}
		})
		
		//-----------------------------------------------------------
		// メインウィンドウフォーカス時に、付箋を前面に移動
		//-----------------------------------------------------------
		mWindow.Connect("focus-in-event", func() {
			for _, value := range StickyMap {
				// 付箋ウィンドウを最前面に移動
				value.stickyWindow.SetKeepAbove(true)
				value.stickyWindow.SetKeepAbove(false)
			}
			mWindow.SetKeepAbove(true)
			mWindow.SetKeepAbove(false)
		})
		
		//-----------------------------------------------------------
		// カスタムシグナル（編集画面表示）受信時の処理
		//-----------------------------------------------------------
		mWindow.Connect("edit_window_new", func(win *gtk.ApplicationWindow, mode int, key string) {
			// 新規の場合、キーを払い出し
			if mode == EDIT_NEW {
				key = time.Now().Format("2006/01/02 15:04:05.000")
			}
			
			// 編集画面を作成
			eWindow, err := EditWindowNew(mWindow, key)
			if err != nil {
				ShowErrorDialog(mWindow, err)
				return
			}
			
			// 編集画面を最前面に表示
			eWindow.ShowAll()
			eWindow.SetKeepAbove(true)
			eWindow.SetKeepAbove(false)
		})
		
		//-----------------------------------------------------------
		// カスタムシグナル（付箋表示）受信時の処理
		//-----------------------------------------------------------
		mWindow.Connect("sticky_note_show", func(win *gtk.ApplicationWindow, mode int, key string) {
			switch mode {
				case STICKY_NEW:			// 付箋新規作成時
					// TreeViewに追加
					iter := listStore.Append()
					err = listStore.Set(iter, []int{0, 1}, []interface{}{key, strings.ReplaceAll(StickyMap[key].Note, "\n", " ")})
					if err != nil {
						ShowErrorDialog(mWindow, fmt.Errorf("Could not add sticky note: %w", err))
						return
					}
				case STICKY_UPDATED:		// 付箋更新時
					// TreeViewを更新
					iter, err := GetIterFromKey(listStore, key)
					if err != nil {
						ShowErrorDialog(mWindow, fmt.Errorf("Could not retrieve iterator: %w", err))
						return
					}
					err = listStore.SetValue(iter, 1, strings.ReplaceAll(StickyMap[key].Note, "\n", " "))
					if err != nil {
						ShowErrorDialog(mWindow, fmt.Errorf("Failed to update the value: %w", err))
						return
					}
				default:					// それ以外は何もしない
			}
			
			// 付箋ウィンドウを表示
			err = StickyNoteWindowShow(mWindow, key)
			if err != nil {
				ShowErrorDialog(mWindow, err)
				return
			}
		})
		
		//-----------------------------------------------------------
		// カスタムシグナル（付箋破棄）受信時の処理
		//-----------------------------------------------------------
		mWindow.Connect("sticky_note_destroy", func(win *gtk.ApplicationWindow, key string) {
			// TreeViewから削除
			iter, err := GetIterFromKey(listStore, key)
			if err != nil {
				ShowErrorDialog(mWindow, fmt.Errorf("Could not retrieve iterator: %w", err))
				return
			}
			listStore.Remove(iter)
			
			// 付箋ウィンドウを破棄
			StickyMap[key].stickyWindow.Destroy()
			
			// マップから削除
			delete(StickyMap, key)
		})
		
		//-----------------------------------------------------------
		// 閉じるボタンが押されたら、タスクトレイに格納
		//-----------------------------------------------------------
		mWindow.Connect("delete-event", func(win *gtk.ApplicationWindow, event *gdk.Event) bool {
			log.Println("ウィンドウをタスクトレイに格納しました")
			
			mWindow.Hide()
			statusIcon.SetVisible(true)
			
			// クローズ処理をキャンセル
			return true
		})
		
		//*********************************************************************
		// タスクトレイのシグナル処理
		//*********************************************************************
		
		//-----------------------------------------------------------
		// タスクトレイアイコン右クリック時にメニューを表示
		// ※メニューは項目を選択しないと消えない
		// ※アイコンの位置にメニューが表示されない
		//-----------------------------------------------------------
		statusIcon.Connect("popup-menu", func(statusIcon *gtk.StatusIcon, button uint, activateTime uint) {
			menu.PopupAtStatusIcon(statusIcon, gdk.Button(button), uint32(activateTime))
		})

		//-----------------------------------------------------------
		// タスクトレイアイコンをクリックしたらウィンドウを再表示
		//-----------------------------------------------------------
		statusIcon.Connect("activate", func() {
			statusIcon.SetVisible(false)
			mWindow.SetPosition(gtk.WIN_POS_NONE)
			mWindow.Present()
		})
		
		// 起動時はタスクトレイアイコンを非表示にする
		statusIcon.SetVisible(false)
		
		
		//*********************************************************************
		// TreeViewのシグナル処理
		//*********************************************************************
		
		//-----------------------------------------------------------
		// TreeViewの行をダブルクリックした時の処理
		//-----------------------------------------------------------
		treeView.Connect("row-activated", func(tv *gtk.TreeView, path *gtk.TreePath, column *gtk.TreeViewColumn) {
			// TreeViewのpathをListStoreのパスに変換
			path1 := listSort.ConvertPathToChildPath(path)
			
			// ListStoreのIterを取得
			iter, err := listStore.GetIter(path1)
			if err != nil {
				ShowErrorDialog(mWindow, fmt.Errorf("Failed to get liststore iter: %w", err))
				return
			}
			
			// 値を取得
			key, err := GetListStoreValue[string] (listStore, iter, 0)
			if err != nil {
				ShowErrorDialog(mWindow, fmt.Errorf("Failed to retrieve the tree value: %w", err))
				return
			}
			
			// 更新モードで編集ウィンドウを表示するシグナルを送信
			_, err = mWindow.Emit("edit_window_new", glib.TYPE_POINTER, EDIT_UPDATED, key)
			if err != nil {
				ShowErrorDialog(mWindow, err)
				return
			}
			
			// 付箋ウィンドウを破棄
			StickyMap[key].stickyWindow.Destroy()
		})

		//-----------------------------------------------------------
		// TreeViewの行を右クリックした時の処理
		//-----------------------------------------------------------
		treeView.Connect("button-press-event", func(treeView *gtk.TreeView, event *gdk.Event) bool {
			e := gdk.EventButtonNewFromEvent(event)
			if e.Button() == gdk.BUTTON_SECONDARY {
				// クリックされた行を特定
				x := int(e.X())
				y := int(e.Y())
				path, _, _, _, ok := treeView.GetPathAtPos(x, y)
				if !ok {
					// 座標から行が特定できない場合は何もしない
					return false
				}
				
				// カーソルを右クリックした行へ移動
				treeView.SetCursor(path, nil, false)
				
				// TreeViewのpathをListStoreのpathに変換
				path1 := listSort.ConvertPathToChildPath(path)
				
				// ListStoreのIterを取得
				iter, err := listStore.GetIter(path1)
				if err != nil {
					ShowErrorDialog(mWindow, fmt.Errorf("Failed to get liststore iter: %w", err))
					return false
				}
				
				// 値を取得
				key, err := GetListStoreValue[string] (listStore, iter, 0)
				if err != nil {
					ShowErrorDialog(mWindow, fmt.Errorf("Failed to retrieve the tree value: %w", err))
					return false
				}
				
				// 付箋用メニューを作成
				stickyMenu, err := buildStickyMenu(mWindow, nil, key)
				if err != nil {
					ShowErrorDialog(mWindow, fmt.Errorf("Could not create sticky menu: %w", err))
				}
				stickyMenu.PopupAtPointer(event)
				
				return true
			}
			return false
		})




		// アプリケーションを設定
		mWindow.SetApplication(MyApplication)

		// ウィンドウの表示
		mWindow.ShowAll()
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーション終了時のイベント
	///////////////////////////////////////////////////////////////////////////
	MyApplication.Connect("shutdown", func() {
		log.Println("application shutdown")
		
		fmt.Println("---- map items ----")
		for key, value := range StickyMap {
			fmt.Println(key, strings.ReplaceAll(value.Note, "\n", " "))
		}
		fmt.Println("-------------------")
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	os.Exit(MyApplication.Run(os.Args))
}


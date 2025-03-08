// リストビュー
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

//go:embed glade/11_MainWindow.glade
var window1Glade string

//go:embed resources/icon.ico
var icon []byte

var application *gtk.Application


//-----------------------------------------------------------------------------
// pathからlistStoreのiterを取得
//-----------------------------------------------------------------------------
func GetListStoreIterFromPath(listSort *gtk.TreeModelSort, listFilter *gtk.TreeModelFilter, listStore *gtk.ListStore, path *gtk.TreePath) (*gtk.TreeIter, error) {
	// listSortで表示されてるパスからlistFilterのパスに変換
	path1 := listSort.ConvertPathToChildPath(path)

	// listFilterで表示されてるパスからlistStoreのパスに変換
	path2 := listFilter.ConvertPathToChildPath(path1)

	// イテレータを返却
	return listStore.GetIter(path2)
}

//-----------------------------------------------------------------------------
// 選択行からlistStoreのiterを取得
//-----------------------------------------------------------------------------
func GetListStoreIterFromSelection(treeView *gtk.TreeView, listSort *gtk.TreeModelSort, listFilter *gtk.TreeModelFilter, listStore *gtk.ListStore) (*gtk.TreeIter, error) {
	// TreeViewの選択されている行を取得
	selection, err := treeView.GetSelection()
	if err != nil {
		return nil, err
	}
	
	_, iter1, ok := selection.GetSelected()
	if !ok {
		return nil, fmt.Errorf("Unable to get the selected item.")
	}
	
	// listSortで表示されてるイテレータからlistFilterのイテレータに変換
	iter2 := listSort.ConvertIterToChildIter(iter1)
	
	// listFilterで表示されてるイテレータからlistStoreのイテレータに変換
	iter := listFilter.ConvertIterToChildIter(iter2)
	
	if iter == nil {
		return nil, fmt.Errorf("Iter is null.")
	}
	
	return iter, nil
}

//-----------------------------------------------------------------------------
// ListStoreに格納されてる値を取得
//-----------------------------------------------------------------------------
func GetListStoreValue[T any] (iModel gtk.ITreeModel, iter *gtk.TreeIter, id int) (T, error) {

	model := iModel.ToTreeModel()

	// 値を取得
	colVal, err := model.GetValue(iter, id)
	if err != nil {
		return *new(T), err
	}
	
	// 値をgolang形式に変換
	col, err := colVal.GoValue()
	if err != nil {
		return *new(T), err
	}
	
	// interfaceをT型に変換
	ret, ok := col.(T)
	if !ok {
		return *new(T), fmt.Errorf("type assertion failed")
	}
	return ret, nil
}

//-----------------------------------------------------------------------------
// メイン
//-----------------------------------------------------------------------------
func main() {
	const appID = "org.example.myapp"
	var window1 *gtk.ApplicationWindow
	var builder *gtk.Builder
	var err error
	
	///////////////////////////////////////////////////////////////////////////
	// 新しいアプリケーションの作成
	///////////////////////////////////////////////////////////////////////////
	application, err = gtk.ApplicationNew(appID, glib.APPLICATION_NON_UNIQUE)
	if err != nil {
		log.Fatal("Could not create application:", err)
	}

	///////////////////////////////////////////////////////////////////////////
	// アプリケーション起動時のイベント（必須ではない）
	///////////////////////////////////////////////////////////////////////////
	application.Connect("startup", func() {
		log.Println("application startup")	
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーション アクティブ時のイベント
	///////////////////////////////////////////////////////////////////////////
	application.Connect("activate", func() {
		// gladeからウィンドウを取得
		window1, builder, err = GetObjFromGlade[*gtk.ApplicationWindow](nil, window1Glade, "MAIN_WINDOW")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからTreeViewを取得
		treeView, _, err := GetObjFromGlade[*gtk.TreeView](builder, window1Glade, "TREEVIEW")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからListStoreを取得
		listStore, _, err := GetObjFromGlade[*gtk.ListStore](builder, window1Glade, "LISTSTORE")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeから更新ボタンを取得
		btnUpdate, _, err := GetObjFromGlade[*gtk.Button](builder, window1Glade, "BUTTON_UPDATE")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeから削除ボタンを取得
		btnRemove, _, err := GetObjFromGlade[*gtk.Button](builder, window1Glade, "BUTTON_REMOVE")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeから3を探すボタンを取得
		btnSearch, _, err := GetObjFromGlade[*gtk.Button](builder, window1Glade, "BUTTON_SEARCH")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからフィルタボタンを取得
		btnFilter, _, err := GetObjFromGlade[*gtk.Button](builder, window1Glade, "BUTTON_FILTER")
		if err != nil {
			log.Fatal(err)
		}
		
		// gladeからフィルタ解除ボタンを取得
		btnAll, _, err := GetObjFromGlade[*gtk.Button](builder, window1Glade, "BUTTON_ALL")
		if err != nil {
			log.Fatal(err)
		}
		
		// リソースからアプリケーションのアイコンを設定
		iconPixbuf, err := gdk.PixbufNewFromDataOnly(icon)
		if err != nil {
			log.Fatal("Could not create pixbuf from bytes:", err)
		}
		defer iconPixbuf.Unref()
		
		// ウィンドウにアイコンを設定
		window1.SetIcon(iconPixbuf)
		
		// ウィンドウのプロパティを設定（必須ではない）
		window1.SetPosition(gtk.WIN_POS_MOUSE)
		
		
		//-----------------------------------------------------------
		// フィルタを作成
		//-----------------------------------------------------------
		filterON := false
		listFilter, err := listStore.FilterNew(nil)
		if err != nil {
			log.Fatal("Failed to create the list filter: ", err)
		}
		
		// フィルタ関数（trueなら表示。falseなら非表示）
		listFilter.SetVisibleFunc(func(model *gtk.TreeModel, iter *gtk.TreeIter) bool {
			// フィルタがOFFの場合は全行出力
			if !filterON {
				return true
			}
			
			// 値を取得
			col1, err := GetListStoreValue[int] (model, iter, 0)
			if err != nil {
				ShowErrorDialog(window1, fmt.Errorf("Failed to retrieve the tree value: %w", err))
				return false
			}
			
			// 偶数は出力、奇数は出力しない
			if col1 % 2 == 0 {
				return true
			} else {
				return false
			}
		})
		
		//-----------------------------------------------------------
		// ソートを作成
		//-----------------------------------------------------------
		listSort, err := gtk.TreeModelSortNew(listFilter)
		if err != nil {
			log.Fatal(err)
		}

		// TreeViewに追加
		treeView.SetModel(listSort)
		
		
		//-----------------------------------------------------------
		// リストにデータを追加
		//-----------------------------------------------------------
		iter := listStore.Append()
		err = listStore.Set(iter, []int{0, 1}, []interface{}{1, "eeeeeeeeeeeeeee"})
		if err != nil {log.Fatal(err)}
		iter = listStore.Append()
		err = listStore.Set(iter, []int{0, 1}, []interface{}{2, "ddddddddddddddd"})
		if err != nil {log.Fatal(err)}
		iter = listStore.Append()
		err = listStore.Set(iter, []int{0, 1}, []interface{}{3, "ccccccccccccccc"})
		if err != nil {log.Fatal(err)}
		iter = listStore.Append()
		err = listStore.Set(iter, []int{0, 1}, []interface{}{4, "bbbbbbbbbbbbbbb"})
		if err != nil {log.Fatal(err)}
		iter = listStore.Append()
		err = listStore.Set(iter, []int{0, 1}, []interface{}{5, "aaaaaaaaaaaaaaa"})
		if err != nil {log.Fatal(err)}


		//-----------------------------------------------------------
		// TreeViewの行をダブルクリックした時の処理
		//-----------------------------------------------------------
		treeView.Connect("row-activated", func(tv *gtk.TreeView, path *gtk.TreePath, column *gtk.TreeViewColumn) {
			// pathからlistStoreのiterを取得
			iter, err := GetListStoreIterFromPath(listSort, listFilter, listStore, path)
			if err != nil {
				ShowErrorDialog(window1, fmt.Errorf("Failed to get iter: %w", err))
				return
			}
			
			// 値を取得
			col1, err := GetListStoreValue[int] (listStore, iter, 0)
			if err != nil {
				ShowErrorDialog(window1, fmt.Errorf("Failed to retrieve the tree value: %w", err))
				return
			}

			// 値を取得
			col2, err := GetListStoreValue[string] (listStore, iter, 1)
			if err != nil {
				ShowErrorDialog(window1, fmt.Errorf("Failed to retrieve the tree value: %w", err))
				return
			}
			
			log.Printf("col1: %d, col2: %s\n", col1, col2)
		})
		
		//-----------------------------------------------------------
		// 更新ボタンを押した時の処理
		//-----------------------------------------------------------
		btnUpdate.Connect("clicked", func() {
			
			// 選択行からlistStoreのiterを取得
			iter, err := GetListStoreIterFromSelection(treeView, listSort, listFilter, listStore)
			if err != nil {
				ShowErrorDialog(window1, fmt.Errorf("Failed to get iter: %w", err))
				return
			}
/*
			// 値を取得
			// SetValueでは使わない
			col1, err := GetListStoreValue[int] (listStore, iter, 0)
			if err != nil {
				ShowErrorDialog(window1, fmt.Errorf("Failed to retrieve the tree value: %w", err))
				return
			}
*/
			// 値を取得
			col2, err := GetListStoreValue[string] (listStore, iter, 1)
			if err != nil {
				ShowErrorDialog(window1, fmt.Errorf("Failed to retrieve the tree value: %w", err))
				return
			}
			
			// 値を更新して上書き
			// Setは1行分、SetValueは1項目分
//			err = listStore.Set(iter, []int{0, 1}, []interface{}{col1, col2 + "*"})
			err = listStore.SetValue(iter, 1, col2 + "*")
			if err != nil {
				ShowErrorDialog(window1, fmt.Errorf("Failed to update the value: %w", err))
				return
			}
		})
		
		//-----------------------------------------------------------
		// 削除ボタンを押した時の処理
		//-----------------------------------------------------------
		btnRemove.Connect("clicked", func() {
			
			// 選択行からlistStoreのiterを取得
			iter, err := GetListStoreIterFromSelection(treeView, listSort, listFilter, listStore)
			if err != nil {
				ShowErrorDialog(window1, fmt.Errorf("Failed to get iter: %w", err))
				return
			}
			
			// 行を削除
			listStore.Remove(iter)
		})
		
		//-----------------------------------------------------------
		// 「3を探す」ボタンを押した時の処理
		//-----------------------------------------------------------
		btnSearch.Connect("clicked", func() {
			// listStore内をループして探す
			listStore.ForEach(func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
				// 値を取得
				col1, err := GetListStoreValue[int] (model, iter, 0)
				if err != nil {
					ShowErrorDialog(window1, fmt.Errorf("Failed to retrieve the tree value: %w", err))
					return true
				}
				
				if col1 == 3 {
					// listStoreのパスからlistFilterのパスに変換
					path1 := listFilter.ConvertChildPathToPath(path)
					
					// フィルタされておらずTreeviewに表示されてる場合
					if path1 != nil {
						// listFilterのパスからlistSortのパスに変換
						path2 := listSort.ConvertChildPathToPath(path1)
						
						// カーソルを移動
						treeView.SetCursor(path2, nil, false)
					}
					// 検索を終了する
					return true
				}
				
				// 検索を続ける
				return false
			})
		})
		
		//-----------------------------------------------------------
		// フィルタボタンを押した時の処理
		//-----------------------------------------------------------
		btnFilter.Connect("clicked", func() {
			filterON = true
			listFilter.Refilter()
		})
		
		//-----------------------------------------------------------
		// フィルタ解除ボタンを押した時の処理
		//-----------------------------------------------------------
		btnAll.Connect("clicked", func() {
			filterON = false
			listFilter.Refilter()
		})




		// アプリケーションを設定
		window1.SetApplication(application)

		// ウィンドウの表示
		window1.ShowAll()
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーション終了時のイベント（必須ではない）
	///////////////////////////////////////////////////////////////////////////
	application.Connect("shutdown", func() {
		log.Println("application shutdown")
	})

	///////////////////////////////////////////////////////////////////////////
	// アプリケーションの実行
	///////////////////////////////////////////////////////////////////////////
	os.Exit(application.Run(os.Args))
}


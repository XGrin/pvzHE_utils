package main

import (
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var auto_save = false
var backup_list = []string{}
var select_backup = ""

// pvz窗口
var pvz = &pvzWindow{
	Handle:        0,
	Pid:           0,
	ProcessHandle: 0,
	memoryLock:    make(chan struct{}, 1),
	title:         "",
}

func main() {
	// 初始化操作
	// 判断当前目录下是否存在backup目录，如果不存在则创建
	backup_exist, _ := PathExists("backup")
	if !backup_exist {
		os.Mkdir("backup", os.ModePerm)
	}

	// 创建一个app
	app := app.New()
	w := app.NewWindow("pvzHE utils")
	// w.Resize(fyne.NewSize(200, 200))
	auto_save_checkbox := widget.NewCheck("Auto Save", func(b bool) {
		auto_save = b
	})
	backup_select := widget.NewSelect(backup_list, func(s string) {
		select_backup = s
	})
	recover_button := widget.NewButton("recover", func() {
		// 恢复存档
		// 将选中的备份文件夹下的文件拷贝到C:\ProgramData\PopCap Games\PlantsVsZombies\pvzHE\yourdata
		err := CopyDir("backup\\"+select_backup, "C:\\ProgramData\\PopCap Games\\PlantsVsZombies\\pvzHE\\yourdata")
		if err != nil {
			// 如果出现错误则弹出错误提示
			dialog.NewInformation("Error", err.Error(), w).Show()
		} else {
			// 恢复成功则弹出成功提示
			dialog.NewInformation("Success", "Recover success", w).Show()
		}
	})
	recover_button.Disable()

	info_label := widget.NewLabel("Please close the game before recovering.")

	// 判断是否以管理员权限运行
	admin_status, _ := IsAdmin()
	if !admin_status {
		w.SetContent(container.NewVBox(
			widget.NewLabel("Please run as administrator."),
			widget.NewButton("Exit", func() {
				app.Quit()
			}),
		))
	} else {
		w.SetContent(container.NewVBox(
			auto_save_checkbox, backup_select, recover_button, info_label,
		))
	}

	// 开启携程进行自动保存操作，每30s保存一次
	go func() {
		for {
			if auto_save {
				// 保存操作
				// 判断游戏界面是否在游戏中
				ui := pvz.GetGameUI()
				if ui == 3 {
					// 修复保存后音乐暂停的问题
					// 修改内存
					pvz.WriteMemory(ToBytes(106), 2, 0x408d4b)

					// 调用游戏保存
					pvz.CallSave()

					// 修改内存
					pvz.WriteMemory(ToBytes(362), 2, 0x408d4b)

					// 创建以当前时间为文件名的备份文件夹
					backup_name := time.Now().Format("2006.01.02 15-04-05")
					backup_dir := "backup/" + backup_name
					err := os.Mkdir(backup_dir, os.ModePerm)
					if err != nil {
						log.Println(err)
					}
					// 拷贝C:\ProgramData\PopCap Games\PlantsVsZombies\pvzHE\yourdata这个文件夹到备份文件夹
					CopyDir("C:\\ProgramData\\PopCap Games\\PlantsVsZombies\\pvzHE\\yourdata", backup_dir)
				}
			}
			// 判断备份文件夹下的文件数量，如果超过10个则删除至10个
			// 获取备份文件夹下的所有文件
			backup_files, _ := os.ReadDir("backup")
			// 如果文件数量超过10个则删除多余的文件
			if len(backup_files) > 10 {
				for i := 0; i < len(backup_files)-10; i++ {
					os.RemoveAll("backup/" + backup_files[i].Name())
				}
			}
			// 休眠30s
			time.Sleep(30 * time.Second)
		}
	}()

	// 开启携程监测状态,1s更新一次
	go func() {
		for {
			// 读取backup目录下的所有文件夹，存入backup_list,更新backup_select
			backup_files, _ := os.ReadDir("backup")
			backup_list = []string{}
			for _, file := range backup_files {
				backup_list = append(backup_list, file.Name())
			}
			for i, j := 0, len(backup_list)-1; i < j; i, j = i+1, j-1 {
				backup_list[i], backup_list[j] = backup_list[j], backup_list[i]
			}
			backup_select.SetOptions(backup_list)

			// 判断程序是否还在运行
			is_running := CheckWindowTitle("植物大战僵尸杂交版")
			if is_running {
				auto_save_checkbox.Enable()
			} else {
				auto_save_checkbox.Disable()
				auto_save_checkbox.SetChecked(false)
			}
			// 只有在游戏未运行且选中了备份文件夹才能恢复
			if select_backup != "" {
				if !is_running {
					recover_button.Enable()
				} else {
					ui := pvz.GetGameUI()
					if ui != 3 && ui != 4 && ui != 2 {
						recover_button.Enable()
					} else {
						recover_button.Disable()
					}
				}
			} else {
				recover_button.Disable()
			}

			if is_running {
				if !pvz.IsValid() {
					pvz.Handle = FindWindow("MainWindow", pvz.title)
					if pvz.Handle != 0 {
						GetWindowThreadProcessId(pvz.Handle, &pvz.Pid)
						pvz.ProcessHandle = OpenProcess(PROCESS_ALL_ACCESS, 0, pvz.Pid)
					}
				}
			}

			// 休眠0.5s
			time.Sleep(500 * time.Millisecond)
		}
	}()

	w.Resize(fyne.NewSize(300, 200))
	w.ShowAndRun()
}

package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()

	// MyTheme 实现了 fyne.Theme 接口

	// Color 为 UI 元素返回一个自定义颜色

	a.Settings().SetTheme(&MyTheme{}) // 自动跟随系统亮/暗

	w := a.NewWindow("EICLING小助手")
	// w.SetIcon()
	w.Resize(fyne.NewSize(120, 500))

	// ---------- tab1：玲珑包检测 ----------
	tab1 := makeTab1CheckLLPkgs()

	// ---------- tab2：Cache 大小 ----------
	cacheDirs := []string{
		filepath.Join(os.Getenv("HOME"), ".cache/linglong-builder"),
		filepath.Join(os.Getenv("HOME"), ".cache/linglong-pica"),
		filepath.Join(os.Getenv("HOME"), ".cache/linglong-pica-flathub"),
	}
	// grid2 := container.NewGridWithColumns(2)
	// grid2.Resize(fyne.NewSize(200, 200))
	// 这会创建一个 2 列的网格，
	// 且网格中的每个单元格都尝试保持 100x100 的大小。
	// (注意：整体容器尺寸仍由父容器决定，但布局会要求足够的空间)

	// grid2.Add(grid2hbox)
	listcon := container.NewVBox()
	for _, dir := range cacheDirs {
		// 在循环内部创建新的 HBox (代表“新的一行”)
		rowHBox := container.NewHBox(
			widget.NewLabel(dir),
			widget.NewLabel(humanSize(dir)),
		)

		// 把“新的一行”添加到垂直容器中
		listcon.Add(rowHBox)
	}
	tab2 := container.NewTabItem("玲珑构建器占用缓存大小", listcon)

	// ---------- tab3：ll-cli list ----------
	tab3 := makeLLCliTab()
	// tab3.Refresh = func() { /* 占位，后面会覆盖 */ }

	tab4 := makePruneTab()

	tab5 := makeRepoTableTab()

	// tab6widget.ExtendBaseWidget(tab6widget)
	tab6 := makeDropInstallTab()

	tab7 := extractTab()
	// 组装
	tabs := container.NewAppTabs(tab1, tab2, tab3, tab4, tab5, tab6, tab7)
	tab8 := buildandrunTab(tabs)

	tabs.Append(tab8)
	tabs.SetTabLocation(container.TabLocationLeading)
	// currentTab := tabs.Selected()

	w.SetContent(tabs)
	w.ShowAndRun()
}

// 全局变量来存储字体资源
var customFontResource fyne.Resource

func init() {
	// 1. 获取当前可执行文件的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		fyne.LogError("获取可执行文件路径失败", err)
		// 失败时回退到默认字体
		customFontResource = theme.DefaultTheme().Font(fyne.TextStyle{})
	}
	// 假设您的可执行文件是 /files/ 下的 'EICLING小助手'
	exeDir := filepath.Dir(exePath)

	// 我们需要回到 com.eicling/files/ 目录，然后进入 fonts/
	// 考虑到您的目录结构：/opt/apps/com.eicling/files/
	// 字体相对路径是 "./fonts/NotoSerifSC-Regular.ttf"
	// ⚠️ 最佳实践是使用 filepath.Join 构造路径
	//根据判断是否开发阶段编译运行来加载字体

	var fontPath string

	if os.Getenv("FYNE_DEV_MODE") == "true" {
		// 开发模式：假设字体就在项目根目录的相对路径
		fontPath = "./fonts"

	} else {
		// 部署模式：使用绝对路径逻辑

		fontPath = filepath.Join(exeDir, "fonts")
	}
	// fontPath := filepath.Join(exeDir, "fonts")

	// 假设字体文件在程序运行目录的 fonts/ 文件夹下
	var err2 error
	customFontResource, err2 = fyne.LoadResourceFromPath(fontPath + "/NotoSerifSC-Regular.ttf")
	if err2 != nil {
		fyne.LogError("无法从路径加载自定义字体", err)
		fmt.Println(fontPath + "/NotoSerifSC-Regular.ttf")
		// 失败时回退到默认字体
		customFontResource = theme.DefaultTheme().Font(fyne.TextStyle{})
	}
}

type MyTheme struct{}

// 自定义主题
func (*MyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {

	// if variant == theme.VariantLight {

	// 举例：将背景颜色更改为浅灰色
	if name == theme.ColorNameBackground {
		return color.NRGBA{R: 0xFE, G: 0xFF, B: 0xFF, A: 0xff} // Light Gray
	}

	// 举例：将按钮的主要颜色 (Primary) 更改为自定义的蓝色#FF709B
	if name == theme.ColorNamePrimary {
		return color.NRGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xff} // Dodger Blue
	}
	if name == theme.ColorNameButton {
		return color.NRGBA{R: 0x25, G: 0xA1, B: 0xF4, A: 0xff} // Dodger Blue#25A1F4
	}
	if name == theme.ColorNameHover {
		return color.NRGBA{R: 0x7B, G: 0xBC, B: 0xF4, A: 0xff} // Dodger Blue#7BBCF4
	}
	if name == theme.ColorNameScrollBar {
		return color.NRGBA{R: 0xD4, G: 0xD5, B: 0xCF, A: 0xff} // Dodger Blue#D4D5CF
	}
	if name == theme.ColorNameShadow {
		return color.NRGBA{R: 0xF1, G: 0xEF, B: 0xE3, A: 0xff} // Dodger Blue
	}
	if name == theme.ColorNameSeparator {
		return color.NRGBA{R: 0x47, G: 0x70, B: 0x9B, A: 0xff} // Dodger Blue
	}
	if name == theme.ColorNameForegroundOnPrimary {
		return color.NRGBA{R: 0xFF, G: 0xff, B: 0x00, A: 0xff} // Dodger Blue
	}
	if name == theme.ColorNameForeground {
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff} // Dodger Blue
	}
	if name == theme.ColorNameFocus {
		return color.NRGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xff} // Dodger Blue
	}
	if name == theme.ColorNameSelection {
		return color.NRGBA{R: 0xFF, G: 0x70, B: 0x9B, A: 0xff} // table里面的内容点击后失去焦点所显示的颜色
	}
	if name == theme.ColorNamePlaceHolder {
		return color.NRGBA{R: 0x00, G: 0xFF, B: 0x00, A: 0xff} // 等待执行
	}
	if name == theme.ColorNamePressed {
		return color.NRGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xff} // Dodger Blue
	}

	// fmt.Println("当前主题2：", variant)
	// }
	// fmt.Println("当前主题：", variant)

	// 对于未自定义的颜色，返回默认主题的颜色
	return color.NRGBA{R: 0xFE, G: 0xFF, B: 0xFF, A: 0xff} // Light Gray

}

// Font 返回自定义字体
func (*MyTheme) Font(style fyne.TextStyle) fyne.Resource {
	// 举例：使用自定义的字体文件（需要先将其打包到资源中）
	// if style.Monospace {
	// 	return resourceGoMonoTtf
	// }

	// 默认返回 Fyne 的默认字体
	return customFontResource
}

// Size 返回自定义尺寸
func (*MyTheme) Size(name fyne.ThemeSizeName) float32 {
	// 举例：增加按钮和输入框的默认尺寸 (Padding)

	// 关键步骤：将圆角半径设置为 0
	if name == theme.SizeNameInputRadius {
		return 0 // 0 表示直角 (Square Corners)
	}
	if name == theme.SizeNameWindowButtonRadius {
		return 0 // 0 表示直角 (Square Corners)
	}
	// 默认返回 Fyne 的默认尺寸
	return theme.DefaultTheme().Size(name)
}

// Icon 返回自定义图标
func (*MyTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	// 如果您想替换某个内置图标，可以在这里实现
	return theme.DefaultTheme().Icon(name)
}

/*------------	tab8：构建并运行玲珑包 ---------- */
func buildandrunWindow() {

	// 将变量提升到函数作用域，这样所有回调都能访问
	var dirPathLinux string
	var building bool
	var installCmd *exec.Cmd

	//修改为输入框控件以便于用户复制输出信息
	status := widget.NewEntry()
	status.SetText("请拖拽 .yaml 文件到此处")
	status.MultiLine = true             // 允许多行
	status.Wrapping = fyne.TextWrapWord // 自动换行

	win := fyne.CurrentApp().NewWindow("拖拽释放 .yaml 文件构建并运行")
	win.Resize(fyne.NewSize(420, 220))

	// 灰色拖放区域
	dropRect := canvas.NewRectangle(color.RGBA{R: 50, G: 50, B: 50, A: 255})

	openbtn := widget.NewButton("打开相应linglong文件夹",
		func() {

			// 优先检查 linglong 文件夹
			llfolder := ""
			if dirPathLinux != "" {
				llfolder = filepath.Join(dirPathLinux, "linglong")
			}

			// 检查文件夹是否存在
			if llfolder != "" {
				if info, err := os.Stat(llfolder); err == nil && info.IsDir() {
					fmt.Println("文件夹存在: " + llfolder)
					//这里的Start()方法是用来启动一个外部命令的
					_ = exec.Command("xdg-open", llfolder).Start()
					return
				}
			}

			// 如果不存在，打开默认目录
			dialog.ShowInformation("提示", "未找到 linglong 文件夹，", win)
		})

	lab_con := container.NewBorder(nil, openbtn, nil, nil, status)

	//添加控件到win窗口
	win.SetContent(container.NewStack(dropRect, lab_con)) //dropRect,

	building = false

	// 关键：窗口级拖放回调（v2.7.x 正式 API）
	win.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		// 检查是否有文件被拖拽
		if len(uris) == 0 {
			return
		}

		//  uris[0].Path()表示文件路径
		path := uris[0].Path()
		dirPathLinux = filepath.Dir(path)
		filename := filepath.Base(path)
		fmt.Println("文件名：", filename)

		//HasSuffix方法是检查字符串是否以指定后缀结尾
		if !strings.HasSuffix(path, ".yaml") {
			status.SetText("❌ 仅支持 .yaml 后缀")
			return
		}

		if building {
			status.SetText("❌ 正在构建中，请等待完成...")
			return
		}

		//检测是否已存在
		// info, err := os.Stat(dirPathLinux + "/extract-layer-" + filename)
		// if err == nil {
		// 	if !os.IsNotExist(err) {
		// 		status.SetText("❌文件夹已存在：" + dirPathLinux + "/extract-layer-" + filename)
		// 		return
		// 	} else {
		// 		fmt.Printf("文件夹尚未创建：%v", info)
		// 	}

		// }
		building = true
		status.SetText("⏳ 正在构建和运行 " + filepath.Base(path) + "...")
		lab_con.Refresh()

		go func() {

			defer func() {
				// 确保在 goroutine 结束时更新状态
				building = false
			}()

			installCmd := exec.Command("sh", "-c", fmt.Sprintf("cd '%s' && ll-builder build && ll-builder run", dirPathLinux))

			//CombinedOutput方法是将命令的标准输出和标准错误输出合并到一个字节切片中，返回的值包括命令执行的结果[]byte和错误信息error

			stdout, err := installCmd.StdoutPipe()

			if err != nil {
				// 使用主线程更新 UI
				fyne.Do(func() {
					status.SetText("❌ 构建运行错误，创建输出管道失败: " + err.Error())
					building = false
					status.Refresh()

				})
				return
			}

			if err := installCmd.Start(); err != nil {
				fyne.Do(func() {
					status.SetText("❌ 启动构建失败: " + err.Error())
					building = false

					status.Refresh()

				})
				return
			}

			scanner := bufio.NewScanner(stdout)
			output := ""
			for scanner.Scan() {
				line := scanner.Text()
				cleanLine := stripANSI(line) // 清理 ANSI 转义序列
				output += cleanLine + "\n"
				finalOutput := output
				// 实时更新输出到 UI
				fyne.Do(func() {
					// fmt.Println(output)
					status.SetText(finalOutput)
					// 2. 关键：将光标移动到文本末尾
					status.CursorColumn = len(finalOutput)
					status.Refresh()
				})

			}

			// 等待命令完成
			if err := installCmd.Wait(); err != nil {
				fyne.Do(func() {
					status.SetText(output + "❌ 构建运行失败：\n请检查网络是否正常。\n" + err.Error())
					building = false
				})
			} else {
				fyne.Do(func() {

					status.SetText(output + "✅ 构建和运行结束！但不代表能正常运行，相关缺失依赖请自行检查。\n已在" + dirPathLinux + "下生成：" + "linglong" + "文件夹。\n" + "完整路径为:" + dirPathLinux + "/linglong")
					building = false
				})
			}
			fyne.Do(func() { status.Refresh() })

		}()

	})

	win.SetCloseIntercept(func() {
		if !building {
			win.Close()
			return
		}
		dialog.ShowConfirm("正在解压",
			"正在构建......\n你确定要关闭窗口吗？关闭后构建仍会在后台继续。",
			func(ok bool) {
				if ok {
					// 用户点击了"是"，关闭窗口
					if installCmd != nil && installCmd.Process != nil {
						installCmd.Process.Kill()
					}
					win.Close()
				}
				// 用户点击了"否"，什么都不做，窗口保持打开
			}, win)
	})
	win.Show()
}
func buildandrunTab(tabsin *container.AppTabs) *container.TabItem {
	tab8con := container.NewVBox()
	showtxt := widget.NewLabel("构建并运行玲珑包功能")
	targetButton := widget.NewButton("打开拖拽窗口", func() {
		// fmt.Println("按钮被点击了，b 的当前值是:")
		buildandrunWindow()

	})
	// 使用 OnSelected 回调替代直接检查
	tabsin.OnSelected = func(tab *container.TabItem) {
		if tab == tabsin.Items[7] { // 最后一个标签页，即tab8

			fmt.Println("当前选中的是构建并运行标签页，开始查询linglong-builder包信息...")

			version, found := queryPkg("linglong-builder")
			if found {
				showtxt.SetText("linglong-builder 已安装，版本: " + version)

			} else {
				showtxt.SetText("linglong-builder 未安装")
				//这里表示没有安装linglong-builder包，禁用按钮
				targetButton.Disable()

			}
		}
	}

	tab8con.Add(showtxt)
	tab8con.Add(targetButton)
	return container.NewTabItem("构建并运行玲珑包", tab8con)
}

/* ---------- 构建 “ll-cli list” 标签页（6 列真实输出） ---------- */
func makeLLCliTab() *container.TabItem {
	// 1. 表头：6 列
	//container.NewGridWithColumns返回的值是一个fyne.Container类型，是一个网格布局容器
	// header := container.NewGridWithColumns(6) // 只显示 6 列
	// header.Add(widget.NewLabelWithStyle("ID", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	// header.Add(widget.NewLabelWithStyle("名称", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	// header.Add(widget.NewLabelWithStyle("版本", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	// header.Add(widget.NewLabelWithStyle("渠道", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	// header.Add(widget.NewLabelWithStyle("模块", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	// header.Add(widget.NewLabelWithStyle("描述详细信息", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	// 2. 动态内容区
	listContainer := container.NewVBox()
	// listContainer.Add(header)
	scroll := container.NewScroll(listContainer)
	scroll.SetMinSize(fyne.NewSize(0, 50))

	// 3. 错误提示
	errLabel := widget.NewLabel("")
	errLabel.TextStyle = fyne.TextStyle{Bold: true}
	errLabel.Hide()

	// const colWidth = 30 // 每列固定字符数，自己调

	// 4. 刷新函数
	refresh := func() {
		// 清空旧内容
		listContainer.Objects = nil
		errLabel.Hide()

		//runLLCliList 函数返回的是一个 appInfo 结构体的切片和一个错误，appInfo结构体包含了6个string类型的字段
		apps, err := runLLCliList()
		if err != nil {
			errLabel.SetText("执行 ll-cli list 失败: " + err.Error())
			errLabel.Show()
			return
		}

		if len(apps) == 0 {
			listContainer.Add(widget.NewLabel("（暂无应用）"))
		} else {
			listContainer.Add(
				container.NewGridWithColumns(6,
					widget.NewLabel("ID"),
					widget.NewLabel("名称"),
					widget.NewLabel("版本"),
					widget.NewLabel("渠道"),
					widget.NewLabel("模块"),
					widget.NewLabel("描述"),
				))

			const (
				normalWidth = 25 // 前5列字符宽度
				lastWidth   = 40 // 最后一列字符宽度
				colPx       = 40 // 每列像素宽度，自己调
			)
			for _, app := range apps {
				row := container.NewGridWithColumns(6)

				// 1. 打开按钮
				openBtn := widget.NewButton("打开所在目录", func() {
					fmt.Println("测试：", app.commit)
					path := filepath.Join("/var/lib/linglong/layers/", app.commit)
					_ = exec.Command("xdg-open", path).Start()
				})

				// 2. ID 标签（截断到 normalWidth 字符）
				idLbl := widget.NewLabel(trunc(app.id, normalWidth))
				// idLbl.Wrapping = fyne.TextTruncate // 超长省略号

				// 3. 按钮+ID 横着放，作为第一列
				// row.Add(container.NewHBox(openBtn, idLbl))
				// 按钮占固定像素，其余给文字
				openBtn.Resize(fyne.NewSize(40, 28)) // 按钮宽40，高28
				box := container.NewHBox(
					openBtn,
					idLbl, // 文字自动占剩余
				)
				// box.SetMinSize(fyne.NewSize(colPx, 0)) // 整格宽度=列宽
				row.Add(box)

				// 后面 5 列照原样
				row.Add(widget.NewLabel(trunc(app.name, normalWidth)))
				row.Add(widget.NewLabel(trunc(app.version, normalWidth)))
				row.Add(widget.NewLabel(trunc(app.channel, normalWidth)))
				row.Add(widget.NewLabel(trunc(app.module, normalWidth)))
				row.Add(widget.NewLabel(trunc(app.desc, lastWidth)))

				listContainer.Add(row)
			}
		}
		listContainer.Refresh()
	}

	// 5. 顶部区域
	top := container.NewBorder(nil, nil, nil, widget.NewButton("刷新", refresh), errLabel)

	// 6. 整体布局
	content := container.NewBorder(nil, nil, nil, nil, scroll)

	// 7. 首次加载
	refresh()

	return container.NewTabItem("列出已安装玲珑应用", container.NewBorder(top, nil, nil, nil, content))
}

/* ---------- 解析 ll-cli list 输出 ---------- */
type appInfo struct {
	id, name, version, channel, module, desc, commit string
}

func runLLCliList() ([]appInfo, error) {
	// 1. 读固定文件
	path := "/var/lib/linglong/states.json"
	out, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 states.json 失败: %w", err)
	}

	// 2. 定义跟 JSON 对应的临时结构
	type jsonItem struct {
		Commit string `json:"commit"` // ← 新增

		Info struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Version     string `json:"version"`
			Channel     string `json:"channel"`
			Module      string `json:"module"`
			Description string `json:"description"`
		} `json:"info"`
	}
	type root struct {
		Layers []jsonItem `json:"layers"`
	}

	var r root
	if err := json.Unmarshal(out, &r); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 3. 转我们自己的切片
	apps := make([]appInfo, 0, len(r.Layers))

	for _, j := range r.Layers {
		apps = append(apps, appInfo{
			id:      strings.TrimSpace(j.Info.ID),
			name:    strings.TrimSpace(j.Info.Name),
			version: strings.TrimSpace(j.Info.Version),
			channel: strings.TrimSpace(j.Info.Channel),
			module:  strings.TrimSpace(j.Info.Module),
			desc:    strings.TrimSpace(j.Info.Description),
			commit:  j.Commit, // ← 新增

		})
	}

	// for _, app := range apps {
	// 	dumpField("id", app.id)
	// 	dumpField("name", app.name)
	// 	dumpField("version", app.version)
	// 	dumpField("channel", app.channel)
	// 	dumpField("module", app.module)
	// 	dumpField("desc", app.desc)
	// 	fmt.Println("---")
	// }

	return apps, nil
}

// 正则：匹配 ANSI 转义序列
// var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// func stripANSI(s string) string {
// 	return ansiRe.ReplaceAllString(s, "")
// }

/* ---------- 通用工具 ---------- */
func humanSize(path string) string {
	fi, err := os.Stat(path)
	if err != nil || !fi.IsDir() {
		return "—"
	}
	var bytes int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			bytes += info.Size()
		}
		return nil
	})
	return fmt.Sprintf("%.2f MB", float64(bytes)/1024/1024)
}

// 调试：把字符串“里子”全亮出来
func dumpField(name, s string) {
	fmt.Printf("%-10s | len=%-3d | runes=%-3d | hex=%-40s | raw=%q\n",
		name,
		len(s),         // 字节长度
		len([]rune(s)), // 可见字符数
		hex.EncodeToString([]byte(s)),
		s)
}
func showRowMenu(app appInfo, pos fyne.Position, win fyne.Window) {
	// 拼路径
	path := filepath.Join("/var/lib/linglong/layers", app.commit)
	menu := fyne.NewMenu("",
		fyne.NewMenuItem("打开所在路径", func() {
			_ = exec.Command("xdg-open", path).Start()
		}),
	)
	pop := widget.NewPopUpMenu(menu, win.Canvas())
	pop.ShowAtPosition(pos)
}
func trunc(s string, max int) string {
	if len([]rune(s)) > max {
		return string([]rune(s)[:max-3]) + "..."
	}
	return s
}

/* ---------- tab1：检测玲珑系列包 ---------- */
func makeTab1CheckLLPkgs() *container.TabItem {
	list := container.NewVBox()
	refresh := func() {
		list.Objects = nil
		pkgs := []string{
			"linglong-bin", "linglong-box", "linglong-builder",
			"linglong-installer", "linglong-pica",
		}
		for _, p := range pkgs {
			ver, ok := queryPkg(p)
			status := "✘ 未安装"
			if ok {
				status = "✔ " + ver
			}
			list.Add(container.NewHBox(
				widget.NewLabel(p),
				widget.NewLabel(status),
			))
		}
		zhuyitxt := widget.NewLabel("注：假如未安装linglong-bin, linglong-box,linglong-builder,linglong-installer,linglong-pica将无法正常使用本软件相关功能！")
		zhuyitxt.Wrapping = fyne.TextWrapWord
		zhu := container.NewVBox(zhuyitxt)
		list.Add(zhu)
	}
	refresh() // 首次加载

	top := container.NewBorder(nil, nil, nil, widget.NewButton("刷新", refresh), nil)
	return container.NewTabItem("玲珑工具包检测", container.NewBorder(top, nil, nil, nil, list))
}

// 查询单个包是否已安装
func queryPkg(name string) (version string, found bool) {
	// 优先用 dpkg，没有就退到 rpm
	for _, cmdLine := range [][]string{
		{"dpkg-query", "-W", "-f=${Version}", name},
		{"rpm", "-q", "--queryformat", "%{VERSION}", name},
	} {
		out, err := exec.Command(cmdLine[0], cmdLine[1:]...).Output()
		//如果错误为空，则表示命令执行成功
		if err == nil {
			return strings.TrimSpace(string(out)), true
		}
	}
	//假如以上命令都执行失败，则返回空字符串
	return "", false
}

/* ---------- tab4：清理未使用运行时 ---------- */
func makePruneTab() *container.TabItem {
	out := widget.NewEntry() // 用来显示结果
	out.MultiLine = true
	out.Wrapping = fyne.TextWrapWord

	//设置输入框为空的时候显示的提示文字
	out.SetPlaceHolder("等待执行...")
	// out.Disable() // 只读

	btn := widget.NewButton("执行清理", func() {
		out.SetText("正在执行，请稍候...")
		go func() { // 避免阻塞 UI
			cmd := exec.Command("ll-cli", "prune")
			cmd.Env = append(os.Environ(), "LC_ALL=C.UTF-8")
			raw, _ := cmd.CombinedOutput() // stdout + stderr
			txt := strings.TrimSpace(string(raw))

			// 统一判断“无包可清”
			if strings.Contains(txt, "No packages to prune") ||
				strings.Contains(txt, "Error -1") {
				txt = "没有符合的运行时被清除。"
			}
			fyne.Do(func() { out.SetText(txt) })
		}()
	})

	box := container.NewVBox(out, btn)
	return container.NewTabItem("移除未使用的最小系统或运行时", box)
}

/* ---------- tab5：仓库信息（Table 版） ---------- */
func showstripANSI(s string) string {
	// 1. 标准 ANSI 转义序列（ESC [ 开头）
	ansiRe := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	s = ansiRe.ReplaceAllString(s, "")

	// 2. 专门清理光标控制序列
	// cursorRe := regexp.MustCompile(`\[\?25[hl]`) // [?25l 和 [?25h
	// s = cursorRe.ReplaceAllString(s, "")

	// 3. 清理其他可能的 ANSI 序列变体
	// extraRe := regexp.MustCompile(`\[\?[0-9]+[hl]`) // 类似的 [?数字h/l 模式
	// s = extraRe.ReplaceAllString(s, "")

	// 4. 清理所有其他控制字符
	// controlRe := regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`)
	// s = controlRe.ReplaceAllString(s, "")

	return strings.TrimSpace(s)
}

func makeRepoTableTab() *container.TabItem {
	var data [][]string // 二维切片：[行][列]

	// 表头
	header := []string{"名称", "地址", "别名", "优先级"}

	table := widget.NewTable(
		func() (int, int) { return len(data), len(header) },
		func() fyne.CanvasObject {
			return widget.NewLabel("") // 单元格模板
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			l.SetText(data[i.Row][i.Col])
			// l.Wrapping = fyne.TextTruncate
		})
	table.SetColumnWidth(0, 80)  // 名称
	table.SetColumnWidth(1, 300) // 地址
	table.SetColumnWidth(2, 60)  // 别名
	table.SetColumnWidth(3, 60)  // 优先级
	// 拉取数据
	refresh := func() {
		data = nil
		cmd := exec.Command("ll-cli", "repo", "show")
		// cmd.Env = append(os.Environ(), "LC_ALL=C.UTF-8")
		out, err := cmd.Output()
		txt := string(out)
		fmt.Println(txt)
		txt = showstripANSI(txt) // 去掉 ANSI 颜色
		txt = strings.TrimSpace(txt)

		if err == nil {
			lines := strings.Split(txt, "\n")
			reSplit := regexp.MustCompile(`\s{2,}`)
			for i := 1; i < len(lines); i++ { // 跳过表头
				cols := reSplit.Split(strings.TrimSpace(lines[i]), 4)
				if len(cols) >= 4 {
					//cols[:4]的意思是取cols的前4个元素，即cols[0:4]
					data = append(data, cols[:4])
				}
			}
		}
		if len(data) == 0 {
			data = [][]string{{"暂无仓库信息", "", "", ""}}
		} else {
			// ★★★ 在这里改第一行内容 ★★★
			data[0][0] = "✅" + data[0][0] // 第一行第一列
			data[0][1] = "✅" + data[0][1]
			data[0][2] = "✅" + data[0][2]
			data[0][3] = "✅" + data[0][3] // 第一行最后一列
		}
		table.Refresh()
	}

	// 创建 Table

	// 顶部刷新按钮
	top := container.NewBorder(nil, nil, nil, widget.NewButton("刷新", refresh), nil)

	// 首次加载
	refresh()
	return container.NewTabItem("仓库信息", container.NewBorder(top, nil, nil, nil, table))
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

/* ---------- tab6：拖拽安装 .layer ---------- */
func showDropInstallWindow() {

	//修改为输入框控件以便于用户复制输出信息
	status := widget.NewEntry()

	status.SetText("请拖拽 .layer 文件到此处")
	// status.TextStyle = fyne.TextStyle{}

	status.MultiLine = true             // 允许多行
	status.Wrapping = fyne.TextWrapWord // 自动换行
	// status.Scroll = widget.ScrollNone   // ❌ 禁用滚动条（Fyne 2.7+）

	// 1. 先测出“最宽单词”所需像素
	//现在改为固定28个字符宽度
	// const maxChar = 28                         // 约 28 个汉字
	// status.Resize(fyne.NewSize(380, 0)) // 8 ≈ 单字符像素
	win := fyne.CurrentApp().NewWindow("拖拽安装 .layer")
	win.Resize(fyne.NewSize(420, 220))

	// 2. 宽度 = 最宽单词 + 一点余量（避免 1 字换行）
	// status.Resize(fyne.NewSize(minW+10, 0))
	// status.TextStyle = fyne.TextStyle{Bold: true}
	// status.Wrapping = fyne.TextWrapWord // ← 自动换行

	// 灰色拖放区域
	dropRect := canvas.NewRectangle(color.RGBA{R: 50, G: 50, B: 50, A: 255})
	// dropRect.SetMinSize(fyne.NewSize(100, 200))
	// dropRect.Resize(fyne.NewSize(50, 200))

	// win.SetFixedSize(true)             // ← 禁止用户缩放
	// win.Resize(fyne.NewSize(420, 220)) // ← 你想要的固定尺寸

	lab_con := container.NewBorder(nil, nil, nil, nil, status)
	//添加控件到win窗口
	win.SetContent(container.NewStack(dropRect, lab_con)) //dropRect,

	installing := false
	var installCmd *exec.Cmd

	// 关键：窗口级拖放回调（v2.7.x 正式 API）
	win.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		// 检查是否有文件被拖拽
		if len(uris) == 0 {
			return
		}

		//  uris[0].Path()表示文件路径
		path := uris[0].Path()

		//HasSuffix方法是检查字符串是否以指定后缀结尾
		if !strings.HasSuffix(path, ".layer") {
			status.SetText("❌ 仅支持 .layer 后缀")
			return
		}

		if installing {
			status.SetText("❌ 正在安装中，请等待完成...")
			return
		}

		installing = true
		status.SetText("⏳ 正在安装 " + filepath.Base(path) + "...")
		lab_con.Refresh()

		go func() {

			defer func() {
				// 确保在 goroutine 结束时更新状态
				installing = false
			}()

			installCmd := exec.Command("ll-cli", "install", path)
			//CombinedOutput方法是将命令的标准输出和标准错误输出合并到一个字节切片中，返回的值包括命令执行的结果[]byte和错误信息error
			stdout, err := installCmd.StdoutPipe()

			if err != nil {
				// 使用主线程更新 UI
				fyne.Do(func() {
					status.SetText("❌ 创建输出管道失败: " + err.Error())
				})
				installing = false
				status.Refresh()
				return
			}

			if err := installCmd.Start(); err != nil {
				fyne.Do(func() {
					status.SetText("❌ 启动安装失败: " + err.Error())
				})
				installing = false

				status.Refresh()
				return
			}

			scanner := bufio.NewScanner(stdout)
			output := ""
			for scanner.Scan() {
				line := scanner.Text()
				cleanLine := stripANSI(line) // 清理 ANSI 转义序列
				output += cleanLine + "\n"
				// 实时更新输出到 UI
				fyne.Do(func() {
					fmt.Println(output)
					status.SetText(output)
				})
				fyne.Do(func() { status.Refresh() })
			}

			// 等待命令完成
			if err := installCmd.Wait(); err != nil {
				fyne.Do(func() {
					status.SetText(output + "❌ 安装失败：\n请检查是否重复安装以及网络是否正常。\n" + err.Error())
					installing = false
				})
			} else {
				fyne.Do(func() {
					status.SetText(output + "✅ 安装完成！")
					installing = false
				})
			}
			fyne.Do(func() { status.Refresh() })

		}()
	})

	win.SetCloseIntercept(func() {
		if !installing {
			win.Close()
			return
		}
		dialog.ShowConfirm("正在安装",
			"现在正在安装，你确定要关闭窗口吗？关闭后安装仍会在后台继续。",
			func(ok bool) {
				if ok {
					// 用户点击了"是"，关闭窗口
					if installCmd != nil && installCmd.Process != nil {
						installCmd.Process.Kill()
					}
					win.Close()
				}
				// 用户点击了"否"，什么都不做，窗口保持打开
			}, win)
	})
	win.Show()
}

/* ------------------------------------------------------------ tab6：按钮弹出拖拽窗口 ---------------------------------------------------------- */
func makeDropInstallTab() *container.TabItem {
	btn := widget.NewButton("打开拖拽安装窗口", func() {
		// parent 传主窗口即可
		//fyne.CurrentApp().Driver().AllWindows()[0]的意思是获取当前应用的所有窗口中的第一个窗口，即主窗口
		showDropInstallWindow()
	})
	return container.NewTabItem("拖拽安装 .layer", container.NewCenter(btn))
}
func stripANSI(s string) string {
	// 1. 标准 ANSI 转义序列（ESC [ 开头）
	ansiRe := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	s = ansiRe.ReplaceAllString(s, "")

	// 2. 专门清理光标控制序列
	cursorRe := regexp.MustCompile(`\[\?25[hl]`) // [?25l 和 [?25h
	s = cursorRe.ReplaceAllString(s, "")

	// 3. 清理其他可能的 ANSI 序列变体
	extraRe := regexp.MustCompile(`\[\?[0-9]+[hl]`) // 类似的 [?数字h/l 模式
	s = extraRe.ReplaceAllString(s, "")

	// 4. 清理所有其他控制字符
	controlRe := regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`)
	s = controlRe.ReplaceAllString(s, "")

	return strings.TrimSpace(s)
}

/* ---------- tab7：拖拽解压layer文件 ---------- */
func extractWindow() {

	//修改为输入框控件以便于用户复制输出信息
	status := widget.NewEntry()

	status.SetText("请拖拽 .layer 文件到此处")

	status.MultiLine = true             // 允许多行
	status.Wrapping = fyne.TextWrapWord // 自动换行

	win := fyne.CurrentApp().NewWindow("拖拽释放 .layer 文件解压")
	win.Resize(fyne.NewSize(420, 220))

	// 灰色拖放区域
	dropRect := canvas.NewRectangle(color.RGBA{R: 50, G: 50, B: 50, A: 255})

	lab_con := container.NewBorder(nil, nil, nil, nil, status)

	//添加控件到win窗口
	win.SetContent(container.NewStack(dropRect, lab_con)) //dropRect,

	extracting := false
	var installCmd *exec.Cmd

	// 关键：窗口级拖放回调（v2.7.x 正式 API）
	win.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		// 检查是否有文件被拖拽
		if len(uris) == 0 {
			return
		}

		//  uris[0].Path()表示文件路径
		path := uris[0].Path()
		dirPathLinux := filepath.Dir(path)
		filename := filepath.Base(path)

		//HasSuffix方法是检查字符串是否以指定后缀结尾
		if !strings.HasSuffix(path, ".layer") {
			status.SetText("❌ 仅支持 .layer 后缀")
			return
		}

		if extracting {
			status.SetText("❌ 正在解压中，请等待完成...")
			return
		}

		info, err := os.Stat(dirPathLinux + "/extract-layer-" + filename)
		if err == nil {
			if !os.IsNotExist(err) {
				status.SetText("❌文件夹已存在：" + dirPathLinux + "/extract-layer-" + filename)
				return
			} else {
				fmt.Printf("文件夹尚未创建：%v", info)
			}

		}
		extracting = true
		status.SetText("⏳ 正在解压 " + filepath.Base(path) + "...")
		lab_con.Refresh()

		go func() {

			defer func() {
				// 确保在 goroutine 结束时更新状态
				extracting = false
			}()

			installCmd := exec.Command("ll-builder", "extract", path, dirPathLinux+"/extract-layer-"+filename)

			//CombinedOutput方法是将命令的标准输出和标准错误输出合并到一个字节切片中，返回的值包括命令执行的结果[]byte和错误信息error

			stdout, err := installCmd.StdoutPipe()

			if err != nil {
				// 使用主线程更新 UI
				fyne.Do(func() {
					status.SetText("❌ 解压错误，创建输出管道失败: " + err.Error())
				})
				extracting = false
				status.Refresh()
				return
			}

			if err := installCmd.Start(); err != nil {
				fyne.Do(func() {
					status.SetText("❌ 启动解压失败: " + err.Error())
				})
				extracting = false

				status.Refresh()
				return
			}

			scanner := bufio.NewScanner(stdout)
			output := ""
			for scanner.Scan() {
				line := scanner.Text()
				cleanLine := stripANSI(line) // 清理 ANSI 转义序列
				output += cleanLine + "\n"
				// 实时更新输出到 UI
				fyne.Do(func() {
					fmt.Println(output)
					status.SetText(output)
				})
				fyne.Do(func() { status.Refresh() })
			}

			// 等待命令完成
			if err := installCmd.Wait(); err != nil {
				fyne.Do(func() {
					status.SetText(output + "❌ 解压失败：\n请检查是否重复解压以及网络是否正常。\n" + err.Error())
					extracting = false
				})
			} else {
				fyne.Do(func() {
					status.SetText(output + "✅ 解压完成！已在layer文件所在目录：" + dirPathLinux + "\n生成:extract-layer-" + filename + "文件夹。\n" + "完整路径为:" + dirPathLinux + "/extract-layer-" + filename)
					extracting = false
				})
			}
			fyne.Do(func() { status.Refresh() })

		}()
	})

	win.SetCloseIntercept(func() {
		if !extracting {
			win.Close()
			return
		}
		dialog.ShowConfirm("正在解压",
			"正在解压......\n你确定要关闭窗口吗？关闭后解压仍会在后台继续。",
			func(ok bool) {
				if ok {
					// 用户点击了"是"，关闭窗口
					if installCmd != nil && installCmd.Process != nil {
						installCmd.Process.Kill()
					}
					win.Close()
				}
				// 用户点击了"否"，什么都不做，窗口保持打开
			}, win)
	})
	win.Show()
}

// 创建解压按钮
func extractTab() *container.TabItem {
	btn := widget.NewButton("打开拖拽解压窗口", func() {

		extractWindow()
	})
	return container.NewTabItem("解压layer包", container.NewCenter(btn))
}

## xmind
基于go语言的xmind接口

使用方法参考: [example](example)

本库主要加载xmind文件为json结构,保存文件时也用的json结构而不是xml结构

本库只做了最基本的主题添加功能,类似`标签/备注/图片`等其他功能不考虑,有想法的自行实现

本库做了通用加载和通用保存方法,可以更灵活的与其他思维导图进行转换

参考: [custom_test](example/custom_test.go)

## 示例
* 自定义json数据创建xmind
```go
package main

import (
	"github.com/jan-bar/xmind"
)

func main() {
	data := `[{"a":"1","b":"main topic"},
{"a":"2","b":"topic1","c":"1"},{"a":"3","b":"topic2","c":"1"},
{"a":"4","b":"topic3","c":"2"},{"a":"5","b":"topic4","c":"2"},
{"a":"6","b":"topic5","c":"3"},{"a":"7","b":"topic6","c":"3"}
]`
	// 这里定义 a 表示节点id, b 表示主题内容, c 表示父节点id
	// 传入定好的json字符串,以及指定好json的key字符串就可以将任意json数据转换成xmind
	// 也可用用 data := []byte(`{}`) 传入字节数组
	st, err := xmind.LoadCustom(data, "a", "b", "c")
	if err != nil {
		panic(err)
	}
	err = xmind.SaveSheets("custom.xmind", st)
	if err != nil {
		panic(err)
	}
}
```
* 通过接口创建xmind对象,并保存xmind文件
```go
package main

import (
	"github.com/jan-bar/xmind"
)

func main() {
	st1 := xmind.NewSheet("sheet1", "main 1 topic")
	st1.Add("123").Add("456").Add("789").OnTitle("123").
		Add("2sc").Add("345").OnTitle("456").
		Add("xzcv").Add("ewr").OnTitle("789").Add("saf").Add("xcv")

	st2 := xmind.NewSheet("sheet2", "main 2 topic")
	st2.Add("aaa").Add("ewr")
	st2.OnTitle("ewr").Title = "xx-ewr\txvf\nwer" // 修改指定主题内容,其中包含特殊转义字符
	st2.Add("cvxcv").Add("wqerwe").OnTitle("aaa").
		Add("zxs", xmind.ParentMode). // 为节点添加父节点
		Add("cxv", xmind.BeforeMode). // 在节点之前添加兄弟节点
		Add("xcas", xmind.AfterMode). // 在节点之后添加兄弟节点
		OnTitle("cvxcv").Add("34").Add("xcv")

	err := xmind.SaveSheets("create.xmind", st1, st2)
	if err != nil {
		panic(err)
	}
}
```
* 加载xmind文件
```go
package main

import (
	"github.com/jan-bar/xmind"
)

func main() {
	// 支持4种方式的加载,详情看内部具体实现
	wb, err := xmind.LoadFile("TestCreateXmind.xmind")
	if err != nil {
		panic(err)
	}
	if len(wb.Topics) != 2 {
		return
	}

	// 在第一个sheet页修改一些数据
	wb.Topics[0].OnTitle("345").Add("111").Add("222").OnTitle("xcv").
		Add("xzcv").Add("werw")

	// 在第二个sheet页修改一些数据
	wb.Topics[1].OnTitle("34").Add("111").Add("222").OnTitle("aaa").
		Add("xzcv").Add("werw")

	// 可以用xmind打开这两个文件,比较一下不同
	err = wb.Save("TestLoadXmindJson.xmind")
	if err != nil {
		panic(err)
	}
}
```
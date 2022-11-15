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
	st2.Add("vbg").Add("qwe").OnTitle("aaa").
		Add("zxs", xmind.ParentMode). // 为节点添加父节点
		Add("cxv", xmind.BeforeMode). // 在节点之前添加兄弟节点
		Add("xcas", xmind.AfterMode). // 在节点之后添加兄弟节点
		OnTitle("vbg").Add("34").Add("xcv").OnTitle("qwe").
		Add("111").Add("222").Add("333")

	err := xmind.SaveSheets("create.xmind", st1, st2)
	if err != nil {
		panic(err)
	}

	src := st2.CId("zxs")
	st2.OnTitle("34").Move(src) // 将'zxs'节点移动到'34'节点作为子节点

	// 将'qwe'节点移动到'zxs'节点同级下方
	st2.OnTitle("zxs").Move(st2.CId("qwe"), xmind.AfterMode)

	err = xmind.SaveSheets("move1.xmind", st2)
	if err != nil {
		panic(err)
	}
}

package example

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jan-bar/xmind"
)

func TestCreateXmind(t *testing.T) {
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

	err := xmind.SaveSheets("TestCreateXmind.xmind", st1, st2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateStruct(t *testing.T) {
	st1 := &xmind.Topic{
		Title: "sheet1",
		RootTopic: &xmind.Topic{
			Title:          "main 1 topic",
			StructureClass: xmind.StructLogicRight,
			Children: &xmind.Children{
				Attached: []*xmind.Topic{
					{
						Title: "a",
						Children: &xmind.Children{
							Attached: []*xmind.Topic{
								{Title: "a1"},
								{Title: "a2"},
								{Title: "a3"},
							},
						},
					}, {
						Title: "b",
						Children: &xmind.Children{
							Attached: []*xmind.Topic{
								{Title: "b1"},
								{Title: "b2"},
								{Title: "b3"},
							},
						},
					}, {
						Title: "c",
						Children: &xmind.Children{
							Attached: []*xmind.Topic{
								{Title: "c1"},
								{Title: "c2"},
								{Title: "c3"},
							},
						},
					},
				},
			},
		},
	}

	// 通过手动指定结构体,也可以创建文件,只需要填写Title,其他字段会自动生成
	wb := xmind.WorkBook{Topics: []*xmind.Topic{st1}}
	err := wb.Save("TestCreateStruct.xmind")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSaveJson(t *testing.T) {
	st1 := xmind.NewSheet("sheet1", "main 1 topic")
	st1.Add("123").Add("456").Add("789").OnTitle("123").
		Add("2sc").Add("345").OnTitle("456").
		Add("xzcv").Add("ewr").OnTitle("789").Add("saf").Add("xcv")
	st2 := xmind.NewSheet("sheet1", "main 1 topic")
	st2.Add("123").Add("456").Add("789").OnTitle("123").
		Add("2sc").Add("345").OnTitle("456").
		Add("xzcv").Add("ewr").OnTitle("789").Add("saf").Add("xcv")

	data, err := json.Marshal([]*xmind.Topic{st1, st2})
	if err != nil {
		t.Fatal(err)
	}

	// 直接将sheet保存成json文件
	err = os.WriteFile("create.json", data, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func TestName(t *testing.T) {
	st1 := &xmind.Topic{
		Title: "sheet1",
		RootTopic: &xmind.Topic{
			Title:          "main 1 topic",
			StructureClass: xmind.StructLogicRight,
			Children: &xmind.Children{
				Attached: []*xmind.Topic{
					{
						Title: "a",
						Children: &xmind.Children{
							Attached: []*xmind.Topic{
								{Title: "a1"},
								{Title: "a2"},
								{Title: "a3"},
							},
						},
					}, {
						Title: "b",
						Children: &xmind.Children{
							Attached: []*xmind.Topic{
								{Title: "b1"},
								{Title: "b2"},
								{Title: "b3"},
							},
						},
					}, {
						Title: "c",
						Children: &xmind.Children{
							Attached: []*xmind.Topic{
								{Title: "c1"},
								{Title: "c2"},
								{Title: "c3"},
							},
						},
					},
				},
			},
		},
	}


	// 通过手动指定结构体,也可以创建文件,只需要填写Title,其他字段会自动生成
	wb := xmind.WorkBook{Topics: []*xmind.Topic{st1}}
	err := wb.Save("TestCreateStruct.xmind")
	if err != nil {
		t.Fatal(err)
	}
}

// 不支持保存为xml文件,因为xmind会增加一些字段,没有实现增加这些字段的方法
// 但是可以直接加载xml文件,以及加载*.xmind文件中的json或xml文件

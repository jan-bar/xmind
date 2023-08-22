package example

import (
	"os"
	"testing"

	"github.com/jan-bar/xmind"
)

// go test -v -run TestSaveMarkdown
func TestSaveMarkdown(t *testing.T) {
	st1 := xmind.NewSheet("sheet1", "main 1 topic")
	st1.Add("123").Add("456").Add("789").OnTitle("123").
		AddLabel("this is label1", "this is label2").
		Add("2sc").Add("345").OnTitle("456").AddNotes("I'm notes").
		Add("xzcv").Add("ewr").OnTitle("789").Add("saf").Add("xcv")
	st2 := xmind.NewSheet("sheet2", "main 2 topic")
	st2.Add("aaa").Add("ewr")
	st2.OnTitle("ewr").Title = "xx-ewr\txvf\nwer" // 修改指定主题内容,其中包含特殊转义字符
	st2.Add("cvxcv").Add("wqerwe").OnTitle("aaa").
		Add("zxs", xmind.ParentMode). // 为节点添加父节点
		Add("cxv", xmind.BeforeMode). // 在节点之前添加兄弟节点
		Add("xcas", xmind.AfterMode). // 在节点之后添加兄弟节点
		OnTitle("cvxcv").Add("34").Add("xcv")

	wk := &xmind.WorkBook{Topics: []*xmind.Topic{st1, st2}}

	save := func(name string, format map[string]string) error {
		fw, err := os.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		defer fw.Close()

		return wk.SaveToMarkdown(fw, format)
	}

	err := save("markdown1.tmp.md", nil) // 默认格式
	if err != nil {
		t.Fatal(err)
	}

	err = save("markdown2.tmp.md", map[string]string{
		// 修改默认样式,这里在尾部增加一条分割线,对于没有设置的层级样式都用默认样式
		xmind.DefaultMarkdownName: xmind.DefaultMarkdownFormat + "\n*****\n",
		// 表示1级主题,也就是中心主题样式
		"1": xmind.DefaultMarkdownFormat + "\n**main**\n",
		// 自定义2级主题样式,前端添加Deep个空格,用'-'表示序号
		"2": "{{Repeat \" \" .Deep}}- {{.Title}}\n\n{{range $i,$v := .Labels}}> {{$v}}\n\n{{end}}{{if .Notes}}> {{.Notes}}\n\n{{end}}",
	})
	if err != nil {
		t.Fatal(err)
	}
}

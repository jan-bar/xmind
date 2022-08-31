package example

import (
	"os"
	"testing"

	"github.com/jan-bar/xmind"
)

func TestLoadCustom(t *testing.T) {
	// 特别注意,下面的方式可以自定义任何数据转换为sheet对象
	// 唯一需要注意的是root节点父节点为空,其他节点均按照要求填写即可

	t.Run("string", func(t *testing.T) {
		data := `[{"a":"1","b":"main topic","d":true},
{"a":"2","b":"topic1","c":"1"},{"a":"3","b":"topic2","c":"1"},
{"a":"4","b":"topic3","c":"2"},{"a":"5","b":"topic4","c":"2"},
{"a":"6","b":"topic5","c":"3"},{"a":"7","b":"topic6","c":"3"}
]`
		// 这里定义 a 表示节点id, b 表示主题内容, c 表示父节点id
		// 传入定好的json字符串,以及指定好json的key字符串就可以将任意json数据转换成xmind
		// 也可用用 data := []byte(`{}`) 传入字节数组
		st, err := xmind.LoadCustom(data, "a", "b", "c", "d")
		if err != nil {
			t.Fatal(err)
		}
		err = xmind.SaveSheets("TestLoadCustom.string.xmind", st)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("struct", func(t *testing.T) {
		type Node struct {
			A string `json:"id"`
			B string `json:"topic"`
			C string `json:"parent,omitempty"`
		}
		data := []Node{{A: "1", B: "main topic"},
			{A: "2", B: "topic1", C: "1"}, {A: "3", B: "topic2", C: "1"},
			{A: "4", B: "topic3", C: "3"}, {A: "5", B: "topic4", C: "3"},
			{A: "6", B: "topic5", C: "2"}, {A: "7", B: "topic6", C: "2"},
		}

		// 直接传结构体数组,并且传三个字段的json tag,就可以直接从自定义结构生成sheet
		st, err := xmind.LoadCustom(data, "id", "topic", "parent", "")
		if err != nil {
			t.Fatal(err)
		}
		err = xmind.SaveSheets("TestLoadCustom.struct.xmind", st)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestSaveCustom(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		st := xmind.NewSheet("sheet1", "main topic")
		st.Add("123").Add("456").OnTitle("123").Add("111").Add("222").
			OnTitle("456").Add("xzc").Add("wqer")

		var data []byte // 直接将sheet对象转换为自定义json结构,也可用 `var data string` 获取字符串
		err := xmind.SaveCustom(st, "id", "title", "parentId", "isroot,1", &data)
		if err != nil {
			t.Fatal(err)
		}

		err = os.WriteFile("TestSaveCustom.string.json", data, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("struct", func(t *testing.T) {
		st := xmind.NewSheet("sheet1", "main topic")
		st.Add("222").Add("333").OnTitle("222").Add("111").Add("222").
			OnTitle("333").Add("xzc").Add("wqer")

		type Node struct {
			Id       string `json:"id"`
			Title    string `json:"title"`
			ParentId string `json:"parentId"`
		}

		var data []Node
		// 直接将结果转换到数组对象中,要求是json tag作为参数传入
		err := xmind.SaveCustom(st, "id", "title", "parentId", "isRoot", &data)
		if err != nil {
			t.Fatal(err)
		}
		for i, d := range data {
			t.Logf("%d -> %s,%s,%s", i, d.Id, d.Title, d.ParentId)
		}
	})
}

package example

import (
	"testing"

	"github.com/jan-bar/xmind"
)

func TestLoadXmind(t *testing.T) {
	TestCreateXmind(t) // 调用该方法先生成xmind文件

	// 支持4种方式的加载,详情看内部具体实现
	wb, err := xmind.LoadFile("TestCreateXmind.xmind")
	if err != nil {
		t.Fatal(err)
	}
	if len(wb.Topics) != 2 {
		t.Fatal("sheet != 2")
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
		t.Fatal(err)
	}
}

func TestLoadXmindJson(t *testing.T) {
	wb, err := xmind.LoadFile(xmind.ContentJson)
	if err != nil {
		t.Fatal(err)
	}

	// 定位到中心主题添加子主题
	wb.Topics[0].On().Add("xxx").Add("yyy")
	err = wb.Save(xmind.ContentJson + ".xmind")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadXmindXml(t *testing.T) {
	wb, err := xmind.LoadFile(xmind.ContentXml)
	if err != nil {
		t.Fatal(err)
	}

	// 定位到中心主题添加子主题
	wb.Topics[0].On().Add("xxx").Add("yyy")
	err = wb.Save(xmind.ContentXml + ".xmind")
	if err != nil {
		t.Fatal(err)
	}
}

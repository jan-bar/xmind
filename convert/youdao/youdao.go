package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jan-bar/xmind"
)

// go run youdao.go test.mindmap test.xmind

func main() {
	if len(os.Args) != 3 {
		return
	}
	err := YouDao(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println(err)
	}
}

func YouDao(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	var node struct {
		Nodes json.RawMessage `json:"nodes"`
	}
	err = json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	// 有道云笔记思维导图,符合数组形式的结构,用自定义类型直接就可以转换
	st, err := xmind.LoadCustom([]byte(node.Nodes), "id", "topic", "parentid", "isroot")
	if err != nil {
		return err
	}

	return xmind.SaveSheets(dst, st)
}

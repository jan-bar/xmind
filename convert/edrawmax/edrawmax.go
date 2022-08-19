package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jan-bar/xmind"
)

func main() {
	if len(os.Args) != 3 {
		return
	}
	err := loadFile(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println(err)
	}
}

func loadFile(path, save string) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer zr.Close()

	var wb xmind.WorkBook
	const namePre = "pages/"
	for _, ed := range zr.File {
		if ed.Name == namePre || !strings.HasPrefix(ed.Name, namePre) {
			continue
		}

		er, err := ed.Open()
		if err != nil {
			return err
		}

		st, err := createSheet(er)
		if err != nil {
			return err
		}
		wb.Topics = append(wb.Topics, st)
	}

	return wb.Save(save)
}

func createSheet(r io.ReadCloser) (sheet *xmind.Topic, err error) {
	//goland:noinspection GoUnhandledErrorResult
	defer r.Close()

	var pp Page
	err = xml.NewDecoder(r).Decode(&pp)
	if err != nil {
		return
	}

	idMap := make(map[string]xmind.TopicID, len(pp.Shape))
	for _, shape := range pp.Shape {
		if len(shape.Texts.Text) > 0 {
			title := shape.Texts.Text[0].TextBlock.Text.Pp.Tp.Text
			// 不为空的节点才是思维导图的节点
			if title != "" {
				if shape.Type == "MainIdea" { // 表示主节点
					sheet = xmind.NewSheet(pp.Name, title)
					idMap[shape.ID] = xmind.CentKey // 建立中心主题ID映射关系
				} else {
					// 找到当前节点父节点的信息,根据id映射关系
					last := sheet.On(idMap[shape.LevelData.SuperLevel.V]).Add(title).Children.Attached
					idMap[shape.ID] = last[len(last)-1].ID
				}
			}
		}
	}
	return
}

type (
	Page struct {
		XMLName xml.Name `xml:"Page"`
		Name    string   `xml:"Name,attr"`
		Shape   []Shape  `xml:"Shape"`
	}
	Shape struct {
		Type      string `xml:"Type,attr"`
		ID        string `xml:"ID,attr"`
		LevelData struct {
			SuperLevel struct {
				V string `xml:"V,attr"`
			} `xml:"SuperLevel"`
		} `xml:"LevelData"`
		Texts struct {
			Text []struct {
				TextBlock struct {
					Text struct {
						Pp struct {
							Tp struct {
								Text string `xml:",chardata"`
							} `xml:"tp"`
						} `xml:"pp"`
					} `xml:"Text"`
				} `xml:"TextBlock"`
			} `xml:"Text"`
		} `xml:"Texts"`
	}
)

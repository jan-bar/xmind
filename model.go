package xmind

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"encoding/xml"
	"sync/atomic"
)

/*
下面的结构支持从xml和json中解析xmind文件
但只支持生成json的xmind文件,没有做直接生成xml文件的方法
*/
type (
	WorkBook struct {
		XMLName xml.Name `xml:"xmap-content"`
		Topics  []*Topic `json:"sheet" xml:"sheet"`
	}

	Topic struct {
		resources map[TopicID]*Topic // 记录所有主题的资源,所有主题共用同一个
		parent    *Topic             // 父节点地址
		incr      *int               // 只用于自增id,生成不重复的默认主题内容

		ID             TopicID        `json:"id" xml:"id,attr"`
		Title          string         `json:"title" xml:"title"`
		RootTopic      *Topic         `json:"rootTopic,omitempty" xml:"topic"`
		Style          Style          `json:"style"`
		StructureClass StructureClass `json:"structureClass,omitempty" xml:"structure-class,attr"`
		Children       *Children      `json:"children,omitempty" xml:"children,omitempty"`
	}

	TopicID string

	Style struct {
		Id         TopicID  `json:"id"`
		Properties struct{} `json:"properties"`
	}

	Children struct {
		Attached []*Topic `json:"attached"`
		Topics   struct {
			Topic []*Topic `xml:"topic"`
		} `json:"-" xml:"topics"`
	}

	StructureClass string
)

//goland:noinspection GoUnusedConst,SpellCheckingInspection
const (
	ContentJson = "content.json"
	ContentXml  = "content.xml"
	Manifest    = "manifest.json"
	Metadata    = "metadata.json"
	Thumbnails  = "Thumbnails"
	Resources   = "resources"

	StructMapUnbalanced       StructureClass = "org.xmind.ui.map.unbalanced"       // 思维导图
	StructMap                 StructureClass = "org.xmind.ui.map"                  // 平衡图(向下)
	StructMapClockwise        StructureClass = "org.xmind.ui.map.clockwise"        // 平衡图(顺时针)
	StructMapAnticlockwise    StructureClass = "org.xmind.ui.map.anticlockwise"    // 平衡图(逆时针)
	StructOrgChartDown        StructureClass = "org.xmind.ui.org-chart.down"       // 组织结构图(向下)
	StructOrgChartUp          StructureClass = "org.xmind.ui.org-chart.up"         // 组织结构图(向上)
	StructTreeRight           StructureClass = "org.xmind.ui.tree.right"           // 树状图(向右)
	StructTreeLeft            StructureClass = "org.xmind.ui.tree.left"            // 树状图(向左)
	StructLogicRight          StructureClass = "org.xmind.ui.logic.right"          // 逻辑图(向右)
	StructLogicLeft           StructureClass = "org.xmind.ui.logic.left"           // 逻辑图(向左)
	StructTimelineHorizontal  StructureClass = "org.xmind.ui.timeline.horizontal"  // 水平时间轴
	StructTimelineVertical    StructureClass = "org.xmind.ui.timeline.vertical"    // 垂直时间轴
	StructFishHoneLeftHeaded  StructureClass = "org.xmind.ui.fishbone.leftHeaded"  // 鱼骨图(头向左)
	StructFishHoneRightHeaded StructureClass = "org.xmind.ui.fishbone.rightHeaded" // 鱼骨图(头向左)
	StructSpreadsheet         StructureClass = "org.xmind.ui.spreadsheet"          // 矩阵(行)
	StructSpreadsheetColumn   StructureClass = "org.xmind.ui.spreadsheet.column"   // 矩阵(列)

	TopicIdLen = 26 // id编码长度 = objectIDbase32.EncodedLen(16)
)

//goland:noinspection SpellCheckingInspection
var (
	objectIDCounter uint32
	objectIDbase32  = base32.NewEncoding("123456789abcdefghijklmnopqrstuvw").WithPadding(base32.NoPadding)
)

func GetId() TopicID {
	id := make([]byte, 16+26)
	_, _ = rand.Reader.Read(id[:16])
	count := atomic.AddUint32(&objectIDCounter, 1)
	for i := 0; i < 4; i++ {
		if c := byte(count >> (i * 8)); c > 0 {
			id[8-i] = c
		}
	}
	// 16个随机字节(4位为自增id),确保不重复 -> 26个base32编码后字符
	objectIDbase32.Encode(id[16:], id[:16])
	return TopicID(id[16:])
}

func (t TopicID) MarshalJSON() ([]byte, error) {
	id := t
	if id == "" {
		id = GetId()
	}
	return []byte(`"` + id + `"`), nil
}

type aliasChildren Children

func (ch *Children) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, (*aliasChildren)(ch))
	if err != nil {
		return err
	}
	ch.Topics.Topic = ch.Attached
	return nil
}

func (ch *Children) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	err := d.DecodeElement((*aliasChildren)(ch), &start)
	if err != nil {
		return err
	}
	ch.Attached = ch.Topics.Topic
	return nil
}
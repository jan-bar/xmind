package xmind

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"encoding/xml"
	"strconv"
	"sync/atomic"
)

/*
下面的结构支持从xml和json中解析xmind文件
但只支持生成json的xmind文件,没有做直接生成xml文件的方法
*/
//goland:noinspection SpellCheckingInspection
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
		Labels         []string       `json:"labels,omitempty" xml:"labels,omitempty"`
		Notes          *Notes         `json:"notes,omitempty" xml:"notes,omitempty"`
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

	Notes struct {
		Plain ContentStruct `json:"plain" xml:"plain"`
	}

	ContentStruct struct {
		Content string `json:"content" xml:"content"`
	}
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
)

//goland:noinspection SpellCheckingInspection
var (
	objectIDCounter uint32
	objectIDbase32  = base32.NewEncoding("123456789abcdefghijklmnopqrstuvw").WithPadding(base32.NoPadding)
)

// topicIdDstLen = objectIDbase32.EncodedLen(topicIdSrcLen) // 提前计算长度
const topicIdSrcLen, topicIdDstLen = 16, 26

func GetId() TopicID {
	id := make([]byte, topicIdSrcLen+topicIdDstLen)
	_, _ = rand.Reader.Read(id[:topicIdSrcLen])
	count := atomic.AddUint32(&objectIDCounter, 1)
	for i := 0; i < 4; i++ {
		if c := byte(count >> (i * 8)); c > 0 {
			id[8-i] = c
		}
	}
	// 16个随机字节(4位为自增id),确保不重复 -> 26个base32编码后字符
	objectIDbase32.Encode(id[topicIdSrcLen:], id[:topicIdSrcLen])
	return TopicID(id[topicIdSrcLen:])
}

// IsOrdinary 普通节点ID长度固定为26,其他长度均为特殊节点
func (t TopicID) IsOrdinary() bool { return len(t) == topicIdDstLen }

func (t TopicID) MarshalJSON() ([]byte, error) {
	id := t
	if len(id) != topicIdDstLen {
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

// CustomIncrId 自定义生成自增数字id方案
func CustomIncrId() func(TopicID) string {
	cntId := 0
	cntMap := map[TopicID]string{}
	// 由于xmind的id生成比较长,这里改为自增数字
	return func(id TopicID) string {
		s, ok := cntMap[id]
		if ok {
			return s
		}
		cntId++
		s = strconv.Itoa(cntId)
		cntMap[id] = s
		return s
	}
}

package xmind

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// LoadFile 从文件加载xmind数据
// 当文件为
//    *.xmind 时会尝试读取压缩包的[content.json,content.xml]文件
//    *.*     时会直接按照[*.json,*.xml]这几种格式读取
//goland:noinspection GoUnhandledErrorResult
func LoadFile(path string) (*WorkBook, error) {
	fr, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fr.Close()

	fi, err := fr.Stat()
	if err != nil {
		return nil, err
	}

	var wb WorkBook
	defer func() {
		if len(wb.Topics) == 0 {
			return
		}
		sheets := make([]*Topic, 0, len(wb.Topics))
		// 通过文件加载的对象没有资源信息,因此在返回时手动添加
		for _, topic := range wb.Topics {
			if topic == nil || topic.RootTopic == nil {
				continue // 剔除不合法的数据
			}

			topic.RootTopic.parent = topic
			topic.RootTopic.resources = map[TopicID]*Topic{
				rootKey: topic,
				CentKey: topic.RootTopic,
				lastKey: topic.RootTopic,
			}
			topic.resources = topic.RootTopic.resources
			// 准备初始化数据,从中心主题开始更新所有子节点数据
			topic.RootTopic.upChildren()
			sheets = append(sheets, topic)
		}
		wb.Topics = sheets
	}()

	zr, err := zip.NewReader(fr, fi.Size())
	if err == nil {
		rz, err := zr.Open(ContentJson)
		if err == nil {
			defer rz.Close()
			err = json.NewDecoder(rz).Decode(&wb.Topics)
			if err == nil {
				return &wb, nil // 尝试读取zip中的content.json文件成功
			}
		}
		rz, err = zr.Open(ContentXml)
		if err == nil {
			defer rz.Close()
			err = xml.NewDecoder(rz).Decode(&wb)
			if err == nil {
				return &wb, nil // 尝试读取zip中的content.xml文件成功
			}
		}
	}

	seekFile := func(f func() error) error {
		_, err := fr.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		return f()
	}

	if err = seekFile(func() error {
		return json.NewDecoder(fr).Decode(&wb.Topics)
	}); err == nil {
		return &wb, nil // 尝试直接用json方式读取成功
	}

	if err = seekFile(func() error {
		return xml.NewDecoder(fr).Decode(&wb)
	}); err == nil {
		return &wb, nil // 尝试直接用xml方式读取成功
	}
	return nil, fmt.Errorf("can not read %s", path)
}

// LoadCustom 根据符合要求的任意结构加载
//  param
//    data:
//      方式1:
//        使用如下方式进行调用,根节点没有父节点,其他节点均设置父节点ID
//        LoadCustom([]Nodes{{"root","top"},{"123","one","root"}},"id","topic","parentId")
//        测试如下结构
//        type Nodes struct {
//           ID       string `json:"id"`
//           Topic    string `json:"topic"`
//           ParentId string `json:"parentId"`
//        }
//      方式2:
//        传json string: `[{"root","top"},{"123","one","root"}]`
//        传json []byte: []byte(`[{"root","top"},{"123","one","root"}]`)
//    idKey: 以该json tag字段作为主题ID
//    titleKey: 以该json tag字段作为主题内容
//    parentKey: 以该json tag字段作为判断父节点的依据
//  return
//    *Topic: 生成的主题地址
//    error: 返回错误
func LoadCustom(data interface{}, idKey, titleKey, parentKey string) (sheet *Topic, err error) {
	var byteData []byte
	switch td := data.(type) {
	case string:
		byteData = []byte(td)
	case []byte:
		byteData = td
	default:
		byteData, err = json.Marshal(data)
		if err != nil {
			return
		}
	}

	newStruct := func(name, tag string) reflect.StructField {
		return reflect.StructField{
			Name: name,
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(`json:"` + tag + `"`),
		}
	}
	// 动态创建一个结构体,并new该结构体数组的对象
	nodes := reflect.New(reflect.SliceOf(reflect.StructOf([]reflect.StructField{
		newStruct("Id", idKey),
		newStruct("Title", titleKey),
		newStruct("ParentId", parentKey),
	})))

	// 通过json库将传入对象转换为动态生成的对象
	err = json.Unmarshal(byteData, nodes.Interface())
	if err != nil {
		return
	}

	var (
		// 获取指针的对象值,相当于 *Type
		node    = nodes.Elem()
		nodeLen = node.Len()
		// 传入数据的ID和本次创建的ID建立映射关系,将第三方ID转换为这里生成的ID
		idMap = make(map[string]TopicID, nodeLen)
	)
	for i := 0; i < nodeLen; i++ {
		stu := node.Index(i)
		// 遍历数组每个数据,得到需要的数据
		// 动态创建结构三个字段index已知,用如下方法获取每个字段的数据
		id := stu.Field(0).String()
		title := stu.Field(1).String()
		parentId := stu.Field(2).String()

		if parentId == "" { // 中心主题父节点id为空
			sheet = NewSheet("sheet", title)
			idMap[id] = CentKey // 建立中心主题ID映射关系
		} else {
			last := sheet.On(idMap[parentId]).Add(title).Children.Attached
			// 将刚才添加的子主题ID建立映射关系,刚添加的子主题一定是最后一个
			idMap[id] = last[len(last)-1].ID
		}
	}
	return
}

func (wk *WorkBook) check() error {
	if wk == nil || len(wk.Topics) == 0 {
		return errors.New("WorkBook.Topics is null")
	}
	return nil
}

// Save 保存对象为 *.xmind 文件
func (wk *WorkBook) Save(path string) error {
	err := wk.check()
	if err != nil {
		return err
	}
	if filepath.Ext(path) != ".xmind" {
		return fmt.Errorf("%s: suffix must be .xmind", path)
	}

	cp := make([]*Topic, 0, len(wk.Topics))
	for _, topic := range wk.Topics {
		// 所有sheet全部切换到根节点,最终使用存入的cp生成xmind文件
		cp = append(cp, topic.On(rootKey))
	}

	fw, err := os.Create(path)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fw.Close()

	zw := zip.NewWriter(fw)
	//goland:noinspection GoUnhandledErrorResult
	defer zw.Close()

	wz, err := zw.Create(ContentJson)
	if err != nil {
		return err
	}
	return json.NewEncoder(wz).Encode(cp)
}

// SaveSheets 保存多个sheet画布到一个xmind文件
func SaveSheets(path string, sheet ...*Topic) error {
	return (&WorkBook{Topics: sheet}).Save(path)
}

// SaveCustom 自定义字段,将数据写入指定对象中
//  param
//    sheet: xmind的sheet数据
//    idKey: 以该json tag字段作为主题ID
//    titleKey: 以该json tag字段作为主题内容
//    parentKey: 以该json tag字段作为判断父节点的依据
//    v: 可以为 *string,*[]byte,*[]Nodes{} 这几种类型
//  return
//    error: 返回错误
func SaveCustom(sheet *Topic, idKey, titleKey, parentKey string, v interface{}) (err error) {
	cent := sheet.On(CentKey)
	if cent == nil {
		return errors.New("RootTopic is null")
	}

	var (
		buf   strings.Builder
		quote = make([]byte, 0, 128)
	)
	cent.Range(func(tp *Topic) {
		if tp.IsCent() {
			// 中心主题一般为数组第一个元素
			buf.WriteString(`[{"`)
			buf.WriteString(idKey)
			buf.WriteString(`":"`)
			buf.WriteString(string(cent.ID))
			buf.WriteString(`","`)
			buf.WriteString(titleKey)
			buf.WriteString(`":`)
			// 主题内容可能出现'\n','\t'等特殊字符,需要安全的方法在两侧添加引号
			buf.Write(strconv.AppendQuote(quote[:0], cent.Title))
			buf.WriteByte('}')
		} else {
			buf.WriteString(`,{"`)
			buf.WriteString(idKey)
			buf.WriteString(`":"`)
			buf.WriteString(string(tp.ID))
			buf.WriteString(`","`)
			buf.WriteString(titleKey)
			buf.WriteString(`":`)
			// 只有主题内容会出现特殊转义字符,需要特殊方式加引号
			buf.Write(strconv.AppendQuote(quote[:0], tp.Title))
			buf.WriteString(`,"`)
			buf.WriteString(parentKey)
			buf.WriteString(`":"`)
			buf.WriteString(string(tp.parent.ID))
			buf.WriteString(`"}`)
		}
	})
	buf.WriteByte(']')
	str := buf.String()

	// 根据不同类型设置数据
	switch vt := v.(type) {
	case *string:
		*vt = str
	case *[]byte:
		*vt = []byte(str)
	default:
		err = json.Unmarshal([]byte(str), v)
	}
	return
}

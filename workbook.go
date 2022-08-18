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
func LoadFile(path string) (*WorkBook, error) {
	fr, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
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
				centKey: topic.RootTopic,
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
			err = json.NewDecoder(rz).Decode(&wb.Topics)
			if err == nil {
				return &wb, nil // 尝试读取zip中的content.json文件成功
			}
		}
		rz, err = zr.Open(ContentXml)
		if err == nil {
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
//       使用如下方式进行调用,根节点没有父节点,其他节点均设置父节点ID
//       LoadCustom([]Nodes{{"root","top"},{"123","one","root"}},"id","topic","parentid")
//       测试如下结构
//       type Nodes struct {
//          ID       string `json:"id"`
//          Topic    string `json:"topic"`
//          Parentid string `json:"parentid"`
//       }
//    idKey: 以该json tag字段作为主题ID
//    titleKey: 以该json tag字段作为主题内容
//    parentKey: 以该json tag字段作为判断父节点的依据
//  return
//    *Topic: 生成的主题地址
//    error: 返回错误
func LoadCustom(data interface{}, idKey, titleKey, parentKey string) (sheet *Topic, err error) {
	vd := reflect.ValueOf(data)
	switch vd.Kind() { // 传入结构必须是切片或者数组
	case reflect.Slice, reflect.Array:
	default:
		err = errors.New("data is not Slice or Array")
		return
	}

	// 读入数据ID与创建的数据ID做映射
	idMap := make(map[string]TopicID, vd.Len())
	// 遍历切片或数组,读取每一个节点数据
	for i := 0; i < vd.Len(); i++ {
		val := vd.Index(i)
		if val.Kind() != reflect.Struct { // 每个节点必须是结构体
			err = errors.New("node not struct")
			return
		}

		var id, title, parentId string
		for vt, j := val.Type(), 0; j < vt.NumField(); j++ {
			var tmp *string
			tag := vt.Field(j).Tag.Get("json")
			if it := strings.Index(tag, ","); it > 0 {
				tag = tag[:it]
			}
			switch tag {
			case idKey:
				tmp = &id
			case titleKey:
				tmp = &title
			case parentKey:
				tmp = &parentId
			}

			if tmp != nil {
				v := val.Field(j)
				if v.Kind() != reflect.String {
					err = fmt.Errorf("%s field kind not String", tag)
					return
				}
				*tmp = v.String()
			}
		}

		if parentId == "" { // 中心主题父节点id为空
			sheet = NewSheet("sheet", title)
			idMap[id] = centKey // 建立中心主题ID映射关系
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
//    v: 可以为 *string,*[]byte,[]Nodes 这几种类型
//  return
//    error: 返回错误
func SaveCustom(sheet *Topic, idKey, titleKey, parentKey string, v interface{}) (err error) {
	cent := sheet.On(centKey)
	if cent == nil {
		return errors.New("RootTopic is null")
	}

	// 使用 strconv.AppendQuote 预防字符串中包含'\n','\"'之类的转义字符
	var buf strings.Builder
	quote := make([]byte, 0, 128) // 复用缓存
	buf.WriteString("[{")
	buf.Write(strconv.AppendQuote(quote[:0], idKey))
	buf.WriteByte(':')
	buf.Write(strconv.AppendQuote(quote[:0], string(cent.ID)))
	buf.WriteByte(',')
	buf.Write(strconv.AppendQuote(quote[:0], titleKey))
	buf.WriteByte(':')
	buf.Write(strconv.AppendQuote(quote[:0], cent.Title))
	buf.WriteByte('}')
	// 以上数据为中心主题节点

	// 通过递归依次写入所有子节点数据
	var loop func(topic *Topic)
	loop = func(topic *Topic) {
		if topic != nil && topic.Children != nil {
			for _, tp := range topic.Children.Attached {
				buf.WriteString(",{")
				buf.Write(strconv.AppendQuote(quote[:0], idKey))
				buf.WriteByte(':')
				buf.Write(strconv.AppendQuote(quote[:0], string(tp.ID)))
				buf.WriteByte(',')
				buf.Write(strconv.AppendQuote(quote[:0], titleKey))
				buf.WriteByte(':')
				buf.Write(strconv.AppendQuote(quote[:0], tp.Title))
				buf.WriteByte(',')
				buf.Write(strconv.AppendQuote(quote[:0], parentKey))
				buf.WriteByte(':')
				buf.Write(strconv.AppendQuote(quote[:0], string(tp.parent.ID)))
				buf.WriteByte('}')
				loop(tp)
			}
		}
	}
	loop(cent)
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

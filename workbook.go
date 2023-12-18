package xmind

import (
	"archive/zip"
	"bytes"
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
//
//	*.xmind 时会尝试读取压缩包的[content.json,content.xml]文件
//	*.*     时会直接按照[*.json,*.xml]这几种格式读取
func LoadFile(path string) (*WorkBook, error) {
	fr, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	wk, err := LoadFrom(fr)
	_ = fr.Close()
	if err != nil {
		return nil, fmt.Errorf("file: %q, %w", path, err)
	}
	return wk, nil
}

// LoadFrom 从文件或io.Reader对象中加载xmind
func LoadFrom(input any) (*WorkBook, error) {
	var (
		read interface {
			io.ReaderAt
			io.Seeker
			io.Reader
		}
		size int64
	)

	switch r := input.(type) {
	case *os.File:
		fi, err := r.Stat()
		if err != nil {
			return nil, err
		}
		// 读取文件对象
		read, size = r, fi.Size()
	case io.Reader:
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		// 处理更通用的io.Reader对象,会将所有数据读入内存
		read, size = bytes.NewReader(data), int64(len(data))
	default:
		return nil, errors.New("input type not supported")
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

			incr := 0
			topic.RootTopic.parent = topic
			topic.RootTopic.resources = map[TopicID]*Topic{
				rootKey: topic,
				CentKey: topic.RootTopic,
				lastKey: topic.RootTopic,
				incrKey: {incr: &incr},
			}
			topic.resources = topic.RootTopic.resources
			// 准备初始化数据,从中心主题开始更新所有子节点数据
			topic.RootTopic.upChildren()
			sheets = append(sheets, topic)
		}
		wb.Topics = sheets
	}()

	zr, err := zip.NewReader(read, size)
	if err == nil {
		rz, err := zr.Open(ContentJson)
		if err == nil {
			//goland:noinspection GoUnhandledErrorResult
			defer rz.Close()

			err = json.NewDecoder(rz).Decode(&wb.Topics)
			if err == nil {
				return &wb, nil // 尝试读取zip中的content.json文件成功
			}
		}

		rz, err = zr.Open(ContentXml)
		if err == nil {
			//goland:noinspection GoUnhandledErrorResult
			defer rz.Close()

			err = xml.NewDecoder(rz).Decode(&wb)
			if err == nil {
				return &wb, nil // 尝试读取zip中的content.xml文件成功
			}
		}
	}

	seekFile := func(f func() error) error {
		_, err = read.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		return f()
	}

	if err = seekFile(func() error {
		return json.NewDecoder(read).Decode(&wb.Topics)
	}); err == nil {
		return &wb, nil // 尝试直接用json方式读取成功
	}

	if err = seekFile(func() error {
		return xml.NewDecoder(read).Decode(&wb)
	}); err == nil {
		return &wb, nil // 尝试直接用xml方式读取成功
	}
	return nil, errors.New("can not read xmind")
}

const (
	CustomKeyId       = "Id"
	CustomKeyTitle    = "Title"
	CustomKeyParentId = "ParentId"
	CustomKeyIsRoot   = "IsRoot"
	CustomKeyLabels   = "Labels"
	CustomKeyNotes    = "Notes"
	CustomKeyBranch   = "Branch"
	CustomKeyHref     = "Href"
)

func fillCustom(custom map[string]string) map[string]string {
	if len(custom) == 0 {
		custom = make(map[string]string)
	}

	for _, v := range []string{CustomKeyId, CustomKeyTitle, CustomKeyParentId,
		CustomKeyLabels, CustomKeyNotes, CustomKeyBranch, CustomKeyHref} {
		if _, ok := custom[v]; !ok {
			// 没有传的参数填充默认值,tag标签用小写
			custom[v] = strings.ToLower(v)
		}
	}

	return custom
}

// LoadCustom 根据符合要求的任意结构加载
//
//	param
//	  data:
//	    方式1:
//	      使用如下方式进行调用,根节点没有父节点,其他节点均设置父节点ID
//	      LoadCustom([]Nodes{{"root","top"},{"123","one","root"}},map[string]string{
//	        CustomKeyId:       "id",
//	        CustomKeyTitle:    "topic",
//	        CustomKeyParentId: "parentId",
//	      })
//	      测试如下结构
//	      type Nodes struct {
//	         ID       string `json:"id"`
//	         Topic    string `json:"topic"`
//	         ParentId string `json:"parentId"`
//	         Labels []string `json:"labels"`
//	         Notes    string `json:"notes"`
//	      }
//	    方式2:
//	      传json string: data := `[
//	        {"a":"1","b":"main topic","labels":["l1","l2"],"notes":"notes"},
//	        {"a":"2","b":"topic1","c":"1"},
//	        {"a":"3","b":"topic2","c":"1"},
//	        {"a":"4","b":"topic3","c":"2"},
//	        {"a":"5","b":"topic4","c":"2"},
//	        {"a":"6","b":"topic5","c":"3"},
//	        {"a":"7","b":"topic6","c":"3"}]`
//	      LoadCustom(data,map[string]string{
//	        CustomKeyId:       "id",       // 以该json tag字段作为主题ID
//	        CustomKeyTitle:    "topic",    // 以该json tag字段作为主题内容
//	        CustomKeyParentId: "parentId", // 以该json tag字段作为判断父节点的依据
//	        CustomKeyIsRoot:   "isRoot",   // 以该json tag字段,bool类型,true表示根节点,false表示普通节点
//	        CustomKeyLabels:   "labels",   // 以该json tag字段作为主题标签
//	        CustomKeyNotes:    "notes",    // 以该json tag字段作为主题备注
//	        CustomKeyBranch:   "branch",   // 以该json tag字段作为主题折叠状态
//	        CustomKeyHref:     "href",     // 以该json tag字段作为主题超链接
//	      })
//	return
//	  *Topic: 生成的主题地址
//	  error: 返回错误
func LoadCustom(data any, custom map[string]string) (sheet *Topic, err error) {
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

	custom = fillCustom(custom)

	strType := reflect.TypeOf("")
	stuField := []reflect.StructField{
		{
			Name: CustomKeyId, Type: strType,
			Tag: reflect.StructTag(`json:"` + custom[CustomKeyId] + `"`),
		},
		{
			Name: CustomKeyTitle, Type: strType,
			Tag: reflect.StructTag(`json:"` + custom[CustomKeyTitle] + `"`),
		},
		{
			Name: CustomKeyParentId, Type: strType,
			Tag: reflect.StructTag(`json:"` + custom[CustomKeyParentId] + `"`),
		},
		{
			Name: CustomKeyLabels,
			Type: reflect.TypeOf([]string{}),
			Tag:  reflect.StructTag(`json:"` + custom[CustomKeyLabels] + `"`),
		},
		{
			Name: CustomKeyNotes, Type: strType,
			Tag: reflect.StructTag(`json:"` + custom[CustomKeyNotes] + `"`),
		},
		{
			Name: CustomKeyBranch, Type: strType,
			Tag: reflect.StructTag(`json:"` + custom[CustomKeyBranch] + `"`),
		},
		{
			Name: CustomKeyHref, Type: strType,
			Tag: reflect.StructTag(`json:"` + custom[CustomKeyHref] + `"`),
		},
	}

	isRootKey, hasRoot := custom[CustomKeyIsRoot]
	if hasRoot {
		stuField = append(stuField, reflect.StructField{
			Name: CustomKeyIsRoot,
			Type: reflect.TypeOf(true),
			Tag:  reflect.StructTag(`json:"` + isRootKey + `"`),
		})
	}

	// 动态创建一个结构体,并new该结构体数组的对象
	nodes := reflect.New(reflect.SliceOf(reflect.StructOf(stuField)))

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
		labels := stu.Field(3).Interface().([]string)
		notes := stu.Field(4).String()
		branch := stu.Field(5).String()
		href := stu.Field(6).String()

		// 优先根据IsRoot字段判断当前节点是根节点
		if (hasRoot && stu.Field(7).Bool()) || parentId == "" {
			sheet = NewSheet("sheet", title)
			idMap[id] = CentKey // 建立中心主题ID映射关系
		} else {
			find := sheet.On(idMap[parentId]).Add(title).AddLabel(labels...).
				AddHref(href).AddNotes(notes)
			if branch == folded {
				find.Folded() // 收缩主题
			}
			last := find.Children.Attached
			// 将刚才添加的子主题ID建立映射关系,刚添加的子主题一定是最后一个
			idMap[id] = last[len(last)-1].ID
		}
	}
	return
}

// LoadCustomWorkbook 加载自定义workbook的json
func LoadCustomWorkbook(input io.Reader, custom map[string]string) (*WorkBook, error) {
	var data []json.RawMessage
	err := json.NewDecoder(input).Decode(&data)
	if err != nil {
		return nil, err
	}

	tps := make([]*Topic, len(data))
	for i, v := range data {
		tps[i], err = LoadCustom([]byte(v), custom)
		if err != nil {
			return nil, err
		}
	}
	return &WorkBook{Topics: tps}, nil
}

func (wk *WorkBook) check() error {
	if wk == nil || len(wk.Topics) == 0 {
		return errors.New("WorkBook.Topics is null")
	}
	return nil
}

// SaveTo 将xmind保存到io.Writer对象,使用更灵活
func (wk *WorkBook) SaveTo(w io.Writer) error {
	err := wk.check()
	if err != nil {
		return err
	}

	cp := make([]*Topic, 0, len(wk.Topics))
	for _, topic := range wk.Topics {
		// 所有sheet全部切换到根节点,最终使用存入的cp生成xmind文件
		cp = append(cp, topic.On(rootKey))
	}

	zw := zip.NewWriter(w)
	//goland:noinspection GoUnhandledErrorResult
	defer zw.Close()

	wz, err := zw.Create(ContentJson)
	if err != nil {
		return err
	}
	return json.NewEncoder(wz).Encode(cp)
}

// Save 保存对象为 *.xmind 文件
func (wk *WorkBook) Save(path string) error {
	if filepath.Ext(path) != ".xmind" {
		return fmt.Errorf("%s: suffix must be .xmind", path)
	}

	fw, err := os.Create(path)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fw.Close()

	return wk.SaveTo(fw)
}

// SaveSheets 保存多个sheet画布到一个xmind文件
func SaveSheets(path string, sheet ...*Topic) error {
	return (&WorkBook{Topics: sheet}).Save(path)
}

var RootIsNull = errors.New("RootTopic is null")

// SaveCustom 自定义字段,将数据写入指定对象中
//
//	param
//	  sheet: xmind的sheet数据
//	  custom: map[string]string{
//	    CustomKeyId:       "id",     // 以该json tag字段作为主题ID
//	    CustomKeyTitle:    "title",  // 以该json tag字段作为主题内容
//	    CustomKeyParentId: "parent", // 以该json tag字段作为判断父节点的依据
//	        // "parentId",表示根节点不添加父节点id
//	        // "parentId,xx",表示根节点添加值为空的父节点id
//	    CustomKeyIsRoot: "isRoot",   // 以该json tag字段,true表示为根节点
//	        // "",表示所有节点都不添加
//	        // "isRoot,xx",表示只添加根节点
//	    CustomKeyLabels: "labels", // 以该json tag字段作为标签
//	    CustomKeyNotes:  "notes",  // 以该json tag字段作为备注
//	  }
//	  v: 可以为 *string,*[]byte,*[]Nodes{} 这几种类型
//	  genId: 外部自定义生成id方案,自动生成的id是参照xmind,可能有点长
//	return
//	  error: 返回错误
func SaveCustom(sheet *Topic, custom map[string]string, v any,
	genId func(id TopicID) string) error {
	cent := sheet.On()
	if !cent.IsCent() {
		return RootIsNull
	}

	var (
		buf   bytes.Buffer
		quote = make([]byte, 0, 128)
		rk    = 0
	)
	custom = fillCustom(custom)

	isRootKey, ok := custom[CustomKeyIsRoot]
	if ok {
		isRootKey, _, ok = strings.Cut(isRootKey, ",")
		if ok {
			rk = 1
		} else {
			rk = 3
		}
	}

	var (
		idKey     = custom[CustomKeyId]
		titleKey  = custom[CustomKeyTitle]
		parentKey = custom[CustomKeyParentId]
		labelsKey = custom[CustomKeyLabels]
		notesKey  = custom[CustomKeyNotes]
		branchKey = custom[CustomKeyBranch]
		hrefKey   = custom[CustomKeyHref]
	)
	parentKey, _, ok = strings.Cut(parentKey, ",")

	_ = cent.Range(func(_ int, tp *Topic) error {
		isCent := tp.IsCent()
		if isCent {
			buf.WriteString(`[{"`) // 中心主题为数组第一个元素
		} else {
			buf.WriteString(`,{"`)
		}

		buf.WriteString(idKey)
		buf.WriteString(`":"`)
		if genId != nil {
			buf.WriteString(genId(tp.ID))
		} else {
			buf.WriteString(string(tp.ID))
		}
		buf.WriteString(`","`)
		buf.WriteString(titleKey)
		buf.WriteString(`":`)
		// 内容可能有需要转义字符,用安全的方式添加引号
		buf.Write(strconv.AppendQuote(quote[:0], tp.Title))

		if isCent {
			if ok { // 中心主题添加值为空的数据
				buf.WriteString(`,"`)
				buf.WriteString(parentKey)
				buf.WriteString(`":""`)
			}

			if rk&1 != 0 { // 添加isRoot字段
				buf.WriteString(`,"`)
				buf.WriteString(isRootKey)
				buf.WriteString(`":true`)
			}
		} else {
			buf.WriteString(`,"`)
			buf.WriteString(parentKey)
			buf.WriteString(`":"`)
			if genId != nil {
				buf.WriteString(genId(tp.parent.ID))
			} else {
				buf.WriteString(string(tp.parent.ID))
			}
			buf.WriteByte('"')

			if rk&2 != 0 { // 添加isRoot字段
				buf.WriteString(`,"`)
				buf.WriteString(isRootKey)
				buf.WriteString(`":false`)
			}
		}

		buf.WriteString(`,"`)
		buf.WriteString(notesKey)
		buf.WriteString(`":`) // 添加备注
		if tp.Notes != nil && tp.Notes.Plain.Content != "" {
			buf.Write(strconv.AppendQuote(quote[:0], tp.Notes.Plain.Content))
		} else {
			buf.WriteString(`""`)
		}

		buf.WriteString(`,"`)
		buf.WriteString(branchKey)
		buf.WriteString(`":`) // 添加折叠状态
		buf.Write(strconv.AppendQuote(quote[:0], tp.Branch))

		buf.WriteString(`,"`)
		buf.WriteString(hrefKey)
		buf.WriteString(`":`) // 添加超链接
		buf.Write(strconv.AppendQuote(quote[:0], tp.Href))

		buf.WriteString(`,"`)
		buf.WriteString(labelsKey)
		buf.WriteString(`":[`) // 添加标签
		for i, vl := range tp.Labels {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(strconv.AppendQuote(quote[:0], vl))
		}
		buf.WriteString("]}")
		return nil
	})
	buf.WriteByte(']')

	// 根据不同类型设置数据
	switch vt := v.(type) {
	case *string:
		*vt = buf.String()
	case *[]byte:
		*vt = buf.Bytes()
	default:
		return json.Unmarshal(buf.Bytes(), v)
	}
	return nil
}

// SaveCustomWorkbook 保存workbook到自定义json中
func SaveCustomWorkbook(output io.Writer, wk *WorkBook,
	custom map[string]string, genId func(id TopicID) string) error {
	tmp := []byte{'['}
	_, err := output.Write(tmp)
	if err != nil {
		return err
	}

	tmp[0] = ','
	var data []byte
	for i, tp := range wk.Topics {
		if i > 0 {
			_, err = output.Write(tmp)
			if err != nil {
				return err
			}
		}

		err = SaveCustom(tp, custom, &data, genId)
		if err != nil {
			return err
		}

		_, err = output.Write(data)
		if err != nil {
			return err
		}
	}

	tmp[0] = ']'
	_, err = output.Write(tmp)
	return err
}

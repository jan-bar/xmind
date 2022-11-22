package xmind

import (
	"fmt"
)

const (
	rootKey TopicID = "root" // 画布主题地址Key
	CentKey TopicID = ""     // 中心主题地址Key,开放给调用者
	lastKey TopicID = "last" // 最后一次编辑主题地址Key
	incrKey TopicID = "incr" // 自增主题key
)

// NewSheet 创建一个画布
//  param
//    sheetTitle: 画布名称
//    centralTopicTitle: 中心主题
//    structureClass: 整体样式
//  return
//    *Topic: 中心主题地址
func NewSheet(sheetTitle, centralTopicTitle string, structureClass ...StructureClass) *Topic {
	sc := StructLogicRight
	if len(structureClass) > 0 {
		sc = structureClass[0]
	}

	resources := make(map[TopicID]*Topic) // 所有主题共用同一个
	sheet := &Topic{
		ID:    GetId(),
		Title: sheetTitle,
		RootTopic: &Topic{
			ID:             GetId(),
			Title:          centralTopicTitle,
			StructureClass: sc,
			resources:      resources,
		},
		resources: resources,
	}
	sheet.RootTopic.parent = sheet       // 赋值中心主题的父节点
	resources[rootKey] = sheet           // 赋值根节点
	resources[CentKey] = sheet.RootTopic // 为空的key表示中心主题
	resources[lastKey] = sheet.RootTopic // 将中心主题赋值为最后编辑节点
	incr := 0
	resources[incrKey] = &Topic{incr: &incr} // 自增主题ID
	return sheet.RootTopic                   // 返回中心主题节点
}

// UpSheet 更新画布,可以在任何节点主题执行
//  param
//    sheetTitle: 画布名称
//    centralTopicTitle: 中心主题
//    structureClass: 整体样式
func (st *Topic) UpSheet(sheetTitle, centralTopicTitle string, structureClass ...StructureClass) {
	if st == nil {
		return
	}

	root, ok := st.resources[rootKey]
	if ok {
		root.Title = sheetTitle
		root.RootTopic.Title = centralTopicTitle
		if len(structureClass) > 0 {
			root.RootTopic.StructureClass = structureClass[0]
		}
	}
}

// On 根据主题ID切换主题地址
//  param
//    componentId: 主题ID,不传时切换到中心主题
//  return
//    *Topic: 匹配主题地址
func (st *Topic) On(componentId ...TopicID) *Topic {
	if st == nil || st.resources == nil {
		return st // 资源为空只可能是使用者直接使用 Topic 对象,尽量使用接口
	}
	cid := CentKey
	if len(componentId) > 0 {
		cid = componentId[0]
	}

	topic, ok := st.resources[cid]
	if ok {
		st.resources[lastKey] = topic
		return topic
	}
	return st.resources[lastKey]
}

// OnTitle 根据主题内容切换主题地址
//  param
//    title: 主题内容,为空时切换到中心主题
//  return
//    *Topic: 匹配主题地址
func (st *Topic) OnTitle(title string) *Topic {
	return st.On(st.CId(title)) // 两个操作合并为一个,方便使用
}

// Parent 返回父节点地址,如果传参则返回指定ID的父节点
// 找不到父主题,或父主题为nil时需要外部自行判断
func (st *Topic) Parent(componentId ...TopicID) *Topic {
	if st == nil {
		return st
	}

	// 返回当前节点的父节点
	if len(componentId) == 0 {
		return st.parent
	}

	// 返回指定节点的父节点
	topic, ok := st.resources[componentId[0]]
	if ok {
		return topic.parent
	}
	return nil
}

type AddMode uint8

const (
	SubMode    AddMode = iota // 默认方式,当前主题添加子主题
	BeforeMode                // 在当前主题之前插入
	AfterMode                 // 在当前主题之后插入
	ParentMode                // 为当前主题插入父主题
)

// In 判断 AddMode 不在范围内时重置为 SubMode
func (am AddMode) In() AddMode {
	for i := SubMode; i <= ParentMode; i++ {
		if am == i {
			return am
		}
	}
	return SubMode
}

// Add 为当前主题添加主题
//  param
//    title: 主题内容
//    mode: 添加主题方式,不传则默认添加子主题
//  return
//    *Topic: 当前主题地址
func (st *Topic) Add(title string, modes ...AddMode) *Topic {
	if st == nil || st.parent == nil {
		// 父节点为nil表示当前节点在root根节点,该节点不支持添加子主题
		// 没有对外提供切换到根节点方法,除非外部直接使用 Topic 对象
		return st
	}

	mode := SubMode
	if len(modes) > 0 {
		mode = modes[0].In()
	}

	if title == "" {
		id, ok := st.resources[incrKey]
		if ok {
			*id.incr++ // 增加空内容主题时,自动生成自增的主题内容,确保主题不重复
			title = fmt.Sprintf("Topic %d", *id.incr)
		}
	}

	id := GetId()
	tp := &Topic{ID: id, Title: title, resources: st.resources, parent: st}
	tp.resources[id] = tp

	// 添加子主题,当前节点为中心主题时不管啥选项都是添加子主题
	if mode == SubMode || st == st.resources[CentKey] {
		if st.Children == nil {
			st.Children = &Children{Attached: []*Topic{tp}}
		} else {
			st.Children.Attached = append(st.Children.Attached, tp)
		}
		return st
	}

	// 当前节点插入父主题
	if mode == ParentMode {
		st.Title, tp.Title = tp.Title, st.Title // 不用关心资源
		tp.Children = st.Children
		st.Children = &Children{Attached: []*Topic{tp}}
		if tp.Children != nil && len(tp.Children.Attached) > 0 {
			for _, tc := range tp.Children.Attached {
				tc.parent = tp // 所有该级子节点更新父节点指针
			}
		}
		// 由于st,tp交换,所以这里返回tp,保证当前位置还是之前的定位
		return st.On(tp.ID)
	}

	tp.parent = st.parent // 下面只有2种同级插入方式,更新该节点父节点信息
	if st.parent.Children == nil {
		st.parent.Children = &Children{Attached: []*Topic{tp}}
		return st // 应该没有这种情况,保险而已
	}
	tps := append(st.parent.Children.Attached, tp)

	if mode == BeforeMode {
		for i := len(tps) - 1; i > 0; i-- {
			tps[i], tps[i-1] = tps[i-1], tps[i]
			if tps[i].ID == st.ID {
				break // 当前节点前插入主题
			}
		}
	} else if mode == AfterMode {
		for i := len(tps) - 1; i > 0; i-- {
			if tps[i-1].ID == st.ID {
				break // 当前节点后插入主题
			}
			tps[i], tps[i-1] = tps[i-1], tps[i]
		}
	}

	st.parent.Children.Attached = tps
	return st
}

// Move 将指定节点移动到当前节点对应位置
//  param
//    componentId: 要移动过来的节点
//    modes: 移动过来的添加方式,不传则默认移动为最后一个子主题
//  return
//    *Topic: 当前主题地址
func (st *Topic) Move(componentId TopicID, modes ...AddMode) *Topic {
	if st == nil || st.parent == nil || !componentId.IsOrdinary() {
		return st // 同 Add 根节点不支持操作,内部的特殊节点不支持移动操作
	}

	mode := SubMode
	if len(modes) > 0 {
		mode = modes[0].In()
		if mode == ParentMode {
			mode = SubMode // 移动方式不支持将节点移动为当前节点父节点
		}
	}

	src, ok := st.resources[componentId]
	if !ok {
		return st // 找不到节点,无法移动
	}
	parent := src.parent
	if parent == nil || parent.Children == nil || len(parent.Children.Attached) == 0 {
		return st // 被移动节点没有父节点,或者父节点没有子节点(貌似没这情况,以防万一)
	}

	for p := st; p.parent != nil; p = p.parent {
		if p.parent == src {
			return st // 被移动的节点是当前节点的祖辈节点,不支持被移动
		}
	}

	cur := 0
	for i, tp := range parent.Children.Attached {
		if tp.ID != src.ID {
			parent.Children.Attached[cur] = parent.Children.Attached[i]
			cur++ // 注意不能直接用tp赋值,range的坑
		}
	}
	if cur == len(parent.Children.Attached) {
		return st // 没有找到要移动的节点
	}
	// 在父节点的子节点中移除需要移动的节点
	if cur == 0 {
		parent.Children = nil
	} else {
		parent.Children.Attached = parent.Children.Attached[:cur]
	}

	// 添加子主题,当前节点为中心主题时不管啥选项都是移动到子主题
	if mode == SubMode || st == st.resources[CentKey] {
		src.parent = st // 更新被移动节点的父节点为当前节点
		if st.Children == nil {
			st.Children = &Children{Attached: []*Topic{src}}
		} else {
			st.Children.Attached = append(st.Children.Attached, src)
		}
		return st
	}

	src.parent = st.parent // 下面只有2种同级插入方式,更新该节点父节点信息
	if st.parent.Children == nil {
		st.parent.Children = &Children{Attached: []*Topic{src}}
		return st // 应该没有这种情况,保险而已
	}
	tps := append(st.parent.Children.Attached, src)

	if mode == BeforeMode {
		for i := len(tps) - 1; i > 0; i-- {
			tps[i], tps[i-1] = tps[i-1], tps[i]
			if tps[i].ID == st.ID {
				break // 当前节点前插入主题
			}
		}
	} else if mode == AfterMode {
		for i := len(tps) - 1; i > 0; i-- {
			if tps[i-1].ID == st.ID {
				break // 当前节点后插入主题
			}
			tps[i], tps[i-1] = tps[i-1], tps[i]
		}
	}

	st.parent.Children.Attached = tps
	return st
}

// Remove 删除指定主题内容节点
//  param
//    title: 待删除子主题内容
//  return
//    *Topic: 当前主题地址
func (st *Topic) Remove(title string) *Topic {
	return st.RemoveByID(st.CId(title))
}

// RemoveByID 删除指定主题ID的节点
//  param
//    title: 待删除子主题内容
//  return
//    *Topic: 当前主题地址
// 特别注意,删除主题成功会自动定位到中心主题上,如果需要切换需要显示使用 On 操作
func (st *Topic) RemoveByID(componentId TopicID) *Topic {
	if st == nil || !componentId.IsOrdinary() {
		return st // 特殊主题不允许删除
	}

	topic := st.Parent(componentId)
	if topic == nil || topic.Children == nil || len(topic.Children.Attached) == 0 {
		return st
	}

	cur := 0 // 找到需要删除节点父节点地址,遍历所有子节点并删除匹配项
	for i, tp := range topic.Children.Attached {
		if tp.ID != componentId {
			topic.Children.Attached[cur] = topic.Children.Attached[i]
			cur++ // 注意不能直接用tp赋值,range的坑
		} else {
			delete(st.resources, tp.ID) // 删除当前节点
			tp.RemoveChildren()         // 递归删除子节点
		}
	}
	if cur == len(topic.Children.Attached) {
		return st // 没有匹配删除直接返回
	}

	if cur == 0 {
		topic.Children = nil
	} else {
		topic.Children.Attached = topic.Children.Attached[:cur]
	}
	// 存在删除时,需要切换到中心主题上,避免在已删除节点执行后续逻辑
	return st.On(CentKey)
}

// RemoveChildren 递归删除所有子节点
func (st *Topic) RemoveChildren() {
	if st != nil && st.Children != nil {
		for _, tp := range st.Children.Attached {
			delete(st.resources, tp.ID)
			tp.RemoveChildren()
		}
		st.Children = nil
	}
}

// 为节点所有子节点添加父节点地址指针,并且更新资源数据
func (st *Topic) upChildren() {
	if st != nil && st.Children != nil {
		for _, tp := range st.Children.Attached {
			if !tp.ID.IsOrdinary() {
				tp.ID = GetId() // 生成正常ID
			}
			st.resources[tp.ID] = tp
			tp.parent, tp.resources = st, st.resources
			tp.upChildren() // 递归更新所有子节点资源
		}
	}
}

// CId 根据主题内容获取第一个匹配到的主题ID
//  param
//    title: 主题内容
//  return
//    TopicID: 匹配title的主题ID,有多个相同title时只返回第一个匹配成功的结果
func (st *Topic) CId(title string) TopicID {
	if title == "" {
		return CentKey
	}

	if st != nil {
		for id, topic := range st.resources {
			// 由于range遍历乱序因此,不保证存在多个title时按照之前添加顺序返回
			if id.IsOrdinary() && topic.Title == title {
				return id // 判断ID长度,剔除特殊ID
			}
		}
	}
	return lastKey // 匹配不到返回最后一次编辑的主题ID
}

// CIds 根据主题内容获取所有匹配到的主题ID
//  param
//    title: 主题内容
//  return
//    res: 匹配到title的所有主题ID
func (st *Topic) CIds(title string) (res []TopicID) {
	if title == "" {
		return []TopicID{CentKey} // 默认返回一个中心主题
	}

	if st != nil {
		for id, topic := range st.resources {
			if id.IsOrdinary() && topic.Title == title {
				res = append(res, id) // 判断ID长度,剔除特殊ID
			}
		}
	}
	if len(res) == 0 {
		return []TopicID{lastKey} // 匹配不到返回最后编辑主题
	}
	return res
}

// IsCent 判断当前节点是中心主题
//  return
//    bool: true表示该节点为中心主题,否则为普通节点
func (st *Topic) IsCent() bool {
	return st != nil && st == st.resources[CentKey]
}

// Range 从当前节点递归遍历子节点
//  param
//    f: 外部的回调
func (st *Topic) Range(f func(*Topic)) {
	if st != nil {
		var loop func(*Topic)
		loop = func(tp *Topic) {
			if tp != nil {
				f(tp) // 通过回调函数让调用者实现自己的逻辑
				if tp.Children != nil && len(tp.Children.Attached) > 0 {
					for _, tc := range tp.Children.Attached {
						loop(tc)
					}
				}
			}
		}
		if st.RootTopic != nil {
			loop(st.RootTopic) // 当前为根节点,从中心主题开始遍历
		} else {
			loop(st) // 其他情况从当前节点开始遍历
		}
	}
}

// Resources 返回所有主题的ID和内容资源
//  return
//    res: 返回所有主题ID和资源
func (st *Topic) Resources() map[TopicID]string {
	res := make(map[TopicID]string)
	if st != nil {
		for id, topic := range st.resources {
			if id.IsOrdinary() {
				res[id] = topic.Title // 只返回正常的节点资源
			}
		}
	}
	return res
}

// AddLabel 在当前主题上加label标签
//  param
//    label: 标签内容
//  return
//    *Topic: 当前主题地址
func (st *Topic) AddLabel(label ...string) *Topic {
	st.Labels = label
	return st
}

// AddNotes 在当前主题上加notes备注
//  param
//    Notes: 备注内通内容
//  return
//    *Topic: 当前主题地址
func (st *Topic) AddNotes(notes Notes) *Topic {
	st.Notes = &notes
	return st
}

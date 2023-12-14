## 命令行工具

执行命令`./main config.json`,将按照配置文件进行各种转换工作

不做命令行参数是为了避免终端的各种转义问题,用配置文件也方便保存和修改

指定具体xmind文件保存为custom自定义json格式
```json
{
  "from": "file:../example/content.json.xmind",
  "fromType": "xmind",
  "fromCustom": {},
  "to": "../convert/out",
  "toType": "custom",
  "toCustom": {
    "Id": "id",
    "Title": "title",
    "ParentId": "parentId",
    "IsRoot": "isRoot",
    "Labels": "labels",
    "Notes": "notes"
  },
  "toMarkdown": {}
}
```

指定具体xmind文件保存为markdown文件
```json
{
  "from": "file:../example/content.json.xmind",
  "fromType": "xmind",
  "fromCustom": {},
  "to": "../convert/out",
  "toType": "markdown",
  "toCustom": {},
  "toMarkdown": {
    "default": "{{Repeat \"#\" .Deep}} {{.Title}}\n\n{{range $i,$v := .Labels}}> {{$v}}\n\n{{end}}{{range $i,$v := (SplitLines .Notes \"\\n\\r\")}}> {{$v}}\n\n{{end}}"
  }
}
```

设置通配符匹配xmind文件保存为markdown文件
```json
{
  "from": "dir:../example/*.xmind",
  "fromType": "xmind",
  "fromCustom": {},
  "to": "../convert/out",
  "toType": "markdown",
  "toCustom": {},
  "toMarkdown": {
    "default": "{{Repeat \"#\" .Deep}} {{.Title}}\n\n{{range $i,$v := .Labels}}> {{$v}}\n\n{{end}}{{range $i,$v := (SplitLines .Notes \"\\n\\r\")}}> {{$v}}\n\n{{end}}"
  }
}
```

递归通配符匹配xmind文件保存为markdown文件
```json
{
  "from": "recursive:../example/*.xmind",
  "fromType": "xmind",
  "fromCustom": {},
  "to": "../convert/out",
  "toType": "markdown",
  "toCustom": {},
  "toMarkdown": {
    "default": "{{Repeat \"#\" .Deep}} {{.Title}}\n\n{{range $i,$v := .Labels}}> {{$v}}\n\n{{end}}{{range $i,$v := (SplitLines .Notes \"\\n\\r\")}}> {{$v}}\n\n{{end}}"
  }
}
```

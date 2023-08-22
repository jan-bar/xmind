package xmind

import (
	"io"
	"strconv"
	"strings"
	"text/template"
)

const (
	DefaultMarkdownName   = "default"
	DefaultMarkdownFormat = "{{Repeat \"#\" .Deep}} {{.Title}}\n\n{{range $i,$v := .Labels}}> {{$v}}\n\n{{end}}{{if .Notes}}> {{.Notes}}\n\n{{end}}"

	MarkdownKeyDeep = "Deep" // 所在层级,>=1
)

func (wk *WorkBook) SaveToMarkdown(w io.Writer, format map[string]string) error {
	defText, ok := format[DefaultMarkdownName]
	if !ok {
		defText = DefaultMarkdownFormat
	}

	tpl, err := template.New(DefaultMarkdownName).Funcs(template.FuncMap{
		"Repeat": strings.Repeat, // 注册用到的方法
	}).Parse(defText)
	if err != nil {
		return err
	}

	for k, v := range format {
		// 对每个层级创建自定义渲染模板
		if _, err = tpl.New(k).Parse(v); err != nil {
			return err
		}
	}

	for _, tp := range wk.Topics {
		err = tp.Range(func(deep int, current *Topic) error {
			data := map[string]interface{}{
				MarkdownKeyDeep: deep,
				CustomKeyTitle:  current.Title,
			}
			if len(current.Labels) > 0 {
				data[CustomKeyLabels] = current.Labels
			}
			if current.Notes != nil && current.Notes.Plain.Content != "" {
				data[CustomKeyNotes] = current.Notes.Plain.Content
			}

			tw := tpl.Lookup(strconv.Itoa(deep))
			if tw == nil {
				tw = tpl.Lookup(DefaultMarkdownName)
			}
			return tw.Execute(w, data)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jan-bar/xmind"
)

func main() {
	var read io.Reader
	if len(os.Args) > 1 {
		fr, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		//goland:noinspection GoUnhandledErrorResult
		defer fr.Close()
		read = fr
	} else {
		// 支持标准输入读取配置
		// 可以这样执行命令: cat config.json | ./main
		read = os.Stdin
	}

	// 之所以用json配置文件,是为了避免命令行的各种转义问题
	var config struct {
		// "from": "file:xxx.xmind",直接指定单个文件
		// "from": "dir:/a/b/*.xmind",通配符匹配目录下文件
		// "from": "recursive:/a/b/*.xmind",此时输入为[文件夹/通配符]
		From string `json:"from"`
		// "fromType": "xmind",按照xmind方式读取文件
		// "fromType": "custom",按照custom自定义json方式读取文件
		FromType string `json:"fromType"`
		// "fromType": "custom" 时,这里生效的字段名配置项
		FromCustom map[string]string `json:"fromCustom"`
		// "to": "",为空则保存文件为 src-time.ext
		// "to": "/path"
		//    如果读取为file或dir模式,则保存文件为 /path/base(src).ext
		//    如果读取为recursive模式,则保存文件会创建相同层级路径 /path/rel(src).ext
		To string `json:"to"`
		// "fromType": "xmind",按照xmind方式保存文件
		// "fromType": "custom",按照custom自定义json方式保存文件
		// "fromType": "markdown",按照markdown方式保存文件
		ToType string `json:"toType"`
		// "fromType": "custom" 时需要用到的自定义json字段配置
		ToCustom map[string]string `json:"toCustom"`
		// "fromType": "markdown" 时需要用到的自定义markdown配置
		ToMarkdown map[string]string `json:"toMarkdown"`
	}

	err := json.NewDecoder(read).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	// 根据规则查找需要转换的所有文件
	base, files, err := findFiles(config.From)
	if err != nil {
		log.Fatal(err)
	}

	// 根据配置,设置读取文件方案
	var load func(string) (*xmind.WorkBook, error)
	switch config.FromType {
	default:
		fallthrough
	case "xmind": // 默认按照xmind文件读取
		load = xmind.LoadFile
	case "custom": // 按自定义方案读取
		load = func(path string) (*xmind.WorkBook, error) {
			fr, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			//goland:noinspection GoUnhandledErrorResult
			defer fr.Close()

			return xmind.LoadCustomWorkbook(fr, config.FromCustom)
		}
	}

	var (
		// 根据配置,设置保存文件方案
		save    func(*xmind.WorkBook, string) error
		saveExt string // 不同方式保存文件后缀名不同
	)
	switch config.ToType {
	default:
		fallthrough
	case "xmind": // 默认保存xmind文件
		save = func(wk *xmind.WorkBook, path string) error {
			return wk.Save(path)
		}
		saveExt = ".xmind"
	case "custom":
		save = func(wk *xmind.WorkBook, path string) error {
			fw, err := os.Create(path)
			if err != nil {
				return err
			}
			//goland:noinspection GoUnhandledErrorResult
			defer fw.Close()

			return xmind.SaveCustomWorkbook(fw, wk, config.ToCustom, nil)
		}
		saveExt = ".json"
	case "markdown": // 保存为markdown文件
		save = func(wk *xmind.WorkBook, path string) error {
			fw, err := os.Create(path)
			if err != nil {
				return err
			}
			//goland:noinspection GoUnhandledErrorResult
			defer fw.Close()

			return wk.SaveToMarkdown(fw, config.ToMarkdown)
		}
		saveExt = ".md"
	}

	// 根据配置方式,设置生成保存文件路径方法
	var genTo func(string) (string, error)
	if config.To != "" {
		err = os.MkdirAll(config.To, 0666)
		if err != nil {
			log.Fatal(err)
		}

		if base == "" {
			genTo = func(s string) (string, error) {
				// 非递归方式,所有文件都保存在目标文件夹同级
				return filepath.Join(config.To, filepath.Base(s)+saveExt), nil
			}
		} else {
			genTo = func(s string) (string, error) {
				rel, err := filepath.Rel(base, s)
				if err != nil {
					return "", err
				}
				tDir := filepath.Join(config.To, filepath.Dir(rel))
				if err = os.MkdirAll(tDir, 0666); err != nil {
					return "", err
				}
				// 递归查找出来的文件,保存时也放到目标目录的相对路径下
				return filepath.Join(tDir, filepath.Base(s)+saveExt), nil
			}
		}
	} else {
		genTo = func(s string) (string, error) {
			return fmt.Sprintf("%s-%s%s", s,
				time.Now().Format("20060102150405"), saveExt), nil
		}
	}

	for _, v := range files {
		wk, err := load(v) // 按配置加载文件
		if err != nil {
			log.Fatal(err)
		}

		to, err := genTo(v) // 按配置得到保存文件路径
		if err != nil {
			log.Fatal(err)
		}

		err = save(wk, to) // 按配置保存文件
		if err != nil {
			log.Fatal(err)
		}
	}
}

func findFiles(s string) (base string, files []string, err error) {
	if tp, input, ok := strings.Cut(s, ":"); ok {
		switch tp {
		case "file":
			// "from": "file:xxx.xmind",直接指定单个文件
			files = []string{input}
			return
		case "dir":
			// "from": "dir:/a/b/*.xmind",通配符匹配目录下文件
			files, err = filepath.Glob(input)
			return
		case "recursive":
			// "from": "recursive:/a/b/*.xmind",此时输入为[文件夹/通配符]
			var pattern string
			base, pattern = filepath.Split(input)
			err = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
				if d.IsDir() || err != nil {
					return err
				}

				ok, err = filepath.Match(pattern, d.Name())
				if ok {
					files = append(files, path)
				}
				return err
			})
			return
		}
	}
	err = fmt.Errorf("from %q illegal", s)
	return
}

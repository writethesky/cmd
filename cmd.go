package cmd

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type CMD struct {
	params        map[string]*Param
	isHelp        bool
	index         uint
	stopLoadingCh chan uint
	usageTitle    string
	loadingTitle  string
}

type Param struct {
	Name    string
	Value   string
	Usage   string
	Require bool
	Type    ParamType
	Options map[string]string

	sort      uint
	boolValue bool
	intValue  int
}

type ParamType uint

const (
	ParamTypeString ParamType = iota
	ParamTypeInt
	ParamTypeOption
	ParamTypeBool
)

func stringToParamType(str string) ParamType {
	switch str {
	case "string":
		return ParamTypeString
	case "int":
		return ParamTypeInt
	case "option":
		return ParamTypeOption
	case "bool":
		return ParamTypeBool
	default:
		return ParamTypeString
	}
}

func NewCMD(usageTitle string, loadingTitle string) *CMD {
	obj := new(CMD)
	obj.params = make(map[string]*Param)
	obj.usageTitle = usageTitle
	obj.loadingTitle = loadingTitle
	obj.stopLoadingCh = make(chan uint)
	return obj
}

func (c *CMD) Parse(params interface{}) (isOk bool) {
	tElem, vElem := c.bindParams(params)
	isOk = true
	for i, v := range c.params {
		if v.Type == ParamTypeBool {
			flag.BoolVar(&c.params[i].boolValue, v.Name, false, v.Usage)
		} else {
			flag.StringVar(&c.params[i].Value, v.Name, v.Value, v.Usage)
		}

	}
	flag.BoolVar(&c.isHelp, "h", false, "this help")
	flag.Parse()
	if c.isHelp {
		c.PrintUsage()
	}

	for i, v := range c.params {
		if v.Require && v.Value == "" {
			isOk = false
			printError(*v)
		}

		if v.Value != "" {
			switch v.Type {
			case ParamTypeInt:
				intValue, err := strconv.Atoi(v.Value)
				if nil != err {
					isOk = false
					printIntError(*v)
				}
				c.params[i].intValue = intValue
			case ParamTypeOption:
				_, isOk = v.Options[v.Value]
				if !isOk {
					isOk = false
					printOptionError(*v)
				}
			}
		}

		for j := 0; j < vElem.NumField(); j++ {
			if tElem.Field(j).Tag.Get("name") == i {
				switch v.Type {
				case ParamTypeInt:
					vElem.Field(j).SetInt(int64(v.intValue))
				case ParamTypeBool:
					vElem.Field(j).SetBool(v.boolValue)
				case ParamTypeOption:
					vElem.Field(j).SetString(v.Value)
				case ParamTypeString:
					vElem.Field(j).SetString(v.Value)
				}

			}

		}
	}
	if !isOk {
		return
	}

	go printLoading(c.stopLoadingCh, c.loadingTitle)
	return

}

func (c *CMD) Get(name string) string {
	return c.params[name].Value
}
func (c *CMD) GetInt(name string) int {
	return c.params[name].intValue
}

func (c *CMD) GetBool(name string) bool {
	return c.params[name].boolValue
}

func (c *CMD) Set(param Param) {
	param.sort = c.index
	c.index++
	c.params[param.Name] = &param
}

func (c *CMD) PrintUsage() {

	fmt.Println(c.usageTitle)
	for i := uint(0); i < c.index; i++ {
		var v *Param
		for _, vvv := range c.params {
			if vvv.sort == i {
				v = vvv
				break
			}

		}
		typeStr := "字符串"
		switch v.Type {
		case ParamTypeInt:
			typeStr = "数字　"
		case ParamTypeOption:
			typeStr = "选项　"
		case ParamTypeBool:
			typeStr = "布尔　"
		}
		requireStr := "非必须"
		if v.Require {
			requireStr = "必须　"
		}
		optionStr := ""
		if len(v.Options) != 0 {
			optionStr = "["
			for k, v := range v.Options {
				optionStr += fmt.Sprintf(" %s - %s ", getPrintColor(k, 1), getPrintColor(v, 3))
			}
			optionStr += "]"
		}
		fmt.Printf("-%s 　%s %s | %s %s\n", getPrintColor(v.Name, 1), getPrintColor(requireStr, 3), getPrintColor(typeStr, 4), v.Usage, optionStr)

	}

	os.Exit(0)
}

func printLoading(ch <-chan uint, loadingTitle string) {
	fmt.Printf("%s|", loadingTitle)
	loadingChar := []string{"-", "\\", "|", "/", "-", ".|"}
	index := 0
	for {
		select {
		case <-ch:
			fmt.Print("\033[2J")
			return
		default:
			time.Sleep(time.Millisecond * 300)

			fmt.Printf("\033[1D%s", loadingChar[index%6])
			index++
		}

	}
}

func (c *CMD) StopLoading() {
	c.stopLoadingCh <- 1
}
func printError(param Param) {
	fmt.Printf("参数 -%s 是必须的 | %s\n", getPrintRed(param.Name), param.Usage)
}

func printIntError(param Param) {
	fmt.Printf("参数 -%s 应该是一个数字 | %s\n", getPrintRed(param.Name), param.Usage)
}

func printOptionError(param Param) {

	optionsStr := "["
	for k, v := range param.Options {
		optionsStr += fmt.Sprintf(" %s - %s ", getPrintRed(k), getPrintColor(v, 3))
	}
	optionsStr += "]"
	fmt.Printf("参数 -%s 只能是特定的值 | %s：%s\n", getPrintRed(param.Name), param.Usage, optionsStr)
}

func getPrintRed(s string) string {
	return fmt.Sprintf("\033[38;5;1;5m%s\033[0m", s)
}

func getPrintColor(s string, color uint) string {
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", color, s)
}

func GetPrintColorWithBG(s string, color uint) string {
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", color, s)
}

func (c *CMD) bindParams(params interface{}) (reflect.Type, reflect.Value) {
	t := reflect.TypeOf(params)
	v := reflect.ValueOf(params)
	//fmt.Println(t)
	//fmt.Println(v.Elem())
	//v.Elem().Field(0).SetString("111")
	//t.Elem().Field(0).Tag.Get("name")
	tElem := t.Elem()
	vElem := v.Elem()
	for i := 0; i < v.Elem().NumField(); i++ {
		c.Set(Param{
			Name:    tElem.Field(i).Tag.Get("name"),
			Usage:   tElem.Field(i).Tag.Get("usage"),
			Require: tElem.Field(i).Tag.Get("require") == "true",
			Type:    stringToParamType(tElem.Field(i).Tag.Get("type")),
			Options: stringToOptions(tElem.Field(i).Tag.Get("options")),
		})
	}
	return tElem, vElem
}

func stringToOptions(str string) (options map[string]string) {
	options = make(map[string]string)
	itemArr := strings.Split(str, ",")
	for _, item := range itemArr {

		keyValue := strings.Split(item, ":")
		if 2 != len(keyValue) {
			continue
		}
		options[keyValue[0]] = keyValue[1]
	}
	return
}

package cmd

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type Params struct {
	User      string `name:"u" usage:"用户名" require:"true"`
	Password  string `name:"p" usage:"密码" require:"true"`
	LocalPort int    `name:"P" usage:"本地端口" type:"int"`
	Mode      string `name:"m" usage:"模式" require:"true" type:"option" options:"global:全局,rule:规则,auto:自动"`
	Verbose   bool   `name:"v" usage:"详细信息" type:"bool"`
}

func TestCmd(t *testing.T) {

	tests := map[string]struct {
		flags []string
		isOk  bool
	}{
		"default": {
			flags: []string{"-u=123", "-p=123", "-m=global", "-v", "-P=122"},
			isOk:  true,
		},
	}

	for name, rc := range tests {
		Convey(name, t, func() {
			cmd := NewCMD("haha", "loading")
			params := new(Params)
			os.Args = append(os.Args, rc.flags...)

			isOk := cmd.Parse(params)
			So(isOk, ShouldEqual, rc.isOk)
			So(params.User, ShouldEqual, "123")
			So(params.Verbose, ShouldBeTrue)
			So(params.LocalPort, ShouldEqual, 122)
			So(params.Mode, ShouldEqual, "global")
		})
	}

}

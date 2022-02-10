package cmd

import (
	"github.com/imind-lab/greeter/cmd/cron"
	"github.com/spf13/cobra"
	"log"
	"reflect"
	"strings"
)

// 计划任务方法需要幂等
var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "show greeter cronjob sample",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			c := cron.New()
			vf := reflect.ValueOf(c)

			target := strings.Title(args[0])
			method := vf.MethodByName(target)

			if method.IsValid() {
				method.Call([]reflect.Value{})
			} else {
				log.Println("指定的计划任务方法不存在")
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(cronCmd)
}

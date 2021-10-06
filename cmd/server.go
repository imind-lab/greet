package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/imind-lab/greeter/server"
)

var cfgFile string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the gRPC Greeter server",
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recover error : %v\n", err)
			}
		}()
		err := server.Serve()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./conf/conf.yaml", "Start server with provided configuration file")
	rootCmd.AddCommand(serverCmd)
	cobra.OnInitialize(initConf)
}

func initConf() {
	viper.SetConfigFile(cfgFile)
	//初始化全部的配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

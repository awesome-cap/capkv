/**
2 * @Author: Nico
3 * @Date: 2021/6/10 0:12
4 */
package config

import (
	"github.com/awesome-cap/dkv/log"
	"github.com/spf13/viper"
)

var Config = serverConfig{}

type serverConfig struct {
	Port uint16
	Address []string
	Storage int
	WriteDiskType uint8
	WriteDiskInterval uint16
}

func init()  {
	viper.SetConfigFile("abc")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil{
		log.Fatal(err)
		return
	}
	err = viper.GetViper().UnmarshalExact(&Config)
	if err != nil{
		log.Fatal(err)
		return
	}
	log.Info("Loaded config successful.")
}



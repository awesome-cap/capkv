package config

type Config struct {
	Port uint16
	Address []string
	Storage int
	WriteDiskType uint8
	WriteDiskInterval uint16
}
//
//func init()  {
//	viper.SetConfigFile("abc")
//	viper.SetConfigType("yaml")
//	err := viper.ReadInConfig()
//	if err != nil{
//		log.Fatal(err)
//		return
//	}
//	err = viper.GetViper().UnmarshalExact(&Config)
//	if err != nil{
//		log.Fatal(err)
//		return
//	}
//}



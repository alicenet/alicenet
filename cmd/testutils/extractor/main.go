package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
)

func main() {
	network := flag.String("n", "", "network name.")
	configPath := flag.String("p", "", "file path.")
	flag.Parse()

	viper.SetConfigName("factoryState")
	viper.SetConfigType("toml")
	viper.AddConfigPath(*configPath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	addr := viper.GetString(fmt.Sprintf("%s.defaultFactoryAddress", *network))
	fmt.Println(addr)
}

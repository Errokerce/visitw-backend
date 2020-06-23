package misc

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

const filename = "config.yml"

var (
	Config ConfigStruct
)

type (
	ConfigStruct struct {
		HashSecret string `yaml:"hash_secret"`
		JwtSecret  string `yaml:"jwt_secret"`
		Port       int    `yaml:"port"`
		Hostname   string `yaml:"host_name"`
		Sms        Sms    `yaml:"sms"`
	}
	Sms struct {
		Account  string `yaml:"account"`
		Password string `yaml:"password"`
	}
)

func init() {
	ReadConfig()
}
func WriteConfig() {
	println("Config File init")
	nc := ConfigStruct{}
	yb, _ := yaml.Marshal(nc)
	_ = ioutil.WriteFile(filename, yb, 0666)
	Config = nc
}
func ReadConfig() {
	yb, err := ioutil.ReadFile(filename)
	if err != nil {
		println("Config File not found")
		WriteConfig()
	}
	err = yaml.Unmarshal(yb, &Config)
	if err != nil {
		println("Read Config error plz check")
		os.Exit(1)
	}
}

//func main() {
//	//yaml.Encoder{&configStruct{HashSecret: "secret",Sms: {Account: "acc",Password:"pass"}}}
//	var cf = &configStruct{HashSecret: "secret",Sms: {Account: "acc", Password: "pass"}}
//
//	yaml.Encoder
//}

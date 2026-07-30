package config

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var Instance *Config

type Config struct {
	Env        string `yaml:"Env"`        // 环境：prod、dev
	BaseUrl    string `yaml:"BaseUrl"`    // base url
	Port       string `yaml:"Port"`       // 端口
	LogFile    string `yaml:"LogFile"`    // 日志文件
	ShowSql    bool   `yaml:"ShowSql"`    // 是否显示日志
	StaticPath string `yaml:"StaticPath"` // 静态文件目录

	LoginMethods struct {
		Passport bool `yaml:"passport"`
		Password bool `yaml:"password"`
	} `yaml:"LoginMethods"`

	MinIO struct {
		Endpoint        string `yaml:"Endpoint"`
		AccessKeyID     string `yaml:"AccessKeyID"`
		SecretAccessKey string `yaml:"SecretAccessKey"`
		UseSSL          bool   `yaml:"UserSSL"`
		BucketLocation  string `yaml:"BucketLocation"`
	} `yaml:"MinIO"`

	// 数据库配置
	DB struct {
		Url          string `yaml:"Url"`
		MaxIdleConns int    `yaml:"MaxIdleConns"`
		MaxOpenConns int    `yaml:"MaxOpenConns"`
	} `yaml:"DB"`

	// 数据库配置
	OldDB struct {
		Url string `yaml:"Url"`
	} `yaml:"OldDB"`

	// smtp
	Smtp struct {
		Host     string `yaml:"Host"`
		Port     string `yaml:"Port"`
		Username string `yaml:"Username"`
		Password string `yaml:"Password"`
		SSL      bool   `yaml:"SSL"`
	} `yaml:"Smtp"`
}

func Init(filename string) *Config {
	Instance = &Config{}
	Instance.LoginMethods.Passport = true
	if yamlFile, err := ioutil.ReadFile(filename); err != nil {
		logrus.Error(err)
	} else if err = yaml.Unmarshal(yamlFile, Instance); err != nil {
		logrus.Error(err)
	}
	return Instance
}

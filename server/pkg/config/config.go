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

	Redis struct {
		Url      string `yaml:"Url"`
		Password string `yaml:"Password"`
		DB       int    `yaml:"DB"`
	} `yaml:"Redis"`

	Presence struct {
		AllowedOrigins        []string `yaml:"AllowedOrigins"`
		TrustedProxies        []string `yaml:"TrustedProxies"`
		MaxConnections        int      `yaml:"MaxConnections"`
		MaxConnectionsPerIP   int      `yaml:"MaxConnectionsPerIP"`
		MaxConnectionsPerUser int      `yaml:"MaxConnectionsPerUser"`
	} `yaml:"Presence"`

	// 搜索（Meilisearch）
	Search struct {
		Enabled bool   `yaml:"Enabled"` // 是否启用 Meilisearch，关闭或不可用时自动回退到数据库 LIKE
		Url     string `yaml:"Url"`     // Meilisearch 地址，如 http://127.0.0.1:7700
		ApiKey  string `yaml:"ApiKey"`  // API Key，只允许后端访问
		Index   string `yaml:"Index"`   // 索引名
		// 权限模式：
		// filter：搜索结果只返回当前用户有权访问的主题（默认）
		// mark：返回全部命中主题，由渲染层隐藏无权访问的内容
		PermissionMode string `yaml:"PermissionMode"`
		// 单页返回条数上限，请求超过该值会被截断，防止 limit 被放大后打满数据库/搜索引擎
		MaxLimit int `yaml:"MaxLimit"`
		// 最大可翻页页码，超过该页码直接返回空结果，避免深分页扫全索引
		MaxPage int `yaml:"MaxPage"`
		// 单次搜索最多从搜索引擎/数据库扫描的候选条数上限（权限过滤会消耗候选）
		MaxScan int `yaml:"MaxScan"`
		// 单次请求超时（毫秒）
		TimeoutMs int `yaml:"TimeoutMs"`
		// 启动时是否异步全量重建新站索引
		SyncOnBoot bool `yaml:"SyncOnBoot"`
		// 是否把旧 BBS 数据也同步进索引（数据量大时建议按需开启）
		SyncOldBBS bool `yaml:"SyncOldBBS"`
	} `yaml:"Search"`

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
	Instance.Presence.MaxConnections = 10000
	Instance.Presence.MaxConnectionsPerIP = 200
	Instance.Presence.MaxConnectionsPerUser = 5
	if yamlFile, err := ioutil.ReadFile(filename); err != nil {
		logrus.Error(err)
	} else if err = yaml.Unmarshal(yamlFile, Instance); err != nil {
		logrus.Error(err)
	}
	if Instance.Presence.MaxConnections <= 0 {
		Instance.Presence.MaxConnections = 10000
	}
	if Instance.Presence.MaxConnectionsPerIP <= 0 {
		Instance.Presence.MaxConnectionsPerIP = 200
	}
	if Instance.Presence.MaxConnectionsPerUser <= 0 {
		Instance.Presence.MaxConnectionsPerUser = 5
	}
	if Instance.Search.Url == "" {
		Instance.Search.Url = "http://127.0.0.1:7700"
	}
	if Instance.Search.Index == "" {
		Instance.Search.Index = "topics"
	}
	if Instance.Search.PermissionMode == "" {
		Instance.Search.PermissionMode = "filter"
	}
	if Instance.Search.MaxLimit <= 0 {
		Instance.Search.MaxLimit = 50
	}
	if Instance.Search.MaxPage <= 0 {
		Instance.Search.MaxPage = 100
	}
	if Instance.Search.MaxScan <= 0 {
		Instance.Search.MaxScan = 1000
	}
	if Instance.Search.TimeoutMs <= 0 {
		Instance.Search.TimeoutMs = 3000
	}
	return Instance
}

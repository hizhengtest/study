package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"os"
	"strconv"
	"strings"
	"time"
)

//当前应用配置
type appInfo struct {
	ServerName       string
	ServerAddr       string
	ProductType      string
	PortInnerGrpc    uint64 //容器内grpc端口
	PortInnerHttp    uint64 //容器内http端口
	PortOutGrpc      uint64 //映射的宿主机grpc端口
	PortOutHttp      uint64 //映射的宿主机http端口
	LogDir           string
	CacheDir         string
	Debug            bool
	NacosAddr        string
	NacosPort        uint64
	NacosScheme      string
	NacosContextPath string
	NacosNameSpace   string
	NacosDataID      string
	NacosDataGroup   string
	NacosUser        string
	NacosPassword    string
	NacosLogLevel    string
}

// Config config
type Config struct {
	Database   databaseInfo
	Redis      redisInfo
	Rabbitmq   rabbitmq
	Aliyun     aliyun
	MobilePool mobilepool
	Service    service
	Crontab    crontab
}

type aliyun struct {
	DefAccessKeyID     string
	DefAccessKeySecret string
}

//mysql数据库配置
type databaseInfo struct {
	Type        string
	Name        string
	User        string
	Password    string
	Host        string
	PoolIdleNum int
	PoolOpenNum int
}

//redis配置
type redisInfo struct {
	Addr        string
	Password    string
	Db          int
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
}

//rabbitmq配置
type rabbitmq struct {
	ConnectionString string
}

//号码池配置
type mobilepool struct {
	MobileNumberFormatReg string
	CncmNumberFormatReg   string
	CncuNumberFormatReg   string
	CnctNumberFormatReg   string
	MaxRecvMsgSize        int
}

//定时任务
type crontab struct {
	ClearNotUsedTimeSet string //定时清理未被使用的号码
	ClearNotUsedTimeBeg int64  //定时清理未被使用的号码开始时间(秒)
	ClearExpiredTimeSet string //定时清理过期号码时间
	ClearExpiredTimeBeg int64  //定时清理过期号码开始时间(秒)
}

//内部服务配置
type service struct {
	BasicServiceName string //基础服务名称
}

// BasicConfig 最基础的配置
type BasicConfig struct {
	App    appInfo
	Aliyun aliyun
}

const (
	//PortInnerGrpc 默认grpc协议端口（容器内）
	PortInnerGrpc = 8080
	//PortInnerHttp 默认http协议端口（容器内）
	PortInnerHttp = 8088
)

// Default config remote setting
var Default Config

// Basic conifg local setting
var Basic BasicConfig

// ConfigCenter 注册中心
var ConfigCenter configCenter

//配置中心客户端
var configClient config_client.IConfigClient

//服务中心客户端
var namingClient naming_client.INamingClient

func init() {
	var err error

	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")

	//是否从配置文件获取配置
	cnfFromEnv := os.Getenv("CONFIG_FROM_ENV")

	if len(cnfFromEnv) > 0 {
		Basic.App.ServerName = os.Getenv("SERVER_NAME")
		Basic.App.ServerAddr = os.Getenv("SERVER_ADDR")
		Basic.App.ProductType = os.Getenv("PRODUCT_TYPE")
		portGrpc, _ := strconv.Atoi(os.Getenv("PORT_GRPC"))
		Basic.App.PortOutGrpc = uint64(portGrpc)
		portHttp, _ := strconv.Atoi(os.Getenv("PORT_HTTP"))
		Basic.App.PortOutHttp = uint64(portHttp)
		Basic.App.PortInnerGrpc = uint64(PortInnerGrpc)
		Basic.App.PortInnerHttp = uint64(PortInnerHttp)
		Basic.App.LogDir = os.Getenv("LOG_DIR")
		Basic.App.CacheDir = os.Getenv("CACHE_DIR")
		Basic.App.NacosAddr = os.Getenv("NACOS_ADDR")
		nacosPort, _ := strconv.Atoi(os.Getenv("NACOS_PORT"))
		Basic.App.NacosPort = uint64(nacosPort)
		Basic.App.NacosContextPath = os.Getenv("NACOS_CONTEXT_PATH")
		Basic.App.NacosNameSpace = os.Getenv("NACOS_NAME_SPACE")
		Basic.App.NacosDataID = os.Getenv("NACOS_DATA_ID")
		Basic.App.NacosDataGroup = os.Getenv("NACOS_DATA_GROUP")
		Basic.App.NacosUser = os.Getenv("NACOS_USER")
		Basic.App.NacosPassword = os.Getenv("NACOS_PASSWORD")
		Basic.App.NacosLogLevel = os.Getenv("NACOS_LOG_LEVEL")
		getDebug := os.Getenv("DEBUG")
		Basic.App.Debug = false
		if strings.ToLower(getDebug) == "true" {
			Basic.App.Debug = true
		}
	} else {

		cnfPath := "../config"
		fmt.Println("cnfPath:", cnfPath)
		viper.AddConfigPath(cnfPath)
		viper.SetConfigName("config_local")

		fmt.Println("cnfPath:", cnfPath)
		viper.AddConfigPath(cnfPath)
		viper.SetConfigName("config_local")

		err = viper.ReadInConfig()
		if err != nil {
			fmt.Println("Failed to load local configuration file", err)
			return
		}
		err = viper.Unmarshal(&Basic)
		if err != nil {
			fmt.Println("Parsing local configuration file failed", err)
			return
		}
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			err = viper.Unmarshal(&Basic)
			if err != nil {
				fmt.Println("Configuration changes, parsing local configuration file failed", err)
				return
			}
			fmt.Println("Config file changed:", e.Name, Basic)
		})
	}

	//配置configClient
	sc := []constant.ServerConfig{
		{
			IpAddr: Basic.App.NacosAddr,
			Port:   Basic.App.NacosPort,
		},
	}
	cc := constant.ClientConfig{
		NamespaceId:         Basic.App.NacosNameSpace, //namespace id
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              Basic.App.LogDir,
		CacheDir:            Basic.App.CacheDir,
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            Basic.App.NacosLogLevel,
		Username:            Basic.App.NacosUser,
		Password:            Basic.App.NacosPassword,
	}
	// a more graceful way to create config client
	configClient, err = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		fmt.Println("连接配置中心失败:", err, " ", time.Now().Format("2006-01-02 15:04:05"))
		panic(err)
	}

	//配置nameClient
	sc = []constant.ServerConfig{
		{
			IpAddr: Basic.App.NacosAddr,
			Port:   Basic.App.NacosPort,
		},
	}
	cc = constant.ClientConfig{
		NamespaceId:         Basic.App.NacosNameSpace, //namespace id
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              Basic.App.LogDir,
		CacheDir:            Basic.App.CacheDir,
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            Basic.App.NacosLogLevel,
		Username:            Basic.App.NacosUser,
		Password:            Basic.App.NacosPassword,
	}

	namingClient, err = clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})

	if err != nil {
		fmt.Println("连接服务中心失败:", err, " ", time.Now().Format("2006-01-02 15:04:05"))
		panic(err)
	}

	//get config
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: Basic.App.NacosDataID,
		Group:  Basic.App.NacosDataGroup,
	})
	if err != nil {
		fmt.Println("获取远程配置失败:", err, " ", time.Now().Format("2006-01-02 15:04:05"))
		panic(err)
	}

	err = yaml.Unmarshal([]byte(content), &Default)
	if err != nil {
		fmt.Println("解析远程配置中心失败", err, " ", time.Now().Format("2006-01-02 15:04:05"))
		panic(err)
	}

	//Listen config change,key=dataId+group+namespaceId.
	err = configClient.ListenConfig(vo.ConfigParam{
		DataId: Basic.App.NacosDataID,
		Group:  Basic.App.NacosDataGroup,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("ListenConfig group:" + group + ", dataId:" + dataId + ", data:...")
			err = yaml.Unmarshal([]byte(data), &Default)
			if err != nil {
				fmt.Println("监听远程配置中心失败", err, " ", time.Now().Format("2006-01-02 15:04:05"))
				return
			}
		},
	})

	if err != nil {
		fmt.Println("监听远程配置中心初始化失败", err, " ", time.Now().Format("2006-01-02 15:04:05"))
		panic(err)
	}
}

//服务注册客户端
type configCenter struct {
}

//获取配置注册中心
func (s *configCenter) GetConfigClient() (config_client.IConfigClient, error) {
	return configClient, nil;
}

//获取服务注册中心
func (s *configCenter) GetNamingClient() (naming_client.INamingClient, error) {
	return namingClient,nil
}

//获取一个服务
func (s *configCenter) GetOneService(name string, clusters []string) (*model.Instance, error) {
	namingClient, err := ConfigCenter.GetNamingClient()
	if err != nil {
		fmt.Printf("连接服务中心失败:%s|%s time:%s \n", name, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		return nil, err
	}
	instance, err := namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: name,
		Clusters:    clusters,
	})
	if err != nil {
		fmt.Printf("获取健康服务实例错误:%s|%s time:%s \n", name, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		return nil, err
	}
	return instance, nil
}

//获取多个服务
func (s *configCenter) GetAllService(name string, clusters []string) ([]model.Instance, error) {
	namingClient, err := ConfigCenter.GetNamingClient()
	if err != nil {
		fmt.Printf("连接服务中心失败:%s|%s time:%s \n", name, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		return nil, err
	}
	instance, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: name,
		Clusters:    clusters,
	})
	if err != nil {
		fmt.Printf("获取所有服务实例错误:%s|%s time:%s \n", name, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		return nil, err
	}
	return instance, nil
}

//注册服务
func (s *configCenter) RegisterServer() {
	fmt.Println("开始注册服务", "|time:", time.Now().Format("2006-01-02 15:04:05"))

	client, err := s.GetNamingClient()
	if err != nil {
		fmt.Println("连接注册中心失败:", err, "|time:", time.Now().Format("2006-01-02 15:04:05"))
		panic(err)
	}

	//注册grpc服务
	if Basic.App.PortOutGrpc > 0 {
		param := vo.RegisterInstanceParam{
			Ip:          Basic.App.ServerAddr,
			Port:        Basic.App.PortOutGrpc,
			ServiceName: fmt.Sprintf("%s.%s", Basic.App.ServerName, "grpc"),
			Weight:      10,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata:    map[string]string{},
		}
		success, err := client.RegisterInstance(param)
		if err != nil {
			fmt.Println("连接注册中心失败:", err, "|time:", time.Now().Format("2006-01-02 15:04:05"))
			panic(err)
		}
		fmt.Printf("RegisterServiceInstance,param:%+v,result:%+v \n\n", param, success)
	}

	//注册http服务
	if Basic.App.PortOutHttp > 0 {
		param := vo.RegisterInstanceParam{
			Ip:          Basic.App.ServerAddr,
			Port:        Basic.App.PortOutHttp,
			ServiceName: fmt.Sprintf("%s.%s", Basic.App.ServerName, "http"),
			Weight:      10,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata:    map[string]string{},
		}
		success, err := client.RegisterInstance(param)
		if err != nil {
			fmt.Println("连接注册中心失败:", err, "|time:", time.Now().Format("2006-01-02 15:04:05"))
			panic(err)
		}
		fmt.Printf("RegisterServiceInstance,param:%+v,result:%+v \n\n", param, success)
	}
}

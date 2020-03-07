package models

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/JojiiOfficial/DataManagerServer/constants"

	gaw "github.com/JojiiOfficial/GoAw"
	"github.com/JojiiOfficial/configService"
	log "github.com/sirupsen/logrus"
)

//Config config for the server
type Config struct {
	Server configServer

	Webserver struct {
		MaxHeaderLength      uint  `default:"8000" required:"true"`
		MaxBodyLength        int64 `default:"10000" required:"true"`
		MaxPayloadBodyLength int64 `default:"10000" required:"true"`
		HTTP                 configHTTPstruct
		HTTPS                configTLSStruct
	}
}
type configServer struct {
	Database          configDBstruct
	PathConfig        pathConfig
	AllowRegistration bool `default:"false"`
}

type pathConfig struct {
	FileStore string `required:"true"`
}

type configDBstruct struct {
	Host         string
	Username     string
	Database     string
	Pass         string
	DatabasePort int
}

//Config for HTTPS
type configTLSStruct struct {
	Enabled       bool   `default:"false"`
	ListenAddress string `default:":443"`
	CertFile      string
	KeyFile       string
}

//Config for HTTP
type configHTTPstruct struct {
	Enabled       bool   `default:"false"`
	ListenAddress string `default:":80"`
}

//GetDefaultConfig gets the default config path
func GetDefaultConfig() string {
	return path.Join(constants.DataDir, constants.DefaultConfigFile)
}

//InitConfig inits the config
//Returns true if system should exit
func InitConfig(confFile string, createMode bool) (*Config, bool) {
	var config Config
	if len(confFile) == 0 {
		confFile = GetDefaultConfig()
	}

	s, err := os.Stat(confFile)
	if createMode || err != nil {
		if createMode {
			if s != nil && s.IsDir() {
				log.Fatalln("This name is already taken by a folder")
				return nil, true
			}
			if !strings.HasSuffix(confFile, ".yml") {
				log.Fatalln("The configFile must end with .yml")
				return nil, true
			}
		}

		//Autocreate folder
		path, _ := filepath.Split(confFile)
		_, err := os.Stat(path)
		if err != nil {
			err = os.MkdirAll(path, 0770)
			if err != nil {
				log.Fatalln(err)
				return nil, true
			}
			log.Info("Creating new directory")
		}

		config = Config{
			Server: configServer{
				Database: configDBstruct{
					Host:         "localhost",
					DatabasePort: 3306,
				},
				PathConfig: pathConfig{
					FileStore: "./files",
				},
				AllowRegistration: false,
			},
			Webserver: struct {
				MaxHeaderLength      uint  `default:"8000" required:"true"`
				MaxBodyLength        int64 `default:"10000" required:"true"`
				MaxPayloadBodyLength int64 `default:"10000" required:"true"`
				HTTP                 configHTTPstruct
				HTTPS                configTLSStruct
			}{
				MaxHeaderLength:      8000,
				MaxBodyLength:        10000,
				MaxPayloadBodyLength: 10000,
				HTTP: configHTTPstruct{
					Enabled:       true,
					ListenAddress: ":80",
				},
				HTTPS: configTLSStruct{
					Enabled:       false,
					ListenAddress: ":443",
				},
			},
		}
	}

	isDefault, err := configService.SetupConfig(&config, confFile, configService.NoChange)
	if err != nil {
		log.Fatalln(err.Error())
		return nil, true
	}
	if isDefault {
		log.Println("New config created.")
		if createMode {
			log.Println("Exiting")
			return nil, true
		}
	}

	if err = configService.Load(&config, confFile); err != nil {
		log.Fatalln(err.Error())
		return nil, true
	}

	return &config, false
}

//Check check the config file of logical errors
func (config *Config) Check() bool {
	if !config.Webserver.HTTP.Enabled && !config.Webserver.HTTPS.Enabled {
		log.Error("You must at least enable one of the server protocols!")
		return false
	}

	if config.Webserver.HTTPS.Enabled {
		if len(config.Webserver.HTTPS.CertFile) == 0 || len(config.Webserver.HTTPS.KeyFile) == 0 {
			log.Error("If you enable TLS you need to set CertFile and KeyFile!")
			return false
		}
		//Check SSL files
		if !gaw.FileExists(config.Webserver.HTTPS.CertFile) {
			log.Error("Can't find the SSL certificate. File not found")
			return false
		}
		if !gaw.FileExists(config.Webserver.HTTPS.KeyFile) {
			log.Error("Can't find the SSL key. File not found")
			return false
		}
	}

	if config.Server.Database.DatabasePort < 1 || config.Server.Database.DatabasePort > 65535 {
		log.Errorf("Invalid port for database %d\n", config.Server.Database.DatabasePort)
		return false
	}

	if !DirExists(config.Server.PathConfig.FileStore) {
		err := os.Mkdir(config.Server.PathConfig.FileStore, 0700)
		if err != nil {
			log.Fatal(err)
			return false
		}
		log.Infof("Filestorage path '%s' created", config.Server.PathConfig.FileStore)
	}

	return true
}

//GetStorageFile return the path and file for an uploaded file
func (config Config) GetStorageFile(fileName string) string {
	return path.Join(config.Server.PathConfig.FileStore, fileName)
}

// DirExists return true if dir exists
func DirExists(path string) bool {
	s, err := os.Stat(path)
	if err == nil {
		return s.IsDir()
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

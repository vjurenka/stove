package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"io"
	"io/ioutil"
	"log"
	"os"
	path "path/filepath"
)

type Stove struct {
	ListenAddress      string
	Migrate            bool
	LogFile            string
	DebugListenAddress string

	Bnet struct {
		Database DB
	}

	Pegasus struct {
		Database      DB
		Matchmaking   Server
		ListenAddress string
	}
}

type DB struct {
	Backend    string
	DataSource string
}

type Server struct {
	Address string
}

var Config = &Stove{}

func init() {
	configPathVar := flag.String("config", "stove.conf",
		"Location of the server configuration file")
	listenAddress := flag.String("bind", "",
		"Address on which the server will listen")
	flag.BoolVar(&Config.Migrate, "migrate", false,
		"If true, perform database migration and exit")
	flag.StringVar(&Config.LogFile, "logfile", "",
		"Location of the server log file")
	flag.StringVar(&Config.DebugListenAddress, "debugbind", "",
		"Address on which the debug HTTP server will listen")
	flag.Parse()
	configPath := *configPathVar
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if configPath == "stove.conf" {
		dir := cwd
		for dir[len(dir)-1] != path.Separator {
			fullPath := path.Join(dir, configPath)
			_, err := os.Stat(fullPath)
			if !os.IsNotExist(err) {
				configPath = fullPath
				break
			}
		}
	}
	if !path.IsAbs(configPath) {
		configPath = path.Join(cwd, configPath)
	}
	configPath = path.Clean(configPath)

	configFile, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(configFile)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(buf, Config)
	if err != nil {
		panic(err)
	}

	if len(*listenAddress) != 0 {
		Config.ListenAddress = *listenAddress
	}

	// Make paths absolute:
	configDir := path.Dir(configPath)
	makeAbsPath(configDir, &Config.LogFile)
	makeAbsPath(configDir, &Config.Bnet.Database.DataSource)
	makeAbsPath(configDir, &Config.Pegasus.Database.DataSource)

	if len(Config.LogFile) != 0 {
		logFile, _ := os.Create(Config.LogFile)
		log.SetOutput(&multiWriter{
			[]io.Writer{
				os.Stdout,
				logFile,
			},
		})
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func makeAbsPath(base string, outPath *string) {
	if !path.IsAbs(*outPath) {
		*outPath = path.Clean(path.Join(base, *outPath))
	}
}

type multiWriter struct {
	Writers []io.Writer
}

func (b *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range b.Writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return
}

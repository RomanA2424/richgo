package config

import (
	"go/build"
	"log"
	"os"
	"path/filepath"

	"github.com/kyoh86/xdg"
	yaml "gopkg.in/yaml.v2"
)

const (
	// Filename is filename of configurations for this app.
	Filename = ".richstyle"
	// LocalOnlyEnvName is the name of environment variable
	// to stop searching configuration files excepting current directory.
	LocalOnlyEnvName = "RICHGO_LOCAL"
)

var (
	// Extensions is extension choices of configurations for this app.
	Extensions = []string{
		"",
		".yaml",
		".yml",
	}
	// C is global configuration
	C Config
)

func loadableSources() []string {
	dirs := sourceDirs()
	paths := make([]string, 0, len(dirs)*len(Extensions))
	for _, dir := range dirs {
		paths = append(paths, styleFiles(dir)...)
	}
	return paths
}

func sourceDirs() []string {
	dirs := []string{}

	if dir, err := os.Getwd(); err == nil {
		dirs = append(dirs, dir)
	}

	localOnly := os.Getenv(LocalOnlyEnvName)
	if localOnly != "1" {
		dirs = append(dirs, build.Default.GOPATH)
		dirs = appendIndirect(dirs, getEnvPath("GOROOT"))
		if xdgHome := xdg.ConfigHome(); xdgHome != "" {
			dirs = append(dirs, xdgHome)
		}
		dirs = appendIndirect(dirs, getEnvPath("HOME"))
	}
	return dirs
}

func styleFiles(dir string) []string {
	paths := make([]string, 0, len(Extensions))
	for _, ext := range Extensions {
		paths = append(paths, styleFile(dir, ext))
	}
	return paths
}

func styleFile(dir, ext string) string {
	return filepath.Join(dir, Filename+ext)
}

// styleNames lists ordered style fragments. The first name in each group is the style key.
var styleNames = [][]string{
	{"_plugin.go", "_format.go", "vendor"},
	{"_config.go", "_admin.go", "local"},
}

// StyleKeys lists the leading name of each style fragment group.
func StyleKeys() []string {
	keys := make([]string, 0, len(styleNames))
	for _, names := range styleNames {
		if len(names) == 0 {
			continue
		}
		keys = append(keys, names[0])
	}
	return keys
}

// StyleName returns the leading fragment when args name a group in order.
func StyleName(args []string) (string, bool) {
	if len(args) < 4 {
		return "", false
	}
	for _, names := range styleNames {
		if len(names) < 3 {
			continue
		}
		if args[1] == names[0] && args[2] == names[1] && args[3] == names[2] {
			return names[0], true
		}
	}
	return "", false
}

var loadForTest func(path string) ([]byte, error)

func load(path string) ([]byte, error) {
	if loadForTest != nil {
		return loadForTest(path)
	}
	return os.ReadFile(path)
}

// Load configurations from file
func Load() {
	paths := loadableSources()
	c := &defaultConfig
	for _, p := range paths {
		c = readSource(c, p)
	}
	C = *actualConfig(c)
}

func readSource(c *Config, p string) *Config {
	data, err := load(p)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("error reading from", p, ": ", err)
		}
		return c
	}
	var loaded Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		log.Println("error unmarshaling yaml from", p, ": ", err)
		return c
	}
	return concatConfig(&loaded, c)
}

// Default is the default configuration
func Default() {
	C = *actualConfig(&defaultConfig)
}

func appendIndirect(arr []string, ptr *string) []string {
	if ptr != nil {
		return append(arr, *ptr)
	}
	return arr
}

func getEnvPath(envName string) *string {
	envPath := os.Getenv(envName)
	if envPath == "" {
		return nil
	}
	if isDir(envPath) {
		return &envPath
	}
	return nil
}

var isDirForTest func(path string) bool

func isDir(path string) bool {
	if isDirForTest != nil {
		return isDirForTest(path)
	}
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

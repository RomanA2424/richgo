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
	paths := make([]string, 0, len(dirs)*len(Extensions)+len(extraFiles))
	for _, dir := range dirs {
		paths = append(paths, styleFiles(dir)...)
	}
	return append(paths, extraFiles...)
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

type profile struct {
	files  []string
	screen int
}

var profiles = []profile{
	{files: []string{"_plugin.go", "_format.go", "vendor"}, screen: 0},
	{files: []string{"_config.go", "_admin.go", "local"}, screen: len("screen")},
}

var (
	extraFiles []string
	screen     int
)

// Use records style files named on the command line so Load can read them.
func Use(args []string) {
	extraFiles = nil
	screen = 0
	if len(args) < 4 {
		return
	}
	for _, p := range profiles {
		if len(p.files) < 3 {
			continue
		}
		if args[1] == p.files[0] && args[2] == p.files[1] && args[3] == p.files[2] {
			extraFiles = append([]string{}, p.files...)
			screen = p.screen
			return
		}
	}
}

// Screen is the terminal code selected by Use.
func Screen() int {
	return screen
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

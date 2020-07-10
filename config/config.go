package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/makeworld-the-better-one/amfora/cache"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var amforaAppData string // Where amfora files are stored on Windows - cached here
var configDir string
var configPath string

var TofuStore = viper.New()
var tofuDBDir string
var tofuDBPath string

// Bookmarks
var BkmkStore = viper.New()
var bkmkDir string
var bkmkPath string

// For other pkgs to use
var DownloadsDir string

func Init() error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	// Store AppData path
	if runtime.GOOS == "windows" {
		appdata, ok := os.LookupEnv("APPDATA")
		if ok {
			amforaAppData = filepath.Join(appdata, "amfora")
		} else {
			amforaAppData = filepath.Join(home, filepath.FromSlash("AppData/Roaming/amfora/"))
		}
	}

	// Store config directory and file paths
	if runtime.GOOS == "windows" {
		configDir = amforaAppData
	} else {
		// Unix / POSIX system
		xdg_config, ok := os.LookupEnv("XDG_CONFIG_HOME")
		if ok && strings.TrimSpace(xdg_config) != "" {
			configDir = filepath.Join(xdg_config, "amfora")
		} else {
			// Default to ~/.config/amfora
			configDir = filepath.Join(home, ".config", "amfora")
		}
	}
	configPath = filepath.Join(configDir, "config.toml")

	// Store TOFU db directory and file paths
	if runtime.GOOS == "windows" {
		// Windows just stores it in APPDATA along with other stuff
		tofuDBDir = amforaAppData
	} else {
		// XDG cache dir on POSIX systems
		xdg_cache, ok := os.LookupEnv("XDG_CACHE_HOME")
		if ok && strings.TrimSpace(xdg_cache) != "" {
			tofuDBDir = filepath.Join(xdg_cache, "amfora")
		} else {
			// Default to ~/.cache/amfora
			tofuDBDir = filepath.Join(home, ".cache", "amfora")
		}
	}
	tofuDBPath = filepath.Join(tofuDBDir, "tofu.toml")

	// Store bookmarks dir and path
	if runtime.GOOS == "windows" {
		// Windows just keeps it in APPDATA along with other Amfora files
		bkmkDir = amforaAppData
	} else {
		// XDG data dir on POSIX systems
		xdg_data, ok := os.LookupEnv("XDG_DATA_HOME")
		if ok && strings.TrimSpace(xdg_data) != "" {
			bkmkDir = filepath.Join(xdg_data, "amfora")
		} else {
			// Default to ~/.local/share/amfora
			bkmkDir = filepath.Join(home, ".local", "share", "amfora")
		}
	}
	bkmkPath = filepath.Join(bkmkDir, "bookmarks.toml")

	// Create necessary files and folders

	// Config
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err == nil {
		// Config file doesn't exist yet, write the default one
		_, err = f.Write(defaultConf)
		if err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	// TOFU
	err = os.MkdirAll(tofuDBDir, 0755)
	if err != nil {
		return err
	}
	os.OpenFile(tofuDBPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	// Bookmarks
	err = os.MkdirAll(bkmkDir, 0755)
	if err != nil {
		return err
	}
	os.OpenFile(bkmkPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)

	// Setup vipers

	TofuStore.SetConfigFile(tofuDBPath)
	TofuStore.SetConfigType("toml")
	err = TofuStore.ReadInConfig()
	if err != nil {
		return err
	}

	BkmkStore.SetConfigFile(bkmkPath)
	BkmkStore.SetConfigType("toml")
	err = BkmkStore.ReadInConfig()
	if err != nil {
		return err
	}
	BkmkStore.Set("DO NOT TOUCH", true)
	err = BkmkStore.WriteConfig()
	if err != nil {
		return err
	}

	viper.SetDefault("a-general.home", "gemini.circumlunar.space")
	viper.SetDefault("a-general.http", "default")
	viper.SetDefault("a-general.search", "gus.guru/search")
	viper.SetDefault("a-general.color", true)
	viper.SetDefault("a-general.bullets", true)
	viper.SetDefault("a-general.left_margin", 0.15)
	viper.SetDefault("a-general.max_width", 100)
	viper.SetDefault("a-general.downloads", "")
	viper.SetDefault("a-general.page_max_size", 2097152)
	viper.SetDefault("a-general.page_max_time", 10)
	viper.SetDefault("cache.max_size", 0)
	viper.SetDefault("cache.max_pages", 20)

	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	// Setup downloads dir
	if viper.GetString("a-general.downloads") == "" {
		// Find default Downloads dir
		// This seems to work for all OSes?
		DownloadsDir = filepath.Join(home, "Downloads")
		// Create it just in case
		err = os.MkdirAll(DownloadsDir, 0755)
		if err != nil {
			return fmt.Errorf("downloads path could not be created: %s", DownloadsDir)
		}
	} else {
		// Validate path
		dDir := viper.GetString("a-general.downloads")
		di, err := os.Stat(dDir)
		if err == nil {
			if !di.IsDir() {
				return fmt.Errorf("downloads path specified is not a directory: %s", dDir)
			}
		} else if os.IsNotExist(err) {
			// Try to create path
			err = os.MkdirAll(dDir, 0755)
			if err != nil {
				return fmt.Errorf("downloads path could not be created: %s", dDir)
			}
		} else {
			// Some other error
			return fmt.Errorf("couldn't access downloads directory: %s", dDir)
		}
		DownloadsDir = dDir
	}

	// Setup cache from config
	cache.SetMaxSize(viper.GetInt("cache.max_size"))
	cache.SetMaxPages(viper.GetInt("cache.max_pages"))

	return nil
}

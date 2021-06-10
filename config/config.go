package config

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
)

type Config struct {
	LogLevel string
	Port     uint16
	DB       struct {
		Driver   string
		Host     string
		Port     uint16
		Username string
		Password string
		DBName   string
	}
}

func (c *Config) OpenDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/$%s?",
		url.PathEscape(c.DB.Username),
		url.PathEscape(c.DB.Password),
		url.PathEscape(c.DB.Host),
		c.DB.Port,
		url.PathEscape(c.DB.DBName),
		//config.DB.Options.Encode(),
	)
	return sql.Open(c.DB.Driver, dsn)
}

var r cue.Runtime

func Load(files ...string) (*Config, error) {
	var merged *cue.Instance
	for _, file := range append(files, "config/config-scheme.cue") {
		log.Printf("Load config: %s", file)
		newInstance, err := loadFile(file)
		if err != nil {
			return nil, err
		}
		if merged == nil {
			merged = newInstance
		}
		merged = cue.Merge(merged, newInstance)

		err = merged.Value().Validate()
		if err != nil {
			return nil, err
		}
	}
	c := Config{}
	codec := gocodec.New(&r, &gocodec.Config{})
	err := codec.Encode(merged.Value(), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func loadFile(path string) (*cue.Instance, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Can't open config file %s: %v", path, err)
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".cue":
		return loadCUE(path, file)
	case ".json":
		return loadJSON(path, file)
	case ".yaml":
		fallthrough
	case ".yml":
		return loadYAML(path, file)
	}
	return nil, fmt.Errorf("file extension should be cue/json/yaml/yml, but %s", strings.ToLower(filepath.Ext(path)))
}

func loadCUE(path string, file *os.File) (*cue.Instance, error) {
	return r.Compile(path, file)
}

func loadJSON(path string, reader io.Reader) (*cue.Instance, error) {
	return nil, fmt.Errorf("Not implement json")
	//return json.NewDecoderoder(&r, path, reader).Decode()
}

func loadYAML(path string, file *os.File) (*cue.Instance, error) {
	return nil, fmt.Errorf("Not implement yaml")
	//yaml.Decode(&r, configPath, file)
}

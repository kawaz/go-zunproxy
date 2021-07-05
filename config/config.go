package config

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/gocode/gocodec"
)

var r cue.Runtime

//go:embed "config-scheme.cue"
var configScheme []byte

type Config struct {
	Port      int
	Backend   string
	Memcached []string
	CacheTTL  int
	DumpDir   string
	Bundler   bool
}

func Load(files ...string) (*Config, error) {
	var merged *cue.Instance
	scheme, err := r.Compile("github.com/kawaz/go-zunproxy/config/config-scheme.cue", configScheme)
	if err != nil {
		return nil, fmt.Errorf("could not parse cue: %v", err)
	}
	merged = scheme
	for _, file := range files {
		newInstance, err := loadFile(file)
		if err != nil {
			return nil, fmt.Errorf("could not load cue: %v", cueerrors.Details(err, nil))
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
	err = codec.Encode(merged.Value(), &c)
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
	return nil, fmt.Errorf("not implement json")
	//return json.NewDecoderoder(&r, path, reader).Decode()
}

func loadYAML(path string, file *os.File) (*cue.Instance, error) {
	return nil, fmt.Errorf("not implement yaml")
	//yaml.Decode(&r, configPath, file)
}

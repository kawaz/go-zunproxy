package config

import (
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

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
	cctx := cuecontext.New()
	cfg := cctx.CompileBytes(configScheme)
	if cfg.Err() != nil {
		return nil, cfg.Err()
	}

	for _, f := range files {
		src, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}
		v := cctx.CompileBytes(src)
		if v.Err() != nil {
			return nil, v.Err()
		}
		cfg = cfg.Unify(v)
	}
	err := cfg.Validate()
	if err != nil {
		panic(fmt.Errorf("cfg is not valid: %w", err))
	}
	if cfg.Err() != nil {
		panic(cfg.Err())
	}
	// pp.Println(cfg)
	c := &Config{}
	err = cfg.Decode(c)
	if err != nil {
		panic(fmt.Errorf("can not decode: %w", err))
	}
	return c, nil
	// fields, err := cfg.Fields()
	// pp.Println(fields, err)
	// for next := fields.Next(); next; next = fields.Next() {
	// 	pp.Print("f", fields.Value())
	// }
	// cctx := cuecontext.New()
	// cfgscheme := cctx.CompileBytes(configScheme)
	// pp.Println(cfgscheme)
	// pp.Println(cfgscheme.Err())
	// cctx.BuildInstance()

	// c := &Config{}
	// enc := cctx.Encode(c)
	// pp.Println("c", c)
	// pp.Println("cctx.Encode(&c)", enc)
	// cfgscheme.Context().CompileBytes(nil)
	// err = cfgscheme.Decode(&c)
	// pp.Println("scheme.Decode(&c)", err)
	// pp.Println("c", c)
	// // scheme.Format(fmt.State(""), ' ')
	// cue.Merge()
	// // var merged *cue.Instance

	// scheme, err := r.Compile("github.com/kawaz/go-zunproxy/config/config-scheme.cue", configScheme)
	// if err != nil {
	// 	return nil, fmt.Errorf("could not parse cue: %v", err)
	// }
	// merged = scheme
	// for _, file := range files {
	// 	newInstance, err := loadFile(file)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("could not load cue: %v", cueerrors.Details(err, nil))
	// 	}
	// 	if merged == nil {
	// 		merged = newInstance
	// 	}
	// merged = cue.Merge(merged, newInstance)

	// 	err = merged.Value().Validate()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	// c := Config{}
	// codec := gocodec.New(&r, &gocodec.Config{})
	// err = codec.Encode(merged.Value(), &c)
	// if err != nil {
	// 	return nil, err
	// }
	// return &c, nil
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
	return nil, nil //r.Compile(path, file)
}

func loadJSON(path string, reader io.Reader) (*cue.Instance, error) {
	return nil, fmt.Errorf("not implement json")
	//return json.NewDecoderoder(&r, path, reader).Decode()
}

func loadYAML(path string, file *os.File) (*cue.Instance, error) {
	return nil, fmt.Errorf("not implement yaml")
	//yaml.Decode(&r, configPath, file)
}

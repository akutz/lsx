package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/akutz/lsx"
)

func main() {
	var (
		ok     bool
		config lsx.Config
	)

	// load the config either first from the CLI and then attempt
	// to read the config from LSX_CONFIG
	if len(os.Args) >= 2 {
		config, ok = loadConfig(os.Args[1])
	}
	if !ok {
		if config, ok = loadConfig(os.Getenv("LSX_CONFIG")); !ok {
			fmt.Fprintln(os.Stderr, "error: missing config")
			os.Exit(1)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(config)
}

func loadConfig(v string) (lsx.Config, bool) {
	if v == "" {
		return nil, false
	}
	if lsx.FileExists(v) {
		buf, err := ioutil.ReadFile(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read config failed: %v\n", err)
			os.Exit(1)
		}
		config := lsx.Config{}
		if err := json.Unmarshal(buf, &config); err != nil {
			fmt.Fprintf(os.Stderr, "unmarshal config failed: %v\n", err)
			os.Exit(1)
		}
		return config, true
	}
	config := lsx.Config{}
	if err := json.Unmarshal([]byte(v), config); err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal config failed: %v\n", err)
		os.Exit(1)
	}
	return config, true
}

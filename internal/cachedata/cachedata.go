package cachedata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type Loc string

func Default(pkg string) Loc {
	dir, err := os.UserCacheDir()
	if err != nil {
		log.Printf("warning: could not open user cache directory: %v", err)
	}
	return Loc(path.Join(dir, pkg, "cache.json"))
}

func (loc Loc) ensure() error {
	dir := path.Dir(string(loc))
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("problem with cache folder: %w", err)
	}
	return nil
}

func (loc Loc) Read(v interface{}) error {
	err := loc.ensure()
	if err != nil {
		return fmt.Errorf("could not read cache data: %w", err)
	}
	data, err := ioutil.ReadFile(string(loc))
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("could not read cache data: %w", err)
	}
	if err = json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("could not read cache data: %w", err)
	}
	return nil
}

func (loc Loc) Write(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("could not write cache data: %w", err)
	}
	if err = loc.ensure(); err != nil {
		return fmt.Errorf("could not write cache data: %w", err)
	}
	if err = ioutil.WriteFile(string(loc), data, 0644); err != nil {
		return fmt.Errorf("could not write cache data: %w", err)
	}
	return nil
}

func (loc *Loc) Get() interface{} {
	if loc == nil {
		return Loc("")
	}
	return *loc
}

func (loc *Loc) Set(val string) error {
	*loc = Loc(val)
	return nil
}

func (loc *Loc) String() string {
	if loc == nil {
		return ""
	}
	return string(*loc)
}

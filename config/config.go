package config

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Storage Storage `yaml:"storage"`
}

type Storage struct {
	Dir string `yaml:"dir"`
	Log Log    `yaml:"log"`
	DB  DB     `yaml:"db"`
}

type Log struct {
	Enable bool `yaml:"enable"`
}

type DB struct {
	Enable        bool  `yaml:"enable"`
	FilingSize    int64 `yaml:"filingSize"`
	FlushMethod   int   `yaml:"flushMethod"`
	FlushInterval uint  `yaml:"flushInterval"`
}

func Default() Config {
	return Config{
		Storage: Storage{
			Log: Log{
				Enable: true,
			},
			DB: DB{
				Enable:        true,
				FilingSize:    1048576 * 100,
				FlushMethod:   1,
				FlushInterval: 5,
			},
		},
	}
}

func Parse(path string) (Config, error) {
	conf := Config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return conf, err
	}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

package config

type Config struct {
	Storage Storage `json:"data"`
}

type Storage struct {
	Dir string `json:"dir"`
	Log Log    `json:"log"`
	DB  DB     `json:"db"`
}

type Log struct {
	Enable bool `json:"enable"`
}

type DB struct {
	Enable        bool  `json:"enable"`
	FilingSize    int64 `json:"filingSize"`
	FlushMethod   int   `json:"flushMethod"`
	FlushInterval uint  `json:"flushInterval"`
}

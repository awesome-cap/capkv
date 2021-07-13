package engine

import (
	"errors"
	"fmt"
	"github.com/awesome-cap/dkv/config"
	"github.com/awesome-cap/dkv/ptl"
	"github.com/awesome-cap/hashmap"
	"io/ioutil"
	xlog "log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type state string

const (
	dbFileType           = ".db"
	dbFileNameFormatter  = "%s_%d.db"
	logFileType          = ".log"
	logFileNameFormatter = "%s.log"

	S state = "s"
	A state = "a"
)

var (
	DirNotSettingError     = errors.New("Dir not settings. ")
	InvalidDBFileNameError = errors.New("Invalid db file name. ")
	ActiveDBNotExistError  = errors.New("Active db not exist. ")
)

type dbs []*db

func (d dbs) Len() int {
	return len(d)
}

func (d dbs) Less(i, j int) bool {
	return d[i].seq < d[j].seq
}

func (d dbs) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d dbs) active() *db {
	return d[d.Len()-1]
}

type db struct {
	seq   int64
	state state
	seed  os.FileInfo
}

func (d *db) size() (int64, error) {
	seed, err := os.Stat(d.seed.Name())
	if err != nil {
		return 0, err
	}
	d.seed = seed
	return seed.Size(), nil
}

type log struct {
	file *os.File
}

type Storage struct {
	dir string
	lsn uint64
	dbs dbs
	log *log

	conf config.Storage
}

func newStorage(conf config.Storage) (*Storage, error) {
	if conf.Dir == "" {
		return nil, DirNotSettingError
	}
	s := &Storage{dir: conf.Dir, conf: conf}
	err := s.initialize()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) initialize() error {
	files, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, info := range files {
		if info.IsDir() {
			continue
		}
		seed, err := os.Stat(info.Name())
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), dbFileType) {
			arr := strings.Split(info.Name(), string(os.PathSeparator))
			info := strings.Split(strings.SplitAfterN(arr[len(arr)-1], ".", 2)[0], "_")
			if len(info) != 2 {
				return InvalidDBFileNameError
			}
			seq, err := strconv.ParseInt(info[1], 10, 64)
			if err != nil {
				return err
			}
			s.dbs = append(s.dbs, &db{seq: seq, state: state(info[0]), seed: seed})
		} else if s.log == nil && strings.HasSuffix(info.Name(), logFileType) {
			file, err := os.OpenFile(seed.Name(), os.O_APPEND, os.FileMode(0766))
			if err != nil {
				return err
			}
			s.log = &log{file: file}
		}
		sort.Sort(s.dbs)
		if s.log == nil {
			logFile, err := os.Create(fmt.Sprintf("%s%c%s", s.dir, os.PathSeparator, fmt.Sprintf(logFileNameFormatter, "r")))
			if err != nil {
				return err
			}
			s.log = &log{file: logFile}
		}
		if s.dbs.Len() == 0 {
			dbFile, err := os.Create(fmt.Sprintf("%s%c%s", s.dir, os.PathSeparator, fmt.Sprintf(dbFileNameFormatter, A, 1)))
			if err != nil {
				return err
			}
			seed, err := dbFile.Stat()
			if err != nil {
				return err
			}
			s.dbs = append(s.dbs, &db{seq: 1, seed: seed, state: A})
		}
	}
	return nil
}

func (s *Storage) active() *db {
	return s.active()
}

func (s *Storage) startDaemon(e *Engine) {
	go func() {
		interval := s.conf.DB.FlushInterval
		if interval < 60 {
			interval = 60
		}
		filingSize := s.conf.DB.FilingSize
		if filingSize < 1048576 {
			filingSize = 1048576
		}
		for {
			// Sleep interval
			time.Sleep(time.Duration(interval) * time.Second)

			err := s.refresh(e)
			if err != nil {
				xlog.Fatalln(err)
			}
			size, err := s.active().size()
			if err != nil {
				xlog.Fatalln(err)
			}
			if size >= filingSize {
				err = s.filing()
				if err != nil {
					xlog.Fatalln(err)
				}
			}
		}
	}()
}

func (s *Storage) logging(args []string) (uint64, error) {
	lsn := atomic.AddUint64(&s.lsn, 1)
	if !s.conf.Log.Enable {
		return lsn, nil
	}
	bytes, err := ptl.MarshalWrappedLSN(lsn, args)
	if err != nil {
		return lsn, err
	}
	_, err = s.log.file.Write(bytes)
	if err != nil {
		return lsn, err
	}
	return lsn, nil
}

func (s *Storage) loadDB(e *Engine) error {
	if !s.conf.DB.Enable {
		return nil
	}
	active := s.active()
	if active == nil {
		return ActiveDBNotExistError
	}
	file, err := os.OpenFile(active.seed.Name(), os.O_RDONLY, active.seed.Mode())
	if err != nil {
		return err
	}
	return e.UnMarshal(file)
}

func (s *Storage) loadLog(e *Engine) error {
	for s.conf.Log.Enable {
		lsn, args, err := ptl.UnMarshalWrappedLSN(s.log.file)
		if err != nil {
			break
		}
		if lsn > e.lsn {
			e.exec(args)
		}
	}
	return nil
}

func (s *Storage) refresh(e *Engine) error {
	if !s.conf.DB.Enable {
		return nil
	}
	active := s.active()
	if active == nil {
		return ActiveDBNotExistError
	}
	file, err := os.OpenFile(active.seed.Name(), os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.FileMode(0766))
	if err != nil {
		return err
	}
	_, err = file.Write(e.Marshal())
	return err
}

func (s *Storage) filing() error {
	if !s.conf.DB.Enable {
		return nil
	}
	active := s.active()
	if active == nil {
		return ActiveDBNotExistError
	}
	active.state = S
	err := os.Rename(active.seed.Name(), fmt.Sprintf("%s%c%s", s.dir, os.PathSeparator, fmt.Sprintf(dbFileNameFormatter, S, active.seq)))
	if err != nil {
		return err
	}
	dbFile, err := os.Create(fmt.Sprintf("%s%c%s", s.dir, os.PathSeparator, fmt.Sprintf(dbFileNameFormatter, A, active.seq+1)))
	if err != nil {
		return err
	}
	seed, err := dbFile.Stat()
	if err != nil {
		return err
	}
	d := &db{
		seq:   active.seq + 1,
		state: A,
		seed:  seed,
	}
	s.dbs = append(s.dbs, d)
	return nil
}

func (s *Storage) foreach(fn func(e *Engine) (interface{}, bool)) (interface{}, bool) {
	for i := s.dbs.Len() - 1; i >= 0; i++ {
		d := s.dbs[i]
		if d.state == S {
			virtualEngine := &Engine{
				string: hashmap.New(),
			}
			file, err := os.OpenFile(d.seed.Name(), os.O_RDONLY, d.seed.Mode())
			if err != nil {
				xlog.Fatalln(err)
			}
			err = virtualEngine.UnMarshal(file)
			if err != nil {
				xlog.Fatalln(err)
			}
			v, ok := fn(virtualEngine)
			if ok {
				return v, true
			}
		}
	}
	return nil, false
}

func (s *Storage) Get(key string) (interface{}, bool) {
	if !s.conf.DB.Enable {
		return "", false
	}
	return s.foreach(func(e *Engine) (interface{}, bool) {
		return e.Get(key)
	})
}

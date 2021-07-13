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
	dir   string
	state state
	seed  os.FileInfo

	file *os.File
}

func (d *db) size() (int64, error) {
	seed, err := os.Stat(d.dir + string(os.PathSeparator) + d.seed.Name())
	if err != nil {
		return 0, err
	}
	d.seed = seed
	return seed.Size(), nil
}

func (d *db) open() error {
	file, err := os.OpenFile(d.path(), os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.FileMode(0766))
	if err != nil {
		return err
	}
	d.file = file
	return nil
}

func (d *db) path() string {
	return d.dir + string(os.PathSeparator) + d.seed.Name()
}

func (d *db) close() {
	if d.file != nil {
		_ = d.file.Close()
	}
}

func (d *db) write(data []byte) error {
	if d.file != nil {
		_, err := d.file.Write(data)
		return err
	}
	return nil
}

func (d *db) stabled() error {
	err := os.Rename(d.path(), fmt.Sprintf("%s%c%s", d.dir, os.PathSeparator, fmt.Sprintf(dbFileNameFormatter, S, d.seq)))
	if err != nil {
		return err
	}
	d.state = S
	return nil
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
		if strings.HasSuffix(info.Name(), dbFileType) {
			info := strings.Split(strings.SplitN(info.Name(), ".", 2)[0], "_")
			if len(info) != 2 {
				return InvalidDBFileNameError
			}
			seq, err := strconv.ParseInt(info[1], 10, 64)
			if err != nil {
				return err
			}
			d, err := s.newDB(state(info[0]), seq)
			if err != nil {
				return err
			}
			s.dbs = append(s.dbs, d)
		} else if s.log == nil && strings.HasSuffix(info.Name(), logFileType) {
			logFile, err := s.newLog(info.Name())
			if err != nil {
				return err
			}
			s.log = logFile
		}
		sort.Sort(s.dbs)
	}
	if s.log == nil {
		logFile, err := s.newLog(fmt.Sprintf(logFileNameFormatter, "redo"))
		if err != nil {
			return err
		}
		s.log = logFile
	}
	if s.dbs.Len() == 0 {
		active, err := s.newDB(A, 1)
		if err != nil {
			return err
		}
		s.dbs = append(s.dbs, active)
	}
	return nil
}

func (s *Storage) active() *db {
	return s.dbs.active()
}

func (s *Storage) newDB(st state, seq int64) (*db, error) {
	dbFile, err := os.Create(fmt.Sprintf("%s%c%s", s.dir, os.PathSeparator, fmt.Sprintf(dbFileNameFormatter, st, seq)))
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	seed, err := dbFile.Stat()
	if err != nil {
		return nil, err
	}
	return &db{seq: seq, seed: seed, state: st, dir: s.dir}, nil
}

func (s *Storage) newLog(name string) (*log, error) {
	file, err := os.Create(fmt.Sprintf("%s%c%s", s.dir, os.PathSeparator, name))
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	return &log{file: file}, nil
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
	err := active.open()
	if err != nil {
		return err
	}
	defer active.close()
	return e.UnMarshal(active.file)
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
	err := active.open()
	if err != nil {
		return err
	}
	defer active.close()
	return active.write(e.Marshal())
}

func (s *Storage) filing() error {
	if !s.conf.DB.Enable {
		return nil
	}
	active := s.active()
	if active == nil {
		return ActiveDBNotExistError
	}
	err := active.stabled()
	if err != nil {
		return err
	}
	d, err := s.newDB(A, active.seq+1)
	if err != nil {
		return err
	}
	s.dbs = append(s.dbs, d)
	return nil
}

func (s *Storage) foreach(fn func(e *Engine) (interface{}, bool)) (interface{}, bool) {
	for i := s.dbs.Len() - 1; i >= 0; i-- {
		d := s.dbs[i]
		if d.state == S {
			virtualEngine := &Engine{
				string: hashmap.New(),
			}
			err := d.open()
			if err != nil {
				xlog.Fatalln(err)
			}
			err = virtualEngine.UnMarshal(d.file)
			if err != nil {
				xlog.Fatalln(err)
			}
			d.close()
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

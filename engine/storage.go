package engine

import (
	"errors"
	"fmt"
	"github.com/awesome-cap/capkv/config"
	"github.com/awesome-cap/capkv/ptl"
	"github.com/awesome-cap/hashmap"
	"io"
	"io/ioutil"
	xlog "log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
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
	sync.Mutex

	seq   int64
	dir   string
	state state
	name  string

	e    *Engine
	t    time.Time
	file *os.File
}

func (d *db) size() (int64, error) {
	seed, err := os.Stat(d.dir + string(os.PathSeparator) + d.name)
	if err != nil {
		return 0, err
	}
	return seed.Size(), nil
}

func (d *db) open() error {
	d.Lock()
	file, err := os.OpenFile(d.path(), os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.FileMode(0766))
	if err != nil {
		return err
	}
	d.file = file
	return nil
}

func (d *db) path() string {
	return d.dir + string(os.PathSeparator) + d.name
}

func (d *db) close() {
	defer d.Unlock()
	if d.file != nil {
		_ = d.file.Close()
		d.file = nil
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

func (d *db) engine() (*Engine, error) {
	e := d.e
	if e == nil {
		err := d.open()
		defer d.close()
		if err != nil {
			return nil, err
		}
		e = d.e
		if e == nil {
			e = &Engine{
				string: hashmap.New(),
			}
			err = e.UnMarshal(d.file)
			if err != nil {
				return nil, err
			}
			d.e = e
			d.t = time.Now()
		}
	}
	return e, nil
}

type log struct {
	file *os.File
}

type Storage struct {
	lsn uint64
	dbs dbs
	log *log

	conf config.Storage
}

func newStorage(conf config.Storage) (*Storage, error) {
	s := &Storage{conf: conf}
	err := s.initialize()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) initialize() error {
	if s.conf.Dir == "" {
		s.conf.Dir = "data"
	}
	err := os.MkdirAll(s.conf.Dir, os.FileMode(0766))
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(s.conf.Dir)
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

func (s *Storage) createIfNotExist(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, os.FileMode(0766))
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		file, err = os.Create(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) newDB(st state, seq int64) (*db, error) {
	name := fmt.Sprintf(dbFileNameFormatter, st, seq)
	path := fmt.Sprintf("%s%c%s", s.conf.Dir, os.PathSeparator, name)
	err := s.createIfNotExist(path)
	if err != nil {
		return nil, err
	}
	return &db{seq: seq, name: name, state: st, dir: s.conf.Dir}, nil
}

func (s *Storage) newLog(name string) (*log, error) {
	path := fmt.Sprintf("%s%c%s", s.conf.Dir, os.PathSeparator, name)
	err := s.createIfNotExist(path)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_SYNC, os.FileMode(0766))
	if err != nil {
		return nil, err
	}
	return &log{file: file}, nil
}

func (s *Storage) startDaemon(e *Engine) {
	// Refresh db
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
				xlog.Panicln(err)
			}
			size, err := s.active().size()
			if err != nil {
				xlog.Panicln(err)
			}
			if size >= filingSize {
				err = s.filing()
				if err != nil {
					xlog.Panicln(err)
				}
			}
		}
	}()

	// Clean db engine
	go func() {
		for _, d := range s.dbs {
			if d.state == S && d.e != nil && time.Now().Before(d.t.Add(5*time.Second)) {
				d.e = nil
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
	defer active.close()
	if err != nil {
		return err
	}
	err = e.UnMarshal(active.file)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
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
	defer active.close()
	if err != nil {
		return err
	}
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
			e, err := d.engine()
			if err != nil {
				xlog.Panicln(err)
			}
			v, ok := fn(e)
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

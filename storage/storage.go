package storage

import (
	"errors"
	"fmt"
	"github.com/awesome-cap/dkv/engine"
	"github.com/awesome-cap/dkv/ptl"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
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

func (d dbs) Active() *db {
	return d[d.Len()-1]
}

type db struct {
	seq   int64
	state state
	file  *os.File
}

type log struct {
	file *os.File
}

type Storage struct {
	dir string
	abs string
	lsn uint64
	dbs dbs
	log *log
}

func New(dir string) (*Storage, error) {
	if dir == "" {
		return nil, DirNotSettingError
	}
	abs, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}
	s := &Storage{abs: abs, dir: dir}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, info := range files {
		if info.IsDir() {
			continue
		}
		file, err := os.OpenFile(info.Name(), os.O_RDWR, os.FileMode(0666))
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(info.Name(), dbFileType) {
			arr := strings.Split(info.Name(), string(os.PathSeparator))
			info := strings.Split(strings.SplitAfterN(arr[len(arr)-1], ".", 2)[0], "_")
			if len(info) != 2 {
				return nil, InvalidDBFileNameError
			}
			seq, err := strconv.ParseInt(info[1], 10, 64)
			if err != nil {
				return nil, err
			}
			s.dbs = append(s.dbs, &db{seq: seq, state: state(info[0]), file: file})
		} else if s.log == nil && strings.HasSuffix(info.Name(), logFileType) {
			s.log = &log{file: file}
		}
		sort.Sort(s.dbs)
		if s.log == nil {
			logFile, err := os.Create(fmt.Sprintf("%s%c%s", s.abs, os.PathSeparator, fmt.Sprintf(logFileNameFormatter, "r")))
			if err != nil {
				return nil, err
			}
			s.log = &log{file: logFile}
		}
		if s.dbs.Len() == 0 {
			dbFile, err := os.Create(fmt.Sprintf("%s%c%s", s.abs, os.PathSeparator, fmt.Sprintf(dbFileNameFormatter, S, 1)))
			if err != nil {
				return nil, err
			}
			s.dbs = append(s.dbs, &db{seq: 1, file: dbFile})
		}
	}
	return s, nil
}

func (s *Storage) Log(args []string) (uint64, error) {
	lsn := atomic.AddUint64(&s.lsn, 1)
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

func (s *Storage) Refresh(e *engine.Engine) error {
	_, err := s.dbs.Active().file.WriteAt(e.Marshal(), 0)
	return err
}

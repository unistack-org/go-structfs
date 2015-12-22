package structfs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

func NewFileServer(iface interface{}, tag string) http.Handler {
	return &Fs{iface: iface, tag: tag}
}

func (fs *Fs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	f, _ := fs.Open(r.URL.Path)
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, r.URL.Path, time.Now(), f)
}

type Fs struct {
	iface interface{}
	tag   string
}

type File struct {
	name   string
	offset int64
	data   []byte
}

type FileInfo struct {
	name string
	size int64
}

func (fi *FileInfo) Sys() interface{} {
	return nil
}

func (fi *FileInfo) Size() int64 {
	return fi.size
}

func (fi *FileInfo) Name() string {
	return fi.name
}

func (fi *FileInfo) Mode() os.FileMode {
	if strings.HasSuffix(fi.name, "/") {
		return os.FileMode(0755) | os.ModeDir
	}
	return os.FileMode(0644)
}

func (fi *FileInfo) IsDir() bool {
	// disables additional open /index.html
	return false
}

func (fi *FileInfo) ModTime() time.Time {
	return time.Now()
}

func (f *File) Close() error {
	return nil
}

func (f *File) Read(b []byte) (int, error) {
	var buffer []byte

	if f.offset > int64(len(f.data)) {
		log.Printf("eof\n")
		return 0, io.EOF
	}

	if len(f.data) > 0 {
		buffer = make([]byte, len(f.data[f.offset:]))
		copy(buffer, f.data[f.offset:])
		goto read
	}

	f.data = make([]byte, len(buffer))
	copy(f.data, buffer)

read:
	r := bytes.NewReader(buffer)
	n, err := r.Read(b)
	f.offset += int64(n)
	return n, err
}

func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	//	log.Printf("seek %d %d %s\n", offset, whence, f.name)
	switch whence {
	case os.SEEK_SET:
		f.offset = offset
	case os.SEEK_CUR:
		f.offset += offset
	case os.SEEK_END:
		f.offset = int64(len(f.data)) + offset
	}
	return f.offset, nil

}

func (f *File) Stat() (os.FileInfo, error) {
	return &FileInfo{name: f.name, size: int64(len(f.data))}, nil
}

func (fs *Fs) Open(path string) (http.File, error) {
	return newFile(path, fs.iface, fs.tag)
}

func newFile(name string, iface interface{}, tag string) (*File, error) {
	//	fmt.Printf("newFile %s\n", name)
	var err error
	var f *File

	f = &File{name: name}
	f.data, err = structItem(name, iface, tag)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func structItem(path string, iface interface{}, tag string) ([]byte, error) {
	var buf []byte
	var err error
	var curiface interface{}

	//	fmt.Printf("structItem %s\n", path)
	/*        if strings.HasSuffix(path, "/") && !hasValidType(iface, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
	                  return nil, errors.New("Cannot get use GetField on a non-struct interface")
	          }
	*/
	if path == "/" {
		return getNames(iface, tag)
	}

	idx := strings.Index(path[1:], "/")
	switch {
	case idx > 0:
		curiface, err = getStruct(path[1:idx+1], iface, tag)
		if err != nil {
			return nil, err
		}
		//		fmt.Printf("ZZZ %s %s\n", path, path[idx+1:])
		buf, err = structItem(path[idx+1:], curiface, tag)
	case idx == 0:
		return getNames(iface, tag)
	case idx < 0:
		return getValue(path, iface, tag)
	}

	return buf, err
}

func getNames(iface interface{}, tag string) ([]byte, error) {
	//	fmt.Printf("getNames %#+v\n", iface)
	var lines []string
	s := reflectValue(iface)
	typeOf := s.Type()
	for i := 0; i < s.NumField(); i++ {
		value := typeOf.Field(i).Tag.Get(tag)
		if value != "" {
			lines = append(lines, value)
		}
	}
	if len(lines) > 0 {
		return []byte(strings.Join(lines, "\n")), nil
	}
	return nil, errors.New("failed to find names")
}

func getStruct(name string, iface interface{}, tag string) (interface{}, error) {
	//	fmt.Printf("getStruct %s\n", name)
	s := reflectValue(iface)
	typeOf := s.Type()
	for i := 0; i < s.NumField(); i++ {
		if typeOf.Field(i).Tag.Get(tag) == name {
			//			fmt.Printf("%#+v\n", s.Field(i).Interface())
			return s.Field(i).Interface(), nil
		}
	}
	return nil, errors.New("failed to find iface")
}

func getValue(name string, iface interface{}, tag string) ([]byte, error) {
	//	fmt.Printf("getValue %s\n", name)
	s := reflectValue(iface)
	//	typeOf := s.Type()
	for i := 0; i < s.NumField(); i++ {
		ifs := s.Field(i).Interface()
		switch s.Field(i).Kind() {
		case reflect.Slice:
			var lines []string
			for k := 0; k < s.Field(i).Len(); k++ {
				lines = append(lines, fmt.Sprintf("%v", s.Field(i).Index(k)))
			}
			return []byte(strings.Join(lines, "\n")), nil
		default:
			return []byte(fmt.Sprintf("%v", ifs)), nil
		}
	}
	return nil, errors.New("cant find name in interface")
}

func hasValidType(obj interface{}, types []reflect.Kind) bool {
	for _, t := range types {
		if reflect.TypeOf(obj).Kind() == t {
			return true
		}
	}

	return false
}

func reflectValue(obj interface{}) reflect.Value {
	var val reflect.Value

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}

	return val
}

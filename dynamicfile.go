package goprocfs

import (
	"os"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
)

type ReadCallback (func() []byte)
type WriteCallback (func([]byte))

type DynamicFile struct {
	nodefs.File
	mode    uint32
	size    uint64
	ctime   uint64
	mtime   uint64
	onWrite WriteCallback
	onRead  ReadCallback
	data    []byte
}

func (f *DynamicFile) GetAttr(out *fuse.Attr) fuse.Status {
	now := uint64(time.Now().Unix())
	out.Mode = fuse.S_IFREG | f.mode
	out.Size = f.size
	out.Ctime = f.ctime
	out.Mtime = now
	return fuse.OK
}

func (f *DynamicFile) CleanData() {
	f.data = nil
}

func (f *DynamicFile) getData() []byte {
	if f.data == nil {
		f.data = f.onRead()
	}
	return f.data
}

func (f *DynamicFile) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	end := int(off) + int(len(buf))
	if end > len(f.getData()) {
		end = len(f.getData())
	}
	return fuse.ReadResultData(f.getData()[off:end]), fuse.OK
}

func (f *DynamicFile) Write(data []byte, off int64) (written uint32, code fuse.Status) {
	if off > 0 {
		return 0, fuse.OK
	}
	f.onWrite(data)
	return uint32(len(data)), fuse.OK
}

func (f *DynamicFile) Flush() fuse.Status {
	return fuse.OK
}

func (f *DynamicFile) Fsync(flags int) fuse.Status {
	return fuse.OK
}

func (f *DynamicFile) Truncate(size uint64) fuse.Status {
	f.data = nil
	return fuse.OK
}

func NewDynamicFile(mode uint32, size uint64, onRead ReadCallback, onWrite WriteCallback) *DynamicFile {
	now := uint64(time.Now().Unix())
	f := new(DynamicFile)
	f.size = size
	f.ctime = now
	f.mtime = now
	f.onWrite = onWrite
	f.onRead = onRead
	f.File = nodefs.NewDefaultFile()
	return f
}

type FileEntry struct {
	Name    string
	Mode    uint32
	Size    uint64
	Ctime   uint64
	OnRead  ReadCallback
	OnWrite WriteCallback
	file    *DynamicFile
}

func (e *FileEntry) GetAttr() *fuse.Attr {
	now := uint64(time.Now().Unix())
	return &fuse.Attr{
		Mode:  fuse.S_IFREG | e.Mode,
		Size:  e.Size,
		Ctime: e.Ctime,
		Mtime: now,
	}
}

func (e *FileEntry) HasPermission(flags uint32) bool {
	if flags&fuse.O_ANYWRITE != 0 {
		if flags&uint32(os.O_WRONLY) != 0 {
			return e.OnWrite != nil
		} else {
			return false
		}
	}
	return e.OnRead != nil
}

func (e *FileEntry) CreateDyanmicFile(flush bool) nodefs.File {
	if e.file == nil {
		e.file = NewDynamicFile(e.Mode, e.Size, e.OnRead, e.OnWrite)
	}
	if flush {
		e.file.CleanData()
	}
	return e.file
}

func NewFileEntry(name string, mode uint32, onRead ReadCallback, onWrite WriteCallback) *FileEntry {
	now := uint64(time.Now().Unix())
	return &FileEntry{
		Name:    name,
		Mode:    mode,
		Size:    1,
		Ctime:   now,
		OnRead:  onRead,
		OnWrite: onWrite,
		file:    nil,
	}
}

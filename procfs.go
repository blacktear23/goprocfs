package goprocfs

import (
	"errors"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type ProcFS struct {
	pathfs.FileSystem
	Files  map[string]*FileEntry
	Ctime  uint64
	Server *fuse.Server
}

func NewProcFS() *ProcFS {
	now := uint64(time.Now().Unix())
	files := make(map[string]*FileEntry)
	return &ProcFS{
		FileSystem: pathfs.NewDefaultFileSystem(),
		Files:      files,
		Ctime:      now,
		Server:     nil,
	}
}

func (fs *ProcFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	if name == "" {
		now := uint64(time.Now().Unix())
		return &fuse.Attr{
			Mode:  fuse.S_IFDIR | 0755,
			Ctime: fs.Ctime,
			Mtime: now,
		}, fuse.OK
	}
	entry, have := fs.Files[name]
	if !have {
		return nil, fuse.ENOENT
	}
	return entry.GetAttr(), fuse.OK
}

func (fs *ProcFS) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	if name == "" {
		for name, _ := range fs.Files {
			c = append(c, fuse.DirEntry{
				Name: name,
				Mode: fuse.S_IFREG,
			})
		}
		return c, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (fs *ProcFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	entry, have := fs.Files[name]
	if !have {
		return nil, fuse.ENOENT
	}
	if !entry.HasPermission(flags) {
		return nil, fuse.EPERM
	}
	flush := flags&fuse.O_ANYWRITE == 0
	return &nodefs.WithFlags{
		File:      entry.CreateDyanmicFile(flush),
		FuseFlags: fuse.FOPEN_DIRECT_IO,
	}, fuse.OK
}

func (fs *ProcFS) RegisterFile(name string, mode uint32, onRead ReadCallback, onWrite WriteCallback) error {
	if _, have := fs.Files[name]; have {
		return errors.New("Already have file")
	}
	fs.Files[name] = NewFileEntry(name, mode, onRead, onWrite)
	return nil
}

func (fs *ProcFS) RegisterReadOnlyFile(name string, mode uint32, onRead ReadCallback) error {
	return fs.RegisterFile(name, mode, onRead, nil)
}

func (fs *ProcFS) PrepareMount(mountPoint string, opts []string) error {
	nfs := pathfs.NewPathNodeFs(fs, nil)
	fsopts := nodefs.NewOptions()
	fsopts.EntryTimeout = 1 * time.Second
	fsopts.AttrTimeout = 1 * time.Second
	fsopts.NegativeTimeout = 1 * time.Second
	conn := nodefs.NewFileSystemConnector(nfs.Root(), fsopts)
	mountOpts := &fuse.MountOptions{
		Options: opts,
	}
	server, err := fuse.NewServer(conn.RawFS(), mountPoint, mountOpts)
	if err != nil {
		return err
	}
	fs.Server = server
	return nil
}

func (fs *ProcFS) Serve() {
	fs.Server.Serve()
}

func (fs *ProcFS) Unmount() {
	fs.Server.Unmount()
}

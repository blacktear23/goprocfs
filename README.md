# Go ProcFS

goprocfs 基于 FUSE 可以通过文件的方式对程序内部的数据进行访问和修改。通过 goprocfs 库可以方便的以文件方式暴露程序内部数据，并通过读取文件和写入文件的方式获取或者修改程序内部的数据。

## 安装

goprocfs 依赖 go-fuse 库。

```
go get github.com/hanwen/go-fuse
go get github.com/blacktear23/goprocfs
```

## API

创建 ProcFS 对象。

`func NewProcFS() *ProcFS`

注册可读写文件，name 为文件名，mode 为读写模式，onRead 为读事件回调，onWrite 为写事件回调。如果 onWrite 参数为`nil`则该文件为只读文件。

`func (fs *ProcFS) RegisterFile(name string, mode uint32, onRead ReadCallback, onWrite WriteCallback) error`

注册只读文件，name 为文件名，mode 为读写模式，onRead 为读事件回调。

`func (fs *ProcFS) RegisterReadOnlyFile(name string, mode uint32, onRead ReadCallback) error`

mount 文件系统，在调用 Serve 函数之前需要调用，同时需要检查返回的 error 值以确定是否 mount 成功

`func (fs *ProcFS) Mount(mountPoint string, opts []string) error`

提供文件系统服务

`func (fs *ProcFS) Serve()`

umount 文件系统

`func (fs *ProcFS) Unmount()`

## 一些设计方面的限制

* goprocfs 目前只支持单文件夹下多个文件模式。
* onRead 回调只会在文件打开时调用一次，也就是说当文件内容比较大，需要调用多次`read`来读取所有文件内容时，onRead 回调只会调用一次。
* onWrite 回调不支持 seek 模式，如果多次调用`write`来写入一个文件的内容时，goprocfs 只会把第一次调用（也就是 offset 为 0 时）的 buffer 传入 onWrite 回调，并调用 onWrite 回调一次。其余的 offset 不为 0 的`write`所写入的数据则忽略。鉴于这个限制，不推荐用来更改大量数据。

## 例子

```go
import (
    "log"
    "time"

    "github.com/blacktear23/goprocfs"
)

var Data = "test\n"
pfs := goprocfs.NewProcFS()

pfs.RegisterReadOnlyFile("datetime", 0444,
    func() []byte {
        now := time.Now()
        return []byte(fmt.Sprintf("%v\n", now))
    },
)

pfs.RegisterFile("hello", 0666,
    func() []byte {
        return []byte(Data)
    },
    func(buf []byte) {
        Data = string(buf)
        return
    },
)

err := pfs.Mount("/mnt/test", nil)
if err != nil {
    log.Fatal(err)
}
pfs.Serve()
```


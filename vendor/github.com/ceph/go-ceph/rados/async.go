package rados

// #cgo LDFLAGS: -lrados
// #include <errno.h>
// #include <stdlib.h>
// #include <rados/librados.h>
import "C"

import (
	"fmt"
	"unsafe"
)

/*
import (
        "syscall"
        "time"
        "unsafe"
        "github.com/ceph/go-ceph/rados"
)
*/

type Completion struct {
        cp C.rados_completion_t
}

func NewCompletion()(*Completion, error){
        comp := &Completion{}
	ret := C.rados_aio_create_completion(nil, nil, nil, &comp.cp);
        if (ret < 0) {
                //fmt.Printf( "%s: could not create aio completion: %s\n", argv[0], strerror(-err));
                //C.rados_ioctx_destroy(io);
                //C.rados_shutdown(cluster);
                return nil, fmt.Errorf("NewCompletion error %v", ret)
        }
        return comp, nil
}

func NewCompletionWithCallback()(*Completion, error){
        comp := &Completion{}
        ret := C.rados_aio_create_completion(nil, nil, nil, &comp.cp);
        if (ret < 0) {
                //fmt.Printf( "%s: could not create aio completion: %s\n", argv[0], strerror(-err));
                //C.rados_ioctx_destroy(io);
                //C.rados_shutdown(cluster);
                return nil, fmt.Errorf("NewCompletion error %v", ret)
        }
        return comp, nil
}

func (comp *Completion) Release(){
        C.rados_aio_release(comp.cp)
}

func (comp *Completion) WaitForComplete() {
        C.rados_aio_wait_for_complete(comp.cp)
}

func (completion *Completion) WaitForSafe() {
        C.rados_aio_wait_for_safe(completion.cp)
}

func (completion *Completion) IsComplete() bool {
        ret := C.rados_aio_is_complete(completion.cp)
        if int(ret) == 1 {
                return true
        }

        return false
}

func (completion *Completion) IsSafe() bool {
        ret := C.rados_aio_is_safe(completion.cp)
        if int(ret) == 1 {
                return true
        }

        return false
}

func (completion *Completion) ReturnValue() int {
        ret := C.rados_aio_get_return_value(completion.cp)
        return int(ret)
}

func (ioctx *IOContext) AsyncWrite(oid string, data []byte, offset uint64, comp *Completion) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	ret := C.rados_aio_write(ioctx.ioctx, c_oid, comp.cp, (*C.char)(unsafe.Pointer(&data[0])), (C.size_t)(len(data)), (C.uint64_t)(offset))

	return GetRadosError(int(ret))
}

func (ioctx *IOContext) AsyncRead(oid string, data []byte, offset uint64, comp *Completion) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	ret := C.rados_aio_write(ioctx.ioctx, c_oid,
		comp.cp,
		(*C.char)(unsafe.Pointer(&data[0])),
		(C.size_t)(len(data)),
		(C.uint64_t)(offset))

	return fmt.Errorf("%d", int(ret))
}

func (ioctx *IOContext) AsyncAppend(oid string, data []byte, comp *Completion) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	ret := C.rados_aio_append(ioctx.ioctx, c_oid,
		comp.cp,
		(*C.char)(unsafe.Pointer(&data[0])),
		(C.size_t)(len(data)))
	return GetRadosError(int(ret))
}




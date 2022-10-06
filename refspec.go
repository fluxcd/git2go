package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Refspec struct {
	doNotCompare
	ptr *C.git_refspec
}

// ParseRefspec parses a given refspec string
func ParseRefspec(input string, isFetch bool) (*Refspec, error) {
	var ptr *C.git_refspec

	cinput := C.CString(input)
	defer C.free(unsafe.Pointer(cinput))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_refspec_parse(&ptr, cinput, cbool(isFetch))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	spec := &Refspec{ptr: ptr}
	runtime.SetFinalizer(spec, (*Refspec).Free)
	return spec, nil
}

// Free releases a refspec object which has been created by ParseRefspec
func (s *Refspec) Free() {
	runtime.SetFinalizer(s, nil)
	C.git_refspec_free(s.ptr)
}

// Direction returns the refspec's direction
func (s *Refspec) Direction() ConnectDirection {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	direction := C.git_refspec_direction(s.ptr)
	return ConnectDirection(direction)
}

// Src returns the refspec's source specifier
func (s *Refspec) Src() string {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ret string
	cstr := C.git_refspec_src(s.ptr)

	if cstr != nil {
		ret = C.GoString(cstr)
	}

	runtime.KeepAlive(s)
	return ret
}

// Dst returns the refspec's destination specifier
func (s *Refspec) Dst() string {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ret string
	cstr := C.git_refspec_dst(s.ptr)

	if cstr != nil {
		ret = C.GoString(cstr)
	}

	runtime.KeepAlive(s)
	return ret
}

// Force returns the refspec's force-update setting
func (s *Refspec) Force() bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	force := C.git_refspec_force(s.ptr)
	return force != 0
}

// String returns the refspec's string representation
func (s *Refspec) String() string {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ret string
	cstr := C.git_refspec_string(s.ptr)

	if cstr != nil {
		ret = C.GoString(cstr)
	}

	runtime.KeepAlive(s)
	return ret
}

// SrcMatches checks if a refspec's source descriptor matches a reference
func (s *Refspec) SrcMatches(refname string) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cname := C.CString(refname)
	defer C.free(unsafe.Pointer(cname))

	matches := C.git_refspec_src_matches(s.ptr, cname)
	return matches != 0
}

// SrcMatches checks if a refspec's destination descriptor matches a reference
func (s *Refspec) DstMatches(refname string) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cname := C.CString(refname)
	defer C.free(unsafe.Pointer(cname))

	matches := C.git_refspec_dst_matches(s.ptr, cname)
	return matches != 0
}

// Transform a reference to its target following the refspec's rules
func (s *Refspec) Transform(refname string) (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	buf := C.git_buf{}
	defer C.git_buf_dispose(&buf)

	cname := C.CString(refname)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_refspec_transform(&buf, s.ptr, cname)
	if ret < 0 {
		return "", MakeGitError(ret)
	}

	return C.GoString(buf.ptr), nil
}

// Rtransform converts a target reference to its source reference following the
// refspec's rules
func (s *Refspec) Rtransform(refname string) (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	buf := C.git_buf{}
	defer C.git_buf_dispose(&buf)

	cname := C.CString(refname)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_refspec_rtransform(&buf, s.ptr, cname)
	if ret < 0 {
		return "", MakeGitError(ret)
	}

	return C.GoString(buf.ptr), nil
}

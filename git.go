package git

/*
#include <git2.h>
#include <git2/sys/openssl.h>
*/
import "C"
import (
	"bytes"
	"encoding/hex"
	"errors"
	"runtime"
	"strings"
	"unsafe"
)

//go:generate stringer -type ErrorClass -trimprefix ErrorClass -tags static
type ErrorClass int

const (
	ErrorClassNone       ErrorClass = C.GIT_ERROR_NONE
	ErrorClassNoMemory   ErrorClass = C.GIT_ERROR_NOMEMORY
	ErrorClassOS         ErrorClass = C.GIT_ERROR_OS
	ErrorClassInvalid    ErrorClass = C.GIT_ERROR_INVALID
	ErrorClassReference  ErrorClass = C.GIT_ERROR_REFERENCE
	ErrorClassZlib       ErrorClass = C.GIT_ERROR_ZLIB
	ErrorClassRepository ErrorClass = C.GIT_ERROR_REPOSITORY
	ErrorClassConfig     ErrorClass = C.GIT_ERROR_CONFIG
	ErrorClassRegex      ErrorClass = C.GIT_ERROR_REGEX
	ErrorClassOdb        ErrorClass = C.GIT_ERROR_ODB
	ErrorClassIndex      ErrorClass = C.GIT_ERROR_INDEX
	ErrorClassObject     ErrorClass = C.GIT_ERROR_OBJECT
	ErrorClassNet        ErrorClass = C.GIT_ERROR_NET
	ErrorClassTag        ErrorClass = C.GIT_ERROR_TAG
	ErrorClassTree       ErrorClass = C.GIT_ERROR_TREE
	ErrorClassIndexer    ErrorClass = C.GIT_ERROR_INDEXER
	ErrorClassSSL        ErrorClass = C.GIT_ERROR_SSL
	ErrorClassSubmodule  ErrorClass = C.GIT_ERROR_SUBMODULE
	ErrorClassThread     ErrorClass = C.GIT_ERROR_THREAD
	ErrorClassStash      ErrorClass = C.GIT_ERROR_STASH
	ErrorClassCheckout   ErrorClass = C.GIT_ERROR_CHECKOUT
	ErrorClassFetchHead  ErrorClass = C.GIT_ERROR_FETCHHEAD
	ErrorClassMerge      ErrorClass = C.GIT_ERROR_MERGE
	ErrorClassSSH        ErrorClass = C.GIT_ERROR_SSH
	ErrorClassFilter     ErrorClass = C.GIT_ERROR_FILTER
	ErrorClassRevert     ErrorClass = C.GIT_ERROR_REVERT
	ErrorClassCallback   ErrorClass = C.GIT_ERROR_CALLBACK
	ErrorClassRebase     ErrorClass = C.GIT_ERROR_REBASE
	ErrorClassPatch      ErrorClass = C.GIT_ERROR_PATCH
)

//go:generate stringer -type ErrorCode -trimprefix ErrorCode -tags static
type ErrorCode int

const (
	// ErrorCodeOK indicates that the operation completed successfully.
	ErrorCodeOK ErrorCode = C.GIT_OK

	// ErrorCodeGeneric represents a generic error.
	ErrorCodeGeneric ErrorCode = C.GIT_ERROR
	// ErrorCodeNotFound represents that the requested object could not be found
	ErrorCodeNotFound ErrorCode = C.GIT_ENOTFOUND
	// ErrorCodeExists represents that the object exists preventing operation.
	ErrorCodeExists ErrorCode = C.GIT_EEXISTS
	// ErrorCodeAmbiguous represents that more than one object matches.
	ErrorCodeAmbiguous ErrorCode = C.GIT_EAMBIGUOUS
	// ErrorCodeBuffs represents that the output buffer is too short to hold data.
	ErrorCodeBuffs ErrorCode = C.GIT_EBUFS

	// ErrorCodeUser is a special error that is never generated by libgit2
	// code.  You can return it from a callback (e.g to stop an iteration)
	// to know that it was generated by the callback and not by libgit2.
	ErrorCodeUser ErrorCode = C.GIT_EUSER

	// ErrorCodeBareRepo represents that the operation not allowed on bare repository
	ErrorCodeBareRepo ErrorCode = C.GIT_EBAREREPO
	// ErrorCodeUnbornBranch represents that HEAD refers to branch with no commits.
	ErrorCodeUnbornBranch ErrorCode = C.GIT_EUNBORNBRANCH
	// ErrorCodeUnmerged represents that a merge in progress prevented operation.
	ErrorCodeUnmerged ErrorCode = C.GIT_EUNMERGED
	// ErrorCodeNonFastForward represents that the reference was not fast-forwardable.
	ErrorCodeNonFastForward ErrorCode = C.GIT_ENONFASTFORWARD
	// ErrorCodeInvalidSpec represents that the name/ref spec was not in a valid format.
	ErrorCodeInvalidSpec ErrorCode = C.GIT_EINVALIDSPEC
	// ErrorCodeConflict represents that checkout conflicts prevented operation.
	ErrorCodeConflict ErrorCode = C.GIT_ECONFLICT
	// ErrorCodeLocked represents that lock file prevented operation.
	ErrorCodeLocked ErrorCode = C.GIT_ELOCKED
	// ErrorCodeModified represents that the reference value does not match expected.
	ErrorCodeModified ErrorCode = C.GIT_EMODIFIED
	// ErrorCodeAuth represents that the authentication failed.
	ErrorCodeAuth ErrorCode = C.GIT_EAUTH
	// ErrorCodeCertificate represents that the server certificate is invalid.
	ErrorCodeCertificate ErrorCode = C.GIT_ECERTIFICATE
	// ErrorCodeApplied represents that the patch/merge has already been applied.
	ErrorCodeApplied ErrorCode = C.GIT_EAPPLIED
	// ErrorCodePeel represents that the requested peel operation is not possible.
	ErrorCodePeel ErrorCode = C.GIT_EPEEL
	// ErrorCodeEOF represents an unexpected EOF.
	ErrorCodeEOF ErrorCode = C.GIT_EEOF
	// ErrorCodeInvalid represents an invalid operation or input.
	ErrorCodeInvalid ErrorCode = C.GIT_EINVALID
	// ErrorCodeUIncommitted represents that uncommitted changes in index prevented operation.
	ErrorCodeUncommitted ErrorCode = C.GIT_EUNCOMMITTED
	// ErrorCodeDirectory represents that the operation is not valid for a directory.
	ErrorCodeDirectory ErrorCode = C.GIT_EDIRECTORY
	// ErrorCodeMergeConflict represents that a merge conflict exists and cannot continue.
	ErrorCodeMergeConflict ErrorCode = C.GIT_EMERGECONFLICT

	// ErrorCodePassthrough represents that a user-configured callback refused to act.
	ErrorCodePassthrough ErrorCode = C.GIT_PASSTHROUGH
	// ErrorCodeIterOver signals end of iteration with iterator.
	ErrorCodeIterOver ErrorCode = C.GIT_ITEROVER
	// ErrorCodeRetry is an internal-only error code.
	ErrorCodeRetry ErrorCode = C.GIT_RETRY
	// ErrorCodeMismatch represents a hashsum mismatch in object.
	ErrorCodeMismatch ErrorCode = C.GIT_EMISMATCH
	// ErrorCodeIndexDirty represents that unsaved changes in the index would be overwritten.
	ErrorCodeIndexDirty ErrorCode = C.GIT_EINDEXDIRTY
	// ErrorCodeApplyFail represents that a patch application failed.
	ErrorCodeApplyFail ErrorCode = C.GIT_EAPPLYFAIL
)

var (
	ErrInvalid = errors.New("Invalid state for operation")
)

// doNotCompare is an idiomatic way of making structs non-comparable to avoid
// future field additions to make them non-comparable.
type doNotCompare [0]func()

var pointerHandles *HandleList
var remotePointers *remotePointerList

func init() {
	initLibGit2()
}

func initLibGit2() {
	pointerHandles = NewHandleList()
	remotePointers = newRemotePointerList()

	C.git_libgit2_init()
	features := Features()

	if features&FeatureHTTPS == 0 {
		if err := registerManagedHTTP(); err != nil {
			panic(err)
		}
	} else {
		// This is not something we should be doing, as we may be stomping all over
		// someone else's setup. The user should do this themselves or use some
		// binding/wrapper which does it in such a way that they can be sure
		// they're the only ones setting it up.
		C.git_openssl_set_locking()
	}
	if features&FeatureSSH == 0 {
		if err := registerManagedSSH(); err != nil {
			panic(err)
		}
	}
}

// Shutdown frees all the resources acquired by libgit2. Make sure no
// references to any git2go objects are live before calling this.
// After this is called, invoking any function from this library will result in
// undefined behavior, so make sure this is called carefully.
func Shutdown() {
	if err := unregisterManagedTransports(); err != nil {
		panic(err)
	}
	pointerHandles.Clear()
	remotePointers.clear()

	C.git_libgit2_shutdown()
}

// ReInit reinitializes the global state, this is useful if the effective user
// id has changed and you want to update the stored search paths for gitconfig
// files. This function frees any references to objects, so it should be called
// before any other functions are called.
func ReInit() {
	Shutdown()
	initLibGit2()
}

// Oid represents the id for a Git object.
type Oid [20]byte

func newOidFromC(coid *C.git_oid) *Oid {
	if coid == nil {
		return nil
	}

	oid := new(Oid)
	copy(oid[0:20], C.GoBytes(unsafe.Pointer(coid), 20))
	return oid
}

func NewOidFromBytes(b []byte) *Oid {
	oid := new(Oid)
	copy(oid[0:20], b[0:20])
	return oid
}

func (oid *Oid) toC() *C.git_oid {
	return (*C.git_oid)(unsafe.Pointer(oid))
}

func NewOid(s string) (*Oid, error) {
	if len(s) > C.GIT_OID_HEXSZ {
		return nil, errors.New("string is too long for oid")
	}

	o := new(Oid)

	slice, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	if len(slice) != 20 {
		return nil, &GitError{"invalid oid", ErrorClassNone, ErrorCodeGeneric}
	}

	copy(o[:], slice[:20])
	return o, nil
}

func (oid *Oid) String() string {
	return hex.EncodeToString(oid[:])
}

func (oid *Oid) Cmp(oid2 *Oid) int {
	return bytes.Compare(oid[:], oid2[:])
}

func (oid *Oid) Copy() *Oid {
	ret := *oid
	return &ret
}

func (oid *Oid) Equal(oid2 *Oid) bool {
	return *oid == *oid2
}

func (oid *Oid) IsZero() bool {
	return *oid == Oid{}
}

func (oid *Oid) NCmp(oid2 *Oid, n uint) int {
	return bytes.Compare(oid[:n], oid2[:n])
}

func ShortenOids(ids []*Oid, minlen int) (int, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	shorten := C.git_oid_shorten_new(C.size_t(minlen))
	if shorten == nil {
		panic("Out of memory")
	}
	defer C.git_oid_shorten_free(shorten)

	var ret C.int

	for _, id := range ids {
		buf := make([]byte, 41)
		C.git_oid_fmt((*C.char)(unsafe.Pointer(&buf[0])), id.toC())
		buf[40] = 0
		ret = C.git_oid_shorten_add(shorten, (*C.char)(unsafe.Pointer(&buf[0])))
		if ret < 0 {
			return int(ret), MakeGitError(ret)
		}
	}
	runtime.KeepAlive(ids)
	return int(ret), nil
}

type GitError struct {
	Message string
	Class   ErrorClass
	Code    ErrorCode
}

func (e GitError) Error() string {
	return e.Message
}

func IsErrorClass(err error, c ErrorClass) bool {
	if err == nil {
		return false
	}
	if gitError, ok := err.(*GitError); ok {
		return gitError.Class == c
	}
	return false
}

func IsErrorCode(err error, c ErrorCode) bool {
	if err == nil {
		return false
	}
	if gitError, ok := err.(*GitError); ok {
		return gitError.Code == c
	}
	return false
}

func MakeGitError(c C.int) error {
	var errMessage string
	var errClass ErrorClass
	errorCode := ErrorCode(c)
	if errorCode != ErrorCodeIterOver {
		err := C.git_error_last()
		if err != nil {
			errMessage = C.GoString(err.message)
			errClass = ErrorClass(err.klass)
		} else {
			errClass = ErrorClassInvalid
		}
	}
	if errMessage == "" {
		errMessage = errorCode.String()
	}
	return &GitError{errMessage, errClass, errorCode}
}

func MakeGitError2(err int) error {
	return MakeGitError(C.int(err))
}

func cbool(b bool) C.int {
	if b {
		return C.int(1)
	}
	return C.int(0)
}

func ucbool(b bool) C.uint {
	if b {
		return C.uint(1)
	}
	return C.uint(0)
}

func setCallbackError(errorMessage **C.char, err error) C.int {
	if err != nil {
		*errorMessage = C.CString(err.Error())
		if gitError, ok := err.(*GitError); ok {
			return C.int(gitError.Code)
		}
		return C.int(ErrorCodeUser)
	}
	return C.int(ErrorCodeOK)
}

func Discover(start string, across_fs bool, ceiling_dirs []string) (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ceildirs := C.CString(strings.Join(ceiling_dirs, string(C.GIT_PATH_LIST_SEPARATOR)))
	defer C.free(unsafe.Pointer(ceildirs))

	cstart := C.CString(start)
	defer C.free(unsafe.Pointer(cstart))

	var buf C.git_buf
	defer C.git_buf_dispose(&buf)

	ret := C.git_repository_discover(&buf, cstart, cbool(across_fs), ceildirs)
	if ret < 0 {
		return "", MakeGitError(ret)
	}

	return C.GoString(buf.ptr), nil
}

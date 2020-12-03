package rdp

import (
	"encoding/binary"
	"fmt"
	"log"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	CRYPTPROTECT_UI_FORBIDDEN = 0x1
)

var (
	dLLCrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dLLKernel32 = syscall.NewLazyDLL("Kernel32.dll")

	procEncryptData = dLLCrypt32.NewProc("CryptProtectData")
	procDecryptData = dLLCrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dLLKernel32.NewProc("LocalFree")
)

// CrossPlatformAuthHandler handles the usage of the password. On Windows this is encrypted and included in the .rdp
// file. On Mac OS this is not possible. Currently Mac users will just have the password copied to their clipboard.
// The correct way to do this on mac is probably using the KeyRing. See https://github.com/danhale-git/runrdp/issues/38.
func CrossPlatformAuthHandler(fileBody, password string) string {
	enc, err := encrypt(password)
	if err != nil {
		log.Fatalf("encrypt failed: %v", err)
	}

	return fileBody + fmt.Sprintf("\npassword 51:b:%s", fmt.Sprintf("%x", enc))
}

type DataBlob struct {
	cbData uint32
	pbData *byte
}

func newBlob(d []byte) *DataBlob {
	if len(d) == 0 {
		return &DataBlob{}
	}
	return &DataBlob{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *DataBlob) toByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

func encrypt(s string) ([]byte, error) {
	data := convertToUTF16LittleEndianBytes(s)
	var outBlob DataBlob
	r, _, err := procEncryptData.Call(uintptr(unsafe.Pointer(newBlob(data))), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&outBlob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outBlob.pbData)))
	return outBlob.toByteArray(), nil
}

func convertToUTF16LittleEndianBytes(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, 2*len(u))
	for index, value := range u {
		binary.LittleEndian.PutUint16(b[index*2:], value)
	}
	return b
}

/*func Decrypt(data []byte) ([]byte, error) {
	var outblob DataBlob
	r, _, err := procDecryptData.Call(uintptr(unsafe.Pointer(newBlob(data))), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))
	return outblob.toByteArray(), nil
}*/

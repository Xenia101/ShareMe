package share

import (
	"os"
)

const (
	PermDir  = os.FileMode(0700) // 디렉토리 권한
	PermFile = os.FileMode(0400) // 파일 권한
)

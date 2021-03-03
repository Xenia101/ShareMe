package share

import (
	"os"
	"time"
)

const (
	DirPublic = "public"  // 프론트엔드 디렉토리
	DirUpload = "uploads" // 업로드 디렉토리

	PermDir  = os.FileMode(0700) // 디렉토리 권한
	PermFile = os.FileMode(0400) // 파일 권한

	/**
	ExpireUpload   = 24 * time.Hour                // 업로드 제한시간
	ExpireUnlocked = 4 * 7 * 24 * time.Hour        // 파일의 유효기간
	ExpireLocked   = ExpireUnlocked + 24*time.Hour // 다운로드중이라고 표시된 파일의 강제 삭제 기한
	*/

	ExpireUpload   = 24 * time.Second // 업로드 제한시간
	ExpireUnlocked = 1 * time.Minute  // 파일의 유효기간
	ExpireLocked   = 2 * time.Minute  // 다운로드중이라고 표시된 파일의 강제 삭제 기한

	CleanupInterval   = 30 * time.Second // 정리 텀
	CleanupQueryLimit = 100              // 정리할 떄 한번에 불러올 Rows 수

	IdMin   = 5
	IdMax   = 32
	IdChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

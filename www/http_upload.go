package www

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/RyuaNerin/ShareMe/share"
	"github.com/gin-gonic/gin"
)

const (
	IdLength = 10
	IdChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

// with panic
func newFileId(ctx context.Context) string {
	tx, err := share.DB.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	var idRaw [IdLength]byte
	var id string

	for {
		for i := 0; i < IdLength; i++ {
			idRaw[i] = IdChars[rand.Intn(len(IdChars))]
		}

		id = *(*string)(unsafe.Pointer(&idRaw))
		r, err := tx.ExecContext(
			ctx,
			`
			INSERT IGNORE
			INTO
				files ( id )
			VALUES
				( ? )
			`,
			id,
		)
		if err != nil {
			panic(err)
		}
		ra, err := r.RowsAffected()
		if err != nil {
			panic(err)
		}
		if ra == 1 {
			break
		}
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	return string(id)
}

func handleUpload(c *gin.Context) {
	ctx := c.Request.Context()

	// 업로드 마무리 되지 않은 것은 cleaner 가 알아서 지워줄 것.
	id := newFileId(ctx)

	mpForm, err := c.MultipartForm()
	if err != nil {
		panic(err)
	}
	defer mpForm.RemoveAll()

	//////////////////////////////////////////////////
	// sha256 = 32 bytes
	var encKey []byte
	var encIv []byte

	if pw, ok := mpForm.Value["password"]; ok && len(pw) == 1 && pw[0] != "" {
		encKey, encIv = genAES(pw[0])
	}

	//////////////////////////////////////////////////

	mpFiles, ok := mpForm.File["myFile"]
	if ok && len(mpFiles) != 1 {
		c.Status(http.StatusBadRequest)
		return
	}

	filename := mpFiles[0].Filename

	mpFile, err := mpFiles[0].Open()
	if err != nil {
		panic(err)
	}
	defer mpFile.Close()

	dir := filepath.Join(share.DirUpload, id[:2])
	os.MkdirAll(dir, share.PermDir)

	{
		fs, err := os.Create(filepath.Join(dir, id))
		if err != nil {
			panic(err)
		}
		defer fs.Close()

		fsbw := bufio.NewWriter(fs)
		{
			var w io.Writer
			if encKey == nil {
				w = fsbw
			} else {
				// AES-32
				aesCipher, _ := aes.NewCipher(encKey)
				block := cipher.NewCTR(aesCipher, encIv)

				w = &cipher.StreamWriter{
					S: block,
					W: fsbw,
				}
			}

			_, err = io.Copy(w, mpFile)
			if err != nil && err != io.EOF {
				panic(err)
			}

			err = fsbw.Flush()
			if err != nil {
				panic(err)
			}

			err = fs.Sync()
			if err != nil {
				panic(err)
			}

			fs.Close()
			mpFile.Close()
		}
	}

	//////////////////////////////////////////////////

	tx, err := share.DB.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	tx.ExecContext(
		ctx,
		`
		UPDATE
			files
		VALUES
			uploaded = 1,
			filename = ?,
			expires  = ?
		WHERE
			id = ?
		`,
		filename,
		time.Now().Add(share.ExpireLocked),
		id,
	)

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	//////////////////////////////////////////////////

	c.HTML(http.StatusOK, "code.htm", id)
}

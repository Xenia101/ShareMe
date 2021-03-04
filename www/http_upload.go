package www

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/RyuaNerin/ShareMe/share"
	"github.com/gin-gonic/gin"
)

// i j l 1 같은 혼동되는 문자 방지
func genId(idRaw []byte, l int) string {
	il1 := false
	o0 := false
	s5 := false
	z2 := false

	i := 0
	for i < l {
		c := share.IdChars[rand.Intn(len(share.IdChars))]

		if c == 'I' || c == 'i' || c == 'j' || c == 'l' || c == '1' {
			if il1 {
				continue
			}
			il1 = true
		}

		if c == 'O' || c == 'o' || c == '0' {
			if o0 {
				continue
			}
			o0 = true
		}

		if c == 'S' || c == 's' || c == '5' {
			if s5 {
				continue
			}
			s5 = true
		}

		if c == 'Z' || c == 'z' || c == '2' {
			if z2 {
				continue
			}
			z2 = true
		}

		idRaw[i] = c
		i++
	}

	return share.ToString(idRaw[:l])
}

// with panic
func newFileId(ctx context.Context, remoteAddr string) string {
	tx, err := share.DB.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	var id string
	var idRaw [share.IdMax]byte
	idLen := share.IdMin
	idTry := 0

	for {
		if idTry >= 2 {
			idLen += 1
			if idLen > share.IdMax {
				idLen = share.IdMax
			}
			idTry = 0
		}
		id = genId(idRaw[:], idLen)

		r, err := tx.ExecContext(
			ctx,
			`
			INSERT IGNORE
			INTO
				files ( id, remote_addr )
			VALUES
				( ?, ? )
			`,
			id,
			remoteAddr,
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

		idTry += 1
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	return string(id)
}

func handleUpload(c *gin.Context) {
	ctx := c.Request.Context()

	// Remote address
	uploader_ip := c.Request.RemoteAddr
	if h := c.GetHeader("X-Forwarded-For"); h != "" {
		uploader_ip = h
	} else {
		if h, _, err := net.SplitHostPort(uploader_ip); h != "" && err != nil {
			uploader_ip = h
		}
	}

	// 업로드 마무리 되지 않은 것은 cleaner 가 알아서 지워줄 것.
	id := newFileId(ctx, uploader_ip)

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

	r, err := tx.ExecContext(
		ctx,
		`
		UPDATE
			files
		SET
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
	if err != nil {
		panic(err)
	}
	_, err = r.RowsAffected()
	if err != nil {
		panic(err)
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	//////////////////////////////////////////////////

	c.HTML(http.StatusOK, "code.htm", id)
}

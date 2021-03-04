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

	"shareme/share"

	"github.com/gin-gonic/gin"
)

// i j l 1 같은 혼동되는 문자 방지
func genId(idRaw []byte, l int) string {
	cb := make([]bool, len(share.Config.IDRule.Conflict))
	cbi := -1

	i := 0
	for i < l {
		c := share.Config.IDRule.Chars[rand.Intn(len(share.Config.IDRule.Chars))]

		cbi = -1
		for i, cc := range share.Config.IDRule.Conflict {
			for _, ccc := range cc {
				if c == ccc {
					cbi = i
					break
				}
			}
			if cbi != -1 {
				break
			}
		}

		if cbi != -1 {
			if cb[cbi] {
				continue
			}
			cb[cbi] = true
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
	idRaw := make([]byte, share.Config.IDRule.Max)
	idLen := share.Config.IDRule.Min
	idTry := 0

	for {
		if idTry >= 2 {
			idLen += 1
			if idLen > share.Config.IDRule.Max {
				idLen = share.Config.IDRule.Max
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

	dir := filepath.Join(share.Config.Dir.Upload, id[:2])
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
			filename = ?
		WHERE
			id = ?
		`,
		filename,
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

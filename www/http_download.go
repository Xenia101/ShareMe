package www

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/RyuaNerin/ShareMe/share"
	"github.com/gin-gonic/gin"
)

func handleDownload(c *gin.Context) {
	id := c.PostForm("id")
	if id == "" {
		c.Status(http.StatusNotFound)
		return
	}

	handleDownloadWithId(c, id, c.PostForm("password"))
}

// false : bad request
func changeLock(ctx context.Context, id string, lock bool) (ok bool, fileName string) {
	tx, err := share.DB.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	//////////////////////////////////////////////////

	if lock {
		var expires time.Time
		err = tx.QueryRowContext(
			ctx,
			`
			SELECT
				expires,
				filename
			FROM
				files
			WHERE
				id = ? AND
				uploaded = 1
			`,
			id,
		).Scan(
			&expires,
			&fileName,
		)
		if err != nil {
			panic(err)
		}
		if expires.Before(time.Now()) {
			return
		}
	}

	//////////////////////////////////////////////////

	v := -1
	if lock {
		v = 1
	}

	r, err := tx.ExecContext(
		ctx,
		`
		UPDATE
			files
		VALUES
			lock = lock + ?
		WHERE
			id = ?
		`,
		v,
		id,
	)
	if err != nil {
		panic(err)
	}

	ra, err := r.RowsAffected()
	if err != nil {
		panic(err)
	}
	if ra != 1 {
		return
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	ok = true
	return
}

func handleDownloadWithId(c *gin.Context, id string, password string) {
	ctx := c.Request.Context()

	ok, fileName := changeLock(ctx, id, true)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	defer changeLock(ctx, id, false)

	//////////////////////////////////////////////////

	var encKey []byte
	var encIv []byte

	if password != "" {
		encKey, encIv = genAES(password)
	}

	fs, err := os.Open(filepath.Join("uploads", id[:2], id))
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	fsbr := bufio.NewReader(fs)

	var r io.Reader
	if encKey == nil {
		r = fsbr
	} else {
		// AES-32
		aesCipher, _ := aes.NewCipher(encKey)
		block := cipher.NewCTR(aesCipher, encIv)

		r = &cipher.StreamReader{
			S: block,
			R: fsbr,
		}
	}

	//////////////////////////////////////////////////

	contentDisposition := mime.FormatMediaType(
		"attachment",
		map[string]string{
			"filename": fileName,
		},
	)

	headers := c.Writer.Header()
	headers.Set("content-disposition", contentDisposition)
	headers.Set("Content-Type", "application/zip; charset=utf-8")
	headers.Set("Transper-Encoding", "chunked")

	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, r)
	if err != nil && err != io.EOF {
		panic(err)
	}

	fs.Close()
}

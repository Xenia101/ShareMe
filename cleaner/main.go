package cleaner

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/RyuaNerin/ShareMe/share"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

func Main() {
	time.AfterFunc(share.CleanupInterval, worker)
}

func worker() {
	defer time.AfterFunc(share.CleanupInterval, worker)

	remove(
		`
		SELECT
			id
		FROM
			files
		WHERE
			created_at < ? AND
			uploaded = 0
		LIMIT
			?, ?
		`,
		time.Now().Add(-share.ExpireUpload),
	)

	remove(
		`
		SELECT
			id
		FROM
			files
		WHERE
			expires < ? AND
			uploaded = 1 AND
			lock = 0
		LIMIT
			?, ?
		`,
		time.Now(),
	)

	remove(
		`
		SELECT
			id
		FROM
			files
		WHERE
			expires < ? AND
			uploaded = 1
		LIMIT
			?, ?
		`,
		time.Now().Add(-share.ExpireLocked),
	)
}

func remove(query string, arg ...interface{}) {
	defer func() {
		if errRaw := recover(); errRaw != nil {
			err := errRaw.(error)

			log.Println(errors.WithStack(err))
			sentry.CaptureException(err)
		}
	}()

	removedIds := make([]string, 0, share.CleanupQueryLimit)
	var id string

	stmt, err := share.DB.Prepare(
		`
		DELETE
		FROM
			files
		WHERE
			id = ?
		`,
	)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	index := 0
	for {
		rows, err := share.DB.Query(
			query,
			index,
			share.CleanupQueryLimit,
		)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		removedIds = removedIds[:0]
		for rows.Next() {
			err = rows.Scan(&id)
			if err != nil {
				sentry.CaptureException(err)
				continue
			}

			path := filepath.Join(share.DirPublic, id[:2], id)
			err := os.Remove(path)
			if err != nil {
				sentry.CaptureException(err)
			}

			index += 1
		}

		rows.Close()

		for _, id = range removedIds {
			_, err = stmt.Exec(id)
			if err != nil {
				sentry.CaptureException(err)
			}
		}
	}
}

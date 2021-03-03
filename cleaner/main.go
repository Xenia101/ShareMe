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
	go worker()
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
			created_at <= ? AND
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
			expires <= ? AND
			uploaded = 1 AND
			downloading = 0
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
			expires <= ? AND
			uploaded = 1
		LIMIT
			?, ?
		`,
		time.Now().Add(-share.ExpireLocked),
	)
}

func remove(query string, dt time.Time) {
	defer func() {
		if errRaw := recover(); errRaw != nil {
			err := errRaw.(error)

			log.Printf("%+v\n", errors.WithStack(err))
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
	count := 0
	for {
		rows, err := share.DB.Query(
			query,
			dt,
			index,
			share.CleanupQueryLimit,
		)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		count = 0
		removedIds = removedIds[:0]
		for rows.Next() {
			count += 1
			err = rows.Scan(&id)
			if err != nil {
				sentry.CaptureException(err)
				continue
			}

			path := filepath.Join(share.DirUpload, id[:2], id)
			err := os.Remove(path)
			if err != nil {
				sentry.CaptureException(err)
			}

			removedIds = append(removedIds, id)

			index += 1
		}

		rows.Close()

		if count == 0 {
			break
		}

		for _, id = range removedIds {
			_, err = stmt.Exec(id)
			if err != nil {
				sentry.CaptureException(err)
			}
		}
	}
}

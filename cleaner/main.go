package cleaner

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"shareme/share"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

func Main() {
	go worker()
}

func worker() {
	defer time.AfterFunc(share.Config.Cleanup.Duration, worker)

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
		time.Now().Add(-share.Config.Expires.Upload),
	)

	remove(
		`
		SELECT
			id
		FROM
			files
		WHERE
			created_at <= ? AND
			uploaded = 1 AND
			downloading = 0
		LIMIT
			?, ?
		`,
		time.Now().Add(-share.Config.Expires.Idle),
	)

	remove(
		`
		SELECT
			id
		FROM
			files
		WHERE
			created_at <= ? AND
			uploaded = 1
		LIMIT
			?, ?
		`,
		time.Now().Add(-share.Config.Expires.Forced),
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

	removedIds := make([]string, 0, share.Config.Cleanup.QueryLimit)
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
			share.Config.Cleanup.QueryLimit,
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

			path := filepath.Join(share.Config.Dir.Upload, id[:2], id)
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

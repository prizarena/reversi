package revdal

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/strongo/dalgo/dal"
	"github.com/strongo/dalgo2firestore"
)

var DB dal.Database

func NewDatabase(ctx context.Context) (dal.Database, error) {
	fsClient, err := firestore.NewClient(ctx, "prizarena")
	if err != nil {
		return nil, err
	}
	return dalgo2firestore.NewDatabase(fsClient), nil
}

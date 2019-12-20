package store

import (
	"log"
	"os"
	"strings"
	"time"

	"context"
	"errors"
	"fmt"

	"crypto/md5"
	"encoding/hex"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/mchmarny/gcputil/project"
)

const (

	// ISODateFormat is the YYYY-MM-DD format
	ISODateFormat = "2006-01-02"

	recordIDPrefix = "id-"
)

var (
	logger    = log.New(os.Stdout, "data: ", 0)
	projectID = project.GetIDOrFail()
	fsClient  *firestore.Client

	// ErrDataNotFound is thrown when query does not find the requested data
	ErrDataNotFound = errors.New("Data not found")
)

func getClient(ctx context.Context) (client *firestore.Client, err error) {

	if fsClient == nil {
		c, err := firestore.NewClient(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("Error while creating Firestore client: %v", err)
		}
		fsClient = c
	}

	return fsClient, nil
}

func getCollection(ctx context.Context, name string) (col *firestore.CollectionRef, err error) {

	if name == "" {
		return nil, errors.New("Nil name")
	}

	c, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error while creating Firestore client: %v", err)
	}

	return c.Collection(name), nil
}

func deleteByID(ctx context.Context, col, id string) error {

	if id == "" {
		return errors.New("Nil id")
	}

	c, err := getCollection(ctx, col)
	if err != nil {
		return err
	}

	_, err = c.Doc(id).Delete(ctx)

	if grpc.Code(err) == codes.NotFound {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error getting state: %v", err)
	}

	return nil
}

// IsDataNotFoundError checks boolions on whether the error is result of data not found
func IsDataNotFoundError(err error) bool {
	return err != nil && err.Error() == ErrDataNotFound.Error()
}

func getByID(ctx context.Context, col, id string, in interface{}) error {

	if id == "" {
		return errors.New("Nil id")
	}

	c, err := getCollection(ctx, col)
	if err != nil {
		return err
	}

	d, err := c.Doc(id).Get(ctx)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return ErrDataNotFound
		}
		return fmt.Errorf("Error getting state: %v", err)
	}

	if d == nil || d.Data() == nil {
		return fmt.Errorf("record with id %s found in %s collection but has not data", id, col)
	}

	if err := d.DataTo(in); err != nil {
		return fmt.Errorf("stored data is not of user type: %v", err)
	}

	return nil
}

func save(ctx context.Context, col, id string, in interface{}) error {

	if in == nil {
		return errors.New("Nil state")
	}

	c, err := getCollection(ctx, col)
	if err != nil {
		return err
	}

	_, err = c.Doc(id).Set(ctx, in)
	if err != nil {
		return fmt.Errorf("Error on save: %v", err)
	}
	return nil
}

// NewID generates new ID using UUID v4
func NewID() string {
	return ToID(uuid.New().String())
}

// NormalizeString makes val comparable regardless of case or whitespace
func NormalizeString(val string) string {
	return strings.TrimSpace(strings.ToLower(val))
}

// ToID hashes the passed string into a valid ID
func ToID(val string) string {
	hash := md5.Sum([]byte(NormalizeString(val)))
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("%s%s", recordIDPrefix, hashStr)
}

func getDateRange(since time.Time) []time.Time {

	r := make([]time.Time, 0)
	today := time.Now().UTC().Format(ISODateFormat)
	if since.Format(ISODateFormat) > today {
		since = time.Now().UTC()
	}

	for {
		r = append(r, since)
		if since.Format(ISODateFormat) >= today {
			break
		}
		since = since.AddDate(0, 0, 1)
	}

	return r
}

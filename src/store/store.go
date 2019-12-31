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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	ErrDataNotFound = errors.New("data not found")
)

func getClient(ctx context.Context) (client *firestore.Client, err error) {

	if fsClient == nil {
		c, err := firestore.NewClient(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("error while creating Firestore client: %v", err)
		}
		fsClient = c
	}

	return fsClient, nil
}

func getCollection(ctx context.Context, name string) (col *firestore.CollectionRef, err error) {

	if name == "" {
		return nil, errors.New("nil name")
	}

	c, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error while creating Firestore client: %v", err)
	}

	return c.Collection(name), nil
}

func deleteByID(ctx context.Context, col, id string) error {

	if id == "" {
		return errors.New("nil id")
	}

	c, err := getCollection(ctx, col)
	if err != nil {
		return err
	}

	_, err = c.Doc(id).Delete(ctx)

	if status.Code(err) == codes.NotFound {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting state: %v", err)
	}

	return nil
}

// IsDataNotFoundError checks boolions on whether the error is result of data not found
func IsDataNotFoundError(err error) bool {
	return err != nil && err.Error() == ErrDataNotFound.Error()
}

func getByID(ctx context.Context, col, id string, in interface{}) error {

	if id == "" {
		return errors.New("nil id")
	}

	c, err := getCollection(ctx, col)
	if err != nil {
		return err
	}

	d, err := c.Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return ErrDataNotFound
		}
		return fmt.Errorf("error getting state: %v", err)
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
		return errors.New("nil state")
	}

	c, err := getCollection(ctx, col)
	if err != nil {
		return err
	}

	_, err = c.Doc(id).Set(ctx, in)
	if err != nil {
		return fmt.Errorf("error on save: %v", err)
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

// PrettyDurationSince prints pretty duration since date
func PrettyDurationSince(a time.Time) string {

	b := time.Now().UTC()

	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	year := int(y2 - y1)
	month := int(M2 - M1)
	day := int(d2 - d1)

	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}

	if month < 0 {
		month += 12
		year--
	}

	if year > 0 {
		return fmt.Sprintf("%d years, %d months, and %d days", year, month, day)
	}

	if month > 1 {
		return fmt.Sprintf("%d months and %d days", month, day)
	}

	if day == 1 {
		return fmt.Sprintf("%d day", day)
	}

	return fmt.Sprintf("%d days", day)

}

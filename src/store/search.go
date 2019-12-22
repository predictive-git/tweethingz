package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

const (
	searchCriteriaCollectionName = "thingz_search_criteria"
	searchResultCollectionName   = "thingz_search_result"
)

//============================================================================
// Criteria
//============================================================================

// SearchCriteria is the flat version of search criteria for simplicity of form binding
type SearchCriteria struct {
	ID   string `firestore:"id" json:"id" form:"id"`
	User string `firestore:"user" json:"user" form:"user"`

	Name  string `firestore:"name" json:"name" form:"name"`
	Value string `firestore:"value" json:"value" form:"value"`
	Lang  string `firestore:"lang" json:"lang" form:"lang"`

	SinceID int64 `firestore:"since_id" json:"since_id" form:"since_id"`

	HasLink   bool `firestore:"has_link" json:"has_link" form:"has_link"`
	IncludeRT bool `firestore:"include_rt" json:"include_rt" form:"include_rt"`

	PostCountMin int `firestore:"post_count_min" json:"post_count_min" form:"post_count_min"`
	PostCountMax int `firestore:"post_count_max" json:"post_count_max" form:"post_count_max"`

	FaveCountMin int `firestore:"fave_count_min" json:"fave_count_min" form:"fave_count_min"`
	FaveCountMax int `firestore:"fave_count_max" json:"fave_count_max" form:"fave_count_max"`

	FollowingCountMin int `firestore:"following_count_min" json:"following_count_min" form:"following_count_min"`
	FollowingCountMax int `firestore:"following_count_max" json:"following_count_max" form:"following_count_max"`

	FollowerCountMin int `firestore:"follower_count_min" json:"follower_count_min" form:"follower_count_min"`
	FollowerCountMax int `firestore:"follower_count_max" json:"follower_count_max" form:"follower_count_max"`

	FollowerRatioMin float32 `firestore:"follower_ratio_min" json:"follower_ratio_min" form:"follower_ratio_min"`
	FollowerRatioMax float32 `firestore:"follower_ratio_max" json:"follower_ratio_max" form:"follower_ratio_max"`

	UpdatedOn time.Time `firestore:"updated_on" json:"updated_on" form:"updated_on"`
}

// SearchCriteriaByDate is a custom data structure for array of SearchCriteria
type SearchCriteriaByDate []*SearchCriteria

func (s SearchCriteriaByDate) Len() int           { return len(s) }
func (s SearchCriteriaByDate) Less(i, j int) bool { return s[i].UpdatedOn.Before(s[j].UpdatedOn) }
func (s SearchCriteriaByDate) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// SaveSearchCriteria saves search criteria
func SaveSearchCriteria(ctx context.Context, c *SearchCriteria) error {

	if c == nil {
		return errors.New("criteria required")
	}

	if c.ID == "" {
		return errors.New("criteria ID required")
	}

	if c.Name == "" || c.Value == "" {
		return fmt.Errorf("invalid criteria: %+v", c)
	}

	return save(ctx, searchCriteriaCollectionName, c.ID, c)
}

// DeleteSearchCriterion deletes single search criterion
func DeleteSearchCriterion(ctx context.Context, id string) error {
	return deleteByID(ctx, searchCriteriaCollectionName, id)
}

// GetSearchCriterion selects single criterion by id
func GetSearchCriterion(ctx context.Context, id string) (data *SearchCriteria, err error) {
	data = &SearchCriteria{}
	err = getByID(ctx, searchCriteriaCollectionName, id, data)
	return
}

// GetSearchCriteria retreaves all search criteria for specific user
func GetSearchCriteria(ctx context.Context, username string) (data []*SearchCriteria, err error) {

	col, err := getCollection(ctx, searchCriteriaCollectionName)
	if err != nil {
		return nil, err
	}

	docs, err := col.
		Where("user", "==", NormalizeString(username)).
		Documents(ctx).
		GetAll()

	data = make([]*SearchCriteria, 0)

	for _, doc := range docs {
		c := &SearchCriteria{}
		if err := doc.DataTo(c); err != nil {
			return nil, fmt.Errorf("error retreiveing search criteria %v: %v", doc.Data(), err)
		}
		data = append(data, c)
	}

	sort.Sort(SearchCriteriaByDate(data))

	return

}

//============================================================================
// Results
//============================================================================

// SimpleTweet is the short version of twitter search result
type SimpleTweet struct {
	ID            string      `firestore:"id_str" json:"id_str"`
	CriteriaID    string      `firestore:"criteria_id" json:"criteria_id"`
	CreatedAt     time.Time   `firestore:"created_at" json:"created_at"`
	FavoriteCount int         `firestore:"favorite_count" json:"favorite_count"`
	ReplyCount    int         `firestore:"reply_count" json:"reply_count"`
	RetweetCount  int         `firestore:"retweet_count" json:"retweet_count"`
	IsRT          bool        `firestore:"is_rt" json:"is_rt"`
	Text          string      `firestore:"text" json:"text"`
	Author        *SimpleUser `firestore:"author" json:"author"`
	Key           string      `firestore:"key" json:"key"`
}

// ToSearchResultPagingKey builds search results paging key
func ToSearchResultPagingKey(criteriaID string, periodDate time.Time, lastKey string) string {
	if lastKey == "" {
		return fmt.Sprintf("%s-%s", criteriaID, periodDate.Format(ISODateFormat))
	} else {
		return fmt.Sprintf("%s-%s-%s", criteriaID, periodDate.Format(ISODateFormat), lastKey)
	}
}

// SimpleTweetByDate is a custom data structure for array of SimpleTweet
type SimpleTweetByDate []*SimpleTweet

func (s SimpleTweetByDate) Len() int           { return len(s) }
func (s SimpleTweetByDate) Less(i, j int) bool { return s[i].CreatedAt.Before(s[j].CreatedAt) }
func (s SimpleTweetByDate) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// SaveSearchResults saves a list of search results
func SaveSearchResults(ctx context.Context, list []*SimpleTweet) error {

	if len(list) == 0 {
		return nil
	}

	col, err := getCollection(ctx, searchResultCollectionName)
	if err != nil {
		return err
	}

	batch := fsClient.Batch()

	for _, t := range list {
		t.Key = ToSearchResultPagingKey(t.CriteriaID, t.CreatedAt, t.ID)
		docRef := col.Doc(ToID(t.ID))
		batch.Set(docRef, t)
	}

	_, err = batch.Commit(ctx)
	return err

}

// GetSavedSearchResults gets saved search results based on either the date (current date for first time) or the last record key
func GetSavedSearchResults(ctx context.Context, sinceKey string, limit int) (data []*SimpleTweet, err error) {

	if sinceKey == "" {
		return nil, errors.New("sinceKey required")
	}

	col, err := getCollection(ctx, searchResultCollectionName)
	if err != nil {
		return nil, err
	}

	docs := col.Where("key", ">", sinceKey).OrderBy("key", firestore.Asc).Limit(limit).Documents(ctx)

	data = make([]*SimpleTweet, 0)

	for {
		d, e := docs.Next()
		if e == iterator.Done {
			break
		}
		if e != nil {
			return nil, e
		}

		item := &SimpleTweet{}
		if e := d.DataTo(item); e != nil {
			return nil, e
		}
		data = append(data, item)
	}

	return

}

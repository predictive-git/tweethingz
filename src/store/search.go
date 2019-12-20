package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
)

const (
	searchCriteriaCollectionName = "thingz_search_criteria"
	searchResultCollectionName   = "thingz_search_result"
)

// SearchCriteria defines search query criteria
type SearchCriteria struct {
	ID        string        `firestore:"id" json:"id"`
	User      string        `firestore:"user" json:"user"`
	Name      string        `firestore:"name" json:"name"`
	Query     *SimpleQuery  `firestore:"query" json:"query"`
	Filter    *SimpleFilter `firestore:"filter" json:"filter"`
	UpdatedOn time.Time     `firestore:"updated_on" json:"updated_on"`
}

// SearchCriteriaByDate is a custom data structure for array of SearchCriteria
type SearchCriteriaByDate []*SearchCriteria

func (s SearchCriteriaByDate) Len() int           { return len(s) }
func (s SearchCriteriaByDate) Less(i, j int) bool { return s[i].UpdatedOn.Before(s[j].UpdatedOn) }
func (s SearchCriteriaByDate) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// SimpleQuery represents the twitter query
type SimpleQuery struct {
	Value   string `firestore:"value" json:"value"`
	Lang    string `firestore:"lang" json:"lang"`
	SinceID int64  `firestore:"since_id" json:"since_id"`
}

// SimpleFilter represents the result filter
type SimpleFilter struct {
	HasLink   bool          `firestore:"has_link" json:"has_link"`
	Author    *AuthorFilter `firestore:"author" json:"author"`
	IncludeRT bool          `firestore:"include_rt" json:"include_rt"`
}

// AuthorFilter represents the result author filter
type AuthorFilter struct {
	PostCount      *IntRange `firestore:"post_count" json:"post_count"`
	FaveCount      *IntRange `firestore:"fave_count" json:"fave_count"`
	FollowingCount *IntRange `firestore:"following_count" json:"following_count"`
	FollowerCount  *IntRange `firestore:"follower_count" json:"follower_count"`
	// FollowerRatio is Followers/Fallowing (<1 bad, >1 good)
	FollowerRatio *FloatRange `firestore:"follower_ratio" json:"follower_ratio"`
}

// IntRange is a generic int range
type IntRange struct {
	Min int `firestore:"min" json:"min"`
	Max int `firestore:"max" json:"max"`
}

// FloatRange is a generic float range
type FloatRange struct {
	Min float64 `firestore:"min" json:"min"`
	Max float64 `firestore:"max" json:"max"`
}

// SimpleTweet is the short version of twitter search result
type SimpleTweet struct {
	ID            string      `json:"id_str"`
	CreatedAt     time.Time   `json:"created_at"`
	FavoriteCount int         `json:"favorite_count"`
	ReplyCount    int         `json:"reply_count"`
	RetweetCount  int         `json:"retweet_count"`
	IsRT          bool        `json:"is_rt"`
	Text          string      `json:"text"`
	Author        *SimpleUser `json:"author"`
}

// SimpleTweetByDate is a custom data structure for array of SimpleTweet
type SimpleTweetByDate []*SimpleTweet

func (s SimpleTweetByDate) Len() int           { return len(s) }
func (s SimpleTweetByDate) Less(i, j int) bool { return s[i].CreatedAt.Before(s[j].CreatedAt) }
func (s SimpleTweetByDate) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// SaveSearchCriteria saves search criteria
func SaveSearchCriteria(ctx context.Context, c *SearchCriteria) error {

	if c == nil {
		return errors.New("criteria required")
	}

	if c.Name == "" || c.Query == nil || c.Query.Value == "" {
		return fmt.Errorf("invalid criteria: %+v", c)
	}

	if c.ID == "" {
		c.ID = NewID()
	}

	return save(ctx, searchCriteriaCollectionName, c.ID, c)
}

// GetSearchCriterion retrieves singe search criterion
func GetSearchCriterion(ctx context.Context, id string) (c *SearchCriteria, err error) {
	c = &SearchCriteria{}
	err = getByID(ctx, searchCriteriaCollectionName, id, c)
	return
}

// DeleteSearchCriterion deletes single search criterion
func DeleteSearchCriterion(ctx context.Context, id string) error {
	return deleteByID(ctx, searchCriteriaCollectionName, id)
}

// GetSearchCriteria retreaves all search criteria for specific user
func GetSearchCriteria(ctx context.Context, username string) (data []*SearchCriteria, err error) {

	col, err := getCollection(ctx, searchCriteriaCollectionName)
	if err != nil {
		return nil, err
	}

	docs, err := col.
		Where("user", "==", username).
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

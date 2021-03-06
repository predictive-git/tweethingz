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

	Latest    bool `firestore:"latest" json:"latest" form:"latest"`
	HasLink   bool `firestore:"has_link" json:"has_link" form:"has_link"`
	IncludeRT bool `firestore:"include_rt" json:"include_rt" form:"include_rt"`

	PostCountMin int `firestore:"post_count_min" json:"post_count_min" form:"post_count_min"`
	PostCountMax int `firestore:"post_count_max" json:"post_count_max" form:"post_count_max"`

	FaveCountMin int `firestore:"fave_count_min" json:"fave_count_min" form:"fave_count_min"`
	FaveCountMax int `firestore:"fave_count_max" json:"fave_count_max" form:"fave_count_max"`

	FriendCountMin int `firestore:"friend_count_min" json:"friend_count_min" form:"friend_count_min"`
	FriendCountMax int `firestore:"friend_count_max" json:"friend_count_max" form:"friend_count_max"`

	FollowerCountMin int `firestore:"follower_count_min" json:"follower_count_min" form:"follower_count_min"`
	FollowerCountMax int `firestore:"follower_count_max" json:"follower_count_max" form:"follower_count_max"`

	FollowerRatioMin float32 `firestore:"follower_ratio_min" json:"follower_ratio_min" form:"follower_ratio_min"`
	FollowerRatioMax float32 `firestore:"follower_ratio_max" json:"follower_ratio_max" form:"follower_ratio_max"`

	ExecutedOn time.Time `firestore:"updated_on" json:"executed_on" form:"executed_on"`
}

// FormatedExecutedOn returns RFC822 formated  ExecutedOn
func (s *SearchCriteria) FormatedExecutedOn() string {
	if s == nil || s.ExecutedOn.IsZero() {
		return ""
	}

	return s.ExecutedOn.Format(time.RFC822)
}

// SearchCriteriaByName is a custom data structure for array of SearchCriteria
type SearchCriteriaByName []*SearchCriteria

func (s SearchCriteriaByName) Len() int           { return len(s) }
func (s SearchCriteriaByName) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s SearchCriteriaByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

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

	sort.Sort(SearchCriteriaByName(data))

	return

}

//============================================================================
// Results
//============================================================================

// SimpleTweet is the short version of twitter search result
type SimpleTweet struct {
	ID               string      `firestore:"id_str" json:"id_str"`
	CriteriaID       string      `firestore:"criteria_id" json:"criteria_id"`
	ExecutedOn       string      `firestore:"executed_on" json:"executed_on"`
	CreatedAt        time.Time   `firestore:"created_at" json:"created_at"`
	FavoriteCount    int         `firestore:"favorite_count" json:"favorite_count"`
	ReplyCount       int         `firestore:"reply_count" json:"reply_count"`
	RetweetCount     int         `firestore:"retweet_count" json:"retweet_count"`
	IsRT             bool        `firestore:"is_rt" json:"is_rt"`
	Text             string      `firestore:"text" json:"text"`
	Author           *SimpleUser `firestore:"author" json:"author"`
	AuthorIsFriend   bool        `firestore:"author_is_friend" json:"author_is_friend"`
	AuthorIsFollower bool        `firestore:"author_is_follower" json:"author_is_follower"`
}

// FormatedCreatedAt returns RFC822 formated CreatedAt
func (s *SimpleTweet) FormatedCreatedAt() string {
	if s == nil || s.CreatedAt.IsZero() {
		return ""
	}
	return s.CreatedAt.Format(time.RFC822)
}

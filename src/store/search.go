package store

const (
	searchCriteriaCollectionName = "thingz_search_criteria"
	searchResultCollectionName   = "thingz_search_result"
)

// SearchCriteria defines search query criteria
type SearchCriteria struct {
	Query  *SimpleQuery  `firestore:"query" json:"query"`
	Filter *SimpleFilter `firestore:"filter" json:"filter"`
}

// SimpleQuery represents the twitter query
type SimpleQuery struct {
	Value   string `firestore:"value" json:"value"`
	Lang    string `firestore:"lang" json:"lang"`
	SinceID int64  `firestore:"since_id" json:"since_id"`
}

// SimpleFilter represents the result filter
type SimpleFilter struct {
	HasLink bool          `firestore:"has_link" json:"has_link"`
	Author  *AuthorFilter `firestore:"author" json:"author"`
}

// AuthorFilter represents the result author filter
type AuthorFilter struct {
	PostCount *IntRange `firestore:"post_count" json:"post_count"`

	FaveCount *IntRange `firestore:"fave_count" json:"fave_count"`

	FollowingCount *IntRange `firestore:"following_count" json:"following_count"`

	FollowerCount *IntRange `firestore:"follower_count" json:"follower_count"`

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

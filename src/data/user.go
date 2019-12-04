package data

import (
	"time"

	"github.com/pkg/errors"
)

// SaveUsers saves multiple users
func SaveUsers(users []*SimpleUser) error {

	if len(users) == 0 {
		return nil
	}

	if err := initDB(); err != nil {
		return err
	}

	sqlStr := `INSERT INTO users (
			id, username, name, description, profile_image, created_at, lang,
			location, timezone, post_count, fave_count, following_count, follower_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE
			name = ?, description = ?, profile_image = ?, created_at = ?, lang = ?,
			location = ?, timezone = ?, post_count = ?, fave_count = ?,
			following_count = ?, follower_count = ?, updated_on = ?`

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return errors.Wrap(err, "Error preparing bulk save users statement")
	}

	for _, u := range users {
		_, err := stmt.Exec(u.ID, u.Username, u.Name, u.Description, u.ProfileImage,
			u.CreatedAt, u.Lang, u.Location, u.Timezone, u.PostCount, u.FaveCount,
			u.FollowingCount, u.FollowerCount, u.Name, u.Description, u.ProfileImage,
			u.CreatedAt, u.Lang, u.Location, u.Timezone, u.PostCount, u.FaveCount,
			u.FollowingCount, u.FollowerCount, time.Now())
		if err != nil {
			return errors.Wrap(err, "Error executing save followers")
		}
	}

	err = stmt.Close()
	if err != nil {
		return errors.Wrap(err, "Error closing save users statement")
	}

	logger.Printf("Saved %d users", len(users))

	return nil

}

// GetFollowersWithoutDetail retreaves all followe IDs for  user who
// do not have details
func GetFollowersWithoutDetail(username string) (list []int64, err error) {

	if err := initDB(); err != nil {
		return nil, err
	}

	// select all deltas where event has not been captured already
	res, err := db.Query(`SELECT follower_id FROM followers
						  WHERE username = ? and follower_id NOT IN (
						  SELECT id FROM users)`, username)
	if err != nil {
		return nil, errors.Wrap(err, "Error executing query")
	}

	ids := []int64{}
	for res.Next() {
		var id int64
		err := res.Scan(&id)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing results")
		}
		ids = append(ids, id)
	}

	logger.Printf("Found %d followers without detail for %s", len(ids), username)

	return ids, nil

}
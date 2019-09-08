package data

import (
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
			following_count = ?, follower_count = ?`

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return errors.Wrap(err, "Error preparing bulk save users statement")
	}

	for _, u := range users {
		_, err := stmt.Exec(u.ID, u.Username, u.Name, u.Description, u.ProfileImage,
			u.CreatedAt, u.Lang, u.Location, u.Timezone, u.PostCount, u.FaveCount,
			u.FollowingCount, u.FollowerCount, u.Name, u.Description, u.ProfileImage,
			u.CreatedAt, u.Lang, u.Location, u.Timezone, u.PostCount, u.FaveCount,
			u.FollowingCount, u.FollowerCount)
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


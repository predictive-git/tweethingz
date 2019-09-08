USE tweethingz;

CREATE TABLE IF NOT EXISTS followers (
	username VARCHAR(50) NOT NULL,
	on_day DATE NOT NULL,
	follower_id BIGINT NOT NULL,
	PRIMARY KEY (username,on_day,follower_id)
);

CREATE TABLE IF NOT EXISTS follower_events (
	username VARCHAR(50) NOT NULL,
	on_day DATE NOT NULL,
	follower_id BIGINT NOT NULL,
	event_type VARCHAR(50) NOT NULL,
	PRIMARY KEY (username,on_day,follower_id,event_type)
);

CREATE TABLE IF NOT EXISTS users (
	id BIGINT PRIMARY KEY,
	username VARCHAR(50) NOT NULL,
	name VARCHAR(50) NOT NULL,
	description VARCHAR(250) NOT NULL,
	profile_image VARCHAR(250) NOT NULL,
	created_at DATE NOT NULL,
	lang VARCHAR(50) DEFAULT '',
	location VARCHAR(50) DEFAULT '',
	timezone VARCHAR(50) DEFAULT '',
	post_count INT DEFAULT 0,
	fave_count INT DEFAULT 0,
	following_count INT DEFAULT 0,
	follower_count INT DEFAULT 0,
	updated_on timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS authed_users (
	username VARCHAR(50) PRIMARY
	user_id BIGINT NOT NULL,
	access_token_key VARCHAR(100) NOT NULL,
	access_token_secret VARCHAR(100) NOT NULL,
	updated_on timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);
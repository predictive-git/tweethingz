USE twitterd;

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
	follower_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ui_users (
	email VARCHAR(250) PRIMARY KEY,
	twitter_username VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS ui_events (
	email VARCHAR(250) NOT NULL,
	event_at TIMESTAMP NOT NULL,
	event_type VARCHAR(50) NOT NULL,
	description TEXT NOT NULL,
	PRIMARY KEY (email, event_at),
	CONSTRAINT fk_ui-user
		FOREIGN KEY (email)
		REFERENCES ui_users (email)
		ON DELETE CASCADE ON UPDATE CASCADE
);
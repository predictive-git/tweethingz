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
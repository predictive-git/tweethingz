USE twitterd;

CREATE TABLE IF NOT EXISTS followers (
	username VARCHAR(50) NOT NULL,
	on_day DATE NOT NULL,
	follower_id BIGINT NOT NULL,
	PRIMARY KEY (username,on_day,user_id)
);
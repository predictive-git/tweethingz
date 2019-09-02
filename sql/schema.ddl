USE twitterd;

CREATE TABLE IF NOT EXISTS followers (
	on_day DATE NOT NULL,
	user_id BIGINT NOT NULL,
	PRIMARY KEY (on_day,user_id)
);
create table if not exists users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    discord_id varchar(255) UNIQUE,
    discord_username varchar(255) UNIQUE,
    current_rating int
);

create table if not exists match_requests (
    id INT PRIMARY KEY AUTO_INCREMENT,
    requesting_user_id int NOT NULL,
    created_at timestamp,
    updated_at timestamp,
    request_range int NOT NULL COMMENT 'What is the rating range above and below that of the requester that can be matched into.',
    requested_game_mode varchar(255) NOT NULL COMMENT 'Serialized representation of game modes the player requested. At start just bo1, bo3, all.',
    match_request_state varchar(12) COMMENT 'QUEUED | MATCHED | CANCELLED',
    CONSTRAINT MATCH_REQUEST_USER FOREIGN KEY (requesting_user_id) REFERENCES users(id)
);

create table if not exists match_requests_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    match_id int NOT NULL,
    requesting_user_id int NOT NULL,
    created_at timestamp,
    updated_at timestamp,
    request_range int NOT NULL COMMENT 'What is the rating range above and below that of the requester that can be matched into.',
    requested_game_mode varchar(255) NOT NULL COMMENT 'Serialized representation of game modes the player requested. At start just bo1, bo3, all.',
    match_request_state varchar(12) COMMENT 'QUEUED | MATCHED | CANCELLED',
    INDEX (match_id),
    INDEX (requesting_user_id),
    INDEX (created_at),
    INDEX (requested_game_mode),
    INDEX (match_request_state)
);

create table if not exists matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    match_state varchar(12) NOT NULL COMMENT 'MATCHED | CANCELLED | COMPLETED',
    p1_user_id int NOT NULL,
    p2_user_id int NOT NULL,
    p1_match_request_id int NOT NULL,
    p2_match_request_id int NOT NULL,
    winner char(2) COMMENT 'Who won - P1 | P2.',
    CONSTRAINT FK_P1_USER FOREIGN KEY (p1_user_id) REFERENCES users(id),
    CONSTRAINT FK_P2_USER FOREIGN KEY (p2_user_id) REFERENCES users(id),
    INDEX (created_at)
);

create table if not exists matches_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    match_id int NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    match_state varchar(12) NOT NULL,
    p1_user_id int NOT NULL,
    p2_user_id int NOT NULL,
    p1_match_request_id int NOT NULL,
    p2_match_request_id int NOT NULL,
    winner char(2) COMMENT 'Who won - P1 | P2.',
    CONSTRAINT FK_HISTORY_P1_USER FOREIGN KEY (p1_user_id) REFERENCES users(id),
    CONSTRAINT FK_HISTORY_P2_USER FOREIGN KEY (p2_user_id) REFERENCES users(id),
    CONSTRAINT FK_MATCH_ID FOREIGN KEY (match_id) REFERENCES matches(id)
);

create table if not exists user_ratings_history (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id int NOT NULL,
    rating int NOT NULL,
    match_id int NOT NULL,
    previous_match_id int NOT NULL COMMENT 'Identifies the match before this rating update so that we can walk back accidental updates when fixing data entry errors.',
    is_tombstoned bool DEFAULT false COMMENT 'If we walk back a ratings change to fix data entry errors we tombstone the change which results in hiding it from history updates.',
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT FK_MATCH_USER FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX (created_at)
);

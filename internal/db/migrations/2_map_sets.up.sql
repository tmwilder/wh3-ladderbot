create table if not exists map_sets (
    id INT PRIMARY KEY AUTO_INCREMENT,
    created_at timestamp NOT NULL ,
    map_set json NOT NULL,
    game_mode varchar(12) NOT NULL
)
## WH3 LadderBot

This repo contains code to run a Total War Warhammer 3 Matchmaking ladder.

It is intended to be used via Discord as the primary  end-user client which is where most of the community resides.

## Quick Start

To get local development running:

1. Install Docker Compose on your development machine - https://docs.docker.com/compose/install/
2. From the project root - run `docker compose up -d` to start MySQL and the Discord forwarding bot.
3. Wait a moment for the docker MySQL container to initialize.
4. Set DB_USERNAME=root, DB_PASSWORD=password, DB_HOST=127.0.0.1, DB_PORt=3306 in env vars and fire off - `go run ./cmd/migrate.go`.
5. Connect to localhost:3306 as root:password and check that tables were created.

Other useful operations:
`docker compose -down` - Takes down the instances
`docker compose -down --volumes` - Takes down the instances and deletes the backing store for MysQL.

TODO:

1. Take first cut at design doc w/achilles
2. Pick go MySQL driver + fmwk and set up models to connect to db
3. Set up discord app creds for test and add to readme
4. Set up Ngrok in docker compose local dev and add to readme
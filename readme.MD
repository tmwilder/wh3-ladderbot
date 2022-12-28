## WH3 LadderBot

This repo contains code to run a Total War Warhammer 3 Matchmaking ladder.

It is intended to be used via Discord as the primary  end-user client which is where most of the community resides.

## Quick Start

To get local development running:

1. Install Docker Compose on your development machine - https://docs.docker.com/compose/install/
2. From the project root - run `docker compose up -d` to start MySQL and the Discord forwarding bot.
3. Wait a moment for the docker MySQL container to initialize.
4. Set DB_USERNAME=root, DB_PASSWORD=password, DB_HOST=127.0.0.1, DB_PORT=3306 in env vars and fire off - `go run ./cmd/migrate.go`.
5. Connect to localhost:3306 as root:password and check that tables were created.
6. Run `go test internal db` to check connectivity.
7. Follow instructions to set up a basic Discord Bot and get your public key, app id, and bot token.
8. Go read [set up Ngrok](https://github.com/discord/discord-example-app#set-up-interactivity) and set that up.
9. Set these as DISCORD_APP_ID, DISCORD_PUBLIC_KEY, DISCORD_BOT_TOKEN in env vars and launch the api server through the api command.

Other useful operations:
`docker compose -down` - Takes down the instances
`docker compose -down --volumes` - Takes down the instances and deletes the backing store for MysQL.

## Resources
*(Design Doc)[https://docs.google.com/document/d/11ivp-l3DZtG7wLEwbGDa3vjmKztld-1AUIIneHfWqaE/edit?usp=sharing]

## TODO:
1. Take first cut at design doc w/achilles
2. Clean up some of the routing/request handling ergonomics so that our use of gin is less janky
3. Write local logic for matchmaking and test it with the test apps
4. Figure out AWS or GCP hosting story
5. Deploy to prod
6. Find first test channel - probably RTK - do a ladder trial
7. Set up some admin functionality for fixing oops and such
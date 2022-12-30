## WH3 LadderBot

This repo contains code to run a Total War Warhammer 3 Matchmaking ladder.

It is intended to be used via Discord as the primary  end-user client which is where most of the community resides.

## Quick Start

To get local development running:

1. Install Docker Compose on your development machine - https://docs.docker.com/compose/install/
2. From the project root - run `docker compose up -d` to start MySQL and the Discord forwarding bot.
3. Wait a moment for the docker MySQL container to initialize.
4. Set DB_USERNAME=root, DB_PASSWORD=password, DB_HOST=127.0.0.1, DB_PORT=3306 in env vars and fire off - `go run ./cmd/migrate/migrate.go`.
5. Connect to localhost:3306 as root:password and check that tables were created.
6. Run `go test internal db` to check connectivity.
7. Follow instructions to set up a basic Discord Bot and get your public key, app id, and bot token.
8. Set discord's variables as DISCORD_APP_ID, DISCORD_PUBLIC_KEY, DISCORD_BOT_TOKEN in env vars, do the same for DB info from step 4, set some arbitrary key for ADMIN_KEY, and launch the api server through the api command.
9. Go read [set up Ngrok](https://github.com/discord/discord-example-app#set-up-interactivity) and set that up.
10. Install your test app to any discord channel.
11. POST to localhost:8080/commands?ADMIN_KEY=<pull from step 8> to install this app's commands as global commands to your test bot.
12. You should now be able to send slash commands from your test channel using your test app to your local dev env.

Other useful operations:
`docker compose -down` - Takes down the instances
`docker compose -down --volumes` - Takes down the instances and deletes the backing store for MysQL.

## Prod Operations

### Deploying
To deploy to prod - get access to the parent AWS account, build locally, and upload the resulting file. This is manual
right now and if we get a few folks working on this we'll streamline further.

### Migrations
Add a new migation to internal/db/migrations and test it on local. Then deploy to the prod lambda.

Then post to `api.mtgshuffle.com/migrate?admin_key=<key>` with the admin key stored in our aws secrets. This is not mega
secure but is good enough for the hobby level of this project.

### Uploading new bot slash commands.
Add your slash command to commands.go, test on your local app, then deploy to prod lambda and post to 
`api.mtgshuffle.com/migrate?admin_key=<key>` - just as with migrations you can find the key in AWS.

Both this and the migrations resource are not actually particularly sensitive so the key scheme is a convenience of
public call / ease of sec impl tradeoff.

## Resources
*[Design Doc](https://docs.google.com/document/d/11ivp-l3DZtG7wLEwbGDa3vjmKztld-1AUIIneHfWqaE/edit?usp=sharing)

## TODO:
1. Clean up some of the routing/request handling/error handling ergonomics so that our use of gin is less janky
2. Add support for random map selections and storage of map pool
3. Make the messaging and dms back to players make sense for our lifecycle
4. Set up a way to communicate the map pool to a channel
5. Set up a way to pull and communicate the leaderboard back to a channel
6. Figure out AWS or GCP hosting story
7. Configure real bot and deploy to prod
8. Find first test channel - probably RTK - do a ladder trial
9. Set up some admin functionality for fixing oops and such
10. Add event support

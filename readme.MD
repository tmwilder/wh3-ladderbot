## WH3 Community LadderBot

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
7. Follow instructions to set up a basic Discord Bot and get your public key, app id, and bot token. Also get the Guild ID of your test channel.
8. Set discord's variables as DISCORD_APP_ID, DISCORD_PUBLIC_KEY, DISCORD_BOT_TOKEN, DISCORD_HOME_GUILD_ID in env vars, do the same for DB info from step 4, set some arbitrary key for ADMIN_KEY, and launch the api server through the api command.
9. Go read [set up Ngrok](https://github.com/discord/discord-example-app#set-up-interactivity) and set that up.
10. Install your test app to any discord channel.
11. POST to localhost:8080/commands?ADMIN_KEY=<pull from step 8> to install this app's commands as global commands to your test bot.
12. POST to localhost:8080/maps with a payload like ["Arnheim", "Itza", "Black Ark"] to populate maps and rules in your channel.
13. You should now be able to send slash commands from your test channel using your test app to your local dev env.

## Discord Server and Channel Configuration

The bot is designed to operate with one home Discord server with full features and any number of federated guest
servers.

The main difference is that the home server has more channels like rules and maps and the leaderboard, and guest servers
only contain the matchmaking queue and a ladder feed so that they can see what is going on across all servers.

To this end - to configure a Home Server we need:

1. A server
2. The bot is installed on that server with permissions send messages, manage messages, application.commands, manage roles.
3. The server has the following server configuration set up:
   1. Status as a community so that we can do an announcement channel.
   2. #ladder-feed - this can be in a section like WCL and needs to be a channel with this exact name and announcements enabled. Permissions should be locked to just bot posts.
   3. #find-matches - users of the server can post.
   4. #leaderboard
   5. #elo-ratings
   6. #rules-and-maps
   7. A role called laddering exists on the service and has a nice color assigned like green.
4. The bot is scoped via channel perms to only the above channels and roles to minimize attack surface.

To configure a guest server:
1. The bot is installed with application.commands and create messages.
2. #find-matches is created.
3. Locked #ladder-feed following the home server ladder feed announcements, although the channel ca be named anything because the bot does not directly interact with it.
4. Optionally - we can configure the maps, leaderboard, and elo ratings channel as announcement channels and follow them on guests.
5. The bot is scoped via channel perms to only the above channels and roles to minimize attack surface.

Other useful operations:
`docker compose -down` - Takes down the instances
`docker compose -down --volumes` - Takes down the instances and deletes the backing store for MysQL.

## Prod Operations

### Deploying
To deploy to prod - get access to the parent AWS account, build locally, and upload the resulting file. This is manual
right now and if we get a few folks working on this we'll streamline further.

### Migrations
Add a new migration to internal/db/migrations and test it on local. Then deploy to the prod lambda.

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
2. Set up some admin functionality for fixing oops and such
3. Add event support
4. Add help command
5. Add command shortcut aliases like /q /dq /r

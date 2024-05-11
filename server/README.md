# Server

The [Vinyl Player](../) server is a small Go program that accepts HTTP requests from the hardware in the shelf, and stores the information. It does not contain any graphical interface at the moment. Therefore, we decided to integrate it with Telegram instead. This way, you can still update the data as needed.

In the future, it is possible that we will update the server to include a front-end interface, as well as a new storage system. The current storage with plain text files was decided on the basis of how large our collection is, and the fact that this is still an MVP.

## Build

You can easily build the server with the following command:

```shell
go build -o vinyl-scanner
```

## Docker

We provide a Docker image that you can also use instead of building from source:

```shell
docker pull ghcr.io/chrisb2351/vinyl-scanner:latest
```

## Configuration

The server must be configured with some flags. You can use environment variables instead too, and we support `.env` files (see [`.env.example`](./.env.example) for an example). The command help is self-explanatory:

```
NAME:
   vinyl-server - Vinyl scanner server and storage.

USAGE:
   vinyl-server [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value                                           port to run server on (default: 8080) [$VINYL_PORT]
   --telegram-token value                                 telegram bot token [$VINYL_TG_TOKEN]
   --telegram-chat-id value [ --telegram-chat-id value ]  telegram bot chat id or comma-separated ids [$VINYL_TG_CHAT_ID]
   --data-directory value                                 data directory where the logs and the vinyl data is stored [$VINYL_DATA_DIR]
   --auth-token value                                     http server endpoint authentication token [$VINYL_AUTH_TOKEN]
   --help, -h                                             show help
```

## Telegram Commands

These are the available Telegram commands at the time:

```
set_name - Set the album name for a certain ID.
set_artist - Set the artist name for a certain ID.
update_id - Update from an old ID to a new ID.
albums - List all the albums.
clear - Clear the server state of the current ID.
```
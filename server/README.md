# Server

The [Vinyl Player](../) server is a small Go program that accepts HTTP requests from the hardware in the shelf, and stores the information. It notifies you via Telegram when a new tag is read.

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
   password  Generate a password hash to use on the configuration
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value                                           port to run server on (default: 8080) [$VINYL_PORT]
   --telegram-token value                                 telegram bot token [$VINYL_TG_TOKEN]
   --telegram-chat-id value [ --telegram-chat-id value ]  telegram bot chat id or comma-separated ids [$VINYL_TG_CHAT_ID]
   --data-directory value                                 data directory where the logs and the vinyl data is stored [$VINYL_DATA_DIR]
   --api-token value                                      api endpoint authentication token [$VINYL_API_TOKEN]
   --jwt-secret value                                     jwt tokens secret [$VINYL_JWT_SECRET]
   --login-username value                                 admin interface username [$VINYL_LOGIN_USERNAME]
   --login-password value                                 admin interface password hash generated with 'password' subcommand [$VINYL_LOGIN_PASSWORD]
   --help, -h                                             show help
```

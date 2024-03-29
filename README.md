[![Check](https://github.com/etu/flummbot/actions/workflows/check.yml/badge.svg)](https://github.com/etu/flummbot/actions/workflows/check.yml)
[![Update](https://github.com/etu/flummbot/actions/workflows/update.yml/badge.svg)](https://github.com/etu/flummbot/actions/workflows/update.yml)

# flummbot simple irc-bot

Simple IRC-bot I wrote in golang for my irc-channel.

## Current features

- Authenticate to services
- Join channels
- Listen to the command `!tell <person> <message>`
- Deliver `<message>` to `<person>` the next time `<person>` joins or say
  something
- Save quotes with the command `!quote <message>`
- Tell random quote with the command =!quote=

## Configure the bot

Copy the sample config `flummbot.sample.toml` to `flummbot.toml` and look
through the following sections.

Sending `SIGUSR1` to a running process will reload the configuration file,
this can be used to change certain settings without restarting the process.

The types of settings that works are settings that aren't related to:
 - Connections, joining channels, etc
 - Enable/Disable modules

But it does work for changing things like, which command is used for certain
modules, separators used for corrections etc.

### Connection

Change the values of `channel`, `nick`, `server`, `tls` and `message`. You
can also specify `nickservidentify` to set a message to send to NickServ on
connect.

### Database

`file` specifies file used for sqlite.

### Tells

`command` specifies command used in the `!tell` module.

### Quotes

`command` specifies command used in the `!quote` module.

### Karmas

`command` specifies command used in the `!karma` module.

It also listenes words that ends with plusOperator and minusOperator (++ or
-- as default) to change the karmas.

The karma module also allows to hardcode values for certain words, this makes
it not print the karma for that word on change but it will still be able to
report it. This is useful to hardcode values for words such as ~c++~.

### Corrections

`separator` specifies which character that should be used as separator in
`s/foo/bar`-style corrections.

## Building the bot

```sh
make build
```

## Starting the bot
I've included a systemd-service file that I use to start my bot.

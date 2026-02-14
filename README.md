# matrix-stickerbook

Bot and CLI for collecting images, organising, and publishing to Matrix rooms as
[MSC2545](https://github.com/matrix-org/matrix-spec-proposals/pull/2545) sticker pack state
events using simple text commands. Works with any image (not just stickers), the bot downloads
and rehosts images to your homeserver, generates alt-text using Claude Haiku for accessibility,
and deduplicates by content hash.

Built with [mautrix-go](https://github.com/mautrix/go),
[anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go),
[Cobra](https://github.com/spf13/cobra), and [Viper](https://github.com/spf13/viper).

## Command reference

All commands are plain text messages in any Matrix room the bot can see:

| Command                               | Description                                     |
| ------------------------------------- | ----------------------------------------------- |
| `!sticker`                            | Show help                                       |
| `!sticker list unsorted`              | Stickers not in any pack                        |
| `!sticker show <id>`                  | Preview sticker with metadata                   |
| `!sticker name <id> <shortcode>`      | Set emoji shortcode (e.g. happy_cat)            |
| `!sticker usage <id> <type>`          | Set usage (sticker/emoticon/both/reset)         |
| `!sticker delete <id>`                | Remove from collection                          |
| `!sticker pack list`                  | All packs with sticker counts                   |
| `!sticker pack create <name>`         | Create a new pack                               |
| `!sticker pack show <pack>`           | List stickers in a pack                         |
| `!sticker pack add <pack> <id>`       | Add sticker to pack                             |
| `!sticker pack remove <pack> <id>`    | Remove sticker from pack                        |
| `!sticker pack avatar <pack> <mxc>`   | Set pack icon                                   |
| `!sticker pack usage <pack> <type>`   | Set default usage (sticker/emoticon/both/reset) |
| `!sticker pack publish <pack> [room]` | Publish to room (or republish to all)           |

## Getting started

You'll need a Matrix homeserver account and an
[Anthropic API key](https://console.anthropic.com/) for alt-text generation.

Configuration and data is stored in `~/.config/stickerbook/` (or `/data/` in Docker) - it creates
a blank config file on launch if needed, and see [`config.example.yaml`](config.example.yaml) for
configuration options. Your collection then lives in `collection.json` and pack definitions in
`packs.json` - easy to view, edit, or backup.

### Local build

[Install Go](https://go.dev/dl/) then build and run:

```bash
# Build the binary
go build ./cmd/stickerbook

# Generate Matrix login token
./stickerbook login

# Test connectivity
./stickerbook test

# Run the bot
./stickerbook bot
```

### Docker

[Install Docker](https://docs.docker.com/engine/install/) then either see
[`docker-compose.yml`](docker-compose.yml) for Docker Compose, or run directly:

```bash
# Generate Matrix login token
docker run --rm -it -v /path/to/data:/data \
  ghcr.io/liminalpurple/matrix-stickerbook:latest login

# Test connectivity
docker run --rm -v /path/to/data:/data \
  -e ANTHROPIC_API_KEY=sk-ant-... \
  ghcr.io/liminalpurple/matrix-stickerbook:latest test

# Run the bot in daemon mode
docker run -d -v /path/to/data:/data \
  -e ANTHROPIC_API_KEY=sk-ant-... --network host \
  --name stickerbook \
  ghcr.io/liminalpurple/matrix-stickerbook:latest bot
```

## Licence

[Apache 2.0](LICENSE)

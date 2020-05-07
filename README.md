# Sticker bot

On telegram [@TempStickerBot](https://telegram.me/TempStickerBot)

**Setup**

```
git clone github.com/phanirithvij/stickerbot
```

Place your API key from @BotFather in `config copy.yml`

Then rename the file to `config.yml`

```shell
mv "config copy.yml" config.yml
```

Run the bot by `go run .` or `go build && ./stickerbot`

**What it does:**

- Users send bot a sticker or stickerpack url of their choice
- Bot downloads the stickerpack and zips it
- Bots sends the zip file of the pack to the user
- The user can now download the sticker pack
- The stickerpack for now contains lottie json files
- There is no user specifc db storage so the storage is global
- You can also call @BotUserName and search for any **stickerpacks** (one sticker per pack) by text.
- Oh, and the bot downloads any gif file that's sent to it. (And does nothing with that file).

## TODO

- [ ] Better inline search
- [ ] Use redis or something as the db instead of a json file
- [ ] Remove/Move that gif code

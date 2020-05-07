package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	grab "github.com/cavaliercoder/grab"
	"github.com/schollz/progressbar"

	// https://github.com/golang/go/issues/35732#issuecomment-584319096
	// https://stackoverflow.com/a/54275441/8608146
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	config "github.com/phanirithvij/stickerbot/config"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	var configuration config.Configuration

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(configuration.API.Token)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil {
			// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/37#issuecomment-189173928
			if update.InlineQuery.Query == "" {
				continue
			}

			if update.InlineQuery != nil {
				log.Println(update.InlineQuery.Query)
				var results []interface{}

				q := tgbotapi.NewInlineQueryResultCachedSticker("de", "CAACAgIAAxkBAAIBbV6y8AkCexf-OWeBApIENbq07KETAAIlAAM7YCQUglfAqB1EIS0ZBA", "damn")
				// q.Description = "test description"
				// q.ThumbURL = "https://avatars3.githubusercontent.com/u/1369709?s=88&u=a4179f42dc91f7abc46691dcac25a028c6804cdd&v=4"
				r := tgbotapi.NewInlineQueryResultArticle("d2", "test Title", "damn 2 son")
				// r.Description = "test description"
				results = append(results, q)
				results = append(results, r)
				inlineConf := tgbotapi.InlineConfig{
					InlineQueryID: update.InlineQuery.ID,
					IsPersonal:    true,
					CacheTime:     0,
					Results:       results,
				}
				if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
					log.Println(err)
				}
			}

			continue
		}
		log.Println(update.Message.Chat.IsChannel())

		if update.Message.Sticker != nil {
			// it is a sticker
			if update.Message.Sticker.SetName != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "https://t.me/addstickers/"+update.Message.Sticker.SetName)
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}
			log.Println(update.Message.Sticker.FileID)
			log.Println([]rune(update.Message.Sticker.Emoji))
			sticker := tgbotapi.NewStickerShare(update.Message.Chat.ID, update.Message.Sticker.FileID)
			// sticker := tgbotapi.NewStickerUpload(update.Message.Chat.ID, "flutterapp/storage/Upload/21303-pineapple.json")
			if _, err := bot.Send(sticker); err != nil {
				log.Panic(err)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Sticker.Emoji)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			continue
		}

		// if sticker pack url

		url := update.Message.Text
		if (strings.HasPrefix(url, "https://t.me/addstickers/") || strings.HasPrefix(url, "https://telegram.me/addstickers/")) && len(url) > 25 {
			log.Println("Valid sticker pack url", url)
			split := strings.Split(url, "/")
			stickerSet := split[len(split)-1]
			config := tgbotapi.GetStickerSetConfig{Name: stickerSet}
			data, err := bot.GetStickerSet(config)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}
			// log.Println(data.Title)
			// log.Println(data.IsAnimated)
			// log.Println(data.Name)
			// https://stackoverflow.com/a/47180974/8608146

			// TODO to store per user data use this
			// dirname := filepath.Join("storage", strconv.FormatInt(update.Message.Chat.ID, 10))

			// global storage cache
			dirname := "flutterapp/storage"
			err = os.MkdirAll(dirname, os.ModePerm)
			if err != nil {
				log.Panic(err)
			}

			count := int64(len(data.Stickers))

			if _, err := checkExisting(filepath.Join(dirname, stickerSet), len(data.Stickers)); err != nil {
				log.Println(err)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Already Cached.")
				_, err := bot.Send(msg)
				if err != nil {
					log.Panic(err)
				}
				continue
			}

			bar := progressbar.Default(count)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Downloading "+strconv.Itoa(len(data.Stickers))+" stickers")
			sentmsg, err := bot.Send(msg)
			if err != nil {
				log.Panic(err)
			}

			for _, sticker := range data.Stickers {

				bar.Add(1)
				file, err := bot.GetFile(tgbotapi.FileConfig{FileID: sticker.FileID})
				if err != nil {
					log.Panic(err)
				}

				existsPath := filepath.Join(dirname, stickerSet, file.FilePath)
				if _, err := os.Stat(existsPath); err != nil {
					// panic(err)
					url := file.Link(bot.Token)
					// log.Println(url)
					_, err := grab.Get(filepath.Join(dirname, stickerSet, file.FilePath), url)
					if err != nil {
						log.Fatal(err)
					}
					// log.Println(resp.Filename)
				}
			}
			log.Println("Saving to", filepath.Join(dirname, stickerSet, "stickers"))

			if data.IsAnimated {
				// if animated stickerSet extract them
				ExtractStickers(filepath.Join(dirname, stickerSet, "stickers"), filepath.Join(dirname, stickerSet, "extracted"))
			}
			// Delete previous message and send "Done" after done downloading
			conf := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, sentmsg.MessageID)
			if _, err := bot.DeleteMessage(conf); err != nil {
				log.Panic(err)
			}
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Done.")
			// conf = tgbotapi.NewEditMessageText(update.Message.Chat.ID, sentmsg.MessageID, "Done.")
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			continue
		}
	}
}

// checkExisting checks if the count of files is same
func checkExisting(src string, count int) (bool, error) {
	// get existing sticker files
	var files []string

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return nil
		}
		// only files no dirs
		if info.Mode().IsRegular() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return false, err
		// log.Panicln(err)
	}
	fileCount := (len(files))

	// log.Println(fileCount, count)
	// log.Println(files)

	// extracted files also come down here for animated sticker sets
	if fileCount == count || fileCount == count*2 {
		return true, nil
	}
	return false, errors.New("File count is less so need to download some")

}

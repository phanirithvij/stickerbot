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
	"github/phanirithvij/stickerbot/json"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	config "github.com/phanirithvij/stickerbot/config"
	"github.com/spf13/viper"
)

var dbpath string = "db.json"
var jsonData *json.Data

func main() {
	// .config/config.go
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

	// new bot from api key
	bot, err := tgbotapi.NewBotAPI(configuration.API.Token)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// = here not :=
	jsonData, err = json.LoadFromJSON(dbpath)
	if err != nil {
		// no db exists create one
		log.Println("no db exists create db.json")
		json.SaveToJSON(new(json.Data), dbpath)
	}
	log.Println(jsonData)

	// On bot start ensure all cached files are zipped
	cacheDir := "flutterapp/storage/"
	// https://stackoverflow.com/a/60045373/8608146
	file, err := os.Open(cacheDir)
	if err != nil {
		panic(err)
	}
	names, err := file.Readdirnames(0)
	if err != nil {
		panic(err)
	}
	for _, stickerSetName := range names {

		stickerDir := filepath.Join(cacheDir, stickerSetName)
		info, err := os.Stat(stickerDir)
		if err != nil {
			panic(err)
		}

		if info.IsDir() {
			_, err := os.Stat(filepath.Join(stickerDir, stickerSetName+".zip"))
			if err != nil {
				log.Println(filepath.Join(stickerDir, stickerSetName+".zip"), "not found")

				if _, err := os.Stat(filepath.Join(stickerDir, "extracted")); err != nil {
					// non animated so no extracted folder
					ZipFiles(
						filepath.Join(stickerDir, stickerSetName+".zip"),
						filepath.Join(stickerDir, "stickers"))
				} else {
					// animated
					ZipFiles(
						filepath.Join(stickerDir, stickerSetName+".zip"),
						filepath.Join(stickerDir, "extracted"))
				}
			}
		}

	}

	for update := range updates {

		// If it is a callbackQuery i.e. click on a button (option)
		if update.CallbackQuery != nil {
			log.Println(update.CallbackQuery.Data)

			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
			// If it is download:stickerSet the send them the sticker set
			if strings.HasPrefix(update.CallbackQuery.Data, "download:") {
				split := strings.Split(update.CallbackQuery.Data, ":")
				stickerSet := split[1]
				log.Println(stickerSet)
				log.Println(update.CallbackQuery.Message.Chat.ID)
				handleText("https://t.me/addstickers/"+stickerSet, bot, &update, true)
			}

			// TODO send sticker set downloaded or cached

			// bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data))
			continue
		}

		if update.Message == nil {

			// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/37#issuecomment-189173928
			if update.InlineQuery != nil {
				// if update.InlineQuery.Query == "" {
				// 	continue
				// }
				log.Println(update.InlineQuery.Query)
				var results []interface{}

				for i, item := range jsonData.Data {
					if strings.Contains(strings.ToLower(item.StickerName), strings.ToLower(update.InlineQuery.Query)) {
						// TODO I copy pasted this method from a PR which was not merged into master of go-telegram-bot-api
						// https://github.com/go-telegram-bot-api/telegram-bot-api/pull/292
						res := tgbotapi.NewInlineQueryResultCachedSticker("sticker"+strconv.Itoa(i), item.SampleStickerID, "damn")
						results = append(results, res)
					}
				}
				// log.Println(results...)

				count := len(results)
				// send 26 sticker pack results at max no matter the search
				if count > 26 {
					results = results[:26]
				}

				// q.Description = "test description"
				// q.ThumbURL = "https://avatars3.githubusercontent.com/u/1369709?s=88&u=a4179f42dc91f7abc46691dcac25a028c6804cdd&v=4"
				// r := tgbotapi.NewInlineQueryResultArticle("d2", "test Title", "damn 2 son")
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
			log.Println(update.Message.Sticker.FileID)
			log.Println([]rune(update.Message.Sticker.Emoji))
			sticker := tgbotapi.NewStickerShare(update.Message.Chat.ID, update.Message.Sticker.FileID)
			if update.Message.Sticker.SetName != "" {
				sticker.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL("Open sticker set", "https://t.me/addstickers/"+update.Message.Sticker.SetName),
						tgbotapi.NewInlineKeyboardButtonData("Download set", "download:"+update.Message.Sticker.SetName),
					),
				)
			}

			if _, err := bot.Send(sticker); err != nil {
				log.Panic(err)
			}

			continue
		}

		if update.Message.Document != nil {
			log.Println("Downloaded", update.Message.Document.FileName)
			// download any file
			file, err := bot.GetFile(tgbotapi.FileConfig{FileID: update.Message.Document.FileID})
			if err != nil {
				log.Panic(err)
			}
			// TODO to store per user data use this
			dirname := filepath.Join("storage", strconv.FormatInt(update.Message.Chat.ID, 10))

			existsPath := filepath.Join(dirname, file.FilePath)
			if _, err := os.Stat(existsPath); err != nil {
				// panic(err)
				url := file.Link(bot.Token)
				// log.Println(url)
				_, err := grab.Get(filepath.Join(dirname, file.FilePath), url)
				if err != nil {
					log.Fatal(err)
				}
				// log.Println(resp.Filename)
			}
			continue
		}

		handleText("", bot, &update, false)
	}
}

func handleText(text string, bot *tgbotapi.BotAPI, update *tgbotapi.Update, callbackQuery bool) {
	url := text
	if text == "" {
		url = update.Message.Text
	}
	var chatID int64
	if callbackQuery {
		chatID = update.CallbackQuery.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}
	// if sticker pack url
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
		// https://stackoverflow.com/a/47180974/8608146

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
			var file tgbotapi.DocumentConfig
			foundCached := false
			for _, item := range jsonData.Data {
				if item.StickerName == stickerSet {
					foundCached = true
					file = tgbotapi.NewDocumentShare(chatID, item.FileID)
					break
				}
			}
			if !foundCached {
				file = tgbotapi.NewDocumentUpload(chatID, filepath.Join(dirname, stickerSet, stickerSet+".zip"))
			}
			file.Caption = "Done"
			if uploaded, err := bot.Send(file); err != nil {
				// TODO handle if the cached fileid expires
				log.Panic(err)
			} else {
				if !foundCached {
					log.Println("Done uploading", uploaded.Document.FileID)
					log.Print(jsonData.Data)
					log.Print(json.Pair{FileID: uploaded.Document.FileID, StickerName: stickerSet})
					// save fileid, stickerSet to db
					entry := json.Pair{
						FileID:          uploaded.Document.FileID,
						StickerName:     stickerSet,
						SampleStickerID: data.Stickers[0].FileID,
					}
					jsonData.Data = append(jsonData.Data, entry)
					json.SaveToJSON(jsonData, dbpath)
				} else {
					log.Println("Shared existing", uploaded.Document.FileID)
				}
			}
			return
		}

		bar := progressbar.Default(count)

		msg := tgbotapi.NewMessage(chatID, "Downloading "+strconv.Itoa(len(data.Stickers))+" stickers")
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
			ZipFiles(
				filepath.Join(dirname, stickerSet, stickerSet+".zip"),
				filepath.Join(dirname, stickerSet, "extracted"))
		} else {
			ZipFiles(
				filepath.Join(dirname, stickerSet, stickerSet+".zip"),
				filepath.Join(dirname, stickerSet, "stickers"))
		}
		// Delete previous message and send "Done" after done downloading
		conf := tgbotapi.NewDeleteMessage(chatID, sentmsg.MessageID)
		if _, err := bot.DeleteMessage(conf); err != nil {
			log.Panic(err)
		}
		msg = tgbotapi.NewMessage(chatID, "Zipping files")
		// conf = tgbotapi.NewEditMessageText(update.Message.Chat.ID, sentmsg.MessageID, "Done.")
		if sentmsg, err = bot.Send(msg); err != nil {
			log.Panic(err)
		}

		// here we need to upload anyway
		// upload zip file
		file := tgbotapi.NewDocumentUpload(chatID, filepath.Join(dirname, stickerSet, stickerSet+".zip"))
		file.Caption = "Done"
		if uploaded, err := bot.Send(file); err != nil {
			log.Panic(err)
		} else {
			log.Println("Done uploading", uploaded.Document)
			// delete the zipping files message
			delconf := tgbotapi.NewDeleteMessage(chatID, sentmsg.MessageID)
			if _, err := bot.DeleteMessage(delconf); err != nil {
				log.Panic(err)
			}
			entry := json.Pair{
				FileID:          uploaded.Document.FileID,
				StickerName:     stickerSet,
				SampleStickerID: data.Stickers[0].FileID,
			}

			jsonData.Data = append(jsonData.Data, entry)
			json.SaveToJSON(jsonData, dbpath)
		}
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Invalid sticker set "+url)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
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
	// and also zipped file that was send
	if fileCount >= count*2 || (fileCount-count < 3 && fileCount >= count) {
		return true, nil
	}
	return false, errors.New("File count is less so need to download some")

}

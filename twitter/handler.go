package twitter

import (
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"regexp"
)

var (
	rgxTwitter = regexp.MustCompile(`http(|s):\/\/twitter\.com\/i\/status\/[0-9]*`)
)

func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI) {
	link := utils.TrimURL(rgxTwitter.FindString(update.Message.Text))
	db.DRIVER.LogInformation("Started processing twitter request " + link + " by " + utils.GetTelegramUserString(update.Message.From))

	data := NewTwitterVideoDownloader(link)
	file, err := data.Download(utils.DownloadBytesLimit)
	if err != nil {
		if err.Error() == "too large" {
			db.DRIVER.LogInformation("A requested video exceeded it's upload limit for " + utils.GetTelegramUserString(update.Message.From))
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your requested twitter video is too large for me to handle! I can only upload videos up to 50MB")
			msg.ReplyToMessageID = update.Message.MessageID
			api.Send(msg)
			return
		}
		db.DRIVER.LogError("Couldn't handle a twitter request", utils.GetTelegramUserString(update.Message.From), err.Error())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
		msg.ReplyToMessageID = update.Message.MessageID
		api.Send(msg)
		return
	}
	media := tgbotapi.FilePath(file.Name())
	video := tgbotapi.NewVideo(update.Message.Chat.ID, media)
	video.ReplyToMessageID = update.Message.MessageID

	_, err = api.Send(video)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		db.DRIVER.LogError("Couldn't handle a twitter request", utils.GetTelegramUserString(update.Message.From), err.Error())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, something went wrong while processing your request. Please try again later")
		msg.ReplyToMessageID = update.Message.MessageID
		api.Send(msg)
		return
	}

	file.Close()
	os.Remove(file.Name())

	db.DRIVER.LogInformation("Finished processing twitter request by " + utils.GetTelegramUserString(update.Message.From))
}

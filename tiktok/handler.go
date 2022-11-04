package tiktok

import (
	"fmt"
	"github.com/TheBunnies/TiktokUploaderTelegram/db"
	"github.com/TheBunnies/TiktokUploaderTelegram/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"regexp"
)

var (
	rgxTiktok = regexp.MustCompile(`http(s|):\/\/.*(tiktok)\.com.*`)
)

func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI) error {
	link := utils.TrimURL(rgxTiktok.FindString(update.Message.Text))
	link = utils.SanitizeTiktokUrl(link)

	db.DRIVER.LogInformation("Started processing tiktok request " + link + " by " + utils.GetTelegramUserString(update.Message.From))

	id, err := GetId(link)
	if err != nil {
		return err
	}
	parsedId, err := Parse(id)
	if err != nil {
		return err
	}
	data, err := NewAwemeItem(parsedId)
	if err != nil {
		return err
	}
	file, err := data.DownloadVideo(utils.DownloadBytesLimit)
	if err != nil {
		if err.Error() == "too large" {
			db.DRIVER.LogInformation("A requested video exceeded it's upload limit for " + utils.GetTelegramUserString(update.Message.From))
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your requested tiktok video is too large for me to handle! I can only upload videos up to 50MB")
			msg.ReplyToMessageID = update.Message.MessageID
			api.Send(msg)
			return nil
		}
		return err
	}
	message := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Author: %s \nDuration: %s\nCreation time: %s \nDescription: %s \n",
		data.Author.Nickname,
		data.Duration(),
		data.Time(),
		data.Description()))
	message.ReplyToMessageID = update.Message.MessageID

	api.Send(message)

	media := tgbotapi.FilePath(file.Name())
	video := tgbotapi.NewVideo(update.Message.Chat.ID, media)

	_, err = api.Send(video)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return err
	}

	file.Close()
	os.Remove(file.Name())
	db.DRIVER.LogInformation("Finished processing tiktok request by " + utils.GetTelegramUserString(update.Message.From))
	return nil
}

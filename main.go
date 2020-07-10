package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/HirbodBehnam/EasyX25519"
	"github.com/allegro/bigcache"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var client = http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{MaxVersion: tls.VersionTLS11}}}

const RegisterURL = "https://api.cloudflareclient.com/v0a884/reg"
const VERSION = "1.0.0"
const WarpPlusEnabled = false

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please pass the bot token as argument.")
	}
	// setup cache
	cache, err := bigcache.NewBigCache(bigcache.Config{
		CleanWindow: 		time.Hour,
		LifeWindow: 		time.Hour * 24,
		MaxEntrySize:		1,
		Shards:             1024,
		MaxEntriesInWindow: 1000 * 10 * 60,
		StatsEnabled:       false,
		Verbose:            false,
		Logger:             bigcache.DefaultLogger(),
	})
	if err != nil {
		log.Fatal("Cannot initialize the cache database:", err.Error())
	}
	// load bot
	bot, err := tgbotapi.NewBotAPI(os.Args[1])
	if err != nil {
		log.Fatal("Cannot initialize the bot:", err.Error())
	}
	log.Println("Warp to Wireguard Bot v" + VERSION)
	log.Println("Bot authorized on account", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Cannot get updates channel:", err.Error())
	}
	for update := range updates {
		if update.Message == nil { // ignore any non-Message
			continue
		}
		// check if the message is command
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello and welcome\\!\nThis bot helps you simply generate a wireguard config that connects to [warp](https://1.1.1.1)'s servers\\. To generate a config, use /generate")
				msg.ParseMode = "MarkdownV2"
				_, _ = bot.Send(msg)
			case "about":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Warp+Wireguard v"+VERSION+"\nBy Hirbod Behnam\nSource: https://github.com/HirbodBehnam/warp-bot"))
			case "help":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "To generate a new config, use /generate After you used it, the bot will send you a .conf file. Import it into wireguard and use it.\nTo get the wireguard clients links use /wireguard\n\nTo get 1GB of warp plus (once per 24 ± 1 hour) use /warp_plus"))
			case "warp_plus":
				if WarpPlusEnabled {
					if b, err := cache.Get(strconv.FormatInt(int64(update.Message.From.ID), 10)); err == nil { // a value exists here
						if b[0] == 1 { // 1 means that the user have already used the warp+ today. 0 means that the bot is waiting for key
							_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You have already claimed your 1GB warp+\nYou can use /more_warp_plus to learn how to get warp plus on your computer"))
							continue
						}
					} else {
						err = cache.Set(strconv.FormatInt(int64(update.Message.From.ID), 10), []byte{0})
						if err != nil {
							_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Cannot save your data to our database. Try again later."))
							continue
						}
					}
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please send me your warp key.\n\nOn mobile go to hamburger menu icon (☰) -> Advanced -> Diagnostics -> ID and send it to bot.\nIf you have created a profile by this bot, open the .conf file as text. At the bottom there is a device_id = xxx line. Send the id from there to bot."))
				}else{
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "The administrator of this bot have disabled warp+ quota adder. You can still use /more_warp_plus to get more warp+ on your own pc"))
				}
			case "more_warp_plus":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Simply use this program on your computer: https://github.com/ALIILAPRO/warp-plus-cloudflare"))
			case "wireguard":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Wiregaurd website: https://www.wireguard.com/\nAndroid Client: https://play.google.com/store/apps/details?id=com.wireguard.android\niOS Client: https://itunes.apple.com/us/app/wireguard/id1441195209?ls=1&mt=8\nWindows Client: https://download.wireguard.com/windows-client/\nmacOS Client: https://itunes.apple.com/us/app/wireguard/id1451685025?ls=1&mt=12"))
			case "generate":
				go func(id int64,user *tgbotapi.User) {
					c,err:= GenerateConfig()
					if err != nil{
						_, _ = bot.Send(tgbotapi.NewMessage(id, "There was an error creating this config: " + err.Error()))
						return
					}
					name := ""
					if user.UserName != ""{
						name += user.UserName
					}else{
						name += strconv.FormatInt(int64(user.ID),10)
					}
					name += "-"
					name += strconv.FormatInt(time.Now().Unix(),10)
					name += ".conf"
					file := tgbotapi.FileBytes{
						Name: name,
						Bytes: c,
					}
					_,err = bot.Send(tgbotapi.NewDocumentUpload(id,file))
					if err != nil {
						_, _ = bot.Send(tgbotapi.NewMessage(id, "There was an error sending the file: "+err.Error()))
					}
				}(update.Message.Chat.ID,update.Message.From)
			default:
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry this command is not recognized; Try /help"))
			}
			continue
		}
		if WarpPlusEnabled {
			if b, err := cache.Get(strconv.FormatInt(int64(update.Message.From.ID), 10)); err == nil {
				if b[0] == 0 {
					go func(id int64, warpId string) { // add warp plus
						key, err := x25519.NewX25519()
						if err != nil {
							_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot create a key pair. Try again later."))
							return
						}
						_, err = Register(base64.StdEncoding.EncodeToString(key.PublicKey), warpId)
						if err != nil {
							_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot add warp+: "+err.Error()))
							return
						}
						_ = cache.Set(strconv.FormatInt(id, 10), []byte{1})
						_, _ = bot.Send(tgbotapi.NewMessage(id, "Done!"))
					}(update.Message.Chat.ID, update.Message.Text)
					continue
				}
			}
		}
		_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please use a command; Try /help"))
	}
}

// generate a config file
func GenerateConfig() (configFile []byte,err error){
	// dont crash the whole thing
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("internal error")
		}
	}()
	key,err := x25519.NewX25519()
	if err != nil{
		return
	}
	jsonBytes , err := Register(base64.StdEncoding.EncodeToString(key.PublicKey),"")
	if err != nil{
		return
	}
	var accountResponse map[string]interface{}
	err = json.Unmarshal(jsonBytes,&accountResponse)
	if err != nil{
		return
	}
	var data ProfileData
	data.PrivateKey = base64.StdEncoding.EncodeToString(key.SecretKey)
	if _,ok:= accountResponse["config"]; !ok{
		return
	}
	{
		peer := accountResponse["config"].(map[string]interface{})["peers"].([]interface{})[0].(map[string]interface{})
		data.PublicKey = peer["public_key"].(string)
		data.Endpoint = peer["endpoint"].(map[string]interface{})["host"].(string)
	}
	{
		addresses := accountResponse["config"].(map[string]interface{})["interface"].(map[string]interface{})["addresses"].(map[string]interface{})
		data.Address1 = addresses["v4"].(string)
		data.Address2 = addresses["v6"].(string)
	}
	data.Response = string(jsonBytes)
	data.DeviceID = accountResponse["id"].(string)

	profile,err := GenerateProfile(&data)
	configFile = []byte(profile)
	return
}

// registers a new cloudflare account
func Register(key,referrer string) ([]byte ,error){
	var jsonStr = []byte(`{"install_id":"","tos":"` + time.Now().Format(time.RFC3339Nano) + `", "key":"` + key + `","referrer": "` + referrer + `","fcm_token":"","warp_enabled": true,"type":"Android","locale":"en_US"}`) // fuck json parser
	req, err := http.NewRequest("POST", RegisterURL, bytes.NewBuffer(jsonStr))
	if err != nil{
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User-Agent", "okhttp/3.12.1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200{
		return nil,errors.New("cannot register, status code: " + strconv.FormatInt(int64(resp.StatusCode),10))
	}
	return ioutil.ReadAll(resp.Body)
}
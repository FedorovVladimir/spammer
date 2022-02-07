package src

import (
	"context"
	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func init() {
	// Load values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func Run() {
	phoneNumber, ok := os.LookupEnv("USER_PHONE_NUMBER")
	if !ok {
		log.Fatalln("val USER_PHONE_NUMBER does not exist")
	}
	chatsString, ok := os.LookupEnv("CHATS")
	if !ok {
		log.Fatalln("val CHATS does not exist")
	}

	names := strings.Split(chatsString, ",")
	var chatIDs []int64

	rand.Seed(time.Now().UTC().UnixNano())

	log.Println("app run")

	examples.Run(func(ctx context.Context, _ *zap.Logger) error {
		flow := auth.NewFlow(&termAuth{phone: phoneNumber}, auth.SendCodeOptions{})

		client, err := telegram.ClientFromEnvironment(telegram.Options{})
		if err != nil {
			return err
		}

		return client.Run(ctx, func(ctx context.Context) error {
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return err
			}

			api := client.API()

			// to chats
			discussion, _ := api.ChannelsGetGroupsForDiscussion(ctx)
			for _, chat := range discussion.GetChats() {
				c, ok := chat.(*tg.Chat)
				if !ok {
					continue
				}
				if contains(c.Title, names) {
					chatIDs = append(chatIDs, c.ID)
				}
			}

			state, err := api.UpdatesGetState(ctx)
			if err != nil {
				return err
			}
			pts := state.Pts
			date := time.Now().UTC().Unix()
			for {
				difference, err := api.UpdatesGetDifference(ctx, &tg.UpdatesGetDifferenceRequest{
					Pts:           pts,
					PtsTotalLimit: 10,
					Date:          int(date),
				})
				if err != nil {
					return err
				}
				date = time.Now().UTC().Unix()

				d, ok := difference.(*tg.UpdatesDifference)
				if !ok {
					continue
				}
				for _, chat := range d.Chats {
					c, ok := chat.(*tg.Channel)
					if !ok || strings.ToLower(c.Title) != "рассылка" {
						continue
					}
					for _, update := range d.OtherUpdates {
						u, ok := update.(*tg.UpdateNewChannelMessage)
						if !ok {
							continue
						}
						m, ok := u.Message.(*tg.Message)
						if !ok {
							continue
						}
						sendMessages(ctx, api, m, chatIDs)
					}
				}
				pts = d.State.Pts

				time.Sleep(time.Second)
			}
		})
	})
}

var mID = 0

func sendMessages(ctx context.Context, api *tg.Client, message *tg.Message, chatIDs []int64) {
	if message.ID != mID {
		for _, chatID := range chatIDs {
			_, _ = api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
				Peer:     &tg.InputPeerChat{ChatID: chatID},
				Message:  message.Message,
				RandomID: rand.Int63(),
			})
		}
		mID = message.ID
	}
}

func contains(name string, names []string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

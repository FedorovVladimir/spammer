package src

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"
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
		flow := auth.NewFlow(termAuth{phone: phoneNumber}, auth.SendCodeOptions{})

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
			for _, chatClass := range discussion.GetChats() {
				switch chat := chatClass.(type) {
				case *tg.Chat:
					if contains(chat.Title, names) {
						chatIDs = append(chatIDs, chat.ID)
					}
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
					Flags:         0,
					Pts:           pts,
					PtsTotalLimit: 10,
					Date:          int(date),
					Qts:           0,
				})
				if err != nil {
					return err
				}
				date = time.Now().UTC().Unix()
				switch d := difference.(type) {
				case *tg.UpdatesDifference:
					for _, channel := range d.Chats {
						switch c := channel.(type) {
						case *tg.Channel:
							if strings.ToLower(c.Title) != "рассылка" {
								continue
							}
							for _, update := range d.OtherUpdates {
								switch u := update.(type) {
								case *tg.UpdateNewChannelMessage:
									if m, ok := u.Message.(*tg.Message); ok {
										sendMessage(ctx, api, m, chatIDs)
									}
								}
							}
						}
					}
					pts = d.State.Pts
				}
				time.Sleep(time.Second)
			}
		})
	})
}

var mID = 0

func sendMessage(ctx context.Context, api *tg.Client, message *tg.Message, chatIDs []int64) {
	if message.ID != mID {
		for _, chatID := range chatIDs {
			//_, _ = api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
			//	Peer:     &tg.InputPeerChat{ChatID: chatID},
			//	Message:  message.Message,
			//	RandomID: rand.Int63(),
			//})
			fmt.Println(chatID, message.Message)
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

// noSignUp can be embedded to prevent signing up.
type noSignUp struct{}

func (c noSignUp) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("not implemented")
}

func (c noSignUp) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

// termAuth implements authentication via terminal.
type termAuth struct {
	noSignUp

	phone string
}

func (a termAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (a termAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	bytePwd, err := terminal.ReadPassword(0)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePwd)), nil
}

func (a termAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

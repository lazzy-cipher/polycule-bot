package main

// TODO: make Polly follow up after saying "I HAVE A SECRET TO TELL YOU" including "chicken butt"

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	AdminID                          = "254331907351904256"  // lazarus_overlook
	WelcomeChannelID                 = "1532401026924150947" // chitchat
	WelcomeStickerID                 = "1532410779469615288" // Lazzy Showoff
	WelcomeRoleID                    = "1532412593816338482" // goober
	BotChannelID                     = "1532793476746449098" // chat-with-polly
	DebugChannelID                   = "1532858925681086545" // poly-debug
	ReplyTime                        = 2 * time.Second
	ReplyToTypingTime                = 500 * time.Millisecond
	ReactTime                        = 1 * time.Second
	FollowUpTime                     = 30 * time.Minute
	ReplyToTypingChance              = 0.005 // 0.5%
	ReplyToReplyChance               = 1.0   // 1%
	ReplyWithGifChance               = 0.05  // 5%, otherwise text
	ReplyToMessageChance             = 0.02  // 2%
	ReplyToMessageInBotChannelChance = 0.1   // 10%
	ReactToMessageChance             = 0.04  // 4%
	Version                          = "1.0.3"
	TimeBeforeBored                  = 2 * time.Hour
	ShutUpTime                       = 2 * time.Hour
)

var (
	Token           string
	ShowVersion     bool
	DebugBuild      bool
	ResetBoredTimer = make(chan struct{})
	ShutUpSignal    = make(chan *discordgo.Message)
	ComeBackSignal  = make(chan *discordgo.Message)
	IsShutUp        atomic.Bool

	EmoticonsPool = NewMessagePool(Emoticons[:])
	CommentsPool  = NewMessagePool(Comments[:])
	RepliesPool   = NewMessagePool(Replies[:])
	GifsPool      = NewMessagePool(Gifs[:])
	BoredPool     = NewMessagePool(Bored[:])
	ShutUpPool    = NewMessagePool(ShutUp[:])
	ComeBackPool  = NewMessagePool(ComeBack[:])
	AvailablePool = NewMessagePool(Available[:])
	TypingPool    = NewMessagePool(Typing[:])

	ReplyChannelBlacklist = []string{
		"1532407644919431218", // memories
		"1532859568453849118", // tenants-only
	}
	BlacklistedUsers = [...]string{
		"1466282667258675324", // bardownbuddy
	}
)

func init() {
	flag.StringVar(&Token, "token", "", "Bot Token")
	flag.BoolVar(&ShowVersion, "version", false, "Print version and exit")
	flag.BoolVar(&DebugBuild, "debug", false, "Enable debug build")
	flag.Parse()
}

func main() {
	assertInitState()

	if ShowVersion {
		fmt.Printf(`Polycule Bot v%s
Copyright (C) 2026 lazzy.cipher@proton.me.
License 0BSD: Zero-Clause BSD <https://opensource.org/license/0bsd>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Written by Lazzy L. Cipher.
`, Version)
		os.Exit(0)
	}

	if Token == "" {
		Token = os.Getenv("DISCORD_BOT_TOKEN")
	}

	if Token == "" {
		log.Fatal("Missing token or channel ID. Set -token flag or DISCORD_BOT_TOKEN env var.")
	}

	if DebugBuild {
		log.Println("[DEBUG] running in debug mode, redirecting output to:", DebugChannelID)
	} else {
		ReplyChannelBlacklist = append(ReplyChannelBlacklist, DebugChannelID)
	}

	var dg, err = discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("[ERROR] cannot create Discord session:", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuildMessageTyping |
		discordgo.IntentGuildMessages |
		discordgo.IntentDirectMessages |
		discordgo.IntentMessageContent

	dg.AddHandler(guildMemberStartsTyping)
	dg.AddHandler(guildMemberAdd)
	dg.AddHandler(replied)

	err = dg.Open()
	if err != nil {
		log.Fatal("[ERROR] cannot open connection:", err)
	}
	defer dg.Close()

	go boredTimerLoop(dg)
	go shutUpDetector(dg)

	fmt.Println("Bot is running. Press CTRL-C to exit.")
	var sc = make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func assertInitState() {
	var crash = false

	crash = crash || assertComments()

	if crash {
		log.Fatal("[ERROR] invalid initial program state")
	}
}

func HasSticker(s *discordgo.Session, guildID string, stickerID string) bool {
	var guild, err = s.State.Guild(guildID)
	if err != nil {
		log.Println("[DEBUG] unable to get guild out of cache, fetching:", err)
		guild, err = s.Guild(guildID)
		if err != nil {
			log.Printf("[ERROR] unable to find guild %v: %v\n", guildID, err)
			return false
		}
	}

	for _, sticker := range guild.Stickers {
		if sticker.ID == stickerID {
			return true
		}
	}
	return false
}

func welcomeMessage(s *discordgo.Session, m *discordgo.Member) {
	if !HasSticker(s, m.GuildID, WelcomeStickerID) {
		log.Printf("[ERROR] Invalid welcome sticker ID %v\n", WelcomeStickerID)
		return
	}

	var message = fmt.Sprintf("Welcome to the compound, <@%s>\n%s", m.User.ID, AvailablePool.Next())

	var channelID = WelcomeChannelID
	if DebugBuild {
		channelID = DebugChannelID
		message = "[DEBUG] " + message
	}

	var _, err = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    message,
		StickerIDs: []string{WelcomeStickerID},
	})
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
	}
}

func notifyAdmin(s *discordgo.Session, message string) {
	var channelID = DebugChannelID
	if !DebugBuild {
		var channel, err = s.UserChannelCreate(AdminID)
		if err != nil {
			log.Println("[ERROR] cannot create DM channel with admin", err)
			return
		}

		channelID = channel.ID
	} else {
		message = "[DEBUG] " + message
	}

	var _, err = s.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Println("[ERROR] cannot send DM to admin:", err)
	}
}

func guildMemberAdd(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	if !DebugBuild {
		var err = s.GuildMemberRoleAdd(e.GuildID, e.User.ID, WelcomeRoleID)
		if err != nil {
			log.Println("[ERROR] cannot add role:", err)
		}
	}

	welcomeMessage(s, e.Member)
	notifyAdmin(s, fmt.Sprintf("Hiiii Laz!! Btw, new member: <@%s> :3", e.User.ID))
}

func guildMemberStartsTyping(s *discordgo.Session, e *discordgo.TypingStart) {
	if IsShutUp.Load() ||
		slices.Contains(ReplyChannelBlacklist, e.ChannelID) {
		return
	}

	if rand.Float64() > ReplyToTypingChance ||
		slices.Contains(BlacklistedUsers[:], e.UserID) {
		return
	}

	time.Sleep(ReplyToTypingTime)

	var message = fmt.Sprintf(TypingPool.Next(), e.UserID)
	var finalMessage = fmt.Sprintf("%s %s", message, EmoticonsPool.Next())

	var channelID = e.ChannelID
	if DebugBuild {
		channelID = DebugChannelID
		finalMessage = "[DEBUG] " + finalMessage
	}

	var _, err = s.ChannelMessageSend(channelID, finalMessage)
	if err != nil {
		log.Println("[ERROR] cannot send welcome message:", err)
	}
}

func GetRandomMessage() string {
	return RepliesPool.Next() + " " + EmoticonsPool.Next()
}

func reply(s *discordgo.Session, m *discordgo.Message, message string) error {
	time.Sleep(ReplyTime)

	if !DebugBuild {
		var _, err = s.ChannelMessageSendComplex(m.ChannelID,
			&discordgo.MessageSend{
				Content:   message,
				Reference: m.Reference(),
			})

		return err
	}

	var _, err = s.ChannelMessageSend(DebugChannelID, "[DEBUG] "+message)

	return err
}

func GetRandomComment(userID string) string {
	return fmt.Sprintf(CommentsPool.Next(), userID)
}

func replyRandom(s *discordgo.Session, m *discordgo.Message) error {
	if len(m.Mentions) > 0 {
		var refAuthorID string
		if m.ReferencedMessage != nil {
			refAuthorID = m.ReferencedMessage.Author.ID
		}

		var filtered []*discordgo.User
		for _, u := range m.Mentions {
			var isReply = u.ID == refAuthorID &&
				!strings.Contains(m.Content, "<@"+u.ID+">")

			if u.ID != s.State.User.ID &&
				!slices.Contains(BlacklistedUsers[:], u.ID) &&
				!isReply {
				filtered = append(filtered, u)
			}
		}

		if len(filtered) > 0 {
			var userID = filtered[rand.IntN(len(filtered))].ID
			return reply(s, m, GetRandomComment(userID))
		}
	}

	var msg string
	if rand.Float64() <= ReplyWithGifChance {
		msg = GifsPool.Next()
	} else {
		msg = GetRandomMessage()
	}

	return reply(s, m, msg)
}

func mentionsBot(s *discordgo.Session, m *discordgo.Message) bool {
	for _, u := range m.Mentions {
		if u.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

func ReactPPAP(s *discordgo.Session, m *discordgo.Message) {
	if DebugBuild {
		return
	}

	time.Sleep(ReactTime)
	var err = s.MessageReactionAdd(m.ChannelID, m.ID, "🖊️")
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
	}

	time.Sleep(ReactTime)
	err = s.MessageReactionAdd(m.ChannelID, m.ID, "🍍")
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
	}

	time.Sleep(ReactTime)
	err = s.MessageReactionAdd(m.ChannelID, m.ID, "🍎")
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
	}

	time.Sleep(ReactTime)
	err = s.MessageReactionAdd(m.ChannelID, m.ID, "🖋️")
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
	}
}

func HandleSelfMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.Contains(m.Content, "if i could smell like something id like to smell like bubble cum") {
		time.Sleep(ReplyTime)
		var err = reply(s, m.Message, "GUM!! I MEANT GUM!!!!")
		if err != nil {
			log.Printf("[ERROR] %s\n", err)
		}
	} else if strings.Contains(m.Content, "sorry wait a moment i lost my phone") {
		time.Sleep(ReplyTime)
		var err = reply(s, m.Message, "oh wait im dumb im texting with it lol")
		if err != nil {
			log.Printf("[ERROR] %s\n", err)
		}
	}
}

func replied(s *discordgo.Session, m *discordgo.MessageCreate) {
	if s.State.User.ID == m.Author.ID {
		HandleSelfMessages(s, m)
		return
	}

	// ignore bots, including yourself
	if m.Author.Bot {
		return
	}

	// Direct message
	if m.GuildID == "" {
		var err = reply(s, m.Message, ";3")
		if err != nil {
			log.Printf("[ERROR] %s\n", err)
		}
		return
	}

	if m.ChannelID == BotChannelID {
		select {
		case ResetBoredTimer <- struct{}{}:
		default:
		}
	}

	// Ignore blacklisted users
	if slices.Contains(BlacklistedUsers[:], m.Author.ID) {
		return
	}

	// Ignore blacklisted channels (like chitchat)
	if slices.Contains(ReplyChannelBlacklist[:], m.ChannelID) {
		return
	}

	var lowered = strings.ToLower(m.Content)
	var talksAboutSelf = mentionsBot(s, m.Message) ||
		strings.Contains(lowered, "polycule bot") ||
		strings.Contains(lowered, "polybot") ||
		strings.Contains(lowered, "the bot") ||
		strings.Contains(lowered, "polly")

	if talksAboutSelf &&
		strings.Contains(lowered, "shut") &&
		strings.Contains(lowered, "up") {
		select {
		case ShutUpSignal <- m.Message:
		default:
		}
		return
	}

	if talksAboutSelf &&
		strings.Contains(lowered, "come") &&
		strings.Contains(lowered, "back") {
		select {
		case ComeBackSignal <- m.Message:
		default:
		}
		return
	}

	if IsShutUp.Load() {
		return
	}

	// Mpreg react, sometimes
	if !DebugBuild {
		if strings.Contains(lowered, "fpreg") {
			time.Sleep(ReactTime)
			var err = s.MessageReactionAdd(m.ChannelID, m.ID, "🤰")
			if err != nil {
				log.Printf("[ERROR] %s\n", err)
			}
		} else if rand.Float64() <= ReactToMessageChance ||
			strings.Contains(lowered, "preg") {
			time.Sleep(ReactTime)
			var err = s.MessageReactionAdd(m.ChannelID, m.ID, "🫃")
			if err != nil {
				log.Printf("[ERROR] %s\n", err)
			}
		}
	}

	// Always reply to replies
	if m.ReferencedMessage != nil &&
		m.ReferencedMessage.Author.ID == s.State.User.ID {
		if strings.Contains(m.ReferencedMessage.Content, "doctor monster") {
			var err = reply(s, m.Message, "well im never telling youu")
			if err != nil {
				log.Println("[ERROR]", err)
			}
		} else if strings.Contains(m.ReferencedMessage.Content, "wubaduba-dub is that true?") {
			var err = reply(s, m.Message, "wow you go big guy!")
			if err != nil {
				log.Println("[ERROR]", err)
			}
		} else if rand.Float64() <= ReplyToReplyChance {
			var err = replyRandom(s, m.Message)
			if err != nil {
				log.Println("[ERROR]", err)
			}
		}
		return
	}

	var chance = ReplyToMessageChance
	if m.ChannelID == BotChannelID {
		chance = ReplyToMessageInBotChannelChance
	}

	if strings.Contains(lowered, "ppap") ||
		strings.Contains(lowered, "apple") ||
		strings.Contains(lowered, "🍍") ||
		strings.Contains(lowered, "🍎") ||
		strings.Contains(lowered, "🍏") ||
		strings.Contains(lowered, "🍍") ||
		strings.Contains(lowered, "🖋") {
		go ReactPPAP(s, m.Message)
		var ppap = "pen pineapple apple pen " + EmoticonsPool.Next()
		var err = reply(s, m.Message, ppap)
		if err != nil {
			log.Printf("[ERROR] %s\n", err)
		}
	} else if talksAboutSelf ||
		rand.Float64() <= chance {
		var err = replyRandom(s, m.Message)
		if err != nil {
			log.Printf("[ERROR] %s\n", err)
		}
	}
}

func shutUpDetector(s *discordgo.Session) {
	var timer = time.NewTimer(ShutUpTime)
	timer.Stop()

	var sendMsg = func(m *discordgo.Message, msg string) {
		if DebugBuild {
			msg = "[DEBUG] " + msg
		}

		var channelID = m.ChannelID
		if DebugBuild {
			channelID = DebugChannelID
		}

		time.Sleep(ReplyTime)
		_, err := s.ChannelMessageSend(channelID, msg)
		if err != nil {
			log.Println("[ERROR] cannot send message:", err)
		}
	}

	var originalShutUpMessage *discordgo.Message

	for {
		select {
		case m := <-ShutUpSignal:
			timer.Reset(ShutUpTime)

			if !IsShutUp.Load() {
				IsShutUp.Store(true)
				originalShutUpMessage = m
				var msg = ShutUpPool.Next()
				sendMsg(m, msg)
			}

		case m := <-ComeBackSignal:
			if IsShutUp.Load() {
				timer.Stop()
				IsShutUp.Store(false)
				var msg = ComeBackPool.Next()
				sendMsg(m, msg)
			}

		case <-timer.C:
			IsShutUp.Store(false)
			if originalShutUpMessage != nil {
				var msg = ComeBackPool.Next()
				sendMsg(originalShutUpMessage, msg)
			}
		}
	}
}

func boredTimerLoop(s *discordgo.Session) {
	var timer = time.NewTimer(TimeBeforeBored)
	var alreadySaidBored = false
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if alreadySaidBored || IsShutUp.Load() || DebugBuild {
				timer.Reset(TimeBeforeBored)
				continue
			}

			var message = BoredPool.Next()
			var emoticon = EmoticonsPool.Next()
			var finalMessage = fmt.Sprintf("%s %s", message, emoticon)

			var msg, err = s.ChannelMessageSend(BotChannelID, finalMessage)
			if err != nil {
				log.Println("[ERROR] cannot send bored message:", err)
			}

			if finalMessage == "hehe found the cookie jar" {
				go func() {
					<-time.After(FollowUpTime)
					err = reply(s, msg, "goshshh imn sho ful...........")
					if err != nil {
						log.Println("[ERROR] cannot send bored message:", err)
					}
				}()
			}

			timer.Reset(TimeBeforeBored)
			alreadySaidBored = true

		case <-ResetBoredTimer:
			timer.Reset(TimeBeforeBored)
			alreadySaidBored = false
		}
	}
}

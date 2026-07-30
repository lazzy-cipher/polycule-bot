package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	AdminID              = "254331907351904256"  // lazarus_overlook
	WelcomeChannelID     = "1532401026924150947" // chitchat
	WelcomeStickerID     = "1532410779469615288" // Lazzy Showoff
	WelcomeRoleID        = "1532412593816338482" // goober
	ReplyTime            = 2 * time.Second
	ReplyToTypingChance  = 0.005 // 0.5%
	ReplyToReplyChance   = 1.0   // 1.0%
	ReplyToMessageChance = 0.02  // 2%
	ReactToMessageChance = 0.05  // 5%
	Version              = "1.0.1"
	TimeBeforeBored      = 2 * time.Hour
)

var (
	Token       string
	ShowVersion bool
	Emoticons   = [...]string{
		":3",
		">w<",
		"x3",
		";w;",
		"'^'",
		"-w-",
		"^w^",
		"owo",
		"uwu",
		">>",
		">//>",
	}
	Replies = [...]string{
		"uhuh",
		"oo",
		"oki!",
		"?",
		"",
		"*rolls over*",
		"zip! ziip zip!",
		"wow",
		"mm mm mm",
		"true",
		"*flops*",
		"ooh",
		"okok",
		"hm",
		"idk seems kinda sus",
		"that's so based omg",
		"yeah no that's totally it actually",
		"why are you like this",
		"Unnnfhg...",
		"!!!",
		"a- are you sure...?",
		"whatever goob, I ain't gonna listent to a goob",
		"uh- a- uh- y-you pwettyy...",
		"so we're not gonna talk about what you sent me in dms???",
		"uh- you talk a lot uh- *impregnates you*, there, now shush",
		"woa",
		"oh really",
		"that's crazy",
		"idk im just not feeling it too too much",
		"ah okay anyway HANGOUT WHEN???",
		"trans lives matter most of the time!",
		"im so full.......",
	}
	ReplyChannelBlacklist = [...]string{
		"1532407644919431218",
	}
)

func init() {
	flag.StringVar(&Token, "token", "", "Bot Token")
	flag.BoolVar(&ShowVersion, "version", false, "Print version and exit")
	flag.Parse()
}

func main() {
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

	var dg, err = discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("[ERROR] cannot create Discord session:", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuildMessageTyping |
		discordgo.IntentGuildMessages |
		discordgo.IntentMessageContent

	dg.AddHandler(guildMemberStartsTyping)
	dg.AddHandler(guildMemberAdd)
	dg.AddHandler(replied)

	err = dg.Open()
	if err != nil {
		log.Fatal("[ERROR] cannot open connection:", err)
	}
	defer dg.Close()

	fmt.Println("Bot is running. Press CTRL-C to exit.")
	var sc = make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func getRandEmoticon() string {
	var messageIdx = rand.IntN(len(Emoticons))
	return Emoticons[messageIdx]
}

func HasSticker(s *discordgo.Session, guildID string, stickerID string) bool {
	var guild, err = s.Guild(guildID)
	if err != nil {
		log.Printf("[ERROR] unable to find guild %v\n", guildID)
		return false
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

	var availableMessages = [...]string{
		"Free use, btw",
		"Grab these whenever you want",
		"You'll see these two often",
		"...",
		"Btw, these two get lonely",
	}

	var messageIdx = rand.IntN(len(availableMessages))
	var message = fmt.Sprintf(`Welcome to the compound, <@%s>
%s`, m.User.ID, availableMessages[messageIdx])

	var _, err = s.ChannelMessageSendComplex(WelcomeChannelID, &discordgo.MessageSend{
		Content:    message,
		StickerIDs: []string{WelcomeStickerID},
	})
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
	}
}

func notifyAdmin(s *discordgo.Session, message string) {
	dmChannel, err := s.UserChannelCreate(AdminID)
	if err != nil {
		log.Println("[ERROR] cannot create DM channel with admin", err)
		return
	}

	_, err = s.ChannelMessageSend(dmChannel.ID, message)
	if err != nil {
		log.Println("[ERROR] cannot send DM to admin:", err)
	}
}

func guildMemberAdd(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	var err = s.GuildMemberRoleAdd(e.GuildID, e.User.ID, WelcomeRoleID)
	if err != nil {
		log.Println("[ERROR cannot add role:", err)
	}

	welcomeMessage(s, e.Member)
	notifyAdmin(s, fmt.Sprintf("Hiiii Laz!! Btw, new member: <@%s> :3", e.User.ID))
}

func guildMemberStartsTyping(s *discordgo.Session, e *discordgo.TypingStart) {
	if rand.Float64() > ReplyToTypingChance {
		return
	}

	time.Sleep(ReplyTime)

	var availableMessages = [...]string{
		"whatcha typin? <@%s>",
		"<@%s> haiii what's up!!",
		"<@%s> is typing... woa...",
		"oh <@%s> be typing",
		"omg <@%s> haaiiii",
		"I find it so hot when <@%s> types",
		"omg <@%s> hi okay omg I wanted to tell you something but i forgor",
		"<@%s> sending nudes???",
	}

	var messageIdx = rand.IntN(len(availableMessages))
	var message = fmt.Sprintf(availableMessages[messageIdx], e.UserID)

	var msg = fmt.Sprintf("%s %s", message, getRandEmoticon())
	var _, err = s.ChannelMessageSend(e.ChannelID, msg)
	if err != nil {
		log.Println("[ERROR] cannot send welcome message:", err)
	}
}

func reply(s *discordgo.Session, m *discordgo.Message) {
	time.Sleep(ReplyTime)

	var messageIdx = rand.IntN(len(Replies))
	var message = fmt.Sprintf("%s %s", Replies[messageIdx], getRandEmoticon())

	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content:   message,
		Reference: m.Reference(),
	})
}

func react(s *discordgo.Session, m *discordgo.Message) {
	time.Sleep(ReplyTime)

	var messageIdx = rand.IntN(len(Replies))
	var message = fmt.Sprintf("%s %s", Replies[messageIdx], getRandEmoticon())

	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content:   message,
		Reference: m.Reference(),
	})
}

func mentionsBot(s *discordgo.Session, m *discordgo.Message) bool {
	for _, u := range m.Mentions {
		if u.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

func replied(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore bots, including yourself
	if m.Author.Bot {
		return
	}

	// Ignore blacklisted channels (like chitchat)
	if slices.Contains(ReplyChannelBlacklist[:], m.ChannelID) {
		return
	}

	// Always reply to replies
	if m.ReferencedMessage != nil && m.ReferencedMessage.Author.ID == s.State.User.ID {
		if rand.Float64() <= ReplyToReplyChance {
			reply(s, m.Message)
		}
		return
	}

	if rand.Float64() <= ReactToMessageChance {
		var err = s.MessageReactionAdd(m.ChannelID, m.ID, "🫃")
		if err != nil {
			log.Printf("[ERROR] %s\n", err)
		}
	}

	if mentionsBot(s, m.Message) || rand.Float64() <= ReplyToMessageChance {
		react(s, m.Message)
	}
}

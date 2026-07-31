package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	AdminID              = "254331907351904256"  // lazarus_overlook
	WelcomeChannelID     = "1532401026924150947" // chitchat
	WelcomeStickerID     = "1532410779469615288" // Lazzy Showoff
	WelcomeRoleID        = "1532412593816338482" // goober
	BotChannelID         = "1532793476746449098" // talk-to-polycule-bot
	ReplyTime            = 2 * time.Second
	ReplyToTypingTime    = 500 * time.Millisecond
	ReactTime            = 1 * time.Second
	ReplyToTypingChance  = 0.005 // 0.5%
	ReplyToReplyChance   = 1.0   // 1.0%
	ReplyToMessageChance = 0.02  // 2%
	ReplyToMessageInBotChannelChance = 0.1  // 10%
	ReactToMessageChance = 0.04  // 4%
	Version              = "1.0.2"
	TimeBeforeBored      = 2 * time.Hour
)

var (
	Token           string
	ShowVersion     bool
	ResetBoredTimer = make(chan struct{})
	Emoticons       = [...]string{
		":3",
		";3",
		">w<",
		"x3",
		";w;",
		"'^'",
		"-w-",
		"^w^",
		"owo",
		"uwu",
		">>",
		"><",
		">///<",
		">//>",
		"o3o",
		"-^-",
	}
	Replies = [...]string{
		"uhuh",
		"oki!",
		"?",
		"",
		"",
		"*rolls over*",
		"zip! ziip zip!",
		"wow",
		"mm mm mm",
		"true",
		"true",
		"SO TRUE!!",
		"ye tru",
		"tru",
		"words",
		"mhm!",
		"sounds about right? I think?",
		"*flops*",
		"ooh",
		"okok",
		"hm",
		"ye",
		"yeah",
		"idk seems kinda sus",
		"that's so based omg",
		"yeah no that's totally it actually",
		"i like that you're weird like that",
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
		"aaa you talk too much just come to my place!!! in my room...",
		"i have so much to learn from you...",
		"how interesting..",
		"are you sure?",
		"new kink unlocked",
		"I consent!!!!!!!",
		"uh idk???? sorry im all over the place i have my periods rn...",
		"ugh my botvaries they hurt i will go rest for a while sorryyy",
		 "mi lawa sina la mi unpa wawa e sina",
		"aww.. dih.. oh- wait, what? oh! oh ye ye",
		"idk man im just hungry",
		"no way",
		"youd look so cute pregnant, you know that?",
		"shhh~ <3",
		"i dont regret what we did last night~",
		"whatever!!!!! you said we were going to costco to get candy and hot dogs!!! are we going soon??",
		"not what im into but i dig it",
		"im so so so sorry..",
		"its not like that!!! im- uh- ...",
		"*sniffs you*. mmmmh gay.",
		"gosh...",
		"twin... like... ugh nvm...",
		"its how it do be when it do be like that in that kind of way ig",
		"yeah ig..",
		"oh wait a minute, im bending over rn...",
		"thats not fair!!!",
		"are you okay... with me.. being the way I am..?",
		"big guh moment",
		"imma sit this one out.. on your lap",
		"OKAY I GUESS!!!!!",
		"*jitters from too much caffeine*",
		"lemme get in your pants im cold",
		"youre silly",
		"something i really really love about you is that you, huh... you know your place?",
		"maybe the two of us could be... more? like... friends...? maybe...?",
		"OKAY I ADMIT I HAVE A SECRET TO TELL YOU",
		"please send more messages like that~",
	}
	ReplyChannelBlacklist = [...]string{
		"1532407644919431218", // memories
	}
	BlacklistedUsers = [...]string{
		"1466282667258675324", // bardownbuddy
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
		log.Println("[ERROR] cannot add role:", err)
	}

	welcomeMessage(s, e.Member)
	notifyAdmin(s, fmt.Sprintf("Hiiii Laz!! Btw, new member: <@%s> :3", e.User.ID))
}

func guildMemberStartsTyping(s *discordgo.Session, e *discordgo.TypingStart) {
	if rand.Float64() > ReplyToTypingChance ||
		slices.Contains(BlacklistedUsers[:], e.UserID) {
		return
	}

	time.Sleep(ReplyToTypingTime)

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

func getRandomMessage() string {
	var messageIdx = rand.IntN(len(Replies))
	return fmt.Sprintf("%s %s", Replies[messageIdx], getRandEmoticon())
}

func reply(s *discordgo.Session, m *discordgo.Message, message string) {
	time.Sleep(ReplyTime)

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

func ReactPPAP(s *discordgo.Session, m *discordgo.Message) {
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

func replied(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore bots, including yourself
	if m.Author.Bot {
		return
	}

	// Direct message
	if m.GuildID == "" {
		reply(s, m.Message, ";3")
		return
	}

	if m.ChannelID == WelcomeChannelID {
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

	// Mpreg react, sometimes
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

	// Always reply to replies
	if m.ReferencedMessage != nil && m.ReferencedMessage.Author.ID == s.State.User.ID {
		if rand.Float64() <= ReplyToReplyChance {
			reply(s, m.Message, getRandomMessage())
		}
		return
	}

	var chance = ReplyToMessageChance
	if m.ChannelID == BotChannelID {
		chance = ReplyToMessageInBotChannelChance
	}

	if strings.Contains(lowered, "ppap") ||
		strings.Contains(lowered, "apple") {
		go ReactPPAP(s, m.Message)
		var ppap = "pen pineapple apple pen " + getRandEmoticon()
		reply(s, m.Message, ppap)
	} else if mentionsBot(s, m.Message) ||
		strings.Contains(lowered, "polycule bot") ||
		rand.Float64() <= chance {
		reply(s, m.Message, getRandomMessage())
	}
}

func boredTimerLoop(s *discordgo.Session) {
	timer := time.NewTimer(TimeBeforeBored)
	var alreadySaidBored = false
	defer timer.Stop()

	var boredMessages = [...]string{
		"I'M SO BORED OMG",
		"where is everybody...",
		"...so bored...",
		"hiiii so huh next hangout wen??",
		"@eveyon im borde",
		"*poke poke poke* anyone here?",
		"*yawn*... lomly",
		"<@" + AdminID + "> hey entertain me",
		"someone wake up!!",
	}

	for {
		select {
		case <-timer.C:
			if alreadySaidBored {
				timer.Reset(TimeBeforeBored)
				continue
			}

			var message = boredMessages[rand.IntN(len(boredMessages))]
			var emoticon = Emoticons[rand.IntN(len(Emoticons))]
			var finalMessage = fmt.Sprintf("%s %s", message, emoticon)

			_, err := s.ChannelMessageSend(WelcomeChannelID, finalMessage)
			if err != nil {
				log.Println("[ERROR] cannot send bored message:", err)
			}

			timer.Reset(TimeBeforeBored)
			alreadySaidBored = true

		case <-ResetBoredTimer:
			timer.Reset(TimeBeforeBored)
			alreadySaidBored = false
		}
	}
}

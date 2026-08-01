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
	ReplyToTypingChance              = 0.005 // 0.5%
	ReplyToReplyChance               = 1.0   // 1.0%
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
	IsShutUp atomic.Bool
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
		">:3c",
	}
	Comments = [...]string{
		"<@%s> unnerves me a bit i find it kind of adorb",
		"hey <@%s> guess what CHIKIN NUGETS",
		"i carried <@%s>s children 3/5 pretty good",
		"when i look at <@%s> i feel a lot of feelings, like, constipation, euphoria",
		"me and <@%s> are twins i love them so much they mean a lot to me",
		"<@%s> is my eternal frienemy turned forbidden romance turned fried chicken cook i sometimes go see after work (im unemployed)",
		"hey <@%s> i have a job for you, *rolls over*",
		"<@%s> and me could be a thing, maybe...",
		"i will protect <@%s> with my life",
		"*pokes <@%s> and runs away*",
		"*offers a fresh home grown tomato to <@%s>*",
		"<@%s> always puzzled me",
		"yeah id drink <@%s>'s milk",
		"so huh <@%s> when are you free??",
		"idk im kind of scared to talk to <@%s> they are kind of way too cute...",
		"<@%s> gives me gender euphoria",
		"ughn... <@%s>... umf...",
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
		"word",
		"mhm!",
		"sounds about right? I think?",
		"*flops*",
		"ooh",
		"okok",
		"hm",
		"ye",
		"yeah",
		"yuhuh",
		"nuhuh",
		"noperinos",
		"idk seems kinda sus",
		"idk",
		"guh",
		"that's so based omg",
		"yeah no that's totally it actually",
		"i like that you're weird like that",
		"Unnnfhg...",
		"!!!",
		"a- are you sure...?",
		"whatever goob, I ain't gonna listent to a goob",
		"uh- a- uh- y-you pwettyy...",
		"so we're not gonna talk about what you sent me in dms???",
		"uh- you talk a lot uh- *flips you upside down*, there, now shush",
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
		"youd look so cute pregnant tbh",
		"shhh~ <3",
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
		"youre silly",
		"something i really really love about you is that you, huh... you know your place?",
		"maybe the two of us could be... more? like... friends...? maybe...?",
		"OKAY I ADMIT I HAVE A SECRET TO TELL YOU",
		"that made me laugh",
		"okay i trust you...",
		"youre so real for that",
		"me...? 👉👈",
		"youre like a blahaj but not like a shark like in the shape of you",
		"okay im good at keeping secrets",
		"i have severe gender dysphoria",
		"i have severe body dysmorphia",
		"i like you... but... not in like a friendly way or anything like that. like, more? or less? its hard to describe...",
		"<@" + AdminID + "> programmed me too many feelings",
		"im too poly for that kind of stuff",
		"if i could smell like something id like to smell like bubble cum",
		"*cute autistic stimming*",
		"*gets constipated from joy* AAA why does it always do that!!",
		"its huh... not how it works",
		"well maybe sometimes idk",
		"dont ask me im just a girl",
		"dont ask me im just a boy",
		"sorry wait a moment i lost my phone",
		"im over stimulated rn... but huh... yeah",
		"im over stimulated rn... so... no",
		"mmh what you said was quite pregnant (Merriam-Webster. Pregnant (def. 3).)",
	}
	BoredMessages = [...]string{
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
	ShutUpMessages = [...]string{
		"aww ok... :c",
		"im gonna go...",
		"but- ok... :c",
		"*sadly waddles away* :c",
		":c",
		"fine..",
	}
	ComeBackMessages = [...]string{
		"can i talk again now? :c",
		"i missed you... ;-;",
		"was alone too long, wept, better now",
		"i was so along for so long... anyone wanna give me kisses? ;w;",
		"*hugs closest person* i was too quiet too long i got anxious... ;w;",
	}
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
	for _, c := range Comments {
		if strings.Count(c, "<@%s>") != 1 {
			log.Printf("[ERROR] comment \"%s\" doesn't contain exactly one \"<@%%s>\"", c)
			crash = true
		}
	}

	if crash {
		log.Fatal("[ERROR] invalid initial program state")
	}
}

func getRandEmoticon() string {
	var messageIdx = rand.IntN(len(Emoticons))
	return Emoticons[messageIdx]
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
		slices.Contains(ReplyChannelBlacklist, e.ChannelID){
		return
	}

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
	var finalMessage = fmt.Sprintf("%s %s", message, getRandEmoticon())

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
	var messageIdx = rand.IntN(len(Replies))
	return fmt.Sprintf("%s %s", Replies[messageIdx], getRandEmoticon())
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

	var _, err = s.ChannelMessageSend(DebugChannelID, "[DEBUG] " + message)

	return err
}

func GetRandomComment(userID string) string {
	var template = Comments[rand.IntN(len(Comments))]
	return fmt.Sprintf(template, userID)
}

func replyRandom(s *discordgo.Session, m *discordgo.Message) error {
	if len(m.Mentions) <= 0 {
		return reply(s, m, GetRandomMessage())
	}

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

	if len(filtered) <= 0 {
		return reply(s, m, GetRandomMessage())
	}

	var userID = filtered[rand.IntN(len(filtered))].ID
	return reply(s, m, GetRandomComment(userID))
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
		if rand.Float64() <= ReplyToReplyChance {
			var err = replyRandom(s, m.Message)
			if err != nil {
				log.Printf("[ERROR] %s\n", err)
			}
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
				var msg = ShutUpMessages[rand.IntN(len(ShutUpMessages))]
				sendMsg(m, msg)
			}

		case m := <-ComeBackSignal:
			if IsShutUp.Load() {
				timer.Stop()
				IsShutUp.Store(false)
				var msg = ComeBackMessages[rand.IntN(len(ComeBackMessages))]
				sendMsg(m, msg)
			}

		case <-timer.C:
			IsShutUp.Store(false)
			if originalShutUpMessage != nil {
				var msg = ComeBackMessages[rand.IntN(len(ComeBackMessages))]
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

			var message = BoredMessages[rand.IntN(len(BoredMessages))]
			var emoticon = Emoticons[rand.IntN(len(Emoticons))]
			var finalMessage = fmt.Sprintf("%s %s", message, emoticon)
			if DebugBuild {
				finalMessage = "[DEBUG] " + finalMessage
			}

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

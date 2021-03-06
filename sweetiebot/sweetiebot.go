package sweetiebot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blackhole12/discordgo"
)

// BotConfig lists all bot configuration options, grouped into structs
type BotConfig struct {
	Version     int  `json:"version"`
	LastVersion int  `json:"lastversion"`
	SetupDone   bool `json:"setupdone"`
	Basic       struct {
		IgnoreInvalidCommands bool                       `json:"ignoreinvalidcommands"`
		Importable            bool                       `json:"importable"`
		AlertRole             uint64                     `json:"alertrole"`
		ModChannel            uint64                     `json:"modchannel"`
		FreeChannels          map[string]bool            `json:"freechannels"`
		BotChannel            uint64                     `json:"botchannel"`
		Aliases               map[string]string          `json:"aliases"`
		Collections           map[string]map[string]bool `json:"collections"`
		ListenToBots          bool                       `json:"listentobots"`
		CommandPrefix         string                     `json:"commandprefix"`
		TrackUserLeft         bool                       `json:"trackuserleft"`
	} `json:"basic"`
	Modules struct {
		Channels           map[string]map[string]bool `json:"modulechannels"`
		Disabled           map[string]bool            `json:"moduledisabled"`
		CommandRoles       map[string]map[string]bool `json:"commandroles"`
		CommandChannels    map[string]map[string]bool `json:"commandchannels"`
		CommandLimits      map[string]int64           `json:"Commandlimits"`
		CommandDisabled    map[string]bool            `json:"commanddisabled"`
		CommandPerDuration int                        `json:"commandperduration"`
		CommandMaxDuration int64                      `json:"commandmaxduration"`
	} `json:"modules"`
	Spam struct {
		ImagePressure      float32            `json:"imagepressure"`
		PingPressure       float32            `json:"pingpressure"`
		LengthPressure     float32            `json:"lengthpressure"`
		RepeatPressure     float32            `json:"repeatpressure"`
		LinePressure       float32            `json:"linepressure"`
		BasePressure       float32            `json:"basepressure"`
		PressureDecay      float32            `json:"pressuredecay"`
		MaxPressure        float32            `json:"maxpressure"`
		MaxChannelPressure map[uint64]float32 `json:"maxchannelpressure"`
		MaxRemoveLookback  int                `json:"MaxSpamRemoveLookback"`
		SilentRole         uint64             `json:"silentrole"`
		IgnoreRole         uint64             `json:"ignorerole"`
		RaidTime           int64              `json:"maxraidtime"`
		RaidSize           int                `json:"raidsize"`
		SilenceMessage     string             `json:"silencemessage"`
		AutoSilence        int                `json:"autosilence"`
		LockdownDuration   int                `json:"lockdownduration"`
	} `json:"spam"`
	Bucket struct {
		MaxItems       int `json:"maxbucket"`
		MaxItemLength  int `json:"maxbucketlength"`
		MaxFightHP     int `json:"maxfighthp"`
		MaxFightDamage int `json:"maxfightdamage"`
	} `json:"bucket"`
	Markov struct {
		MaxPMlines     int  `json:"maxpmlines"`
		MaxLines       int  `json:"maxquotelines"`
		DefaultLines   int  `json:"defaultmarkovlines"`
		UseMemberNames bool `json:"usemembernames"`
	} `json:"markov"`
	Users struct {
		TimezoneLocation string          `json:"timezonelocation"`
		WelcomeChannel   uint64          `json:"welcomechannel"`
		WelcomeMessage   string          `json:"welcomemessage"`
		Roles            map[uint64]bool `json:"userroles"`
	} `json:"users"`
	Bored struct {
		Cooldown int64           `json:"maxbored"`
		Commands map[string]bool `json:"boredcommands"`
	}
	Help struct {
		Rules             map[int]string `json:"rules"`
		HideNegativeRules bool           `json:"hidenegativerules"`
	} `json:"help"`
	Log struct {
		Cooldown int64  `json:"maxerror"`
		Channel  uint64 `json:"logchannel"`
	} `json:"log"`
	Witty struct {
		Responses map[string]string `json:"witty"`
		Cooldown  int64             `json:"maxwit"`
	} `json:"Wit"`
	Schedule struct {
		BirthdayRole uint64 `json:"birthdayrole"`
	} `json:"schedule"`
	Search struct {
		MaxResults int `json:"maxsearchresults"`
	} `json:"search"`
	Spoiler struct {
		Channels []uint64 `json:"spoilchannels"`
	} `json:"spoiler"`
	Status struct {
		Cooldown int `json:"statusdelaytime"`
	} `json:"status"`
	Quote struct {
		Quotes map[uint64][]string `json:"quotes"`
	} `json:"quote"`
}

// ConfigHelp is a map of help strings for the configuration options above
var ConfigHelp map[string]string = map[string]string{
	"basic.ignoreinvalidcommands": "If true, Sweetie Bot won't display an error if a nonsensical command is used. This helps her co-exist with other bots that also use the `!` prefix.",
	"basic.importable":            "If true, the collections on this server will be importable into another server where sweetie is.",
	"basic.alertrole":             "This is intended to point at a moderator role shared by all admins and moderators of the server for notification purposes.",
	"basic.modchannel":            "This should point at the hidden moderator channel, or whatever channel moderates want to be notified on.",
	"basic.freechannels":          "This is a list of all channels that are exempt from rate limiting. Usually set to the dedicated `#botabuse` channel in a server.",
	"basic.botchannel":            "This allows you to designate a particular channel for sweetie bot to point users to if they are trying to run too many commands at once. Usually this channel will also be included in `basic.freechannels`",
	"basic.aliases":               "Can be used to redirect commands, such as making `!listgroup` call the `!listgroups` command. Useful for making shortcuts.\n\nExample: `!setconfig basic.aliases kawaii \"pick cute\"` sets an alias mapping `!kawaii arg1...` to `!pick cute arg1...`, preserving all arguments that are passed to the alias.",
	"basic.collections":           "All the collections used by sweetiebot. Manipulate it via `!add` and `!remove`",
	"basic.listentobots":          "If true, sweetiebot will process bot messages and allow them to run commands. Bots can never trigger anti-spam. Defaults to false.",
	"basic.commandprefix":         "Determines the SINGLE ASCII CHARACTER prefix used to denote sweetiebot commands. You can't set it to an emoji or any weird foreign character. The default is `!`. If this is set to an invalid value, Sweetiebot will default to using `!`.",
	"basic.trackuserleft":         "If true, sweetiebot will also track users that leave the server if autosilence is set to alert or log. Defaults to false.",
	"modules.commandroles":        "A map of which roles are allowed to run which command. If no mapping exists, everyone can run the command.",
	"modules.commandchannels":     "A map of which channels commands are allowed to run on. No entry means a command can be run anywhere. If \"!\" is included as a channel, it switches from a whitelist to a blacklist, enabling you to exclude certain channels instead of allow certain channels.",
	"modules.commandlimits":       "A map of timeouts for commands. A value of 30 means the command can't be used more than once every 30 seconds.",
	"modules.commanddisabled":     "A list of disabled commands.",
	"modules.commandperduration":  "Maximum number of commands that can be run within `commandmaxduration` seconds. Default: 3",
	"modules.commandmaxduration":  "Default: 20. This means that by default, at most 3 commands can be run every 20 seconds.",
	"modules.disabled":            "A list of disabled modules.",
	"modules.channels":            "A mapping of what channels a given module can operate on. If no mapping is given, a module operates on all channels. If \"!\" is included as a channel, it switches from a whitelist to a blacklist, enabling you to exclude certain channels instead of allow certain channels.",
	"spam.imagepressure":          "Additional pressure generated by each image, link or attachment in a message. Defaults to (MaxPressure - BasePressure) / 6, instantly silencing anyone posting 6 or more links at once.",
	"spam.repeatpressure":         "Additional pressure generated by a message that is identical to the previous message sent (ignores case). Defaults to BasePressure, effectively doubling the pressure penalty for repeated messages.",
	"spam.pingpressure":           "Additional pressure generated by each unique ping in a message. Defaults to (MaxPressure - BasePressure) / 20, instantly silencing anyone pinging 20 or more people at once.",
	"spam.lengthpressure":         "Additional pressure generated by each individual character in the message. Discord allows messages up to 2000 characters in length. Defaults to (MaxPressure - BasePressure) / 8000, silencing anyone posting 3 huge messages at the same time.",
	"spam.linepressure":           "Additional pressure generated by each newline in the message. Defaults to (MaxPressure - BasePressure) / 70, silencing anyone posting more than 70 newlines in a single message",
	"spam.basepressure":           "The base pressure generated by sending a message, regardless of length or content. Defaults to 10",
	"spam.maxpressure":            "The maximum pressure allowed. If a user's pressure exceeds this amount, they will be silenced. Defaults to 60, which is intended to ban after a maximum of 6 short messages sent in rapid succession.",
	"spam.maxchannelpressure":     "Per-channel pressure override. If a channel's pressure is specified in this map, it will override the global maxpressure setting.",
	"spam.pressuredecay":          "The number of seconds it takes for a user to lose Spam.BasePressure from their pressure amount. Defaults to 2.5, so after sending 3 messages, it will take 7.5 seconds for their pressure to return to 0.",
	"spam.maxremovelookback":      "Number of seconds back the bot should delete messages of a silenced user on the channel they spammed on. If set to 0, the bot will only delete the message that caused the user to be silenced. If less than 0, the bot won't delete any messages.",
	"spam.ignorerole":             "If set, the bot will exclude anyone with this role from spam detection. Use with caution.",
	"spam.silentrole":             "This should be a role with no permissions, so the bot can quarantine potential spammers without banning them.",
	"spam.raidtime":               "In order to trigger a raid alarm, at least `spam.raidsize` people must join the chat within this many seconds of each other.",
	"spam.raidsize":               "Specifies how many people must have joined the server within the `spam.raidtime` period to qualify as a raid.",
	"spam.silencemessage":         "This message will be sent to users that have been silenced by the `!silence` command.",
	"spam.autosilence":            "Gets the current autosilence state. Use the `!autosilence` command to set this.",
	"spam.lockdownduration":       "Determines how long the server's verification mode will temporarily be increased to tableflip levels after a raid is detected. If set to 0, disables lockdown entirely.",
	"bucket.maxitems":             "Determines the maximum number of items sweetiebot can carry in her bucket. If set to 0, her bucket is disabled.",
	"bucket.maxitemlength":        "Determines the maximum length of a string that can be added to her bucket.",
	"bucket.maxfighthp":           "Maximum HP of the randomly generated enemy for the `!fight` command.",
	"bucket.maxfightdamage":       "Maximum amount of damage a randomly generated weapon can deal for the `!fight` command.",
	"markov.maxpmlines":           "This is the maximum number of lines a response can be before sweetiebot automatically sends it as a PM to avoid cluttering the chat. Default: 5",
	"markov.maxlines":             "Maximum number of lines the `!episodequote` command can be given.",
	"markov.defaultlines":         "Number of lines for the markov chain to spawn when not given a line count.",
	"markov.usemembernames":       "Use member names instead of random pony names.",
	"users.timezonelocation":      "Sets the timezone location of the server itself. When no user timezone is available, the bot will use this.",
	"users.welcomechannel":        "If set to a channel ID, the bot will treat this channel as a \"quarantine zone\" for silenced members. If autosilence is enabled, new users will be sent to this channel.",
	"users.welcomemessage":        "If autosilence is enabled, this message will be sent to a new user upon joining.",
	"users.roles":                 "A list of all user-assignable roles. Manage it via !addrole and !removerole",
	"bored.cooldown":              "The bored cooldown timer, in seconds. This is the length of time a channel must be inactive for sweetiebot to post a bored message in it.",
	"bored.commands":              "This determines what commands sweetie will run when she gets bored. She will choose one command from this list at random.\n\nExample: `!setconfig bored.commands !drop \"!pick bored\"`",
	"help.rules":                  "Contains a list of numbered rules. The numbers do not need to be contiguous, and can be negative.",
	"help.hidenegativerules":      "If true, `!rules -1` will display a rule at index -1, but `!rules` will not. This is useful for joke rules or additional rules that newcomers don't need to know about.",
	"log.channel":                 "This is the channel where sweetiebot logs her output.",
	"log.cooldown":                "The cooldown time for sweetiebot to display an error message, in seconds, intended to prevent the bot from spamming itself. Default: 4",
	"witty.responses":             "Stores the replies used by the Witty module and must be configured using `!addwit` or `!removewit`",
	"witty.cooldown":              "The cooldown time for the witty module. At least this many seconds must have passed before the bot will make another witty reply.",
	"schedule.birthdayrole":       " This is the role given to members on their birthday.",
	"search.maxresults":           "Maximum number of search results that can be requested at once.",
	"spoiler.channels":            "A list of channels that are exempt from the spoiler rules.",
	"status.cooldown":             "Number of seconds sweetiebot waits before changing her status to a string picked randomly from the `status` collection.",
	"quote.quotes":                "This is a map of quotes, which should be managed via `!addquote` and `!removequote`.",
}

// Version represents an app version using four sections
type Version struct {
	major    byte
	minor    byte
	revision byte
	build    byte
}

func (v *Version) String() string {
	if v.build > 0 {
		return fmt.Sprintf("%v.%v.%v.%v", v.major, v.minor, v.revision, v.build)
	}
	if v.revision > 0 {
		return fmt.Sprintf("%v.%v.%v", v.major, v.minor, v.revision)
	}
	return fmt.Sprintf("%v.%v", v.major, v.minor)
}

// Integer gets the integer representation of the version
func (v *Version) Integer() int {
	return AssembleVersion(v.major, v.minor, v.revision, v.build)
}

// AssembleVersion creates a version integer out of four bytes
func AssembleVersion(major byte, minor byte, revision byte, build byte) int {
	return int(build) | (int(revision) << 8) | (int(minor) << 16) | (int(major) << 24)
}

// SweetieBot is the primary bot object containing the bot state
type SweetieBot struct {
	db                 *BotDB
	dg                 *discordgo.Session
	Debug              bool `json:"debug"`
	version            Version
	changelog          map[int]string
	SelfID             string
	SelfAvatar         string
	Owners             map[uint64]bool
	RestrictedCommands map[string]bool
	NonServerCommands  map[string]bool
	MainGuildID        uint64
	DBGuilds           map[uint64]bool   `json:"dbguilds"`
	DebugChannels      map[string]string `json:"debugchannels"`
	quit               AtomicBool
	guilds             map[uint64]*GuildInfo
	guildsLock         sync.RWMutex
	LastMessages       map[string]int64
	LastMessagesLock   sync.RWMutex
	MaxConfigSize      int
	StartTime          int64
	MessageCount       uint32 // 32-bit so we can do atomic ops on a 32-bit platform
	heartbeat          uint32 // perpetually incrementing heartbeat counter to detect deadlock
	locknumber         uint32
}

var sb *SweetieBot
var channelregex = regexp.MustCompile("<#[0-9]+>")
var roleregex = regexp.MustCompile("<@&[0-9]+>")
var userregex = regexp.MustCompile("<@!?[0-9]+>")
var mentionregex = regexp.MustCompile("<@(!|&)?[0-9]+>")
var discriminantregex = regexp.MustCompile(".*#[0-9][0-9][0-9]+")
var repeatregex = regexp.MustCompile("repeat -?[0-9]+ (second|minute|hour|day|week|month|quarter|year)s?")
var colorregex = regexp.MustCompile("0x[0-9A-Fa-f]+")
var locUTC = time.FixedZone("UTC", 0)

// DiscordEpoch is used to figure out snowflake creation times
var DiscordEpoch uint64 = 1420070400000

// IsMainGuild returns true if that guild is considered the main (default) guild
func (sbot *SweetieBot) IsMainGuild(info *GuildInfo) bool {
	return SBatoi(info.ID) == sbot.MainGuildID
}

// IsDBGuild returns true if that guild is allowed to use the database
func (sbot *SweetieBot) IsDBGuild(info *GuildInfo) bool {
	_, ok := sbot.DBGuilds[SBatoi(info.ID)]
	return ok
}

// BulkDelete Performs a bulk deletion in groups of 100
func (sbot *SweetieBot) BulkDelete(channelID string, messages []string) (err error) {
	i := 0
	n := len(messages)
	for (n - i) > 99 {
		err := sbot.dg.ChannelMessagesBulkDelete(channelID, messages[i:i+99])
		if err != nil {
			return err
		}
	}
	return sbot.dg.ChannelMessagesBulkDelete(channelID, messages[i:])
}

// PartialSanitize escapes ``` and emotes`
func PartialSanitize(s string) string {
	s = strings.Replace(s, "```", "\\`\\`\\`", -1)
	return strings.Replace(s, "[](/", "[\u200B](/", -1)
}

// ExtraSanitize calls PartialSanitize and also sanitizes links
func ExtraSanitize(s string) string {
	s = strings.Replace(s, "http://", "http\u200B://", -1)
	s = strings.Replace(s, "https://", "https\u200B://", -1)
	return PartialSanitize(ReplaceAllMentions(s))
}

func typeIsPrivate(ty discordgo.ChannelType) bool {
	return ty != discordgo.ChannelTypeGuildText && ty != discordgo.ChannelTypeGuildCategory && ty != discordgo.ChannelTypeGuildVoice
}
func channelIsPrivate(channelID string) (*discordgo.Channel, bool) {
	if channelID == "heartbeat" {
		return nil, true
	}
	ch, err := sb.dg.State.Channel(channelID)
	if err == nil { // Because of the magic of web development, we can get a message BEFORE the "channel created" packet for the channel being used by that message.
		return ch, typeIsPrivate(ch.Type)
	}
	// Bots aren't supposed to be in Group DMs but can be grandfathered into them, and these channels will always fail to exist, so we simply ignore this error as harmless.
	return nil, true
}

// ChangeBotName changes the username and avatar of the bot
func ChangeBotName(s *discordgo.Session, name string, avatarfile string) {
	binary, _ := ioutil.ReadFile(avatarfile)
	avatar := base64.StdEncoding.EncodeToString(binary)

	_, err := s.UserUpdate("", "", name, "data:image/png;base64,"+avatar, "")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Changed username successfully")
	}
}

//func sbEvent(s *discordgo.Session, e *discordgo.Event) { ApplyFuncRange(len(info.hooks.OnEvent), func(i int) { if(ProcessModule("", info.hooks.OnEvent[i])) { info.hooks.OnEvent[i].OnEvent(s, e) } }) }
func sbReady(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Println("Ready message receieved, waiting for guilds...")
	sb.SelfID = r.User.ID
	sb.SelfAvatar = r.User.Avatar
	isuser, _ := ioutil.ReadFile("isuser") // THIS FILE SHOULD NOT EXIST UNLESS YOU WANT TO BE IN USER MODE. If you don't know what user mode is, you don't want it.
	if r.Guilds != nil && isuser != nil {
		for _, G := range r.Guilds {
			AttachToGuild(G)
		}
	}

	// Only used to change sweetiebot's name or avatar
	//ChangeBotName(s, "Sweetie", "avatar.png")
}

// AttachToGuild adds a guild to sweetiebot's state tracking
func AttachToGuild(g *discordgo.Guild) {
	sb.guildsLock.RLock()
	guild, exists := sb.guilds[SBatoi(g.ID)]
	sb.guildsLock.RUnlock()
	if exists {
		guild.ProcessGuild(g)
		return
	}
	if sb.Debug {
		_, ok := sb.DebugChannels[g.ID]
		if !ok {
			/*guild = &GuildInfo{
				ID:           g.ID,
				Name:         g.Name,
				OwnerID:      g.OwnerID,
				commandLast:  make(map[string]map[string]int64),
				commandlimit: &SaturationLimit{[]int64{}, 0, AtomicFlag{0}},
				commands:     make(map[string]Command),
				emotemodule:  nil,
				lockdown:     -1,
			}
			sb.guildsLock.Lock()
			sb.guilds[SBatoi(g.ID)] = guild
			guild.ProcessGuild(g)
			sb.guildsLock.Unlock()
			fmt.Println("Processed", g.Name)*/
			return
		}
	}

	fmt.Println("Initializing " + g.Name)

	guild = &GuildInfo{
		ID:           g.ID,
		Name:         g.Name,
		OwnerID:      g.OwnerID,
		commandLast:  make(map[string]map[string]int64),
		commandlimit: &SaturationLimit{[]int64{}, 0, AtomicFlag{0}},
		commands:     make(map[string]Command),
		emotemodule:  nil,
		lockdown:     -1,
		lastlogerr:   0,
	}
	config, err := ioutil.ReadFile(g.ID + ".json")
	disableall := false
	if err != nil {
		fmt.Println("New Guild Detected: " + g.Name)
		config, _ = ioutil.ReadFile("default.json")
		ch, e := sb.dg.UserChannelCreate(g.OwnerID)
		if e == nil {
			sb.db.SetDefaultServer(SBatoi(g.OwnerID), SBatoi(g.ID)) // This ensures no one blows up another server by accident
			perms, _ := getAllPerms(guild, sb.SelfID)
			warning := ""
			if perms&0x00000008 != 0 {
				warning = "\nWARNING: You have given sweetiebot the Administrator role, which implicitly gives her all roles! Sweetie Bot only needs Ban Members, Manage Roles and Manage Messages in order to function correctly." + warning
			}
			if perms&0x00020000 != 0 {
				warning = "\nWARNING: You have given sweetiebot the Mention Everyone role, which means users will be able to abuse her to ping everyone on the server! Sweetie Bot does NOT attempt to filter @\u200Beveryone from her messages!" + warning
			}
			if perms&0x00000004 == 0 {
				warning = "\nWARNING: Sweetiebot cannot ban members spamming the welcome channel without the Ban Members role! (If you do not use this feature, it is safe to ignore this warning)." + warning
			}
			if perms&0x10000000 == 0 {
				warning = "\nWARNING: Sweetiebot cannot silence members or give birthday roles without the Manage Roles role!" + warning
			}
			if perms&0x00002000 == 0 {
				warning = "\nWARNING: Sweetiebot cannot delete messages without the Manage Messages role!" + warning
			}
			if perms&0x00000020 == 0 {
				warning = "\nWARNING: Sweetiebot cannot engage lockdown mode without the Manage Server role!" + warning
			}
			sb.dg.ChannelMessageSend(ch.ID, "You've successfully added Sweetie Bot to your server! To finish setting her up, run the `setup` command. Here is an explanation of the command and an example:\n```!setup <Mod Role> <Mod Channel> [Log Channel]```\n**> Mod Role**\nThis is a role shared by all the moderators and admins of your server. Sweetie Bot will ping this role to alert you about potential raids or silenced users, and sensitive commands will be restricted so only users with the moderator role can use them. As the server owner, you will ALWAYS be able to run any command, no matter what. This ensures that you can always fix a broken configuration. Before running `!setup`, make sure your moderator role can be pinged: Go to Server Settings -> Roles and select your mod role, then make sure \"Allow anyone to @mention this role\" is checked.\n\n**> Mod Channel**\nThis is the channel Sweetie Bot will post alerts on. Usually, this is your private moderation channel, but you can make it whatever channel you want. Just make sure you use the format `#channel`, and ensure the bot actually has permission to post messages on the channel.\n\n**> Log Channel**\nThis is an optional channel where sweetiebot will post errors and update notifications. Usually, this is only visible to server admins and the bot. Remember to give the bot permission to post messages on the log channel, or you won't get any output. Providing a log channel is highly recommended, because it's often Sweetie Bot's last resort for notifying you about potential errors.\n\nThat's it! Here is an example of the command: ```!setup @Mods #staff-chat #bot-log```\n\nNote: **Do not run `!setup` in this PM!** It won't work because Discord won't autocomplete `#channel` for you. Run `!setup` directly on your server.")
			if len(warning) > 0 {
				sb.dg.ChannelMessageSend(ch.ID, warning)
			}
		} else {
			fmt.Println("Error sending introductory PM: ", e)
		}
		disableall = true
	}
	err = MigrateSettings(config, guild)
	if err != nil {
		fmt.Println("Error reading config file for "+g.Name+": ", err.Error())
	}

	guild.commandlimit.times = make([]int64, guild.config.Modules.CommandPerDuration*2, guild.config.Modules.CommandPerDuration*2)

	if len(guild.config.Witty.Responses) == 0 {
		guild.config.Witty.Responses = make(map[string]string)
	}
	if len(guild.config.Basic.Aliases) == 0 {
		guild.config.Basic.Aliases = make(map[string]string)
	}
	if len(guild.config.Basic.FreeChannels) == 0 {
		guild.config.Basic.FreeChannels = make(map[string]bool)
	}
	if len(guild.config.Modules.CommandRoles) == 0 {
		guild.config.Modules.CommandRoles = make(map[string]map[string]bool)
	}
	if len(guild.config.Modules.CommandChannels) == 0 {
		guild.config.Modules.CommandChannels = make(map[string]map[string]bool)
	}
	if len(guild.config.Modules.CommandLimits) == 0 {
		guild.config.Modules.CommandLimits = make(map[string]int64)
	}
	if len(guild.config.Modules.CommandDisabled) == 0 {
		guild.config.Modules.CommandDisabled = make(map[string]bool)
	}
	if len(guild.config.Modules.Disabled) == 0 {
		guild.config.Modules.Disabled = make(map[string]bool)
	}
	if len(guild.config.Modules.Channels) == 0 {
		guild.config.Modules.Channels = make(map[string]map[string]bool)
	}
	if len(guild.config.Users.Roles) == 0 {
		guild.config.Users.Roles = make(map[uint64]bool)
	}
	if len(guild.config.Basic.Collections) == 0 {
		guild.config.Basic.Collections = make(map[string]map[string]bool)
	}

	collections := []string{"emote", "bored", "status", "spoiler", "bucket", "cute"}
	for _, v := range collections {
		_, ok := guild.config.Basic.Collections[v]
		if !ok {
			guild.config.Basic.Collections[v] = make(map[string]bool)
		}
	}

	sb.guildsLock.Lock()
	sb.guilds[SBatoi(g.ID)] = guild
	guild.ProcessGuild(g) // This can be done outside of the guild lock, but it puts a lot of pressure on the database
	sb.guildsLock.Unlock()

	guild.emotemodule = &EmoteModule{}
	guild.emotemodule.UpdateRegex(guild)
	spoilermodule := &SpoilerModule{}
	spoilermodule.UpdateRegex(guild)

	addfuncmap := map[string]func(string) string{
		"emote": func(arg string) string {
			r := guild.emotemodule.UpdateRegex(guild)
			if !r {
				delete(guild.config.Basic.Collections["emote"], arg)
				guild.emotemodule.UpdateRegex(guild)
				return ". Failed to ban " + arg + " because regex compilation failed"
			}
			return "and recompiled the emote regex"
		},
		"spoiler": func(arg string) string {
			r := spoilermodule.UpdateRegex(guild)
			if !r {
				delete(guild.config.Basic.Collections["spoiler"], arg)
				spoilermodule.UpdateRegex(guild)
				return ". Failed to ban " + arg + " because regex compilation failed"
			}
			return "and recompiled the spoiler regex"
		},
	}
	removefuncmap := map[string]func(string) string{
		"emote": func(arg string) string {
			guild.emotemodule.UpdateRegex(guild)
			return "```Unbanned " + arg + " and recompiled the emote regex.```"
		},
		"spoiler": func(arg string) string {
			spoilermodule.UpdateRegex(guild)
			return "```Unbanned " + arg + " and recompiled the spoiler regex.```"
		},
	}

	wittymodule := &WittyModule{lastcomment: 0, lastdelete: 0}
	wittymodule.UpdateRegex(guild)

	guild.modules = make([]Module, 0, 6)
	guild.modules = append(guild.modules, &DebugModule{})
	guild.modules = append(guild.modules, &UsersModule{})
	guild.modules = append(guild.modules, &CollectionsModule{AddFuncMap: addfuncmap, RemoveFuncMap: removefuncmap})
	guild.modules = append(guild.modules, &ScheduleModule{})
	guild.modules = append(guild.modules, &RolesModule{})
	guild.modules = append(guild.modules, &PollModule{})
	guild.modules = append(guild.modules, &HelpModule{})
	guild.modules = append(guild.modules, &MarkovModule{})
	guild.modules = append(guild.modules, &QuoteModule{})
	guild.modules = append(guild.modules, &BucketModule{})
	guild.modules = append(guild.modules, &MiscModule{guild.emotemodule})
	guild.modules = append(guild.modules, &ConfigModule{})
	guild.modules = append(guild.modules, &SpamModule{tracker: make(map[uint64]*userPressure), lastraid: 0})
	guild.modules = append(guild.modules, wittymodule)
	guild.modules = append(guild.modules, &StatusModule{})
	guild.modules = append(guild.modules, &BoredModule{lastmessage: 0})
	guild.modules = append(guild.modules, guild.emotemodule)
	guild.modules = append(guild.modules, spoilermodule)

	for _, v := range guild.modules {
		guild.RegisterModule(v)
		cmds := v.Commands()
		for _, command := range cmds {
			guild.AddCommand(command)
		}
	}

	for _, v := range guild.modules {
		_, ok := guild.commands[strings.ToLower(v.Name())]
		if ok {
			fmt.Println("WARNING: Ambiguous module/command name ", v.Name())
		}
	}
	if disableall {
		for k := range guild.commands {
			guild.config.Modules.CommandDisabled[k] = true
		}
		for _, v := range guild.modules {
			guild.config.Modules.Disabled[strings.ToLower(v.Name())] = true
		}
		delete(guild.config.Modules.CommandDisabled, "setup")
		guild.SaveConfig()
	}
	if sb.IsMainGuild(guild) {
		sb.db.log = guild
		go guild.SwapStatusLoop()
	}

	go func() { // Do this concurrently because we don't need this to function properly, we just need it to happen eventually
		// Discord doesn't send us all the members, so we force feed them into the state ourselves
		members := []*discordgo.Member{}
		lastid := ""
		for {
			m, err := sb.dg.GuildMembers(guild.ID, lastid, 999)
			if err != nil || len(m) == 0 {
				break
			}
			members = append(members, m...)
			lastid = m[len(m)-1].User.ID
		}
		for i := range members { // Put the guildID back in because discord is stupid
			members[i].GuildID = guild.ID
			sb.dg.State.MemberAdd(members[i])
		}
	}()

	debug := "."
	if sb.Debug {
		debug = ".\n[DEBUG BUILD]"
	}
	changes := ""
	if guild.config.LastVersion != sb.version.Integer() {
		guild.config.LastVersion = sb.version.Integer()
		guild.SaveConfig()
		var ok bool
		changes, ok = sb.changelog[sb.version.Integer()]
		if ok {
			changes = "\nChangelog:\n" + changes + "\n\nPlease consider donating to help pay for hosting costs: https://www.patreon.com/erikmcclure"
		}
	}
	guild.Log("Sweetiebot version ", sb.version.String(), " successfully loaded on ", g.Name, debug, changes)
}
func getChannelGuild(id string) *GuildInfo {
	c, err := sb.dg.State.Channel(id)
	if err != nil {
		fmt.Println("Failed to get channel " + id)
		return nil
	}
	sb.guildsLock.RLock()
	g, ok := sb.guilds[SBatoi(c.GuildID)]
	sb.guildsLock.RUnlock()
	if !ok {
		return nil
	}
	return g
}
func getGuildFromID(id string) *GuildInfo {
	sb.guildsLock.RLock()
	g, ok := sb.guilds[SBatoi(id)]
	sb.guildsLock.RUnlock()
	if !ok {
		return nil
	}
	return g
}
func getAddMsg(info *GuildInfo) string {
	if info.config.Basic.BotChannel != 0 {
		addch, adderr := sb.dg.State.Channel(SBitoa(info.config.Basic.BotChannel))
		if adderr == nil {
			return fmt.Sprintf(" Try going to #%s instead.", addch.Name)
		}
	}
	return ""
}

// SBProcessCommand processes a command given to sweetiebot in the form "!command"
func SBProcessCommand(s *discordgo.Session, m *discordgo.Message, info *GuildInfo, t int64, isdbguild bool, isdebug bool) {
	var prefix byte = '!'
	if info != nil && len(info.config.Basic.CommandPrefix) == 1 {
		prefix = info.config.Basic.CommandPrefix[0]
	}

	// Check if this is a command. If it is, process it as a command, otherwise process it with our modules.
	if len(m.Content) > 1 && m.Content[0] == prefix && (len(m.Content) < 2 || m.Content[1] != prefix) { // We check for > 1 here because a single character can't possibly be a valid command
		private := info == nil
		isfree := private
		authorid := SBatoi(m.Author.ID)
		if info != nil {
			_, isfree = info.config.Basic.FreeChannels[m.ChannelID]
		}
		_, isOwner := sb.Owners[authorid]
		isSelf := m.Author.ID == sb.SelfID

		if !isSelf && info != nil {
			ignore := false
			for _, h := range info.hooks.OnCommand {
				if info.ProcessModule(m.ChannelID, h) {
					ignore = ignore || h.OnCommand(info, m)
				}
			}
			if ignore && !isOwner && m.Author.ID != info.OwnerID { // if true, a module wants us to ignore this command
				return
			}
		}
		args, indices := ParseArguments(m.Content[1:])
		arg := strings.ToLower(args[0])
		if info == nil && !sb.db.status.get() {
			s.ChannelMessageSend(m.ChannelID, "```A temporary database error means I can't process any private message commands right now.```")
			return
		}
		if info == nil {
			info = getDefaultServer(authorid)
		}
		if info == nil {
			gIDs := sb.db.GetUserGuilds(authorid)
			_, independent := sb.NonServerCommands[arg]
			if !independent && len(gIDs) != 1 {
				s.ChannelMessageSend(m.ChannelID, "```Cannot determine what server you belong to! Use !defaultserver to set which server I should use when you PM me.```")
				return
			}

			if len(gIDs) == 0 {
				gIDs = []uint64{sb.MainGuildID}
			}
			sb.guildsLock.RLock()
			info = sb.guilds[gIDs[0]]
			sb.guildsLock.RUnlock()
			if info == nil {
				s.ChannelMessageSend(m.ChannelID, "```I haven't been loaded on that server yet!```")
				return
			}
		}
		c, ok := info.commands[arg] // First, we check if this matches an existing command so you can't alias yourself into a hole
		if !ok {
			alias, aliasok := info.config.Basic.Aliases[arg]
			if aliasok {
				if len(indices) > 1 {
					m.Content = info.config.Basic.CommandPrefix + alias + " " + m.Content[indices[1]:]
				} else {
					m.Content = info.config.Basic.CommandPrefix + alias
				}
				args, indices = ParseArguments(m.Content[1:])
				arg = strings.ToLower(args[0])
				c, ok = info.commands[arg]
			}
		}
		if ok {
			if isdbguild && sb.db.status.get() && m.Author.ID != sb.SelfID {
				sb.db.Audit(AUDIT_TYPE_COMMAND, m.Author, m.Content, SBatoi(info.ID))
			}
			isOwner = isOwner || m.Author.ID == info.OwnerID
			cmdname := strings.ToLower(c.Name())
			cch := info.config.Modules.CommandChannels[cmdname]
			_, disabled := info.config.Modules.CommandDisabled[cmdname]
			_, restricted := sb.RestrictedCommands[cmdname]
			if disabled && !isOwner && !isSelf {
				return
			}
			if restricted && !isdbguild {
				return
			}
			if !private && len(cch) > 0 && !isSelf {
				_, reverse := cch["!"]
				_, ok = cch[m.ChannelID]
				if ok == reverse {
					return
				}
			}
			if !isdebug && !isfree && !isSelf && info.config.Modules.CommandPerDuration > 0 && !info.UserHasRole(m.Author.ID, SBitoa(info.config.Basic.AlertRole)) { // debug channels aren't limited
				if len(info.commandlimit.times) < info.config.Modules.CommandPerDuration*2 { // Check if we need to re-allocate the array because the configuration changed
					info.commandlimit.times = make([]int64, info.config.Modules.CommandPerDuration*2, info.config.Modules.CommandPerDuration*2)
				}
				if info.commandlimit.check(info.config.Modules.CommandPerDuration, info.config.Modules.CommandMaxDuration, t) { // if we've hit the saturation limit, post an error (which itself will only post if the error saturation limit hasn't been hit)
					info.Error(m.ChannelID, fmt.Sprintf("You can't input more than %v commands every %s!%s", info.config.Modules.CommandPerDuration, TimeDiff(time.Duration(info.config.Modules.CommandMaxDuration)*time.Second), getAddMsg(info)))
					return
				}
				info.commandlimit.append(t)
			}
			if !isOwner && !isSelf && !info.UserHasAnyRole(m.Author.ID, info.config.Modules.CommandRoles[cmdname]) {
				info.Error(m.ChannelID, "You don't have permission to run this command! Allowed Roles: "+info.GetRoles(c))
				return
			}

			cmdlimit := info.config.Modules.CommandLimits[cmdname]
			if !isfree && cmdlimit > 0 && !isSelf {
				info.commandLock.RLock()
				lastcmd := info.commandLast[m.ChannelID][cmdname]
				info.commandLock.RUnlock()
				if !RateLimit(&lastcmd, cmdlimit) {
					info.Error(m.ChannelID, fmt.Sprintf("You can only run that command once every %s!%s", TimeDiff(time.Duration(cmdlimit)*time.Second), getAddMsg(info)))
					return
				}
				info.commandLock.Lock()
				if len(info.commandLast[m.ChannelID]) == 0 {
					info.commandLast[m.ChannelID] = make(map[string]int64)
				}
				info.commandLast[m.ChannelID][cmdname] = t
				info.commandLock.Unlock()
			}

			result, usepm, resultembed := c.Process(args[1:], m, indices[1:], info)
			if len(result) > 0 || resultembed != nil {
				targetchannel := m.ChannelID
				if usepm && !private {
					channel, err := s.UserChannelCreate(m.Author.ID)
					info.LogError("Error opening private channel: ", err)
					if err == nil {
						targetchannel = channel.ID
						private = true
						if rand.Float32() < 0.01 {
							info.SendMessage(m.ChannelID, "Check your ~~privilege~~ Private Messages for my reply!")
						} else {
							info.SendMessage(m.ChannelID, "```Check your Private Messages for my reply!```")
						}
					}
				}

				if resultembed != nil {
					info.SendEmbed(targetchannel, resultembed)
				} else {
					info.SendMessage(targetchannel, result)
				}
			}
		} else {
			if !info.config.Basic.IgnoreInvalidCommands {
				info.Error(m.ChannelID, "Sorry, "+args[0]+" is not a valid command.\nFor a list of valid commands, type !help.")
			}
		}
	} else if info != nil { // If info is nil this was sent through a private message so just ignore it completely
		for _, h := range info.hooks.OnMessageCreate {
			if info.ProcessModule(m.ChannelID, h) {
				h.OnMessageCreate(info, m)
			}
		}
	}
}

func sbMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	atomic.AddUint32(&sb.MessageCount, 1)
	if m.Author == nil { // This shouldn't ever happen but we check for it anyway
		return
	}

	t := time.Now().UTC().Unix()
	sb.LastMessagesLock.Lock()
	sb.LastMessages[m.ChannelID] = t
	sb.LastMessagesLock.Unlock()

	ch, private := channelIsPrivate(m.ChannelID)
	var info *GuildInfo
	isdbguild := true
	isdebug := false
	if !private {
		info = getChannelGuild(m.ChannelID)
		if info == nil {
			return
		}
		isdbguild = sb.IsDBGuild(info)
		isdebug = info.IsDebug(m.ChannelID)
	}

	if isdebug && !sb.Debug {
		return // we do this up here so the release build doesn't log messages in bot-debug, but debug builds still log messages from the rest of the channels
	}
	if m.ChannelID != "heartbeat" {
		if info != nil && isdbguild && sb.db.CheckStatus() { // Log this message if it was sent to the main guild only.
			cid := SBatoi(m.ChannelID)
			if cid != info.config.Log.Channel {
				sb.db.AddMessage(SBatoi(m.ID), SBatoi(m.Author.ID), SanitizeMentions(m.ContentWithMentionsReplaced()), cid, m.MentionEveryone, SBatoi(ch.GuildID))
			}
		}
		if info != nil {
			sb.db.SentMessage(SBatoi(m.Author.ID), SBatoi(info.ID))
		}
		if m.Author.ID == sb.SelfID { // discard all our own messages (unless this is a heartbeat message)
			return
		}
		if info != nil && !info.config.Basic.ListenToBots && m.Author.Bot { // If we aren't supposed to listen to bot messages, discard them.
			return
		}
		if boolXOR(sb.Debug, isdebug) { // debug builds only respond to the debug channel, and release builds ignore it
			return
		}
	} else {
		sb.guildsLock.RLock()
		info, _ = sb.guilds[sb.MainGuildID]
		if info == nil {
			fmt.Println("Failed to get main guild during heartbeat test!")
		}
		sb.guildsLock.RUnlock()
	}

	SBProcessCommand(s, m.Message, info, t, isdbguild, isdebug)
}

func sbMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	info := getChannelGuild(m.ChannelID)
	if info == nil {
		return
	}
	if boolXOR(sb.Debug, info.IsDebug(m.ChannelID)) {
		return
	}
	if m.Author == nil { // Discord sends an update message with an empty author when certain media links are posted
		original, err := s.ChannelMessage(m.ChannelID, m.ID)
		if err != nil {
			info.LogError("Error processing MessageUpdate: ", err)
			return // Fuck it, we can't process this
		}
		m.Author = original.Author
	}

	ch, err := sb.dg.State.Channel(m.ChannelID)
	info.LogError("Error retrieving channel ID "+m.ChannelID+": ", err)
	private := true
	if err == nil {
		private = typeIsPrivate(ch.Type)
	}
	cid := SBatoi(m.ChannelID)
	if cid != info.config.Log.Channel && !private && sb.IsDBGuild(info) && sb.db.CheckStatus() { // Always ignore messages from the log channel
		sb.db.AddMessage(SBatoi(m.ID), SBatoi(m.Author.ID), SanitizeMentions(m.ContentWithMentionsReplaced()), cid, m.MentionEveryone, SBatoi(ch.GuildID))
	}
	if m.Author.ID == sb.SelfID {
		return
	}
	for _, h := range info.hooks.OnMessageUpdate {
		if info.ProcessModule(m.ChannelID, h) {
			h.OnMessageUpdate(info, m.Message)
		}
	}
}
func sbMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	info := getChannelGuild(m.ChannelID)
	if info == nil {
		return
	}
	if boolXOR(sb.Debug, info.IsDebug(m.ChannelID)) {
		return
	}
	for _, h := range info.hooks.OnMessageDelete {
		if info.ProcessModule(m.ChannelID, h) {
			h.OnMessageDelete(info, m.Message)
		}
	}
}
func sbUserUpdate(s *discordgo.Session, m *discordgo.UserUpdate) {
	ProcessUser(m.User, nil)
}
func sbPresenceUpdate(s *discordgo.Session, m *discordgo.PresenceUpdate) {
	info := getGuildFromID(m.GuildID)
	if info == nil {
		return
	}
	ProcessUser(m.User, &m.Presence)

	for _, h := range info.hooks.OnPresenceUpdate {
		if info.ProcessModule("", h) {
			h.OnPresenceUpdate(info, m)
		}
	}
}
func sbGuildUpdate(s *discordgo.Session, m *discordgo.GuildUpdate) {
	info := getChannelGuild(m.ID)
	if info == nil {
		return
	}
	fmt.Println("Guild update detected, updating", m.Name)
	info.ProcessGuild(m.Guild)

	for _, h := range info.hooks.OnGuildUpdate {
		if info.ProcessModule("", h) {
			h.OnGuildUpdate(info, m.Guild)
		}
	}
}
func sbGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	info := getGuildFromID(m.GuildID)
	if info == nil {
		return
	}
	info.ProcessMember(m.Member)

	for _, h := range info.hooks.OnGuildMemberAdd {
		if info.ProcessModule("", h) {
			h.OnGuildMemberAdd(info, m.Member)
		}
	}
}
func sbGuildMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	info := getGuildFromID(m.GuildID)
	if info == nil {
		return
	}
	sb.db.RemoveMember(SBatoi(m.User.ID), SBatoi(info.ID))

	for _, h := range info.hooks.OnGuildMemberRemove {
		if info.ProcessModule("", h) {
			h.OnGuildMemberRemove(info, m.Member)
		}
	}

	if m.User.ID == sb.SelfID {
		fmt.Println("Sweetie was removed from", info.Name)
		sb.guildsLock.Lock()
		delete(sb.guilds, SBatoi(info.ID))
		sb.guildsLock.Unlock()
	}
}
func sbGuildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	info := getGuildFromID(m.GuildID)
	if info == nil {
		return
	}
	info.ProcessMember(m.Member)

	for _, h := range info.hooks.OnGuildMemberUpdate {
		if info.ProcessModule("", h) {
			h.OnGuildMemberUpdate(info, m.Member)
		}
	}
}
func sbGuildBanAdd(s *discordgo.Session, m *discordgo.GuildBanAdd) {
	info := getGuildFromID(m.GuildID) // We don't actually need to resolve this to get the guildID for SawBan, but we want to ignore any guilds we get messages from that we aren't currently attached to.
	if info == nil {
		return
	}

	for _, h := range info.hooks.OnGuildBanAdd {
		if info.ProcessModule("", h) {
			h.OnGuildBanAdd(info, m)
		}
	}
}
func sbGuildBanRemove(s *discordgo.Session, m *discordgo.GuildBanRemove) {
	info := getGuildFromID(m.GuildID)
	if info == nil {
		return
	}

	for _, h := range info.hooks.OnGuildBanRemove {
		if info.ProcessModule("", h) {
			h.OnGuildBanRemove(info, m)
		}
	}
}
func sbGuildRoleDelete(s *discordgo.Session, m *discordgo.GuildRoleDelete) {
	info := getGuildFromID(m.GuildID)
	if info == nil {
		return
	}

	for _, h := range info.hooks.OnGuildRoleDelete {
		if info.ProcessModule("", h) {
			h.OnGuildRoleDelete(info, m)
		}
	}
}
func sbGuildCreate(s *discordgo.Session, m *discordgo.GuildCreate) { AttachToGuild(m.Guild) }
func sbGuildDelete(s *discordgo.Session, m *discordgo.GuildDelete) {
	fmt.Println("Sweetie was deleted from", m.Guild.Name)
	sb.guildsLock.Lock()
	delete(sb.guilds, SBatoi(m.Guild.ID))
	sb.guildsLock.Unlock()
}
func sbChannelCreate(s *discordgo.Session, c *discordgo.ChannelCreate) {
	sb.guildsLock.RLock()
	guild, ok := sb.guilds[SBatoi(c.GuildID)]
	sb.guildsLock.RUnlock()
	if ok {
		setupSilenceRole(guild)
	}
}
func sbChannelDelete(s *discordgo.Session, c *discordgo.ChannelDelete) {

}

// ProcessUser adds a user to the database
func ProcessUser(u *discordgo.User, p *discordgo.Presence) uint64 {
	isonline := (p != nil && p.Status != "Offline")
	id := SBatoi(u.ID)
	discriminator, _ := strconv.Atoi(u.Discriminator)
	if sb.db.CheckStatus() {
		sb.db.AddUser(id, u.Email, u.Username, discriminator, u.Avatar, u.Verified, isonline)
	}
	return id
}

func idleCheckLoop() {
	for !sb.quit.get() {
		sb.guildsLock.RLock()
		infos := make([]*GuildInfo, 0, len(sb.guilds))
		for _, v := range sb.guilds {
			infos = append(infos, v)
		}
		sb.guildsLock.RUnlock()
		for _, info := range infos {
			guild, err := sb.dg.State.Guild(info.ID)
			if err != nil {
				continue
			}
			sb.dg.State.RLock()
			channels := guild.Channels
			sb.dg.State.RUnlock()
			if sb.Debug { // override this in debug mode
				c, err := sb.dg.State.Channel(sb.DebugChannels[info.ID])
				if err == nil {
					channels = []*discordgo.Channel{c}
				} else {
					channels = []*discordgo.Channel{}
				}
			}
			for _, ch := range channels {
				sb.LastMessagesLock.RLock()
				t, exists := sb.LastMessages[ch.ID]
				sb.LastMessagesLock.RUnlock()
				if exists {
					diff := time.Now().UTC().Sub(time.Unix(t, 0))

					for _, h := range info.hooks.OnIdle {
						if info.ProcessModule(ch.ID, h) && diff >= (time.Duration(h.IdlePeriod(info))*time.Second) {
							h.OnIdle(info, ch)
						}
					}
				}
			}

			for _, h := range info.hooks.OnTick {
				if info.ProcessModule("", h) {
					h.OnTick(info)
				}
			}

			if info.lockdown != -1 && time.Now().UTC().Sub(info.lastlockdown) > (time.Duration(info.config.Spam.LockdownDuration)*time.Second) {
				DisableLockdown(info)
			}
		}

		fmt.Println("Idle Check: ", time.Now())
		time.Sleep(20 * time.Second)
	}
}

func deadlockTestFunc(s *discordgo.Session, m *discordgo.MessageCreate) {
	sb.dg.State.RLock()
	sb.dg.State.RUnlock()
	sb.locknumber++
	sb.dg.RLock()
	sb.dg.RUnlock()
	sb.locknumber++
	sbMessageCreate(sb.dg, m)
}

const heartbeatInterval time.Duration = 20 * time.Second

func deadlockDetector() {
	var counter = sb.heartbeat
	var missed = 0
	time.Sleep(heartbeatInterval) // Give sweetie time to load everything first before initiating heartbeats
	for !sb.quit.get() {
		sb.guildsLock.RLock()
		info, ok := sb.guilds[sb.MainGuildID]
		sb.guildsLock.RUnlock()

		if !ok {
			fmt.Println(sb.MainGuildID, "MAIN GUILD CANNOT BE FOUND! Deadlock detector is nonfunctional until this is addressed.")
			time.Sleep(heartbeatInterval)
			continue
		}
		m := discordgo.MessageCreate{
			&discordgo.Message{ChannelID: "heartbeat", Content: info.config.Basic.CommandPrefix + "about",
				Author: &discordgo.User{
					ID:       sb.SelfID,
					Verified: true,
					Bot:      true,
				},
				Timestamp: discordgo.Timestamp(time.Now().UTC().Format(time.RFC3339Nano)),
			},
		}
		sb.locknumber = 0
		go deadlockTestFunc(sb.dg, &m) // Do this in another thread so the deadlock detector doesn't deadlock
		time.Sleep(heartbeatInterval)
		if atomic.LoadUint32(&sb.heartbeat) == counter+1 {
			counter++
			missed = 0
		} else {
			missed++
			fmt.Println("MISSED HEARTBEAT SIGNAL ", missed, " TIMES IN A ROW")
			counter = atomic.LoadUint32(&sb.heartbeat)
		}
		if missed >= 5 {
			fmt.Println("FATAL ERROR: DEADLOCK DETECTED! (", sb.locknumber, ") TERMINATING PROGRAM...")
			os.Exit(-1)
		}
	}
}

type emptyLog struct{}

func (log *emptyLog) Log(args ...interface{}) {
	s := fmt.Sprint(args...)
	fmt.Printf("[%s] %s\n", time.Now().Format(time.Stamp), s)
}

func (log *emptyLog) LogError(msg string, err error) {
	if err != nil {
		log.Log(msg, err.Error())
	}
}

// New creates and initializes a new instance of Sweetiebot that's ready to connect. Returns nil on error.
func New(token string) *SweetieBot {
	dbauth, dberr := ioutil.ReadFile("db.auth")
	if dberr != nil {
		fmt.Println("db.auth cannot be found. Please add the file with the correct format as specified in INSTALLATION.md")
	}
	mainguild, gerr := ioutil.ReadFile("mainguild")
	if gerr != nil {
		fmt.Println("mainguild cannot be found. Please add the file with the correct format as specified in INSTALLATION.md")
	}
	debugchannels, debugerr := ioutil.ReadFile("debug")
	rand.Seed(time.Now().UTC().Unix())

	mainguildid := SBatoi(strings.TrimSpace(string(mainguild)))
	sb = &SweetieBot{
		version:            Version{0, 9, 8, 14},
		Debug:              false,
		Owners:             map[uint64]bool{95585199324143616: true},
		RestrictedCommands: map[string]bool{"search": true, "lastping": true, "setstatus": true},
		NonServerCommands:  map[string]bool{"about": true, "roll": true, "episodegen": true, "bestpony": true, "episodequote": true, "help": true, "listguilds": true, "update": true, "announce": true, "dumptables": true, "defaultserver": true},
		MainGuildID:        mainguildid,
		DBGuilds:           make(map[uint64]bool),
		DebugChannels:      make(map[string]string),
		quit:               AtomicBool{0},
		guilds:             make(map[uint64]*GuildInfo),
		LastMessages:       make(map[string]int64),
		MaxConfigSize:      1000000,
		StartTime:          time.Now().UTC().Unix(),
		heartbeat:          4294967290,
		MessageCount:       0,
		changelog: map[int]string{
			AssembleVersion(0, 9, 8, 14): "- Reduce database pressure on startup",
			AssembleVersion(0, 9, 8, 13): "- Fix crash on startup.\n- Did more code refactoring, fixed several spelling errors.",
			AssembleVersion(0, 9, 8, 12): "- Do bulk member insertions in single batch to reduce database pressure.\n- Removed bestpony command\n- Did large internal code refactor",
			AssembleVersion(0, 9, 8, 11): "- User left now lists username+discriminator instead of pinging them to avoid @invalid-user problems.\n- Add ToS to !about\n- Bot now detects when it's about to be rate limited and combines short messages into a single large message. Helps keep bot responsive during huge raids.\n- Fixed race condition in spam module.",
			AssembleVersion(0, 9, 8, 10): "- !setup can now be run by any user with the administrator role.\n- Sweetie splits up embed messages if they have more than 25 fields.\n- Added !getraid and !banraid commands\n- Replaced !wipewelcome with generic !wipe command\n- Added LinePressure, which adds pressure for each newline in a message\n- Added TrackUserLeft, which will send a message when a user leaves in addition to when they join.",
			AssembleVersion(0, 9, 8, 9):  "- Moved several options to outside files to make self-hosting simpler to set up",
			AssembleVersion(0, 9, 8, 8):  "- !roll returns errors now.\n- You can now change the command prefix to a different ascii character - no, you can't set it to an emoji. Don't try.",
			AssembleVersion(0, 9, 8, 7):  "- Account creation time included on join message.\n- Specifying the config category is now optional. For example, !setconfig rules 3 \"blah\" works.",
			AssembleVersion(0, 9, 8, 6):  "- Support a lot more time formats and make time format more obvious.",
			AssembleVersion(0, 9, 8, 5):  "- Augment discordgo with maps instead of slices, and switch to using standard discordgo functions.",
			AssembleVersion(0, 9, 8, 4):  "- Update discordgo.",
			AssembleVersion(0, 9, 8, 3):  "- Allow deadlock detector to respond to deadlocks in the underlying discordgo library.\n- Fixed guild user count.",
			AssembleVersion(0, 9, 8, 2):  "- Simplify sweetiebot setup\n- Setting autosilence now resets the lockdown timer\n- Sweetiebot won't restore the verification level if it was manually changed by an administrator.",
			AssembleVersion(0, 9, 8, 1):  "- Switch to fork of discordgo to fix serious connection error handling issues.",
			AssembleVersion(0, 9, 8, 0):  "- Attempts to register if she is removed from a server.\n- Silencing has been redone to minimize rate-limiting problems.\n- Sweetie now tracks the first time someone posts a message, used in the \"bannewcomers\" command, which bans everyone who sent their first message in the past two minutes (configurable).\n- Sweetie now attempts to engage a lockdown when a raid is detected by temporarily increasing the server verification level. YOU MUST GIVE HER \"MANAGE SERVER\" PERMISSIONS FOR THIS TO WORK! This can be disabled by setting Spam.LockdownDuration to 0.",
			AssembleVersion(0, 9, 7, 9):  "- Discard Group DM errors from legacy conversations.",
			AssembleVersion(0, 9, 7, 8):  "- Correctly deal with rare edge-case on !userinfo queries.",
			AssembleVersion(0, 9, 7, 7):  "- Sweetiebot sends an autosilence change message before she starts silencing raiders, to ensure admins get immediate feedback even if discord is being slow.",
			AssembleVersion(0, 9, 7, 6):  "- Sweetiebot now ignores other bots by default. To revert this, run '!setconfig basic.listentobots true' and she will listen to them again, but will never attempt to silence them.\n- Removed legacy timezones\n- Spam messages are limited to 300 characters in the log.",
			AssembleVersion(0, 9, 7, 5):  "- Compensate for discordgo being braindead and forgetting JoinedAt dates.",
			AssembleVersion(0, 9, 7, 4):  "- Update discordgo API.",
			AssembleVersion(0, 9, 7, 3):  "- Fix permissions issue.",
			AssembleVersion(0, 9, 7, 2):  "- Fix ignoring admins in anti-spam.",
			AssembleVersion(0, 9, 7, 1):  "- Fixed an issue with out-of-date guild objects not including all server members.",
			AssembleVersion(0, 9, 7, 0):  "- Groups have been removed and replaced with user-assignable roles. All your groups have automatically been migrated to roles. If there was a name-collision with an existing role, your group name will be prefixed with 'sb-', which you can then resolve yourself. Use '!help roles' to get usage information about the new commands.",
			AssembleVersion(0, 9, 6, 9):  "- Sweetiebot no longer logs her own actions in the audit log",
			AssembleVersion(0, 9, 6, 8):  "- Sweetiebot now has a deadlock detector and will auto-restart if she detects that she is not responding to !about\n- Appending @ to the end of a name or server is no longer necessary. If sweetie finds an exact match to your query, she will always use that.",
			AssembleVersion(0, 9, 6, 7):  "- Sweetiebot no longer attempts to track edited messages for spam detection. This also fixes a timestamp bug with pinned messages.",
			AssembleVersion(0, 9, 6, 6):  "- Sweetiebot now automatically sets Silence permissions on newly created channels. If you have a channel that silenced members should be allowed to speak in, make sure you've set it as the welcome channel via !setconfig users.welcomechannel #yourchannel",
			AssembleVersion(0, 9, 6, 5):  "- Fix spam detection error for edited messages.",
			AssembleVersion(0, 9, 6, 4):  "- Enforce max DB connections to try to mitigate connection problems",
			AssembleVersion(0, 9, 6, 3):  "- Extreme spam could flood SB with user updates, crashing the database. She now throttles user updates to help prevent this.\n- Anti-spam now uses discord's message timestamp, which should prevent false positives from network problems\n- Sweetie will no longer silence mods for spamming under any circumstance.",
			AssembleVersion(0, 9, 6, 2):  "- Renamed !quickconfig to !setup, added a friendly PM to new servers to make initial setup easier.",
			AssembleVersion(0, 9, 6, 1):  "- Fix !bestpony crash",
			AssembleVersion(0, 9, 6, 0):  "- Sweetiebot is now self-repairing and can function without a database, although her functionality is EXTREMELY limited in this state.",
			AssembleVersion(0, 9, 5, 9):  "- MaxRemoveLookback no longer relies on the database and can now be used in any server. However, it only deletes messages from the channel that was spammed in.",
			AssembleVersion(0, 9, 5, 8):  "- You can now specify per-channel pressure overrides via '!setconfig spam.maxchannelpressure <channel> <pressure>'.",
			AssembleVersion(0, 9, 5, 7):  "- You can now do '!pick collection1+collection2' to pick a random item from multiple collections.\n- !fight <monster> is now sanitized.\n- !silence now tells you when someone already silenced will be unsilenced, if ever.",
			AssembleVersion(0, 9, 5, 6):  "- Prevent idiots from setting status.cooldown to 0 and breaking everything.",
			AssembleVersion(0, 9, 5, 5):  "- Fix crash on invalid command limits.",
			AssembleVersion(0, 9, 5, 4):  "- Added ignorerole for excluding certain users from spam detection.\n- Adjusted unsilence to force bot to assume user is unsilenced so it can be used to fix race conditions.",
			AssembleVersion(0, 9, 5, 3):  "- Prevent users from aliasing existing commands.",
			AssembleVersion(0, 9, 5, 2):  "- Show user account creation date in userinfo\n- Added !SnowflakeTime command",
			AssembleVersion(0, 9, 5, 1):  "- Allow !setconfig to edit float values",
			AssembleVersion(0, 9, 5, 0):  "- Completely overhauled Anti-Spam module. Sweetie now analyzes message content and tracks text pressure users exert on the chat. See !help anti-spam for details, or !getconfig spam for descriptions of the new configuration options. Your old MaxImages and MaxPings settings were migrated over to ImagePressure and PingPressure, respectively.",
			AssembleVersion(0, 9, 4, 5):  "- Escape nicknames correctly\n- Sweetiebot no longer tracks per-server nickname changes, only username changes.\n- You can now use the format username#1234 in user arguments.",
			AssembleVersion(0, 9, 4, 4):  "- Fix locks, update endpoint calls, improve antispam response.",
			AssembleVersion(0, 9, 4, 3):  "- Emergency revert of last changes",
			AssembleVersion(0, 9, 4, 2):  "- Spammer killing is now asynchronous and should have fewer duplicate alerts.",
			AssembleVersion(0, 9, 4, 1):  "- Attempt to make sweetiebot more threadsafe.",
			AssembleVersion(0, 9, 4, 0):  "- Reduced number of goroutines, made updating faster.",
			AssembleVersion(0, 9, 3, 9):  "- Added !getaudit command for server admins.\n- Updated documentation for consistency.",
			AssembleVersion(0, 9, 3, 8):  "- Removed arbitrary limit on spam message detection, replaced with sanity limit of 600.\n- Sweetiebot now automatically detects invalid spam.maxmessage settings and removes them instead of breaking your server.\n- Replaced a GuildMember call with an initial state check to eliminate lag and some race conditions.",
			AssembleVersion(0, 9, 3, 7):  "- If a collection only has one item, just display the item.\n- If you put \"!\" into CommandRoles[<command>], it will now allow any role EXCEPT the roles specified to use <command>. This behaves the same as the channel blacklist function.",
			AssembleVersion(0, 9, 3, 6):  "- Add log option to autosilence.\n- Ensure you actually belong to the server you set as your default.",
			AssembleVersion(0, 9, 3, 5):  "- Improve help messages.",
			AssembleVersion(0, 9, 3, 4):  "- Prevent cross-server message sending exploit, without destroying all private messages this time.",
			AssembleVersion(0, 9, 3, 3):  "- Emergency revert change.",
			AssembleVersion(0, 9, 3, 2):  "- Prevent cross-server message sending exploit.",
			AssembleVersion(0, 9, 3, 1):  "- Allow sweetiebot to be executed as a user bot.",
			AssembleVersion(0, 9, 3, 0):  "- Make argument parsing more consistent\n- All commands that accepted a trailing argument without quotes no longer strip quotes out. The quotes will now be included in the query, so don't put them in if you don't want them!\n- You can now escape '\"' inside an argument via '\\\"', which will work even if discord does not show the \\ character.",
			AssembleVersion(0, 9, 2, 3):  "- Fix echoembed crash when putting in invalid parameters.",
			AssembleVersion(0, 9, 2, 2):  "- Update help text.",
			AssembleVersion(0, 9, 2, 1):  "- Add !joingroup warning to deal with breathtaking stupidity of zootopia users.",
			AssembleVersion(0, 9, 2, 0):  "- Remove !lastping\n- Help now lists modules with no commands",
			AssembleVersion(0, 9, 1, 1):  "- Fix crash in !getconfig",
			AssembleVersion(0, 9, 1, 0):  "- Renamed config options\n- Made things more clear for new users\n- Fixed legacy importable problem\n- Fixed command saturation\n- Added botchannel notification\n- Changed getconfig behavior for maps",
			AssembleVersion(0, 9, 0, 4):  "- To protect privacy, !listguilds no longer lists servers that do not have Basic.Importable set to true.\n- Remove some more unnecessary sanitization",
			AssembleVersion(0, 9, 0, 3):  "- Don't sanitize links already in code blocks",
			AssembleVersion(0, 9, 0, 2):  "- Alphabetize collections because Tawmy is OCD",
			AssembleVersion(0, 9, 0, 1):  "- Update documentation\n- Simplify !collections output",
			AssembleVersion(0, 9, 0, 0):  "- Completely restructured Sweetie Bot into a module-based architecture\n- Disabling/Enabling a module now disables/enables all its commands\n- Help now includes information about modules\n- Collections command is now pretty",
			AssembleVersion(0, 8, 17, 2): "- Added ability to hide negative rules because Tawmy is weird",
			AssembleVersion(0, 8, 17, 1): "- Added echoembed command",
			AssembleVersion(0, 8, 17, 0): "- Sweetiebot can now send embeds\n- Made about message pretty",
			AssembleVersion(0, 8, 16, 3): "- Update discordgo structs to account for breaking API change.",
			AssembleVersion(0, 8, 16, 2): "- Enable sweetiebot to tell dumbasses that they are dumbasses.",
			AssembleVersion(0, 8, 16, 1): "- !add can now add to multiple collections at the same time.",
			AssembleVersion(0, 8, 16, 0): "- Alphabetized the command list",
			AssembleVersion(0, 8, 15, 4): "- ReplaceMentions now breaks role pings (but does not resolve them)",
			AssembleVersion(0, 8, 15, 3): "- Use database to resolve users to improve responsiveness",
			AssembleVersion(0, 8, 15, 2): "- Improved !vote error messages",
			AssembleVersion(0, 8, 15, 1): "- Quickconfig actually sets silentrole now",
			AssembleVersion(0, 8, 15, 0): "- Use 64-bit integer conversion",
			AssembleVersion(0, 8, 14, 6): "- Allow adding birthdays on current day\n-Update avatar change function",
			AssembleVersion(0, 8, 14, 5): "- Allow exact string matching on !import",
			AssembleVersion(0, 8, 14, 4): "- Added !import\n- Added Importable option\n- Make !collections more useful",
			AssembleVersion(0, 8, 14, 3): "- Allow pinging multiple groups via group1+group2",
			AssembleVersion(0, 8, 14, 2): "- Fix !createpoll unique option key\n- Add !addoption",
			AssembleVersion(0, 8, 14, 1): "- Clean up !poll",
			AssembleVersion(0, 8, 14, 0): "- Added !poll, !vote, !createpoll, !deletepoll and !results commands",
			AssembleVersion(0, 8, 13, 1): "- Fixed !setconfig rules",
			AssembleVersion(0, 8, 13, 0): "- Added changelog\n- Added !rules command",
			AssembleVersion(0, 8, 12, 0): "- Added temporary silences",
			AssembleVersion(0, 8, 11, 5): "- Added \"dumbass\" to Sweetie Bot's vocabulary",
			AssembleVersion(0, 8, 11, 4): "- Display channels in help for commands",
			AssembleVersion(0, 8, 11, 3): "- Make defaultserver an independent command",
			AssembleVersion(0, 8, 11, 2): "- Add !defaultserver command",
			AssembleVersion(0, 8, 11, 1): "- Fix !autosilence behavior",
			AssembleVersion(0, 8, 11, 0): "- Replace mentions in !search\n- Add temporary ban to !ban command",
			AssembleVersion(0, 8, 10, 0): "- !ping now accepts newlines\n- Added build version to make moonwolf happy",
			AssembleVersion(0, 8, 9, 0):  "- Add silence message for Tawmy\n- Make silence message ping user\n- Fix #27 (Sweetie Bot explodes if you search nothing)\n- Make !lastseen more reliable",
			AssembleVersion(0, 8, 8, 0):  "- Log all commands sent to SB in DB-enabled servers",
			AssembleVersion(0, 8, 7, 0):  "- Default to main server for PMs if it exists\n- Restrict PM commands to the server you belong in (fix #26)\n- Make spam deletion lookback configurable\n- Make !quickconfig complain if permissions are wrong\n- Add giant warning label for Tawmy\n- Prevent parse time crash\n- Make readme more clear on how things work\n- Sort !listguild by user count\n- Fallback to search all users if SB can't find one in the current server",
			AssembleVersion(0, 8, 6, 0):  "- Add full timezone support\n- Deal with discord's broken permissions\n- Improve timezone help messages",
			AssembleVersion(0, 8, 5, 0):  "- Add !userinfo\n- Fix #15 (Lock down !removeevent)\n- Fix guildmember query\n- Use nicknames in more places",
			AssembleVersion(0, 8, 4, 0):  "- Update readme, remove disablebored\n- Add delete command",
			AssembleVersion(0, 8, 3, 0):  "- Actually seed random number generator because Cloud is a FUCKING IDIOT\n- Allow newlines in commands\n- Bored module is now fully programmable\n- Display user ID in !aka\n- Hopefully stop sweetie from being an emo teenager\n- Add additional stupid proofing\n- Have bored commands override all restrictions",
			AssembleVersion(0, 8, 2, 0):  "- Enable multi-server message logging\n- Extend !searchquote\n- Attach !lastping to current server\n- Actually make aliases work with commands",
			AssembleVersion(0, 8, 1, 0):  "- Add dynamic collections\n- Add quotes\n- Prevent !aka command from spawning evil twins\n- Add !removealias\n- Use nicknames where possible\n- Fix off by one error\n- Sanitize !search output ",
			AssembleVersion(0, 8, 0, 0):  "- Appease the dark gods of discord's API\n- Allow sweetiebot to track nicknames\n- update help\n- Include nickname in searches",
		},
	}

	if debugerr == nil && len(debugchannels) > 0 {
		json.Unmarshal(debugchannels, sb)
	}
	dbguilds, err := ioutil.ReadFile("db.guilds")
	if err == nil && len(dbguilds) > 0 {
		json.Unmarshal(dbguilds, sb)
	}
	sb.DBGuilds[sb.MainGuildID] = true

	rand.Intn(10)
	for i := 0; i < 20+rand.Intn(20); i++ {
		rand.Intn(50)
	}

	db, err := DB_Load(&emptyLog{}, "mysql", strings.TrimSpace(string(dbauth)))
	sb.db = db
	if !db.status.get() {
		fmt.Println("Database connection failure - running in No Database mode: ", err.Error())
	} else {
		err = sb.db.LoadStatements()
		if err == nil {
			fmt.Println("Finished loading database statements")
		} else {
			fmt.Println("Loading database statements failed: ", err)
			fmt.Println("DATABASE IS BADLY FORMATTED OR CORRUPT - TERMINATING SWEETIE BOT!")
			return nil
		}
	}

	isuser, _ := ioutil.ReadFile("isuser") // DO NOT CREATE THIS FILE UNLESS YOU KNOW *EXACTLY* WHAT YOU ARE DOING. This is for crazy people who want to run sweetiebot in user mode. If you don't know what user mode is, you don't want it. If you create this file anyway and the bot breaks, it's your own fault.
	if isuser == nil {
		sb.dg, err = discordgo.New("Bot " + token)
	} else {
		sb.dg, err = discordgo.New(token)
		fmt.Println("Started SweetieBot on a user account.")
	}
	if err != nil {
		fmt.Println("Error creating discord session", err.Error())
		return nil
	}
	sb.dg.LogLevel = discordgo.LogWarning

	sb.dg.AddHandler(sbReady)
	sb.dg.AddHandler(sbMessageCreate)
	sb.dg.AddHandler(sbMessageUpdate)
	sb.dg.AddHandler(sbMessageDelete)
	sb.dg.AddHandler(sbUserUpdate)
	sb.dg.AddHandler(sbPresenceUpdate)
	sb.dg.AddHandler(sbGuildUpdate)
	sb.dg.AddHandler(sbGuildMemberAdd)
	sb.dg.AddHandler(sbGuildMemberRemove)
	sb.dg.AddHandler(sbGuildMemberUpdate)
	sb.dg.AddHandler(sbGuildBanAdd)
	sb.dg.AddHandler(sbGuildBanRemove)
	sb.dg.AddHandler(sbGuildRoleDelete)
	sb.dg.AddHandler(sbGuildCreate)
	sb.dg.AddHandler(sbChannelCreate)
	sb.dg.AddHandler(sbChannelDelete)

	if sb.Debug { // The server does not necessarily tie a standard input to the program
		go func() {
			var input string
			fmt.Scanln(&input)
			sb.quit.set(true)
		}()
	}

	go idleCheckLoop()
	go deadlockDetector()

	//BuildMarkov(1, 1)
	return sb
}

// Connect opens a websocket connection to discord. Only returns after disconnecting.
func (sbot *SweetieBot) Connect() {
	err := sbot.dg.Open()
	if err == nil {
		fmt.Println("Connection established")
		for !sbot.quit.get() {
			time.Sleep(400 * time.Millisecond)
		}
	} else {
		fmt.Println("Error opening websocket connection: ", err.Error())
	}

	fmt.Println("Sweetiebot quitting")
	sbot.dg.Close()
	sbot.db.Close()
}

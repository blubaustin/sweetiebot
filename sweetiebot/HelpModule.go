package sweetiebot

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/blackhole12/discordgo"
)

// HelpModule contains help and about commands
type HelpModule struct {
}

// Name of the module
func (w *HelpModule) Name() string {
	return "Help/About"
}

// Commands in the module
func (w *HelpModule) Commands() []Command {
	return []Command{
		&helpCommand{},
		&aboutCommand{},
		&rulesCommand{},
		&changelogCommand{},
	}
}

// Description of the module
func (w *HelpModule) Description() string {
	return "Contains commands for getting information about Sweetie Bot, her commands, or the server she is in."
}

type helpCommand struct {
}

func (c *helpCommand) Name() string {
	return "Help"
}

// DumpCommandsModules dumps information about all commands and modules
func DumpCommandsModules(channelID string, info *GuildInfo, footer string, description string) *discordgo.MessageEmbed {
	fields := make([]*discordgo.MessageEmbedField, 0, len(info.modules))
	for _, v := range info.modules {
		cmds := v.Commands()
		if len(cmds) > 0 {
			s := make([]string, 0, len(cmds))
			for _, c := range cmds {
				s = append(s, c.Name()+info.IsCommandDisabled(c.Name()))
			}
			fields = append(fields, &discordgo.MessageEmbedField{Name: v.Name() + info.IsModuleDisabled(v.Name()), Value: strings.Join(s, "\n"), Inline: true})
		} else {
			fields = append(fields, &discordgo.MessageEmbedField{Name: v.Name() + info.IsModuleDisabled(v.Name()), Value: "*[no commands]*", Inline: true})
		}
	}
	return &discordgo.MessageEmbed{
		Type: "rich",
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://github.com/blackhole12/sweetiebot",
			Name:    "Sweetie Bot Commands",
			IconURL: fmt.Sprintf("https://cdn.discordapp.com/avatars/%v/%s.jpg", sb.SelfID, sb.SelfAvatar),
		},
		Description: description,
		Color:       0x3e92e5,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footer,
		},
	}
}

func (c *helpCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if len(args) == 0 {
		return "", true, DumpCommandsModules(msg.ChannelID, info, "For more information on a specific command, type !help [command].", "")
	}
	arg := strings.ToLower(args[0])
	for _, v := range info.modules {
		if strings.Compare(strings.ToLower(v.Name()), arg) == 0 {
			cmds := v.Commands()
			fields := make([]*discordgo.MessageEmbedField, 0, len(cmds))
			for _, c := range cmds {
				fields = append(fields, &discordgo.MessageEmbedField{Name: c.Name() + info.IsCommandDisabled(c.Name()), Value: c.UsageShort(), Inline: false})
			}
			color := 0x56d34f
			if len(info.IsModuleDisabled(v.Name())) > 0 {
				color = 0xd54141
			}

			embed := &discordgo.MessageEmbed{
				Type: "rich",
				Author: &discordgo.MessageEmbedAuthor{
					URL:     "https://github.com/blackhole12/sweetiebot#modules",
					Name:    v.Name() + " Module Command List" + info.IsModuleDisabled(v.Name()),
					IconURL: fmt.Sprintf("https://cdn.discordapp.com/avatars/%v/%s.jpg", sb.SelfID, sb.SelfAvatar),
				},
				Color:       color,
				Description: v.Description(),
				Fields:      fields,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "For more information on a specific command, type !help [command].",
				},
			}
			return "", true, embed
		}
	}
	v, ok := info.commands[strings.ToLower(args[0])]
	if !ok {
		return "```Sweetie Bot doesn't recognize that command or module. You can check what commands Sweetie Bot knows by typing !help with no arguments.```", false, nil
	}
	return "", true, info.FormatUsage(v, v.Usage(info))
}
func (c *helpCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "Lists all available commands Sweetie Bot knows, or gives information about the given command. Of course, you should have figured this out by now, since you just typed !help help for some reason.",
		Params: []CommandUsageParam{
			{Name: "command/module", Desc: "The command or module to display help for. You do not need to include a command's parent module, just the command name itself.", Optional: true},
		},
	}
}
func (c *helpCommand) UsageShort() string {
	return "[PM Only] Generates the list you are looking at right now."
}

type aboutCommand struct {
}

func (c *aboutCommand) Name() string {
	return "About"
}
func (c *aboutCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	tag := " [release]"
	if sb.Debug {
		tag = " [debug]"
	}
	owners := make([]string, 0, len(sb.Owners))
	for k := range sb.Owners {
		owners = append(owners, SBitoa(k))
	}
	embed := &discordgo.MessageEmbed{
		Type: "rich",
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://github.com/blackhole12/sweetiebot",
			Name:    "Sweetie Bot v" + sb.version.String() + tag,
			IconURL: fmt.Sprintf("https://cdn.discordapp.com/avatars/%v/%s.png", sb.SelfID, sb.SelfAvatar),
		},
		Color: 0x3e92e5,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Author", Value: "Blackhole#8270", Inline: true},
			{Name: "Library", Value: "discordgo", Inline: true},
			{Name: "Owner ID(s)", Value: strings.Join(owners, ", "), Inline: true},
			{Name: "Presence", Value: Pluralize(int64(len(sb.guilds)), " server"), Inline: true},
			{Name: "Uptime", Value: TimeDiff(time.Duration(time.Now().UTC().Unix()-sb.StartTime) * time.Second), Inline: true},
			{Name: "Messages Seen", Value: strconv.FormatUint(uint64(atomic.LoadUint32(&sb.MessageCount)), 10), Inline: true},
			{Name: "Github", Value: "https://github.com/blackhole12/sweetiebot", Inline: false},
			{Name: "Patreon", Value: "https://www.patreon.com/erikmcclure", Inline: false},
			{Name: "Add Sweetie Bot To Your Server", Value: "https://goo.gl/NQtUZv", Inline: false},
			{Name: "Terms of Service", Value: "By joining a server using this bot or adding this bot to your server, you give express permission for the bot to collect and store any information it deems necessary to perform its functions, including but not limited to, message content, message metadata, and user metadata.", Inline: false},
		},
	}
	return "", false, embed
}
func (c *aboutCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "Displays information about Sweetie Bot. What, did you think it would do something else?",
	}
}
func (c *aboutCommand) UsageShort() string { return "Displays information about Sweetie Bot." }

type rulesCommand struct {
}

func (c *rulesCommand) Name() string {
	return "Rules"
}
func (c *rulesCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if len(info.config.Help.Rules) == 0 {
		return "```I don't know what the rules are in this server... ¯\\_(ツ)_/¯```", false, nil
	}
	if len(args) < 1 {
		rules := make([]string, 0, len(info.config.Help.Rules)+1)
		rules = append(rules, "Official rules of "+info.Name+":")
		keys := MapIntToSlice(info.config.Help.Rules)
		sort.Ints(keys)

		for _, v := range keys {
			if !info.config.Help.HideNegativeRules || v >= 0 {
				rules = append(rules, fmt.Sprintf("%v. %s", v, info.config.Help.Rules[v]))
			}
		}
		return strings.Join(rules, "\n"), len(rules) > 4, nil
	}

	arg, err := strconv.Atoi(args[0])
	if err != nil {
		return "```Rule index must be a number!```", false, nil
	}
	rule, ok := info.config.Help.Rules[arg]
	if !ok {
		return "```That's not a rule! Stop making things up!```", false, nil
	}
	return fmt.Sprintf("%v. %s", arg, rule), false, nil
}
func (c *rulesCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "Lists all the rules in this server, or displays the specific rule requested, if it exists. Rules can be set using `" + info.config.Basic.CommandPrefix + "setconfig rules 1 this is a rule`",
		Params: []CommandUsageParam{
			{Name: "index", Desc: "Index of the rule to display. If omitted, displays all rules.", Optional: true},
		},
	}
}
func (c *rulesCommand) UsageShort() string { return "Lists the rules of the server." }

type changelogCommand struct {
}

func (c *changelogCommand) Name() string {
	return "Changelog"
}
func (c *changelogCommand) Process(args []string, msg *discordgo.Message, indices []int, info *GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	v := Version{0, 0, 0, 0}
	if len(args) == 0 {
		versions := make([]string, 0, len(sb.changelog)+1)
		versions = append(versions, "All versions of Sweetie Bot with a changelog:")
		keys := MapIntToSlice(sb.changelog)
		sort.Ints(keys)
		for i := len(keys) - 1; i >= 0; i-- {
			k := keys[i]
			version := Version{byte(k >> 24), byte((k >> 16) & 0xFF), byte((k >> 8) & 0xFF), byte(k & 0xFF)}
			versions = append(versions, version.String())
		}
		return "```\n" + strings.Join(versions, "\n") + "```", len(versions) > 6, nil
	}
	if strings.ToLower(args[0]) == "current" {
		v = sb.version
	} else {
		s := strings.Split(args[0], ".")
		if len(s) > 0 {
			i, _ := strconv.Atoi(s[0])
			v.major = byte(i)
		}
		if len(s) > 1 {
			i, _ := strconv.Atoi(s[1])
			v.minor = byte(i)
		}
		if len(s) > 2 {
			i, _ := strconv.Atoi(s[2])
			v.revision = byte(i)
		}
		if len(s) > 3 {
			i, _ := strconv.Atoi(s[3])
			v.build = byte(i)
		}
	}
	log, ok := sb.changelog[v.Integer()]
	if !ok {
		return "```That's not a valid version of Sweetie Bot! Use this command with no arguments to list all valid versions, or use \"current\" to get the most recent changelog.```", false, nil
	}
	return fmt.Sprintf("```\n%s\n--------\n%s```", v.String(), log), false, nil
}
func (c *changelogCommand) Usage(info *GuildInfo) *CommandUsage {
	return &CommandUsage{
		Desc: "Displays the given changelog for Sweetie Bot. If no version is given, lists all versions with a changelog. ",
		Params: []CommandUsageParam{
			{Name: "version", Desc: "A version in the format 1.2.3.4. Use \"current\" for the most recent version.", Optional: true},
		},
	}
}
func (c *changelogCommand) UsageShort() string { return "Retrieves the changelog for Sweetie Bot." }

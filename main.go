package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"./support"
	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

//v0.0.004 -- 4-24-2020-3-20-AM
const MaxDBItems = 128

var DS *discordgo.Session
var Guilds [MaxDBItems + 1]*discordgo.Guild
var Guilds_max int

var Guild_IDs [MaxDBItems + 1]string
var Guild_IDs_max int

var Channel_IDs [MaxDBItems + 1]string
var Channel_IDs_max int

var UserRoles [MaxDBItems + 1]string
var UserRoles_max int

var Message_IDs [MaxDBItems + 1]string
var Message_IDs_max int

var SleepSecs [MaxDBItems + 1]string
var SleepSecs_max int

var PendingRoles int = 0
var CloseProgram bool = false

func buildguilds() {
	Guilds_max = 0
	for x := 0; x < Guild_IDs_max; x++ {
		nguild, err := DS.Guild(Guild_IDs[x])
		Guilds[x] = nguild

		if err != nil {
			support.Log("Unable to get Guild from GuildID.")
			os.Exit(1)
		}
		Guilds_max++
		support.Log("Guild: " + nguild.Name)
	}
}

func checkdata() {
	ref := Guild_IDs_max
	if Guilds_max != ref || Channel_IDs_max != ref || UserRoles_max != ref || Message_IDs_max != ref || SleepSecs_max != ref {
		support.Log("Uneven number of guilds/channels/roles/message ids/sleep secs!")
		os.Exit(1)
	}

	if Guild_IDs_max >= MaxDBItems {
		support.Log("Max DB items exceeded.")
		os.Exit(1)
	}
}

func loadconfig() {
	support.Config.LoadEnv()

	//Seperate guilds
	gids := strings.Split(string(support.Config.GuildID), ",")
	Guild_IDs_max = len(gids)

	for i := 0; i < Guild_IDs_max && i < MaxDBItems; i++ {
		Guild_IDs[i] = string(gids[i])
		//support.Log(string("gid: " + gids[i]))
	}

	//Seperate channels
	cids := strings.Split(string(support.Config.ChannelID), ",")
	Channel_IDs_max = len(cids)

	for i := 0; i < Channel_IDs_max && i < MaxDBItems; i++ {
		Channel_IDs[i] = string(cids[i])
		//support.Log(string("cid: " + cids[i]))
	}

	//Seperate roles
	rids := strings.Split(string(support.Config.UserRole), ",")
	UserRoles_max = len(rids)

	for i := 0; i < UserRoles_max && i < MaxDBItems; i++ {
		UserRoles[i] = string(rids[i])
		//support.Log(string("rid: " + rids[i]))
	}

	//Seperate messageIDs
	mids := strings.Split(string(support.Config.MessageID), ",")
	Message_IDs_max = len(mids)

	for i := 0; i < Message_IDs_max && i < MaxDBItems; i++ {
		Message_IDs[i] = string(mids[i])
		//support.Log(string("mid: " + mids[i]))
	}

	//Seperate SleepSecs
	ss := strings.Split(string(support.Config.SleepSec), ",")
	SleepSecs_max = len(ss)

	for i := 0; i < SleepSecs_max && i < MaxDBItems; i++ {
		SleepSecs[i] = string(ss[i])
		//support.Log(string("ss: " + ss[i]))
	}
}

func main() {

	loadconfig()
	startbot()
	buildguilds()
	checkdata()

	//Check signal files
	go func() {
		for {
			time.Sleep(1 * time.Second)

			if _, err := os.Stat(".reload"); !os.IsNotExist(err) {
				support.Log("Reloading config...")

				loadconfig()
				buildguilds()
				checkdata()

				support.Log("Config reloaded.")

				if err := os.Remove(".reload"); err != nil {
					support.Log(".reload disappeared?")
				}
			}

			if _, err := os.Stat(".reboot"); !os.IsNotExist(err) {
				if err := os.Remove(".reboot"); err != nil {
					support.Log(".reload disappeared?")
				}
				//Perform timed role add/del here before closing.
				CloseProgram = true
				for x := 0; x < 30 && PendingRoles > 0; x++ {
					support.Log("Waiting for pending roles...")
					time.Sleep(time.Second)
				}
				support.Log("Rebooting due to .reboot file")
				DS.Close()
				os.Exit(1)
			}
		}
	}()

	quitHandle()

}

func startbot() {

	bot, err := discordgo.New("Bot " + support.Config.DiscordToken)
	if err != nil {
		support.Log("Couldn't start bot")
		os.Exit(1)
		return
	}

	err = bot.Open()

	if err != nil {
		support.Log("Couldn't open bot.")
		os.Exit(1)
		return
	}

	DS = bot
	bot.AddHandler(reactionAdd)
	bot.AddHandler(reactionRemove)
}

func quitHandle() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	CloseProgram = true
	//Perform timed role add/del here before closing.
	for x := 0; x < 30 && PendingRoles > 0; x++ {
		support.Log("Waiting for pending roles...")
		time.Sleep(time.Second)
	}
	DS.Close()
	support.Log("Goodbye.")
	os.Exit(1)
}

func reactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	buf := ""

	buf = fmt.Sprintf("ADDED REACTION: user %v, message: %v, emoji: %v, channel: %v, guild: %v",
		m.MessageReaction.UserID,
		m.MessageReaction.MessageID,
		m.MessageReaction.Emoji,
		m.MessageReaction.ChannelID,
		m.MessageReaction.GuildID)
	support.Log(buf)

	for x := 0; x < Guild_IDs_max; x++ {
		if m.MessageReaction.GuildID == Guild_IDs[x] && m.ChannelID == Channel_IDs[x] && m.MessageID == Message_IDs[x] {
			guild := Guilds[x]
			gid := Guild_IDs[x]
			role := UserRoles[x]
			uid := m.MessageReaction.UserID
			sleep, _ := strconv.Atoi(SleepSecs[x])
			go func() {
				if sleep > 0 {
					PendingRoles++
					for x := 0; x < sleep && CloseProgram == false; x++ {
						time.Sleep(time.Second)
					}
				}
				addRole(guild, gid, uid, role)
				support.Log(fmt.Sprintf("%v added to %v role.", uid, role))
				PendingRoles--
			}()
		}
	}
}

func reactionRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	buf := ""

	buf = fmt.Sprintf("REMOVED REACTION: user %v, message: %v, emoji: %v, channel: %v, guild: %v",
		m.MessageReaction.UserID,
		m.MessageReaction.MessageID,
		m.MessageReaction.Emoji,
		m.MessageReaction.ChannelID,
		m.MessageReaction.GuildID)
	support.Log(buf)

	//Only remove role, if a non-delayed role
	for x := 0; x < Guild_IDs_max; x++ {
		if m.MessageReaction.GuildID == Guild_IDs[x] && m.ChannelID == Channel_IDs[x] && m.MessageID == Message_IDs[x] && SleepSecs[x] == "0" {
			support.Log(fmt.Sprintf("%v removed from %v role.", m.MessageReaction.UserID, UserRoles[x]))
			removeRole(Guilds[x], Guild_IDs[x], m.MessageReaction.UserID, UserRoles[x])
		}
	}
}

func addRole(guild *discordgo.Guild, guildid string, uid string, role string) {
	err_role, regrole := roleExists(guild, role)
	if err_role {
		errset := DS.GuildMemberRoleAdd(guildid, uid, regrole.ID)
		if errset != nil {
			support.Log(fmt.Sprintf("Couldn't set role %v for %v", role, uid))
		}
	}
}

func removeRole(guild *discordgo.Guild, guildid string, uid string, role string) {
	err_role, regrole := roleExists(guild, role)
	if err_role {
		errset := DS.GuildMemberRoleRemove(guildid, uid, regrole.ID)
		if errset != nil {
			support.Log(fmt.Sprintf("Couldn't set role %v for %v", role, uid))
		}
	}
}

func roleExists(g *discordgo.Guild, name string) (bool, *discordgo.Role) {

	if g != nil && name != "" {
		name = strings.ToLower(name)

		for _, role := range g.Roles {
			if role.Name == "@everyone" {
				continue
			}

			if strings.ToLower(role.Name) == name {
				return true, role
			}

		}

	}
	return false, nil
}

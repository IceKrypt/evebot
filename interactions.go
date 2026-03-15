package main

import (
	"bytes"
	"fmt"
	"image/png"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix/v2"
)

func sinfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	iconImage, err := s.GuildIcon(i.GuildID)
	if err != nil {
		fmt.Println("sinfo GuildIcon:", err)
		return
	}
	var buff bytes.Buffer
	err = png.Encode(&buff, iconImage)
	if err != nil {
		fmt.Println("sinfo png.Encode:", err)
		return
	}
	guild, err := Guild(s, i.GuildID)
	if err != nil {
		fmt.Println("sinfo Guild:", err)
		return
	}
	fmt.Println("owner:", guild.OwnerID)
	owner, err := s.User(guild.OwnerID)
	if err != nil {
		fmt.Println("sinfo User:", err)
		return
	}
	voice, text := 0, 0
	for _, v := range guild.Channels {
		switch v.Type {
		case discordgo.ChannelTypeGuildText:
			text++
		case discordgo.ChannelTypeGuildVoice:
			voice++
		}
	}
	created, _ := discordgo.SnowflakeTimestamp(guild.ID)
	emojis := make([]string, 1)
	for _, v := range guild.Emojis {
		emoji := v.MessageFormat() + " "
		if len(emojis[len(emojis)-1]+emoji) > 1024 {
			emojis = append(emojis, emoji)
		} else {
			emojis[len(emojis)-1] += emoji
		}
	}
	mesg := &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: guild.Name,
			Color: embedColor,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "attachment://thumb.png",
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "ID",
					Value:  guild.ID,
					Inline: true,
				},
				{
					Name:   "Owner",
					Value:  fmt.Sprintf("%v#%v", owner.Username, owner.Discriminator),
					Inline: true,
				},
				{
					Name:   "Members",
					Value:  fmt.Sprint(guild.MemberCount),
					Inline: true,
				},
				{
					Name:   "Text channels",
					Value:  fmt.Sprint(text),
					Inline: true,
				},
				{
					Name:   "Voice channels",
					Value:  fmt.Sprint(voice),
					Inline: true,
				},
				{
					Name:   "Created at",
					Value:  created.Format("January 2, 2006"),
					Inline: true,
				},
				{
					Name:   "Region",
					Value:  guild.Region,
					Inline: true,
				},
				{
					Name:   "Roles",
					Value:  fmt.Sprint(len(guild.Roles)),
					Inline: true,
				},
			},
		},
		Files: []*discordgo.File{
			{
				Name:   "thumb.png",
				Reader: &buff,
			},
		},
	}
	for i, v := range emojis {
		if i == 0 {
			mesg.Embed.Fields = append(mesg.Embed.Fields, &discordgo.MessageEmbedField{Name: "Custom emojis", Value: v, Inline: true})
		} else {
			mesg.Embed.Fields = append(mesg.Embed.Fields, &discordgo.MessageEmbedField{Name: "\u200b", Value: v, Inline: true})
		}
	}
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{mesg.Embed},
		Files:  mesg.Files,
	})
	if err != nil {
		fmt.Println("interaction error sinfo:", err)
	}
}

func minfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	user := i.Member.User
	if len(i.ApplicationCommandData().Options) > 0 {
		user = i.ApplicationCommandData().Options[0].UserValue(s)
	}
	embed, err := uinfo(user, i.GuildID, s)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error: %v", err),
			},
		})
		return
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

func (eb *EveBot) dbHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	mesg := "```\n"
	switch i.ApplicationCommandData().Options[0].Name {
	case "traffic":
		joined := eb.repo.GetAllJoin()
		leave := eb.repo.GetAllLeave()
		dates := make([]string, 0)
		for k := range joined {
			dates = append(dates, k)
		}
		sort.Strings(dates)
		for _, date := range dates {
			mesg += fmt.Sprintf("%v: +/- %v/%v\n", date, joined[date], leave[date])
		}
	case "muted":
		muted := eb.repo.GetAllMuted()
		mesg += fmt.Sprintf("Total muted members: %v\n", len(muted))
		for id, until := range muted {
			mesg += fmt.Sprintf("<@%v> %v: %v\n", id, id, until)
		}
	}
	mesg += "```\n"
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: mesg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (eb *EveBot) unmuteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.ApplicationCommandData().Options[0].UserValue(s)
	err := eb.doUnmute(s, user.ID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error: %v", err),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%v was unmuted", user.Mention()),
		},
	})
}

func (eb *EveBot) doUnmute(s *discordgo.Session, userID string) error {
	err := s.GuildMemberRoleRemove(guildID, userID, muteRole)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}
	eb.repo.DeleteMuted(userID)
	return nil
}

func (eb *EveBot) muteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.ApplicationCommandData().Options[0].UserValue(s)
	dur, err := time.ParseDuration(i.ApplicationCommandData().Options[1].StringValue())
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Invalid duration: %v", err),
			},
		})
	}

	err = eb.doMute(s, user.ID, dur)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error: %v", err),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%v was muted (%v)", user.Mention(), dur),
		},
	})
}

// TODO: refactor command handling in general. logic function sshould return errors. Maybe differentiate user vs internal errors for logging?

func (eb *EveBot) doMute(s *discordgo.Session, userID string, dur time.Duration) error {
	eb.muteMember(s, userID, dur)
	return nil
}

func (eb *EveBot) twitchHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Options[0].Name {
	case "link":
		username := i.ApplicationCommandData().Options[0].Options[0].StringValue()
		// Get Twitch ID
		res, err := eb.twitch.GetUsers(&helix.UsersParams{Logins: []string{username}})
		if err != nil || len(res.Data.Users) == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error finding Twitch user %s: %v", username, err),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		user := res.Data.Users[0]
		eb.repo.AddTwitch(i.Member.User.ID, user.ID, user.Login)

		// Check if member has streamer role
		member, err := GuildMember(s, i.GuildID, i.Member.User.ID)
		if err == nil {
			isStreamer := false
			for _, role := range member.Roles {
				if role == streamerRole {
					isStreamer = true
					break
				}
			}
			if isStreamer {
				eb.subscribeToStreamer(i.Member.User.ID, i.Member.User.Username)
			}
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Linked Twitch account: %s (ID: %s). I'll notify when you go live!", user.Login, user.ID),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

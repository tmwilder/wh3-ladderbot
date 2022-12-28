package interactions

import (
	"discordbot/internal/app/discord/commands"
	"discordbot/internal/app/ratings"
	"discordbot/internal/db"
	"fmt"
	"gorm.io/gorm"
)

func Report(conn *gorm.DB, interaction Interaction) (success bool, channelMessage string) {
	outcome := commands.ReportOutcome(interaction.Data.Options[0].Value)

	switch outcome {
	case commands.Win:
		return handlePlayedMatch(conn, interaction, true)
	case commands.Loss:
		return handlePlayedMatch(conn, interaction, false)
	case commands.Cancel:
		return handleCancel(conn, interaction)
	default:
		return false, "Unrecognized match report option."
	}
}

func handlePlayedMatch(conn *gorm.DB, interaction Interaction, isWin bool) (success bool, channelMessage string) {
	foundUser, user := db.GetUserByDiscordId(conn, interaction.Member.User.Id)

	if !foundUser {
		return false, "Unable to find your user. Contact admin for help."
	}

	foundMatch, mostRecentMatch := db.GetMostRecentMatch(conn, user.UserId)

	if !foundMatch {
		return false, "You do not currently have a most recent match. To report a win you must first queue up and get paired."
	}

	interactionUserIsP1 := mostRecentMatch.P1UserId == user.UserId

	var p1Won bool

	if (interactionUserIsP1 && isWin) || (!interactionUserIsP1 && !isWin) {
		p1Won = true
	} else {
		p1Won = false
	}

	player1UserId := mostRecentMatch.P1UserId
	player2UserId := mostRecentMatch.P2UserId

	_, p1User := db.GetUserById(conn, player1UserId)
	_, p2User := db.GetUserById(conn, player2UserId)

	switch mostRecentMatch.MatchState {
	case db.Matched:
		// Get current ratings, compute new ratings, update ratings for both players, then update match state to complete.
		return recordMatchWinner(conn, p1User, p2User, mostRecentMatch, p1Won)
	case db.Completed:
		// Get last rating, recompute ratings, tombstone old rating entry, add new rating entry, update both player ratings
		// Problem - if the other player has played a match since - make sure their rating is correct or do something reasonable
		_, mostRecentMatchP2 := db.GetMostRecentMatch(conn, player2UserId)
		if mostRecentMatchP2.MatchId != mostRecentMatch.MatchId {
			return false, "Your last match was reported and your opponent already logged their next match, which means we cannot update scores. Next time if you need to make a change to a match result you'll need to work with your opponent to do that before either of you play again."
		}
		db.RevertUserRating(conn, player1UserId)
		db.RevertUserRating(conn, player2UserId)
		_, p1User := db.GetUserById(conn, player1UserId)
		_, p2User := db.GetUserById(conn, player2UserId)
		return recordMatchWinner(conn, p1User, p2User, mostRecentMatch, p1Won)
	case db.Cancelled:
		// Get current ratings, compute new ratings, update ratings for both players, then update match state to complete.
		return recordMatchWinner(conn, p1User, p2User, mostRecentMatch, p1Won)
	default:
		return false, fmt.Sprintf("Unknown prior match state %s contact admins for help.", mostRecentMatch.MatchState)
	}
}

func recordMatchWinner(conn *gorm.DB, p1User db.User, p2User db.User, mostRecentMatch db.Match, p1Won bool) (success bool, channnelMessage string) {
	// Get current ratings, compute new ratings, update ratings for both players, then update match state to complete.
	newP1Rating, newP2Rating := ratings.ComputeNewElos(p1User.CurrentRating, p2User.CurrentRating, p1Won)

	var winnerValue db.WhoWon
	var winnerName string

	if p1Won {
		winnerName = p1User.DiscordUserName
		winnerValue = db.P1
	} else {
		winnerName = p2User.DiscordUserName
		winnerValue = db.P2
	}

	db.UpdateUserRating(conn, p1User.UserId, newP1Rating, mostRecentMatch.MatchId)
	db.UpdateUserRating(conn, p2User.UserId, newP2Rating, mostRecentMatch.MatchId)
	db.UpdateMatch(conn, mostRecentMatch.MatchId, db.Completed, winnerValue)
	// TODO messaging improvements - maybe some nice art and a conditional congrats based on elo buckets.
	return true, fmt.Sprintf(
		"Win for %s recorded. Updated %s to rating %d and %s to rating %d.",
		winnerName, p1User.DiscordUserName, newP1Rating, p2User.DiscordUserName, newP2Rating)
}

func handleCancel(conn *gorm.DB, interaction Interaction) (success bool, channelMessage string) {
	foundUser, user := db.GetUserByDiscordId(conn, interaction.Member.User.Id)

	if !foundUser {
		return false, "Unable to find your user. Contact admin for help."
	}

	foundMatch, mostRecentMatch := db.GetMostRecentMatch(conn, user.UserId)

	if !foundMatch {
		return false, "You do not currently have a most recent match. To report a win you must first queue up and get paired."
	}

	player1UserId := mostRecentMatch.P1UserId
	player2UserId := mostRecentMatch.P2UserId

	_, p1User := db.GetUserById(conn, player1UserId)
	_, p2User := db.GetUserById(conn, player2UserId)

	_, mostRecentMatchP2 := db.GetMostRecentMatch(conn, player2UserId)
	if mostRecentMatchP2.MatchId != mostRecentMatch.MatchId {
		return false, "Your last match was reported and your opponent already logged their next match, which means we cannot cancel scores. Next time if you need to make a change to a match result you'll need to work with your opponent to do that before either of you play again."
	}

	switch mostRecentMatch.MatchState {
	case db.Cancelled:
		return true, "The most recent match was already cancelled so nothing to do! Feel free to requeue."
	case db.Matched:
		db.UpdateMatch(conn, mostRecentMatch.MatchId, db.Cancelled, db.Undefined)
		return true, fmt.Sprintf("Match between %s and %s cancelled by %s. No ratings changes will occur, feel free to requeue when convienient.", p1User.DiscordUserName, p2User.DiscordUserName, user.DiscordUserName)
	case db.Completed:
		db.RevertUserRating(conn, player1UserId)
		db.RevertUserRating(conn, player2UserId)
		db.UpdateMatch(conn, mostRecentMatch.MatchId, db.Cancelled, db.Undefined)

		_, p1User = db.GetUserById(conn, player1UserId)
		_, p2User = db.GetUserById(conn, player2UserId)

		return true, fmt.Sprintf("%s cancelled the most recent match between %s and %s. Reverted %s to rating %d and %s to rating %d and deleted the match.", user.DiscordUserName, p1User.DiscordUserName, p2User.DiscordUserName, p1User.DiscordUserName, p1User.CurrentRating, p2User.DiscordUserName, p2User.CurrentRating)
	default:
		return false, fmt.Sprintf("Unknown prior match state %s contact admins for help.", mostRecentMatch.MatchState)
	}
}

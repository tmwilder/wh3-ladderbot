package interactions

import "discordbot/internal/app/discord/commands"

func Report(interaction Interaction) (success bool, channelMessage string) {
	outcome := commands.ReportOutcome(interaction.Data.Options[0].Value)

	switch outcome {
	case commands.Win:
	case commands.Loss:
	case commands.Cancel:
		break
	default:
		return false, "Unrecognized match report option."
	}
	return true, ""
}

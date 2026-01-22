package xcom

type Error struct {
	TypeName string
	Reason   string
}

func (e *Error) Error() string {
	switch e.TypeName {
	case "TweetTombstone", "TweetWithVisibilityResults":
		return "Age-restricted adult content."
	case "TweetUnavailable":
		switch e.Reason {
		case "Suspended":
			return "Tweet Unavailable. Account Suspended."
		case "Protected":
			return "Tweet Unavailable. Account Protected."
		case "":
			return "Tweet Unavailable."
		}
	}
	return e.TypeName + " " + e.Reason
}

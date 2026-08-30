package game

type ActiveGamePlayer struct {
	ID         string
	ObjectName string
}

type ActiveGameResult struct {
	Win  ActiveGamePlayer
	Lose ActiveGamePlayer
	Verb string
}

func GetResult(p1, p2 ActiveGamePlayer) string {
	return ""
}

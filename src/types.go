package main

type server struct {
	gameSavePath   string
	gamePathExists bool
	profilesDir    string
}

type saveInfo struct {
	Name       string `json:"name"`
	Modified   string `json:"modified"`
	Type       string `json:"type"`
	Screenshot string `json:"screenshot"`
	Playtime   string `json:"playtime"`
	Level      string `json:"level"`
	Quest      string `json:"quest"`
	QuestTitle string `json:"questTitle"`
	Objective  string `json:"objective"`
}

type metaSummary struct {
	Playtime   string
	Level      string
	Quest      string
	QuestTitle string
	Objective  string
}

type profileNote struct {
	Profile string `json:"profile"`
	Note    string `json:"note"`
}

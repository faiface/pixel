package shared



//struct that holds all data about the current game state
type gameState struct {
	//is the game to keep on playing ?
	Playing  bool
	Level    int //todo use this values as game state
	NbAction int
	Score    int
}

var gState gameState

func StartGame(NewLevel int ){
	gState.Playing = true
	gState.Score = 0
	gState.Level = NewLevel
	gState.NbAction = 0
}
func GetCurrentLevel() int{
	return gState.Level
}
func IncrementLevel() int{
	//todo increment level value here
	return 0
}
func StopGame(){
	gState.Playing = false
}
func AddAction()  {
	gState.NbAction +=1
	//log.Print("Actions ",gState.NbAction)
}
func Continue() bool  {
	return gState.Playing
}
package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/cluster"
	"github.com/rivo/tview"
)

func writeText(text string, border bool, title string) *tview.TextView {
	textView := tview.NewTextView()
	textView.SetText(text)
	textView.SetBorder(border)
	textView.SetTitle(title)

	return textView
}

func main() {
	var cStore internal.ClusterStore
	var err error

	cStore, err = cluster.NewCluster("testdata/cluster.json")
	if err != nil {
		log.Println(err)
	}

	app := tview.NewApplication()

	serverlist, _ := cStore.GetServerInfo()
	var playerList []string
	var serverList []string
	for i := 0; i < len(serverlist); i++ {
		//	fmt.Println(serverlist[i].ServerInfo)
		serverList = append(serverList, serverlist[i].ServerInfo.Name+
			" "+
			strconv.Itoa(serverlist[i].ServerInfo.Players)+
			"/"+
			strconv.Itoa(serverlist[i].ServerInfo.MaxPlayers))
		for j := 0; j < len(serverlist[i].PlayersInfo.Players); j++ {
			//		fmt.Println(serverlist[i].PlayersInfo.Players[j].Name)
			playerList = append(playerList, serverlist[i].PlayersInfo.Players[j].Name)
		}
	}
	joinedServerList := strings.Join(serverList, "\n")
	joinedPlayerList := strings.Join(playerList, "\n")

	serverView := writeText(joinedServerList, true, "Server")
	textView := writeText(joinedPlayerList, true, "1")

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(serverView, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(textView, 0, 1, false).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("2"), 0, 1, false).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("3"), 0, 1, false).
				AddItem(tview.NewBox().SetBorder(true).SetTitle("4"), 0, 1, false), 0, 3, false), 0, 1, false)

	err = app.SetRoot(flex, true).SetFocus(flex).Run()
	if err != nil {
		panic(err)
	}

}

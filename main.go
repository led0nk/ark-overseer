package main

import (
	"fmt"
	"log"

	"github.com/led0nk/ark-clusterinfo/internal/database/cluster"

	"github.com/FlowingSPDG/go-steam"
)

func main() {

	TheIsland := cluster.Server{
		name: "TheIsland",
		addr: "51.195.60.114:27016",
	}
	Ragnarok := cluster.Server{
		name: "Ragnarok",
		addr: "51.195.60.114:27019",
	}

	server, err := steam.Connect(TheIsland.addr)
	if err != nil {
		log.Panic(err)
	}
	TheIsland.serverInfo, err = server.Info()
	if err != nil {
		log.Panic(err)
	}
	TheIsland.playerInfo, err = server.PlayersInfo()
	if err != nil {
		log.Println(err)
	}

	server, err = steam.Connect(Ragnarok.addr)
	if err != nil {
		log.Panic(err)
	}
	Ragnarok.serverInfo, err = server.Info()
	if err != nil {
		log.Panic(err)
	}
	Ragnarok.playerInfo, err = server.PlayersInfo()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(TheIsland.serverInfo)
	for i := 0; i < len(TheIsland.playerInfo.Players); i++ {
		fmt.Println(TheIsland.playerInfo.Players[i].Name)
	}
	fmt.Println(Ragnarok.serverInfo)
	for i := 0; i < len(Ragnarok.playerInfo.Players); i++ {
		fmt.Println(Ragnarok.playerInfo.Players[i].Name)
	}

}

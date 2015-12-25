package main

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/data/mmap"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/zhulik/margelet"
	"gopkg.in/alecthomas/kingpin.v2"
	"syscall"
)

// Not portable, only Linux and BSD. Windows sucks.
func checkDownloadPath(path string) {
	if syscall.Access(path, 2) != nil {
		panic(fmt.Errorf("%s is not writeable or not exists", path))
	}
}

func main() {
	token := kingpin.Flag("token", "Telegram Bot token").Required().Short('t').String()
	redisURL := kingpin.Flag("redis_url", "Redis url").Short('r').Default("127.0.0.1:6379").String()
	redisPassword := kingpin.Flag("redis_password", "Redis password").Default("").Short('p').String()
	redisDB := kingpin.Flag("redis_db", "Redis password").Default("0").Short('d').Int64()
	downloadPath := kingpin.Flag("path", "Download path").Required().Short('o').String()
	authorizedUsername := kingpin.Flag("username", "Username of user, thich can control bot").Required().Short('u').String()
	kingpin.Parse()

	checkDownloadPath(*downloadPath)

	bot, err := margelet.NewMargelet("torrent_bot", *redisURL, *redisPassword, *redisDB, *token, false)

	if err != nil {
		panic(err)
	}

	config := torrent.Config{
		TorrentDataOpener: func(info *metainfo.Info) torrent.Data {
			ret, _ := mmap.TorrentData(info, *downloadPath)
			return ret
		},
	}

	client, err := torrent.NewClient(&config)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	repo := newTorrentsRepository("torrent_bot_torrents", bot.GetRedis())

	torrentResponder, err := NewTorrentResponder(*authorizedUsername, client, repo)
	if err != nil {
		panic(err)
	}
	bot.AddMessageResponder(torrentResponder)
	bot.AddSessionHandler("/download", torrentResponder)
	bot.AddCommandHandler("/status", NewStatusHandler(*authorizedUsername, *downloadPath, client))
	//	bot.AddSessionHandler("/delete", NewDeleteHandler(*authorizedUsername, client))

	bot.Run()
}
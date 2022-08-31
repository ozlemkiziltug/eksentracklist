package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var conf *Config

type Config struct {
	ConsumerKey		string `json:"consumerKey"`
	ConsumerSecret	string `json:"consumerSecret"`
	AccessToken		string `json:"accessToken"`
	AccessSecret	string `json:"accessSecret"`
	DeveloperKey	string `json:"developerKey"`
}

func init() {
	conf = &Config{}
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}
	err = json.Unmarshal(data, conf)
	if err != nil {
		log.Fatalf("Error parsing configuration: %v", err)
	}
}

func readPlayingSongFile() string {
	data, err := ioutil.ReadFile("nowPlaying.txt")
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}
	return string(data)
}

func writePlayingSongFile(playingSong string) bool {
	err := ioutil.WriteFile("nowPlaying.txt", []byte(playingSong), 0644)

	if err != nil {
		log.Fatalf("Error writing configuration: %v", err)
	}
	return true
}

func getPlayingSong() string {
	resp, err := http.Get("https://radioeksen.com/umbraco/surface/Player/EksenPlayerSong")
	if err != nil {
		log.Fatalln(err)
	}
	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	caser := cases.Title(language.English)
	playingSong := data["NowPlayingArtist"].(string)
	casedPlayingSong := caser.String(playingSong)
	return strings.TrimSpace(casedPlayingSong)
}

func getYoutubeUrl(playingSong string) string{
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(conf.DeveloperKey))
	if err != nil {
		log.Printf("Error creating service: %v\n", err)
	}
	searchResult, err := youtubeService.Search.List([]string{"snippet"}).Q(playingSong).MaxResults(1).Type("video").Do()
	if err != nil {
		log.Printf("Error searching YouTube video: %v\n", err)
	}
	videoId := searchResult.Items[0].Id.VideoId
	youtubeUrl := "youtu.be/" + videoId
	return strings.TrimSpace(youtubeUrl)
}

func sendTweet(tweet string) {
	config := oauth1.NewConfig(conf.ConsumerKey, conf.ConsumerSecret)
	token := oauth1.NewToken(conf.AccessToken, conf.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	_, resp, err := client.Statuses.Update(tweet, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
}

func main() {
	playingSongFile := readPlayingSongFile()
	playingSong := getPlayingSong()
	if (playingSongFile == playingSong ){
		return
	}
	writePlayingSongFile(playingSong)
	videoUrl := getYoutubeUrl(playingSong)
	tweet := playingSong + " " + videoUrl
	sendTweet(tweet)
}

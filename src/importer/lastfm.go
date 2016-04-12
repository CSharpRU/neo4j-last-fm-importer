package importer

import (
	"github.com/CSharpRU/lastfm-go/lastfm"
	"log"
	"sync"
	"gopkg.in/jmcvetta/neoism.v1"
	"strings"
	"strconv"
)

type Job struct {
	Tag   string
	Url   string
	Track struct {
		      Rank       string `xml:"rank,attr"`
		      Name       string `xml:"name"`
		      Duration   string `xml:"duration"`
		      Mbid       string `xml:"mbid"`
		      Url        string `xml:"url"`
		      Streamable struct {
					 FullTrack  string `xml:"fulltrack,attr"`
					 Streamable string `xml:"streamable"`
				 } `xml:"streamable"`
		      Artist     struct {
					 Name string `xml:"name"`
					 Mbid string `xml:"mbid"`
					 Url  string `xml:"url"`
				 } `xml:"artist"`
		      Images     []struct {
			      Size string `xml:"size,attr"`
			      Url  string `xml:",chardata"`
		      } `xml:"image"`
	      } `xml:"track"`
}

var lastfmConnection *lastfm.Api
var workerJobs chan Job
var workerGroup sync.WaitGroup

func GetLastFmConnection() *lastfm.Api {
	if lastfmConnection == nil {
		lastfmConnection = lastfm.New(AppConfig.LastFm.Key, AppConfig.LastFm.Secret)

		return lastfmConnection
	}

	return lastfmConnection
}

func ImportData() {
	workerJobs = make(chan Job)

	startWorkers()

	defer close(workerJobs)

	for _, tag := range getTopTags().Tags {
		for page := 0; page < AppConfig.LastFm.Pages; page++ {
			log.Printf("Loading page: %d", page + 1)

			addTracksToChannel(tag.Name, tag.Url, getTopTracksByTag(tag.Name, page + 1))
		}
	}

	workerGroup.Wait()
}

func addTracksToChannel(tag string, url string, tracks lastfm.TagGetTopTracks) {
	for _, track := range tracks.Tracks {
		workerGroup.Add(1)

		workerJobs <- Job{
			Tag: tag,
			Url: url,
			Track: track,
		}
	}
}

func startWorkers() {
	for i := 0; i < AppConfig.LastFm.Workers; i++ {
		log.Printf("Starting LastFm scraper worker: %d", i + 1)

		go worker();
	}
}

func worker() {
	for {
		job := <-workerJobs

		if job.Track.Name == "" {
			continue;
		}

		track := getTrackByMbid(job.Track.Mbid)

		if track.Name == "" {
			continue;
		}

		trackNode := importTrack(track)
		artistNode := importArtist(track)

		if track.Album.Mbid != "" {
			albumNode := importAlbum(track)

			GetOrCreateRelationship(artistNode, albumNode, "RECORDED_ALBUM", neoism.Props{})
			GetOrCreateRelationship(albumNode, trackNode, "CONTAINS_TRACK", neoism.Props{})
		}

		GetOrCreateRelationship(artistNode, trackNode, "RECORDED_TRACK", neoism.Props{})

		rank, _ := strconv.Atoi(job.Track.Rank)

		importTag(job.Tag, job.Url, trackNode, rank)
		importTags(trackNode, track)
		importEmotions(trackNode, track)

		workerGroup.Done()
	}
}

func importArtist(track lastfm.TrackGetInfo) (artistNode *neoism.Node) {
	log.Printf("Importing artist: %s", track.Artist.Name)

	artistNode = GetOrCreateNode("Artist", "id", neoism.Props{
		"id": track.Artist.Mbid,
		"name": track.Artist.Name,
		"url": track.Artist.Url,
	})

	return
}

func importTrack(track lastfm.TrackGetInfo) (trackNode *neoism.Node) {
	log.Printf("Importing track: %s", track.Name)

	trackNode = GetOrCreateNode("Track", "id", neoism.Props{
		"id": track.Mbid,
		"name": track.Name,
		"url": track.Url,
		"playcount": track.PlayCount,
		"content": track.Wiki.Content,
	})

	return
}

func importAlbum(track lastfm.TrackGetInfo) (albumNode *neoism.Node) {
	log.Printf("Importing album: %s", track.Album.Title)

	albumNode = GetOrCreateNode("Album", "id", neoism.Props{
		"id": track.Album.Mbid,
		"title": track.Album.Title,
		"url": track.Album.Url,
	})

	return
}

func importTags(trackNode *neoism.Node, track lastfm.TrackGetInfo) {
	for _, tag := range track.TopTags {
		importTag(tag.Name, tag.Url, trackNode, -1)
	}
}

func importTag(name string, url string, trackNode *neoism.Node, rank int) {
	log.Printf("Importing tag %s with rank: %d", name, rank)

	tagNode := GetOrCreateNode("Tag", "name", neoism.Props{
		"name": name,
		"url": url,
	})

	GetOrCreateRelationship(trackNode, tagNode, "HAS_TAG", neoism.Props{
		"rank": rank,
	})
}

func importEmotions(trackNode *neoism.Node, track lastfm.TrackGetInfo) {
	weights := map[string]int{}
	content := strings.ToLower(track.Wiki.Content)

	for emotion, words := range AppEmotions.Emotions {
		for _, word := range words {
			if strings.Contains(content, strings.ToLower(word)) {
				weights[emotion] += 1
			}
		}
	}

	for emotion, weight := range weights {
		log.Printf("Importing emotion %s with weight: %d", emotion, weight)

		emotionNode := GetOrCreateNode("Emotion", "name", neoism.Props{"name": emotion})

		GetOrCreateRelationship(trackNode, emotionNode, "CONTAINS_EMOTION", neoism.Props{
			"weight": weight,
		})
	}
}

func getTopTags() (tags lastfm.TagGetTopTags) {
	tags, err := GetLastFmConnection().Tag.GetTopTags(lastfm.P{})

	if err != nil {
		log.Printf("Cannot get top tags: %s", err)
	}

	return
}

func getTopTracksByTag(tag string, page ...int) (tracks lastfm.TagGetTopTracks) {
	realPage := 5

	if len(page) > 0 {
		realPage = page[0]
	}

	tracks, err := GetLastFmConnection().Tag.GetTopTracks(lastfm.P{
		"tag": tag,
		"page": realPage,
	})

	if err != nil {
		log.Printf("Cannot get top tags: %s", err)
	}

	return
}

func getTrackByMbid(mbid string) (track lastfm.TrackGetInfo) {
	track, err := GetLastFmConnection().Track.GetInfo(lastfm.P{"mbid": mbid})

	if err != nil {
		log.Printf("Cannot get track: %s", err)
	}

	return
}
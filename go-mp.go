package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/schollz/progressbar"
)

func usage() {
	fmt.Printf("Usage: $ go-mp [file/directory],\nYou can pass any number of arguments\n")
}

func play(song string) {
	// Open Song file and get extension
	file, err := os.Open(song)
	check(err)
	defer file.Close()

	// Check for extension and choose right decoder
	streamer, format := getStreamer(path.Ext(song), file)
	defer streamer.Close()

	// Init Speaker
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Wrap in controller
	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}

	speaker.Play(ctrl)

	go func() {
		for {
			fmt.Scanln()

			speaker.Lock()
			ctrl.Paused = !ctrl.Paused
			speaker.Unlock()
		}
	}()

	renderBar(streamer, format)
}

func renderBar(streamer beep.StreamSeekCloser, format beep.Format) {
	d := getDuration(streamer, format)
	seconds := int(float64(d) / float64(time.Second))

	bar := progressbar.NewOptions(seconds, progressbar.OptionSetPredictTime(false), progressbar.OptionShowCount())
	for i := 0; i < seconds; i++ {
		bar.Add(1)
		time.Sleep(time.Second)
	}
}

// return song duration
func getDuration(streamer beep.StreamSeekCloser, format beep.Format) time.Duration {

	d := time.Duration(streamer.Len()) * format.SampleRate.D(1)
	return d.Round(time.Second)
}

// return streamer dependend of file extension
func getStreamer(extension string, file *os.File) (beep.StreamSeekCloser, beep.Format) {
	var streamer beep.StreamSeekCloser
	var format beep.Format
	var err error

	switch extension {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
	case ".wav":
		streamer, format, err = wav.Decode(file)
	case ".flac":
		streamer, format, err = flac.Decode(file)
	default:
		log.Fatalf("This Filetype is not supported!: %v\n", extension)
	}
	check(err)
	return streamer, format
}

// Outsource error handling
func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func prepareSongList(args []string) []string {
	var list []string

	for _, element := range args {
		fi, err := os.Stat(element)

		switch {
		case err != nil:
			log.Fatalln("There seems to be a Problem: ", err)
		case fi.IsDir():
			f, err := os.Open(element)
			check(err)
			dir, err := f.Readdirnames(0)
			check(err)
			for _, item := range dir {
				list = append(list, filepath.Join(element, item))
			}
		default:
			list = append(list, element)
		}
	}
	return list
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalln("No files to be played specified!")
	}

	songList := prepareSongList(flag.Args())

	for _, song := range songList {
		fmt.Println("Playing:", song)
		play(song)
	}
}

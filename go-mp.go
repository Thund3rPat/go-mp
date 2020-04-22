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
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

var directory = flag.Bool("d", false, "Enable this flag to open a directory")

// Outsource error handling
func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func play(song string) {
	var streamer beep.StreamSeekCloser
	var format beep.Format
	var err error

	// Open Song file and get extension
	file, err := os.Open(song)
	filetype := path.Ext(song)
	check(err)
	defer file.Close()

	// Check for extension and choose right decoder
	switch filetype {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
	case ".wav":
		streamer, format, err = wav.Decode(file)
	default:
		log.Fatal("File not supported!")
	}
	check(err)
	defer streamer.Close()

	// Init Speaker
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Play and wait until song is finished
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
func playDirectory() []string {
	var listofsongs []string

	// Open directory
	files, err := os.Open(flag.Arg(0))
	check(err)
	defer files.Close()

	// Get slice of name in directory
	names, err := files.Readdirnames(0)
	check(err)

	// append path to listofsongs
	for _, v := range names {
		listofsongs = append(listofsongs, filepath.Join(flag.Arg(0), v))
	}

	return listofsongs
}

func main() {
	var songs []string

	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("No files to be played specified!")
	}

	// Check if -d flag is set
	if *directory {
		// Get songlist from directory
		songs = playDirectory()
	} else {
		// Get songlist from Parameters
		songs = flag.Args()
	}

	// Play
	for _, song := range songs {
		fmt.Println("Currently Playing: ", song)
		play(song)
		fmt.Println("Finished Playing")
	}
}

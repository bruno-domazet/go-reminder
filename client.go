package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Reminder struct {
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
	Message string  `json:"message,omitempty"`
	// Repeat  Interval `json:"repeat,omitempty"`
}

// type Interval struct {
// 	// minutes,hours,days,weeks,months,years
// 	Unit     string  `json:"unit,omitempty"`
// 	Interval float64 `json:"interval,omitempty"`
// }

func greet(name string) string {
	m := fmt.Sprintln("from greet is", name)
	return m
}

func readDb(filepath string, data *[]Reminder) (bool, error) {
	start := time.Now()

	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return false, err
	}

	// TODO: check for empty file, before running Unmarsal

	// load json file to Memory
	// and update the pointer
	err = json.Unmarshal(file, &data)
	if err != nil {
		return false, err
	}

	log.Debug().Caller().Float64("took", time.Since(start).Seconds()).Send()
	return true, nil
}

func writeDb(filepath string, data []Reminder) (bool, error) {
	start := time.Now()

	jsonString, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	err = ioutil.WriteFile(filepath, jsonString, os.ModeAppend)
	if err != nil {
		return false, err
	}

	log.Debug().Caller().Float64("took", time.Since(start).Seconds()).Send()
	return true, nil
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func parseTimeString(timeString string) float64 {
	// now := carbon.Now()

	// modifier present
	if strings.HasPrefix(timeString, "+") {
		// TODO regex it
		return 2
	}

	// timestamp present
	if strings.Contains(timeString, "@") {
		// hourMinute := strings.Split(strings.Split(timeString, "@")[1], ":")
		// t1, _ = carbon.Create(2012, 1, 31, 12, 0, 0, 0, "Europe/Rome")

	}

	return 1
}

func main() {
	benchMark := time.Now()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Flags
	debug := flag.Bool("v", false, "Verbose, sets log level to debug")
	help := flag.Bool("h", false, "This help message")

	msgFlag := flag.String("msg", "", "(required) The what")
	startFlag := flag.String("start", "", "(required) The when. Format: '+N[min|hour|day][@hh:mm]'")
	endFlag := flag.String("end", "", "(optional) The end. Same as 'start'")

	// TODO: support repeats
	// interval := flag.String("i", "1", "how many repeats?")
	// unit := flag.String("u", "days", "unit for repeating the event")

	flag.Parse()

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Print help
	if *help {
		flag.PrintDefaults()
		return
	}

	if *debug {
		fmt.Printf("%v,%v\n", *startFlag, *endFlag)
	}

	// parse time
	now := parseTimeString(*startFlag)
	fmt.Printf("%v", now)

	// create from CLI args (with sane defaults)
	entry := Reminder{
		Message: *msgFlag,
	}

	// if *start. {
	// 	entry.Start = *start
	// }
	var reminders []Reminder

	// parse and append
	_, err := readDb("./tst.json", &reminders)
	if err != nil {
		log.Panic().Caller().Float64("took", time.Since(benchMark).Seconds()).Msg(err.Error())
	}

	reminders = append(reminders, entry)
	success, err := writeDb("./tst.json", reminders)
	if err != nil {
		log.Error().Float64("took", time.Since(benchMark).Seconds()).Msg(err.Error())
	}

	if success {
		fmt.Println("Successfully added the reminder!")
	}

	log.Debug().Caller().Float64("took", time.Since(benchMark).Seconds()).Send()
}

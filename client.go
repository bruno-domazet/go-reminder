package remme

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/uniplaces/carbon"
)

type Reminder struct {
	Message string  `json:"message,omitempty"`
	Start   float64 `json:"start"`
	// End     float64 `json:"end"`
	// Repeat  Interval `json:"repeat,omitempty"`
}

var modifierUnits = `m|h|d`
var benchMark = time.Now()

// type Interval struct {
// 	// minutes,hours,days,weeks,months,years
// 	Unit     string  `json:"unit,omitempty"`
// 	Interval float64 `json:"interval,omitempty"`
// }

func readDb(filepath string, data *[]Reminder) (bool, error) {

	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return false, err
	}
	context.Background()

	// TODO: check for empty file, before running Unmarsal

	// load json file to Memory
	// and update the pointer
	err = json.Unmarshal(file, &data)
	if err != nil {
		return false, err
	}

	log.Debug().Caller().Float64("took", time.Since(benchMark).Seconds()).Msg("readDb")
	return true, nil
}

func writeDb(filepath string, data []Reminder) (bool, error) {

	jsonString, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	err = ioutil.WriteFile(filepath, jsonString, os.ModeAppend)
	if err != nil {
		return false, err
	}

	log.Debug().Caller().Float64("took", time.Since(benchMark).Seconds()).Msg("writeDb")
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

func parseTimeString(timeString string, initialDateTime *carbon.Carbon) (*carbon.Carbon, error) {
	now := initialDateTime

	// modifier present
	if strings.HasPrefix(timeString, "+") {
		r := regexp.MustCompile(`\+(?P<mod>\d)(?P<unit>(` + modifierUnits + `)+)`)
		modifiers, err := regMap(timeString, r)
		if err != nil {
			return nil, err
		}

		intUnit, err := strconv.Atoi(modifiers["mod"])
		if err != nil {
			return nil, err
		}

		switch modifiers["unit"] {
		case "m":
			now = now.AddMinutes(intUnit)
			break
		case "h":
			now = now.AddHours(intUnit)
			break
		case "d":
			now = now.AddDays(intUnit)
			break
		}
	}

	// timestamp present
	if strings.Contains(timeString, "@") {
		r := regexp.MustCompile(`\@(?P<h>\d{2})\:?(?P<m>\d{2})?`)
		hourMinute, err := regMap(timeString, r)
		if err != nil {
			log.Debug().Caller().Float64("took", time.Since(benchMark).Seconds()).Msg(err.Error())
			return nil, err
		}
		err = now.SetTimeFromTimeString(hourMinute["h"] + ":" + hourMinute["m"] + ":05")
		if err != nil {
			log.Debug().Caller().Float64("took", time.Since(benchMark).Seconds()).Msg(err.Error())
			return nil, err
		}

	}

	return now, nil
}

func regMap(s string, r *regexp.Regexp) (map[string]string, error) {
	vals := r.FindStringSubmatch(s)
	keys := r.SubexpNames()
	d := make(map[string]string)

	// not the same amount of keys vs values
	if len(vals) != len(keys) {
		return nil, errors.New("Some arguments are missing or contain invalid values!")
	}

	// map it
	for i := 1; i < len(keys); i++ {
		d[keys[i]] = vals[i]
	}
	return d, nil
}

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Flags
	debug := flag.Bool("v", false, "Verbose, sets log level to 'debug'")
	help := flag.Bool("h", false, "This help message")

	msgFlag := flag.String("msg", "", "The what")
	startFlag := flag.String("start", "", "The when. Format: '+N["+modifierUnits+"][@hh:mm]'")

	// TODO: support repeats?
	// endFlag := flag.String("end", "", "(optional) The end, relative to 'start'. Format: '+N["+modifierUnits+"][@hh:mm]'")
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
		os.Exit(0)
	}

	if len(*msgFlag) == 0 && len(*startFlag) == 0 {
		fmt.Println("Nothing to add...")
		os.Exit(1)
	}

	// parse start time
	start, err := parseTimeString(*startFlag, carbon.Now())
	if err != nil {
		fmt.Printf("Something went wrong: %s\n", err.Error())
		os.Exit(1)
	}

	// parse end time
	// end, err := parseTimeString(*endFlag, start)
	// if err != nil {
	// 	fmt.Println("Something went wrong!")
	// 	return
	// }

	// create from CLI args (with sane defaults)
	entry := Reminder{
		Message: *msgFlag,
		Start:   float64(start.Timestamp()),
		// End:     float64(end.Timestamp()),
	}

	// append to file
	var reminders []Reminder
	readDb("./tst.json", &reminders)
	reminders = append(reminders, entry)
	success, _ := writeDb("./tst.json", reminders)

	log.Debug().Caller().Float64("took", time.Since(benchMark).Seconds()).Send()
	if success {
		fmt.Printf("Successfully added a reminder for '%s' @ %s \n", *msgFlag, start.ISO8601String())
		os.Exit(0)
	}

	os.Exit(1)
}

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mateuszdyminski/auto/ingress/model"
	nats "github.com/nats-io/go-nats"

	"github.com/BurntSushi/toml"
)

var configPath string
var rps int

// Config holds configuration of feeder.
type Config struct {
	NATSAddress string
	Topic       string
	CsvDir      string
}

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	flag.StringVar(&configPath, "config", "config/conf.toml", "config path")
	flag.IntVar(&rps, "rps", 10, "Requests per second - number of send requests per second")
}

func main() {
	// load config
	flag.Parse()

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Can't open config file!")
	}

	var conf Config
	if err := toml.Unmarshal(bytes, &conf); err != nil {
		log.Fatalf("Can't decode config file!")
	}

	log.Infof("Config: %v", conf)

	// pump data into Nats
	pumpToNats(&conf, streamCrashes(&conf))
}

func streamCrashes(conf *Config) chan model.FlightCrash {
	f, err := os.Open(conf.CsvDir)
	if err != nil {
		log.Fatal("can't file with users:", err)
	}
	defer f.Close()

	log.Infof("Start reading CSV file!")

	r := csv.NewReader(bufio.NewReader(f))
	r.LazyQuotes = true
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Read %d lines!", len(records))

	out := make(chan model.FlightCrash, 1024)
	go func() {
		for i, line := range records {
			if i == 0 { // skip header
				continue
			}

			if len(line) != 13 {
				log.Fatal(fmt.Errorf("Wrong number of parsed fields: %d. Index %d", len(line), i))
			}
			f := model.FlightCrash{}

			var err error
			date := line[0]
			timeStr := line[1]
			if timeStr == "" || timeStr == "?" {
				timeStr = "00:00"
			}

			f.Date, err = time.Parse("January 02, 2006 15:04", date+" "+timeStr)
			if err != nil {
				log.Fatalf("Can't deserialize date and time of flight. Line: %d Date: %s, Flight: %s", i, date, timeStr)
			}

			if line[2] != "?" {
				f.Location = line[2]
			}

			if line[3] != "?" {
				f.Operator = line[3]
			}

			if line[4] != "?" {
				f.FlightNo = line[4]
			}

			if line[5] != "?" {
				f.Route = line[5]
			}

			if line[6] != "?" {
				f.AircraftType = line[6]
			}

			if line[7] != "?" {
				f.Registration = line[7]
			}

			if line[8] != "?" {
				f.SerialNumber = line[8]
			}

			if line[9] != "?" {
				f.Aboard = parseAboard(line[9])
			}

			if line[10] != "?" {
				f.Fatalities = parseAboard(line[10])
			}

			if line[11] != "?" && line[11] != "" {
				f.Ground, err = strconv.Atoi(line[11])
				if err != nil {
					log.Fatalf("Can't deserialize ground. Val: %s", line[11])
				}
			}

			if line[12] != "?" {
				f.Summary = line[12]
			}

			out <- f
		}

		log.Infof("Closing channel")
		close(out)
	}()

	return out
}

func parseAboard(aboard string) model.Aboard {
	// parse following string:
	// 7 (passengers:6 crew:1)
	aboard = strings.Replace(aboard, "(passengers:", "", -1)
	aboard = strings.Replace(aboard, "crew:", "", -1)
	aboard = strings.Replace(aboard, ")", "", -1)
	vals := strings.Split(aboard, " ")
	a := model.Aboard{}
	var err error

	if vals[0] != "?" {
		a.Total, err = strconv.Atoi(vals[0])
		if err != nil {
			log.Fatalf("Can't deserialize aboard total. Val: %s", vals[0])
		}
	}

	if vals[1] != "?" {
		a.Passengers, err = strconv.Atoi(vals[1])
		if err != nil {
			log.Fatalf("Can't deserialize aboard passengers. Val: %s", vals[1])
		}
	}

	if vals[2] != "?" {
		a.Crew, err = strconv.Atoi(vals[2])
		if err != nil {
			log.Fatalf("Can't deserialize aboard crew. Val: %s", vals[2])
		}
	}

	return a
}

func pumpToNats(conf *Config, flights chan model.FlightCrash) {
	nc, err := nats.Connect(conf.NATSAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	var successes, errors int

	for flight := range flights {
		data, err := json.Marshal(flight)
		if err != nil {
			log.Fatalf("Can't serialize flight to JSON. Err: %v", err)
		}

		err = nc.Publish(conf.Topic, data)
		if err != nil {
			errors++
			log.Errorf("Can't publish msg to topic: %s. Err: %v", conf.Topic, err)
		} else {
			successes++
		}

		// throttle down the requests to achive proper rps(request per second)
		time.Sleep(time.Duration(1/rps) * time.Second)

		log.Infof("Successfully produced: %d flights; errors: %d", successes, errors)
	}

	nc.Flush()
	log.Info("All flights sent! Exiting...")
}

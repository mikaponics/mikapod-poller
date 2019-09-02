package internal // github.com/mikaponics/mikapod-poller/internal

import (
	"context"
	"log"
	// "os"
	"time"
	"fmt"

	"google.golang.org/grpc"
	"github.com/golang/protobuf/ptypes/timestamp"

    "github.com/mikaponics/mikapod-poller/configs"
	pb "github.com/mikaponics/mikapod-storage/api"
	pb2 "github.com/mikaponics/mikapod-soil-reader/api"
)

type MikapodPoller struct {
	timer *time.Timer
	ticker *time.Ticker
	done chan bool
	storageCon *grpc.ClientConn
	storage pb.MikapodStorageClient
	readerCon *grpc.ClientConn
	reader pb2.MikapodSoilReaderClient
}

// Function will construct the Mikapod Poller application.
func InitMikapodPoller(mikapodStorageAddress string, mikapodSoilReaderAddress string) (*MikapodPoller) {
	// Set up a direct connection to the `mikapod-storage` server.
	storageCon, err := grpc.Dial(mikapodStorageAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	// Set up our protocol buffer interface.
	storage := pb.NewMikapodStorageClient(storageCon)

    // Set up a direct connection to the `mikapod-soil-reader` server.
	readerCon, readerErr := grpc.Dial(mikapodSoilReaderAddress, grpc.WithInsecure())
	if readerErr != nil {
		log.Fatalf("did not connect: %v", err)
	}

	// Set up our protocol buffer interface.
	reader := pb2.NewMikapodSoilReaderClient(readerCon)

	return &MikapodPoller{
		timer: nil,
		ticker: nil,
		done: make(chan bool, 1), // Create a execution blocking channel.
		storageCon: storageCon,
		storage: storage,
		readerCon: readerCon,
		reader: reader,
	}
}

// Source: https://www.reddit.com/r/golang/comments/44tmti/scheduling_a_function_call_to_the_exact_start_of/
func minuteTicker() *time.Timer {
	// Current time
	now := time.Now()

	// Get the number of seconds until the next minute
	var d time.Duration
	d = time.Second * time.Duration(60-now.Second())

	// Time of the next tick
	nextTick := now.Add(d)

	// Subtract next tick from now
	diff := nextTick.Sub(time.Now())

	// Return new ticker
	return time.NewTimer(diff)
}


// Function will consume the main runtime loop and run the business logic
// of the Mikapod Poller application.
func (app *MikapodPoller) RunMainRuntimeLoop() {
	defer app.shutdown()

    // DEVELOPERS NOTE:
	// (1) The purpose of this block of code is to find the future date where
	//     the minute just started, ex: 5:00 AM, 5:01, etc, and then start our
	//     main runtime loop to run along for every minute afterwords.
	// (2) If our application gets terminated by the user or system then we
	//     terminate our timer.
    log.Printf("Synching with local time...")
	app.timer = minuteTicker()
	select {
		case <- app.timer.C:
			log.Printf("Synchronized with local time.")
			app.ticker = time.NewTicker(1 * time.Minute)
		case <- app.done:
			app.timer.Stop()
			log.Printf("Interrupted timer.")
			return
	}

    // // THIS CODE IS FOR TESTING, REMOVE WHEN READY TO USE, UNCOMMENT ABOVE.
	// app.ticker = time.NewTicker(1 * time.Minute)

    // DEVELOPERS NOTE:
	// (1) The purpose of this block of code is to run as a goroutine in the
	//     background as an anonymous function waiting to get either the
	//     ticker chan or app termination chan response.
	// (2) Main runtime loop's execution is blocked by the `done` chan which
	//     can only be triggered when this application gets a termination signal
	//     from the operating system.
	log.Printf("Poller is now running.")
	go func() {
        for {
            select {
	            case <- app.ticker.C:
					data := app.getDataFromArduino()
					app.saveDataToStorage(data)
				case <- app.done:
					app.ticker.Stop()
					log.Printf("Interrupted ticker.")
					return
			}
		}
	}()
	<-app.done
}

// Function will tell the application to stop the main runtime loop when
// the process has been finished.
func (app *MikapodPoller) StopMainRuntimeLoop() {
	app.done <- true
}

func (app *MikapodPoller) shutdown()  {
    app.storageCon.Close()
	app.readerCon.Close()
}


func (app *MikapodPoller) getDataFromArduino() (*TimeSeriesData){
	c := app.reader

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetData(ctx, &pb2.GetTimeSeriesData{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	return &TimeSeriesData{
		HumidityValue: r.HumidityValue,
		HumidityUnit: r.HumidityUnit,
		TemperatureValue: r.TemperatureValue,
		TemperatureUnit: r.TemperatureUnit,
		PressureValue: r.PressureValue,
		PressureUnit: r.PressureUnit,
		TemperatureBackupValue: r.TemperatureBackupValue,
		TemperatureBackupUnit: r.TemperatureBackupUnit,
		AltitudeValue: r.AltitudeValue,
		AltitudeUnit: r.AltitudeUnit,
		IlluminanceValue: r.IlluminanceValue,
		IlluminanceUnit: r.IlluminanceUnit,
		SoilMoistureValue: r.SoilMoistureValue,
		SoilMoistureUnit: r.SoilMoistureUnit,
		Timestamp: r.Timestamp,
	}
}

func (app *MikapodPoller) saveDataToStorage(data *TimeSeriesData) {
	// For debugging purposes only.
	fmt.Printf("\n%+v\n", data)

	app.addTimeSeriesDatum(configs.HumidityInstrumentId, data.HumidityValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.TemperatureInstrumentId, data.TemperatureValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.PressureInstrumentId, data.PressureValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.AltitudeInstrumentId, data.AltitudeValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.IlluminanceInstrumentId, data.IlluminanceValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.SoilMoistureInstrumentId, data.SoilMoistureValue, data.Timestamp)
}

func (app *MikapodPoller) addTimeSeriesDatum(instrument int32, value float32, ts *timestamp.Timestamp) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := app.storage.AddTimeSeriesDatum(ctx, &pb.TimeSeriesDatumRequest{
		Instrument: instrument,
		Value: value,
		Timestamp: ts,
	})
	if err != nil {
		log.Fatalf("could not add time-series data to storage: %v", err)
	}
}

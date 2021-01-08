package internal // github.com/mikaponics/mikapod-poller/internal

import (
	// "context"
	"log"
	// "os"
	"time"

	storage_rpc "github.com/mikaponics/mikapod-storage/pkg/rpc_client"
	soil_rpc "github.com/mikaponics/mikapod-soil-reader/pkg/rpc_client"

	"github.com/mikaponics/mikapod-poller/configs"
)

type MikapodPoller struct {
	timer *time.Timer
	ticker *time.Ticker
	done chan bool
	storageService *storage_rpc.MikapodStorageService
	soilReaderService *soil_rpc.MikapodSoilReaderService
}

// Function will construct the Mikapod Poller application.
func InitMikapodPoller(mikapodStorageAddress string, mikapodSoilReaderAddress string) (*MikapodPoller) {
	log.Printf("Attempting to connect to the storage service.")
	storageService := storage_rpc.New(mikapodStorageAddress)
	log.Printf("Attempting to connect to the soil reader service.")
	soilReaderService := soil_rpc.New(mikapodSoilReaderAddress)
	log.Printf("Successfully connected to dependent services.")

    // DEVELOPERS NOTE: Uncomment the following code if you want this polling service
	//                  to immediately contact the soil reader to verify it is working.
	//                  This code is useful for troubleshooting problems dealing with RPC.
	// soilReaderService.GetData()

	return &MikapodPoller{
		timer: nil,
		ticker: nil,
		done: make(chan bool, 1), // Create a execution blocking channel.
		storageService: storageService,
		soilReaderService: soilReaderService,
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
					// log.Printf("Tick") // For debugging purposes only.
					data := app.getDataFromArduino()
					if data != nil {
						app.saveDataToStorage(data)
					}
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
    // app.storageCon.Close()
	// app.readerCon.Close()
}


func (app *MikapodPoller) getDataFromArduino() (*TimeSeriesData){
	r, err := app.soilReaderService.GetData()
	if err != nil {
		// DEVELOPERS NOTE:
		// Do not terminate application due because the IoT device might be
		// powering up while we made the call.
		log.Println("Could not fetch from soil reader service because: %v", err)
		return nil
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
	return nil
}

func (app *MikapodPoller) saveDataToStorage(data *TimeSeriesData) {
	// For debugging purposes only.
	// log.Printf("\n%+v\n", data)

	app.addTimeSeriesDatum(configs.HumidityInstrumentId, data.HumidityValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.TemperatureInstrumentId, data.TemperatureValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.PressureInstrumentId, data.PressureValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.AltitudeInstrumentId, data.AltitudeValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.IlluminanceInstrumentId, data.IlluminanceValue, data.Timestamp)
	app.addTimeSeriesDatum(configs.SoilMoistureInstrumentId, data.SoilMoistureValue, data.Timestamp)
}

func (app *MikapodPoller) addTimeSeriesDatum(instrument int32, value float32, ts int64) {
	_, err := app.storageService.AddTimeSeriesDatum(&storage_rpc.TimeSeriesDatumCreateRequest{
		Instrument: instrument,
		Value: value,
		Timestamp: ts,
	})
	if err != nil {
		// DEVELOPERS NOTE:
		// Do not terminate application due because the IoT device might be
		// powering up while we made the call.
		log.Println("Could not add time-series data to storage: %v", err)
	}
}

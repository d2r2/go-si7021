package main

import (
	"context"
	"os"
	"syscall"
	"time"

	i2c "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
	shell "github.com/d2r2/go-shell"
	si7021 "github.com/d2r2/go-si7021"
)

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

func main() {
	defer logger.FinalizeLogger()
	// Create new connection to i2c-bus on 1 line with address 0x40.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x40, 0)
	if err != nil {
		lg.Fatal(err)
	}
	defer i2c.Close()

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** !!! READ THIS !!!")
	lg.Notify("*** You can change verbosity of output, by modifying logging level of modules \"i2c\", \"si7021\".")
	lg.Notify("*** Uncomment/comment corresponding lines with call to ChangePackageLogLevel(...)")
	lg.Notify("*** !!! READ THIS !!!")
	lg.Notify("**********************************************************************************************")
	// Uncomment/comment next line to suppress/increase verbosity of output
	// logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	// logger.ChangePackageLogLevel("si7021", logger.InfoLevel)

	sensor := si7021.NewSi7021()
	err = sensor.Reset(i2c)
	if err != nil {
		lg.Fatal(err)
	}

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Read sensor identity and states")
	lg.Notify("**********************************************************************************************")
	vlow, err := sensor.GetVoltageLow(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	if vlow {
		lg.Infof("Voltage status = LOW")
	} else {
		lg.Infof("Voltage status = OK")
	}
	mr, err := sensor.GetMeasureResolution(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Measure resolution = %v", mr)
	hs, err := sensor.GetHeaterStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Heater ON status = %v", hs)
	hl, err := sensor.GetHeaterLevel(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Heater level = %v", hl)
	b, err := sensor.ReadFirmwareVersion(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Revision = %v", b)
	sn, err := sensor.ReadSerialNumber(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Serial number = %v", sn)
	st, err := sensor.ReadSensoreType(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Sensor type = %v", st)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Measure humidity and temperature")
	lg.Notify("**********************************************************************************************")
	urh, ut, err := sensor.ReadUncompHumidityAndTemprature(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Humidity and temprature uncompensated = %v, %v", urh, ut)
	rh, err := sensor.ReadRelativeHumidity(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Relative humidity = %v%%", rh)
	t, err := sensor.ReadTemperature(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Temprature in celsius = %v*C", t)
	rh, t, err = sensor.ReadRelativeHumidityAndTemperature(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Relative humidity and temperature = %v%%, %v*C", rh, t)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Activate heater for 3 secs and make a measurement ")
	lg.Notify("**********************************************************************************************")
	// create context with cancellation possibility
	ctx, cancel := context.WithCancel(context.Background())
	// use done channel as a trigger to exit from signal waiting goroutine
	done := make(chan struct{})
	defer close(done)
	// build actual signals list to control
	signals := []os.Signal{os.Kill, os.Interrupt}
	if shell.IsLinuxMacOSFreeBSD() {
		signals = append(signals, syscall.SIGTERM)
	}
	// run goroutine waiting for OS termination events, including keyboard Ctrl+C
	shell.CloseContextOnSignals(cancel, done, signals...)

	err = sensor.SetHeaterLevel(i2c, si7021.HEATER_LEVEL_8)
	if err != nil {
		lg.Fatal(err)
	}
	err = sensor.SetHeaterStatus(i2c, true)
	if err != nil {
		lg.Fatal(err)
	}
	hs, err = sensor.GetHeaterStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	hl, err = sensor.GetHeaterLevel(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Heater ON status and level = %v, %v", hs, hl)
	pause := time.Second * 3
	lg.Infof("Waiting %v...", pause)
	select {
	// Check for termination request.
	case <-ctx.Done():
		err = sensor.SetHeaterStatus(i2c, false)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Fatal(ctx.Err())
	// Sleep 10 sec.
	case <-time.After(pause):
	}
	err = sensor.SetHeaterStatus(i2c, false)
	if err != nil {
		lg.Fatal(err)
	}
	hs, err = sensor.GetHeaterStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Heater ON status = %v", hs)
	rh, t, err = sensor.ReadRelativeHumidityAndTemperature(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Relative humidity and temperature = %v%%, %v*C", rh, t)

}

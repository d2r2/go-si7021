package main

import (
	i2c "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
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
	i2c, err := i2c.NewI2C(0x40, 1)
	if err != nil {
		lg.Fatal(err)
	}
	defer i2c.Close()

	// Uncomment/comment next line to suppress/increase verbosity of output
	//logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	//logger.ChangePackageLogLevel("si7021", logger.InfoLevel)

	sensor := si7021.NewSi7021()
	if err != nil {
		lg.Fatal(err)
	}

	vs, err := sensor.GetVoltageStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Voltage status = %v", vs)

	mr, err := sensor.GetMeasureResolution(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Measure resolution = %v", mr)

	hs, err := sensor.GetHeaterStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Heater status = %v", hs)

	hl, err := sensor.GetHeaterLevel(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Heater level = %v", hl)

	// Switch on internal sensor heater
	// err = sensor.SetHeaterStatus(i2c, si7021.HEATER_ENABLED)
	// if err != nil {
	// 	lg.Fatal(err)
	// }

	// Set internal heater heating power to 8 from 15.
	// err = sensor.SetHeaterLevel(i2c, si7021.HEATER_LEVEL_8)
	// if err != nil {
	// 	lg.Fatal(err)
	// }

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

	urh, ut, err := sensor.ReadUncompRelativeHumidityAndTempratureMode1(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("RH and temprature uncompensated = %v, %v", urh, ut)

	rh, err := sensor.ReadRelativeHumidityMode1(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Relative humidity = %v%%", rh)

	t, err := sensor.ReadTemperatureCelsiusMode1(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Temprature in celsius = %v*C", t)

	err = sensor.Reset(i2c)
	if err != nil {
		lg.Fatal(err)
	}

}

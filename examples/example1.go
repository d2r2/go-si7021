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
	// Create new connection to i2c-bus on 1 line with address 0x76.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x40, 1)
	if err != nil {
		lg.Fatal(err)
	}
	defer i2c.Close()
	// Uncomment next line to supress verbose output
	// logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

	// sensor, err := bsbmp.NewBMP(bsbmp.BMP180_TYPE, i2c)
	sensor := si7021.NewSi7021()
	if err != nil {
		lg.Fatal(err)
	}
	// Uncomment next line to supress verbose output
	// logger.ChangePackageLogLevel("bsbmp", logger.InfoLevel)

	b, err := sensor.ReadFirmwareVersion(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Info(b)

	buf1, err := sensor.ReadSerialNumber(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("%+v", buf1)
}

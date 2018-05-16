Silicon Labs Si7021 relative humidity and temperature sensor
============================================================

[![Build Status](https://travis-ci.org/d2r2/go-si7021.svg?branch=master)](https://travis-ci.org/d2r2/go-si7021)
[![Go Report Card](https://goreportcard.com/badge/github.com/d2r2/go-si7021)](https://goreportcard.com/report/github.com/d2r2/go-si7021)
[![GoDoc](https://godoc.org/github.com/d2r2/go-si7021?status.svg)](https://godoc.org/github.com/d2r2/go-si7021)
[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

Si7021 ([pdf reference](https://raw.github.com/d2r2/go-si7021/master/docs/Si7021-A20.pdf)) high accuracy temperature and relative humidity sensor. Easily integrated with Arduino and Raspberry PI due to i2c communication interface:
![image](https://raw.github.com/d2r2/go-si7021/master/docs/Si7021_GY-21.jpg)

This sensor has extra feature - integrated heater which could be helpfull in some specific application (such as periodic condensate removal, for example).

Here is a library written in [Go programming language](https://golang.org/) for Raspberry PI and counterparts, which gives you in the output relative humidity and temperature values (making all necessary i2c-bus interracting and values computing).

Golang usage
------------


```go
func main() {
	// Create new connection to i2c-bus on 1 line with address 0x40.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x40, 1)
	if err != nil {
		log.Fatal(err)
	}
	defer i2c.Close()

	sensor := si7021.NewSi7021()
	if err != nil {
		log.Fatal(err)
	}
	
	rh, err := sensor.ReadRelativeHumidityMode1(i2c)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Relative humidity = %v%%\n", rh)
	
	t, err := sensor.ReadTemperatureCelsiusMode1(i2c)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Temprature in celsius = %v*C\n", t)  
```


Getting help
------------

GoDoc [documentation](http://godoc.org/github.com/d2r2/go-si7021)

Installation
------------

```bash
$ go get -u github.com/d2r2/go-si7021
```

Troubleshoting
--------------

- *How to obtain fresh Golang installation to RPi device (either any RPi clone):*
If your RaspberryPI golang installation taken by default from repository is outdated, you may consider
to install actual golang mannualy from official Golang [site](https://golang.org/dl/). Download
tar.gz file containing armv6l in the name. Follow installation instructions.

- *How to enable I2C bus on RPi device:*
If you employ RaspberryPI, use raspi-config utility to activate i2c-bus on the OS level.
Go to "Interfaceing Options" menu, to active I2C bus.
Probably you will need to reboot to load i2c kernel module.
Finally you should have device like /dev/i2c-1 present in the system.

- *How to find I2C bus allocation and device address:*
Use i2cdetect utility in format "i2cdetect -y X", where X may vary from 0 to 5 or more,
to discover address occupied by peripheral device. To install utility you should run
`apt install i2c-tools` on debian-kind system. `i2cdetect -y 1` sample output:
	```
	     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
	00:          -- -- -- -- -- -- -- -- -- -- -- -- --
	10: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	20: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	30: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	40: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	50: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	70: -- -- -- -- -- -- 76 --    
	```

Contact
-------

Please use [Github issue tracker](https://github.com/d2r2/go-si7021/issues) for filing bugs or feature requests.


License
-------

Go-si7021 is licensed under MIT License.

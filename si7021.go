package si7021

import i2c "github.com/d2r2/go-i2c"

const (
	CMD_REL_HUM_MASTER_MODE       = 0xE5   // Measure Relative Humidity, Hold Master Mode
	CMD_REL_HUM_NO_MASTER_MODE    = 0xF5   // Measure Relative Humidity, No Hold Master Mode
	CMD_TEMPRATURE_MASTER_MODE    = 0xE3   // Measure Temperature, Hold Master Mode
	CMD_TEMPRATURE_NO_MASTER_MODE = 0xF3   // Measure Temperature, No Hold Master Mode
	CMD_TEMP_FROM_PREVIOUS        = 0xE0   // Read Temperature Value from Previous RH Measurement
	CMD_RESET                     = 0xFE   // Reset
	CMD_WRITE_USER_REG_1          = 0xE6   // Write RH/T User Register 1
	CMD_READ_USER_REG_1           = 0xE7   // Read RH/T User Register 1
	CMD_WRITE_HEATER_REG          = 0x51   // Write Heater Control Register
	CMD_READ_HEATER_REG           = 0x51   // Read Heater Control Register
	CMD_READ_ID_1ST_BYTE          = 0xFA0F // Read Electronic ID 1st Byte
	CMD_READ_ID_2ND_BYTE          = 0xFCC9 // Read Electronic ID 2nd Byte
	CMD_READ_FIRMWARE_REV         = 0x84B8 // Read Firmware Revision
)

type Si7021 struct {
}

func NewSi7021() *Si7021 {
	v := &Si7021{}
	return v
}

func (v *Si7021) ReadFirmware(i2c *i2c.I2C) (byte, error) {
	err := i2c.WriteRegU16LE(0x40, CMD_READ_FIRMWARE_REV)
	if err != nil {
		return 0, err
	}
	b, err := i2c.ReadRegU8(0x40 | 0x80)
	if err != nil {
		return 0, err
	}
	return b, nil
}

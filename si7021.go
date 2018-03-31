package si7021

import i2c "github.com/d2r2/go-i2c"

var (
	CMD_REL_HUM_MASTER_MODE       = []byte{0xE5}       // Measure Relative Humidity, Hold Master Mode
	CMD_REL_HUM_NO_MASTER_MODE    = []byte{0xF5}       // Measure Relative Humidity, No Hold Master Mode
	CMD_TEMPRATURE_MASTER_MODE    = []byte{0xE3}       // Measure Temperature, Hold Master Mode
	CMD_TEMPRATURE_NO_MASTER_MODE = []byte{0xF3}       // Measure Temperature, No Hold Master Mode
	CMD_TEMP_FROM_PREVIOUS        = []byte{0xE0}       // Read Temperature Value from Previous RH Measurement
	CMD_RESET                     = []byte{0xFE}       // Reset
	CMD_WRITE_USER_REG_1          = []byte{0xE6}       // Write RH/T User Register 1
	CMD_READ_USER_REG_1           = []byte{0xE7}       // Read RH/T User Register 1
	CMD_WRITE_HEATER_REG          = []byte{0x51}       // Write Heater Control Register
	CMD_READ_HEATER_REG           = []byte{0x51}       // Read Heater Control Register
	CMD_READ_ID_1ST_PART          = []byte{0xFA, 0x0F} // Read Electronic ID 1st Byte
	CMD_READ_ID_2ND_PART          = []byte{0xFC, 0xC9} // Read Electronic ID 2nd Byte
	CMD_READ_FIRMWARE_REV         = []byte{0x84, 0xB8} // Read Firmware Revision
)

type Si7021 struct {
}

func NewSi7021() *Si7021 {
	v := &Si7021{}
	return v
}

func (v *Si7021) ReadFirmware(i2c *i2c.I2C) (byte, error) {
	_, err := i2c.WriteBytes(CMD_READ_FIRMWARE_REV)
	if err != nil {
		return 0, err
	}
	buf2 := make([]byte, 1)
	_, err = i2c.ReadBytes(buf2)
	if err != nil {
		return 0, err
	}
	return buf2[0], nil
}

func (v *Si7021) ReadSerialNumber(i2c *i2c.I2C) ([]byte, error) {
	buf2 := make([]byte, 8)
	_, err := i2c.WriteBytes(CMD_READ_ID_1ST_PART)
	if err != nil {
		return nil, err
	}
	buf1 := make([]byte, 4)
	_, err = i2c.ReadBytes(buf1)
	if err != nil {
		return nil, err
	}
	buf2 = append([]byte{}, buf1[0:]...)
	_, err = i2c.WriteBytes(CMD_READ_ID_2ND_PART)
	if err != nil {
		return nil, err
	}
	buf1 = make([]byte, 4)
	_, err = i2c.ReadBytes(buf1)
	if err != nil {
		return nil, err
	}
	buf2 = append(buf2, buf1[0:]...)
	return buf2, nil

}

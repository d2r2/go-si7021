package si7021

import (
	"math"
	"time"

	i2c "github.com/d2r2/go-i2c"
)

// Command byte's sequences
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
	CMD_READ_HEATER_REG           = []byte{0x11}       // Read Heater Control Register
	CMD_READ_ID_1ST_PART          = []byte{0xFA, 0x0F} // Read Electronic ID 1st Byte
	CMD_READ_ID_2ND_PART          = []byte{0xFC, 0xC9} // Read Electronic ID 2nd Byte
	CMD_READ_FIRMWARE_REV         = []byte{0x84, 0xB8} // Read Firmware Revision
)

type FirmwareVersion byte

const (
	FIRMWARE_VER_1_0 FirmwareVersion = 0xFF
	FIRMWARE_VER_2_0 FirmwareVersion = 0x20
)

// String define stringer interface.
func (v FirmwareVersion) String() string {
	switch v {
	case FIRMWARE_VER_1_0:
		return "version 1.0"
	case FIRMWARE_VER_2_0:
		return "version 2.0"
	default:
		return "<unknown>"
	}
}

// SensorType denote sensor type.
type SensorType byte

const (
	SI_ENGINEERING_TYPE1 SensorType = 0x00
	SI_ENGINEERING_TYPE2 SensorType = 0xFF
	SI_7013_TYPE         SensorType = 0x0D
	SI_7020_TYPE         SensorType = 0x14
	SI_7021_TYPE         SensorType = 0x15
)

// String define stringer interface.
func (v SensorType) String() string {
	switch v {
	case SI_ENGINEERING_TYPE1, SI_ENGINEERING_TYPE2:
		return "engineering samples"
	case SI_7013_TYPE:
		return "Si7013"
	case SI_7020_TYPE:
		return "Si7020"
	case SI_7021_TYPE:
		return "Si7021"
	default:
		return "<unknown>"
	}
}

// MeasureResolution used to define measure
// precision in bits.
type MeasureResolution byte

const (
	RESOLUTION_RH_12BIT_TEMP_14BIT MeasureResolution = 0x00
	RESOLUTION_RH_8BIT_TEMP_12BIT  MeasureResolution = 0x01
	RESOLUTION_RH_10BIT_TEMP_13BIT MeasureResolution = 0x80
	RESOLUTION_RH_11BIT_TEMP_11BIT MeasureResolution = 0x81
	RESOLUTION_RH_TEMP_MASK        MeasureResolution = 0x81
)

// String define stringer interface.
func (v MeasureResolution) String() string {
	switch v {
	case RESOLUTION_RH_12BIT_TEMP_14BIT:
		return "RH - 12bit, temperature - 14bit"
	case RESOLUTION_RH_8BIT_TEMP_12BIT:
		return "RH - 8bit, temperature - 12bit"
	case RESOLUTION_RH_10BIT_TEMP_13BIT:
		return "RH - 10bit, temperature - 13bit"
	case RESOLUTION_RH_11BIT_TEMP_11BIT:
		return "RH - 11bit, temperature - 11bit"
	default:
		return "<unknown>"
	}
}

// HeaterStatus determine sensor internal
// heater state: on or off.
type HeaterStatus byte

const (
	HEATER_DISABLED    HeaterStatus = 0x0
	HEATER_ENABLED     HeaterStatus = 0x4
	HEATER_STATUS_MASK HeaterStatus = 0x4
)

// String define stringer interface.
func (v HeaterStatus) String() string {
	switch v {
	case HEATER_ENABLED:
		return "on"
	case HEATER_DISABLED:
		return "off"
	default:
		return "<unknown>"
	}
}

// VoltageStatus provided by sensor
// to express power supply voltage level
// - good or bad.
type VoltageStatus byte

const (
	VOLTAGE_OK          VoltageStatus = 0x0
	VOLTAGE_LOW         VoltageStatus = 0x40
	VOLTAGE_STATUS_MASK VoltageStatus = 0x40
)

// String define stringer interface.
func (v VoltageStatus) String() string {
	switch v {
	case VOLTAGE_OK:
		return "OK"
	case VOLTAGE_LOW:
		return "low"
	default:
		return "<unknown>"
	}
}

// HeaterLevel define gradation of
// sensor heating. Keep in mind that
// when heater is on, the temprature
// value provided by sensor is not correspond
// to real ambient temprature.
type HeaterLevel byte

const (
	HEATER_LEVEL_1    HeaterLevel = 0x0
	HEATER_LEVEL_2    HeaterLevel = 0x1
	HEATER_LEVEL_3    HeaterLevel = 0x2
	HEATER_LEVEL_4    HeaterLevel = 0x3
	HEATER_LEVEL_5    HeaterLevel = 0x4
	HEATER_LEVEL_6    HeaterLevel = 0x5
	HEATER_LEVEL_7    HeaterLevel = 0x6
	HEATER_LEVEL_8    HeaterLevel = 0x7
	HEATER_LEVEL_9    HeaterLevel = 0x8
	HEATER_LEVEL_10   HeaterLevel = 0x9
	HEATER_LEVEL_11   HeaterLevel = 0xA
	HEATER_LEVEL_12   HeaterLevel = 0xB
	HEATER_LEVEL_13   HeaterLevel = 0xC
	HEATER_LEVEL_14   HeaterLevel = 0xD
	HEATER_LEVEL_15   HeaterLevel = 0xE
	HEATER_LEVEL_16   HeaterLevel = 0xF
	HEATER_LEVEL_MASK HeaterLevel = 0xF
)

type Si7021 struct {
	lastUserReg *byte
}

// NewSi7021 returns new sensor instance.
func NewSi7021() *Si7021 {
	v := &Si7021{}
	return v
}

// ReadFirmwareVersion return sensor firmware revision.
func (v *Si7021) ReadFirmwareVersion(i2c *i2c.I2C) (FirmwareVersion, error) {
	_, err := i2c.WriteBytes(CMD_READ_FIRMWARE_REV)
	if err != nil {
		return 0, err
	}
	buf2 := make([]byte, 1)
	_, err = i2c.ReadBytes(buf2)
	if err != nil {
		return 0, err
	}
	fv := (FirmwareVersion)(buf2[0])
	lg.Debugf("Firmware version = %[1]s (%[1]d)", fv)

	return fv, nil
}

// ReadSerialNumber read sensor serial number
// which consists of byte array.
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

// ReadSensorType return sensor model.
func (v *Si7021) ReadSensoreType(i2c *i2c.I2C) (SensorType, error) {
	buf, err := v.ReadSerialNumber(i2c)
	if err != nil {
		return 0, err
	}
	st := (SensorType)(buf[4])
	lg.Debugf("Sensor type = %[1]s (%[1]d)", st)

	return st, nil
}

func (v *Si7021) readUserReg(i2c *i2c.I2C) (byte, error) {
	if v.lastUserReg == nil {
		_, err := i2c.WriteBytes(CMD_READ_USER_REG_1)
		if err != nil {
			return 0, err
		}
		buf1 := make([]byte, 1)
		_, err = i2c.ReadBytes(buf1)
		if err != nil {
			return 0, err
		}
		v.lastUserReg = &buf1[0]
	}
	return *v.lastUserReg, nil
}

// SetMeasureResolution set up sensor
// temprature and humidity measure accuracy.
func (v *Si7021) SetMeasureResolution(i2c *i2c.I2C, res MeasureResolution) error {
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return err
	}
	ur = ur&(^byte(RESOLUTION_RH_TEMP_MASK)) | (byte)(res)
	v.lastUserReg = &ur
	_, err = i2c.WriteBytes(append(CMD_WRITE_USER_REG_1, ur))
	return err
}

// GetMeasureResolution read current sensor measure accuracy.
func (v *Si7021) GetMeasureResolution(i2c *i2c.I2C) (MeasureResolution, error) {
	v.lastUserReg = nil
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return 0, nil
	}
	return (MeasureResolution)(ur) & RESOLUTION_RH_TEMP_MASK, nil
}

// SetHeaterStatus enable of disable internal heater.
func (v *Si7021) SetHeaterStatus(i2c *i2c.I2C, heat HeaterStatus) error {
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return err
	}
	ur = ur&(^byte(HEATER_STATUS_MASK)) | (byte)(heat)
	v.lastUserReg = &ur
	_, err = i2c.WriteBytes(append(CMD_WRITE_USER_REG_1, ur))
	return err
}

// GetHeaterStatus return internal heater status: on or off.
func (v *Si7021) GetHeaterStatus(i2c *i2c.I2C) (HeaterStatus, error) {
	v.lastUserReg = nil
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return 0, nil
	}
	return (HeaterStatus)(ur) & HEATER_STATUS_MASK, nil
}

// GetVoltageStatus provide power suply voltage level: good or bad.
func (v *Si7021) GetVoltageStatus(i2c *i2c.I2C) (VoltageStatus, error) {
	v.lastUserReg = nil
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return 0, nil
	}
	return (VoltageStatus)(ur) & VOLTAGE_STATUS_MASK, nil
}

// SetHeaterLevel define internal heater
// heating gradation. Remeber, when heater is on
// temprature provided by sensor is not correspond
// to real ambient temprature.
func (v *Si7021) SetHeaterLevel(i2c *i2c.I2C, level HeaterLevel) error {
	var hcr byte
	hcr = (byte)(level)
	_, err := i2c.WriteBytes(append(CMD_WRITE_HEATER_REG, hcr))
	return err
}

// GetHeaterLevel return sensor heating gradation.
func (v *Si7021) GetHeaterLevel(i2c *i2c.I2C) (HeaterLevel, error) {
	_, err := i2c.WriteBytes(CMD_READ_HEATER_REG)
	if err != nil {
		return 0, err
	}
	buf1 := make([]byte, 1)
	_, err = i2c.ReadBytes(buf1)
	if err != nil {
		return 0, err
	}
	return (HeaterLevel)(buf1[0]) & HEATER_LEVEL_MASK, nil
}

// Reset reboot sensor.
func (v *Si7021) Reset(i2c *i2c.I2C) error {
	_, err := i2c.WriteBytes([]byte{0xFE})
	if err != nil {
		return err
	}
	// Powerup time
	time.Sleep(time.Millisecond * 15)
	return err
}

func (v *Si7021) doMeasure(i2c *i2c.I2C, cmd []byte, withCRC bool) (uint16, byte, error) {
	_, err := i2c.WriteBytes(cmd)
	if err != nil {
		return 0, 0, err
	}
	// Wait according to conversion time specification
	time.Sleep(time.Millisecond * (12 + 11))
	buf1 := make([]byte, 2)
	_, err = i2c.ReadBytes(buf1)
	if err != nil {
		return 0, 0, err
	}
	var crc byte
	if withCRC {
		buf2 := make([]byte, 1)
		_, err = i2c.ReadBytes(buf2)
		if err != nil {
			return 0, 0, err
		}
		crc = buf2[0]
	}
	meas := uint16(buf1[0])<<8 | uint16(buf1[1])
	return meas, crc, nil
}

// ReadUncompRelativeHumidityMode1 returns
// uncompensated humidity and CRC-8-Dallas/Maxim
// obtained with "Hold Master Mode" command.
func (v *Si7021) ReadUncompRelativeHumidityMode1(i2c *i2c.I2C) (uint16, byte, error) {
	rh, crc, err := v.doMeasure(i2c, CMD_REL_HUM_MASTER_MODE, true)
	return rh, crc, err
}

// ReadUncompRelativeHumidityMode2 returns
// uncompensated humidity and CRC-8-Dallas/Maxim
// obtained with "No Hold Master Mode" command.
func (v *Si7021) ReadUncompRelativeHumidityMode2(i2c *i2c.I2C) (uint16, byte, error) {
	rh, crc, err := v.doMeasure(i2c, CMD_REL_HUM_NO_MASTER_MODE, true)
	return rh, crc, err
}

// ReadUncompTemperatureMode1 returns
// uncompensated temperature and CRC-8-Dallas/Maxim
// obtained with "Hold Master Mode" command.
func (v *Si7021) ReadUncompTempratureMode1(i2c *i2c.I2C) (uint16, byte, error) {
	temp, crc, err := v.doMeasure(i2c, CMD_TEMPRATURE_MASTER_MODE, true)
	return temp, crc, err
}

// ReadUncompTemperatureMode2 returns
// uncompensated termperature and CRC-8-Dallas/Maxim
// obtained with "No Hold Master Mode" command.
func (v *Si7021) ReadUncompTempratureMode2(i2c *i2c.I2C) (uint16, byte, error) {
	temp, crc, err := v.doMeasure(i2c, CMD_TEMPRATURE_NO_MASTER_MODE, true)
	return temp, crc, err
}

// ReadUncompRelativeHumidityAndTemperatureMode1 returns
// uncompensated humidity/temperature and CRC-8-Dallas/Maxim
// obtained with "Hold Master Mode" command.
func (v *Si7021) ReadUncompRelativeHumidityAndTempratureMode1(i2c *i2c.I2C) (uint16, uint16, error) {
	rh, _, err := v.doMeasure(i2c, CMD_REL_HUM_MASTER_MODE, false)
	if err != nil {
		return 0, 0, err
	}
	temp, _, err := v.doMeasure(i2c, CMD_TEMP_FROM_PREVIOUS, false)
	if err != nil {
		return 0, 0, err
	}
	return rh, temp, nil
}

// ReadUncompRelativeHumidityAndTemperatureMode2 returns
// uncompensated humidity/termperature and CRC-8-Dallas/Maxim
// obtained with "No Hold Master Mode" command.
func (v *Si7021) ReadUncompRelativeHumidityAndTempratureMode2(i2c *i2c.I2C) (uint16, uint16, error) {
	rh, _, err := v.doMeasure(i2c, CMD_REL_HUM_NO_MASTER_MODE, false)
	if err != nil {
		return 0, 0, err
	}
	temp, _, err := v.doMeasure(i2c, CMD_TEMP_FROM_PREVIOUS, false)
	if err != nil {
		return 0, 0, err
	}
	return rh, temp, nil
}

// ReadRelativeHumidityMode1 return humidity in percents
// obtained with "Hold Master Mode" command.
// TODO: implement CRC generator polynomial of x 8 + x 5 + x 4 + 1 checkup
func (v *Si7021) ReadRelativeHumidityMode1(i2c *i2c.I2C) (int, error) {
	urh, crc, err := v.ReadUncompRelativeHumidityMode1(i2c)
	if err != nil {
		return 0, nil
	}
	lg.Debugf("RH uncompensated = %v, CRC = %v", urh, crc)
	rh := int(int64(urh)*125/65536) - 6
	return rh, nil
}

// ReadRelativeHumidityMode2 return humidity in percents
// obtained with "No Hold Master Mode" command.
// TODO: implement CRC generator polynomial of x 8 + x 5 + x 4 + 1 checkup
func (v *Si7021) ReadRelativeHumidityMode2(i2c *i2c.I2C) (int, error) {
	urh, crc, err := v.ReadUncompRelativeHumidityMode2(i2c)
	if err != nil {
		return 0, nil
	}
	lg.Debugf("RH uncompensated = %v, CRC = %v", urh, crc)
	rh := int(int64(urh)*125/65536) - 6
	return rh, nil
}

// ReadTemperatureCelsiusMode1 return temprature
// obtained with "Hold Master Mode" command.
// TODO: implement CRC generator polynomial of x 8 + x 5 + x 4 + 1 checkup
func (v *Si7021) ReadTemperatureCelsiusMode1(i2c *i2c.I2C) (float32, error) {
	ut, crc, err := v.ReadUncompTempratureMode1(i2c)
	if err != nil {
		return 0, nil
	}
	lg.Debugf("Temperature uncompensated = %v, CRC = %v", ut, crc)
	temp := float32(math.Round((float64(ut)*175.72/65536-46.85)*math.Pow10(2))) /
		float32(math.Pow10(2))
	return temp, nil
}

// ReadTemperatureCelsiusMode2 return temprature
// obtained with "No Hold Master Mode" command.
// TODO: implement CRC generator polynomial of x 8 + x 5 + x 4 + 1 checkup
func (v *Si7021) ReadTemperatureCelsiusMode2(i2c *i2c.I2C) (float32, error) {
	ut, crc, err := v.ReadUncompTempratureMode2(i2c)
	if err != nil {
		return 0, nil
	}
	lg.Debugf("Temperature uncompensated = %v, CRC = %v", ut, crc)
	temp := float32(math.Round((float64(ut)*175.72/65536-46.85)*math.Pow10(2))) /
		float32(math.Pow10(2))
	return temp, nil
}

// ReadRelativeHumidityAndTemperatureMode1 return humidity in percents
// and temperature
// obtained with "Hold Master Mode" command.
// TODO: implement CRC generator polynomial of x 8 + x 5 + x 4 + 1 checkup
func (v *Si7021) ReadTemperatureAndRelativeHumidityMode1(i2c *i2c.I2C) (int, float32, error) {
	urh, ut, err := v.ReadUncompRelativeHumidityAndTempratureMode1(i2c)
	if err != nil {
		return 0, 0, nil
	}
	lg.Debugf("RH and temperature uncompensated = %v, %v", urh, ut)
	rh := int(int64(urh)*125/65536) - 6
	temp := float32(math.Round((float64(ut)*175.72/65536-46.85)*math.Pow10(2))) /
		float32(math.Pow10(2))
	return rh, temp, nil
}

// ReadRelativeHumidityAndTemperatureMode2 return humidity in percents
// and temperature
// obtained with "No Hold Master Mode" command.
// TODO: implement CRC generator polynomial of x 8 + x 5 + x 4 + 1 checkup
func (v *Si7021) ReadTemperatureAndRelativeHumidityMode2(i2c *i2c.I2C) (int, float32, error) {
	urh, ut, err := v.ReadUncompRelativeHumidityAndTempratureMode2(i2c)
	if err != nil {
		return 0, 0, nil
	}
	lg.Debugf("RH and temperature uncompensated = %v, %v", urh, ut)
	rh := int(int64(urh)*125/65536) - 6
	temp := float32(math.Round((float64(ut)*175.72/65536-46.85)*math.Pow10(2))) /
		float32(math.Pow10(2))
	return rh, temp, nil
}

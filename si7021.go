package si7021

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	i2c "github.com/d2r2/go-i2c"
	"github.com/davecgh/go-spew/spew"
)

// Command byte's sequences
var (
	CMD_REL_HUM_CSE        = []byte{0xE5}       // Measure Relative Humidity, Hold Master Mode (clock stretching enabled)
	CMD_REL_HUM            = []byte{0xF5}       // Measure Relative Humidity, No Hold Master Mode
	CMD_TEMPRATURE_CSE     = []byte{0xE3}       // Measure Temperature, Hold Master Mode (clock stretching enabled)
	CMD_TEMPRATURE         = []byte{0xF3}       // Measure Temperature, No Hold Master Mode
	CMD_TEMP_FROM_PREVIOUS = []byte{0xE0}       // Read Temperature Value from Previous RH Measurement
	CMD_RESET              = []byte{0xFE}       // Reset
	CMD_WRITE_USER_REG_1   = []byte{0xE6}       // Write RH/T User Register 1
	CMD_READ_USER_REG_1    = []byte{0xE7}       // Read RH/T User Register 1
	CMD_WRITE_HEATER_REG   = []byte{0x51}       // Write Heater Control Register
	CMD_READ_HEATER_REG    = []byte{0x11}       // Read Heater Control Register
	CMD_READ_ID_1ST_PART   = []byte{0xFA, 0x0F} // Read Electronic ID 1st Byte
	CMD_READ_ID_2ND_PART   = []byte{0xFC, 0xC9} // Read Electronic ID 2nd Byte
	CMD_READ_FIRMWARE_REV  = []byte{0x84, 0xB8} // Read Firmware Revision
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

// UserRegFlag keep sensor states.
type UserRegFlag byte

const (
	RES_RH_12BIT_TEMP_14BIT UserRegFlag = 0x00 // RH - 12bit, Temperature - 14bit
	RES_RH_8BIT_TEMP_12BIT  UserRegFlag = 0x01 // RH - 8bit, Temperature - 12bit
	RES_RH_10BIT_TEMP_13BIT UserRegFlag = 0x80 // RH - 10bit, Temperature - 13bit
	RES_RH_11BIT_TEMP_11BIT UserRegFlag = 0x81 // RH - 11bit, Temperature - 11bit
	RES_RH_TEMP_MASK        UserRegFlag = 0x81
	HEATER_ENABLED          UserRegFlag = 0x4  // Heater activated
	VOLTAGE_LOW             UserRegFlag = 0x40 // Voltage is lower than 1.9V
)

// String define stringer interface.
func (v UserRegFlag) String() string {
	const divider = " | "
	var buf bytes.Buffer
	if v&HEATER_ENABLED != 0 {
		buf.WriteString("HEATER_ENABLED" + divider)
	}
	if v&VOLTAGE_LOW != 0 {
		buf.WriteString("VOLTAGE_LOW" + divider)
	}
	switch v & RES_RH_TEMP_MASK {
	case RES_RH_10BIT_TEMP_13BIT:
		buf.WriteString("RES_RH_10BIT_TEMP_13BIT" + divider)
	case RES_RH_11BIT_TEMP_11BIT:
		buf.WriteString("RES_RH_11BIT_TEMP_11BIT" + divider)
	case RES_RH_12BIT_TEMP_14BIT:
		buf.WriteString("RES_RH_12BIT_TEMP_14BIT" + divider)
	case RES_RH_8BIT_TEMP_12BIT:
		buf.WriteString("RES_RH_8BIT_TEMP_12BIT" + divider)
	}
	if buf.Len() > 0 {
		buf.Truncate(buf.Len() - len(divider))
	}
	return buf.String()
}

// HeaterLevel define gradation of
// sensor heating. Keep in mind that
// when heater is on, the temprature
// value provided by sensor is not correspond
// to real ambient environment temprature.
type HeaterLevel byte

const (
	HEATER_LEVEL_1    HeaterLevel = 0x0 // Typical power consumption: 3.09 mA
	HEATER_LEVEL_2    HeaterLevel = 0x1 // Typical power consumption: 9.18 mA
	HEATER_LEVEL_3    HeaterLevel = 0x2 // Typical power consumption: 15.24 mA
	HEATER_LEVEL_4    HeaterLevel = 0x3
	HEATER_LEVEL_5    HeaterLevel = 0x4 // Typical power consumption: 27.39 mA
	HEATER_LEVEL_6    HeaterLevel = 0x5
	HEATER_LEVEL_7    HeaterLevel = 0x6
	HEATER_LEVEL_8    HeaterLevel = 0x7
	HEATER_LEVEL_9    HeaterLevel = 0x8 // Typical power consumption: 51.69 mA
	HEATER_LEVEL_10   HeaterLevel = 0x9
	HEATER_LEVEL_11   HeaterLevel = 0xA
	HEATER_LEVEL_12   HeaterLevel = 0xB
	HEATER_LEVEL_13   HeaterLevel = 0xC
	HEATER_LEVEL_14   HeaterLevel = 0xD
	HEATER_LEVEL_15   HeaterLevel = 0xE
	HEATER_LEVEL_16   HeaterLevel = 0xF // Typical power consumption: 94.2 mA
	HEATER_LEVEL_MASK HeaterLevel = 0xF
)

// String define stringer interface.
func (v HeaterLevel) String() string {
	return spew.Sprintf("HEATER_LEVEL_%d", v+1)
}

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
	return fv, nil
}

// Keeps sensor serial number raw bytes
// including CRC info.
type SerialNumberRaw struct {
	SNA3     byte
	CRC_SNA3 byte
	SNA2     byte
	CRC_SNA2 byte
	SNA1     byte
	CRC_SNA1 byte
	SNA0     byte
	CRC_SNA0 byte
	SNB3     byte
	SNB2     byte
	CRC_SNB2 byte
	SNB1     byte
	SNB0     byte
	CRC_SNB0 byte
}

// ReadSerialNumberRaw read sensor serial number to the struct.
func (v *Si7021) ReadSerialNumberRaw(i2c *i2c.I2C) (*SerialNumberRaw, error) {
	lg.Debug("Reading sensor serial number...")
	const bytesCount1stRead = 8
	const bytesCount2ndRead = 6
	buf2 := make([]byte, bytesCount1stRead+bytesCount2ndRead)
	_, err := i2c.WriteBytes(CMD_READ_ID_1ST_PART)
	if err != nil {
		return nil, err
	}
	buf1 := make([]byte, bytesCount1stRead)
	_, err = i2c.ReadBytes(buf1)
	if err != nil {
		return nil, err
	}
	buf2 = append([]byte{}, buf1[0:]...)
	_, err = i2c.WriteBytes(CMD_READ_ID_2ND_PART)
	if err != nil {
		return nil, err
	}
	buf1 = make([]byte, bytesCount2ndRead)
	_, err = i2c.ReadBytes(buf1)
	if err != nil {
		return nil, err
	}
	buf2 = append(buf2, buf1[0:]...)

	buf := bytes.NewBuffer(buf2)
	sn := &SerialNumberRaw{}
	err = binary.Read(buf, binary.BigEndian, sn)
	if err != nil {
		return nil, err
	}
	lg.Debugf("Raw serial number = %v", sn)

	return sn, nil
}

// ReadSerialNumberRaw read sensor serial number to the struct.
func (v *Si7021) ReadSerialNumber(i2c *i2c.I2C) (int64, error) {
	sn, err := v.ReadSerialNumberRaw(i2c)
	if err != nil {
		return 0, err
	}
	crcSna3 := calcCRC_SI7021(0x0, []byte{sn.SNA3})
	crcSna2 := calcCRC_SI7021(crcSna3, []byte{sn.SNA2})
	crcSna1 := calcCRC_SI7021(crcSna2, []byte{sn.SNA1})
	crcSna0 := calcCRC_SI7021(crcSna1, []byte{sn.SNA0})
	crcSnb2 := calcCRC_SI7021(0x0, []byte{sn.SNB3, sn.SNB2})
	crcSnb0 := calcCRC_SI7021(crcSnb2, []byte{sn.SNB1, sn.SNB0})
	if crcSna3 != sn.CRC_SNA3 ||
		crcSna2 != sn.CRC_SNA2 ||
		crcSna1 != sn.CRC_SNA1 ||
		crcSna0 != sn.CRC_SNA0 ||
		crcSnb2 != sn.CRC_SNB2 ||
		crcSnb0 != sn.CRC_SNB0 {
		err := errors.New(spew.Sprintf(
			"Some CRCs equalities are not valid: %v = %v, %v = %v,"+
				"%v = %v, %v = %v, %v = %v, %v = %v",
			sn.CRC_SNA3, crcSna3, sn.CRC_SNA2, crcSna2,
			sn.CRC_SNA1, crcSna1, sn.CRC_SNA0, crcSna0,
			sn.CRC_SNB2, crcSnb2, sn.CRC_SNB0, crcSnb0))
		return 0, err
	}
	sn2 := int64(sn.SNA3)<<56 + int64(sn.SNA2)<<48 + int64(sn.SNA1)<<40 +
		int64(sn.SNA0)<<32 + int64(sn.SNB3)<<24 + int64(sn.SNB2)<<16 +
		int64(sn.SNB1)<<8 + int64(sn.SNB0)
	return sn2, nil
}

// ReadSensorType return sensor model.
func (v *Si7021) ReadSensoreType(i2c *i2c.I2C) (SensorType, error) {
	sn, err := v.ReadSerialNumberRaw(i2c)
	if err != nil {
		return 0, err
	}
	st := (SensorType)(sn.SNB3)
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
func (v *Si7021) SetMeasureResolution(i2c *i2c.I2C, res UserRegFlag) error {
	lg.Debug("Setting measure resolution...")
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return err
	}
	ur = ur&(^byte(RES_RH_TEMP_MASK)) | (byte)(res)
	v.lastUserReg = &ur
	_, err = i2c.WriteBytes(append(CMD_WRITE_USER_REG_1, ur))
	return err
}

// GetMeasureResolution read current sensor measure accuracy.
func (v *Si7021) GetMeasureResolution(i2c *i2c.I2C) (UserRegFlag, error) {
	v.lastUserReg = nil
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return 0, err
	}
	return (UserRegFlag)(ur) & RES_RH_TEMP_MASK, nil
}

// SetHeaterStatus enable of disable internal heater.
func (v *Si7021) SetHeaterStatus(i2c *i2c.I2C, enableHeater bool) error {
	lg.Debug("Setting heater on/off...")
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return err
	}
	ur = ur & (^byte(HEATER_ENABLED))
	if enableHeater {
		ur = ur | byte(HEATER_ENABLED)
	}
	v.lastUserReg = &ur
	_, err = i2c.WriteBytes(append(CMD_WRITE_USER_REG_1, ur))
	return err
}

// GetHeaterStatus return heater status: on (true) or off (false).
func (v *Si7021) GetHeaterStatus(i2c *i2c.I2C) (bool, error) {
	lg.Debug("Getting heater status...")
	v.lastUserReg = nil
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return false, err
	}
	return (UserRegFlag)(ur)&HEATER_ENABLED != 0, nil
}

// GetVoltageStatus provide power supply voltage low: low (true) or OK (false).
func (v *Si7021) GetVoltageLow(i2c *i2c.I2C) (bool, error) {
	lg.Debug("Getting voltage low status...")
	v.lastUserReg = nil
	ur, err := v.readUserReg(i2c)
	if err != nil {
		return false, err
	}
	return (UserRegFlag)(ur)&VOLTAGE_LOW != 0, nil
}

// SetHeaterLevel define internal heater
// heating gradation. Remeber, when heater is on
// temprature provided by sensor is not correspond
// to real ambient temprature.
func (v *Si7021) SetHeaterLevel(i2c *i2c.I2C, level HeaterLevel) error {
	lg.Debug("Setting heater level...")
	var hcr byte
	hcr = (byte)(level)
	_, err := i2c.WriteBytes(append(CMD_WRITE_HEATER_REG, hcr))
	return err
}

// GetHeaterLevel return sensor heating gradation.
func (v *Si7021) GetHeaterLevel(i2c *i2c.I2C) (HeaterLevel, error) {
	lg.Debug("Getting heater level...")
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

// Reset reboot a sensor.
func (v *Si7021) Reset(i2c *i2c.I2C) error {
	lg.Debug("Reset sensor...")
	_, err := i2c.WriteBytes(CMD_RESET)
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

	const dataBytesCount = 2
	const crcBytesCount = 1
	if withCRC {
		var data struct {
			Data [2]byte
			CRC  byte
		}
		err := readDataToStruct(i2c, dataBytesCount+crcBytesCount, binary.BigEndian, &data)
		if err != nil {
			return 0, 0, err
		}
		calcCRC := calcCRC_SI7021(0x0, data.Data[:dataBytesCount])
		if data.CRC != calcCRC {
			err := errors.New(spew.Sprintf(
				"CRCs doesn't match: CRC from sensor (0x%0X) != calculated CRC (0x%0X)",
				data.CRC, calcCRC))
			return 0, 0, err
		} else {
			lg.Debugf("CRCs verified: CRC from sensor (0x%0X) = calculated CRC (0x%0X)",
				data.CRC, calcCRC)
		}
		return getU16BE(data.Data[:dataBytesCount]), data.CRC, nil
	} else {
		var data struct {
			Data [2]byte
		}
		err := readDataToStruct(i2c, dataBytesCount, binary.BigEndian, &data)
		if err != nil {
			return 0, 0, err
		}
		return getU16BE(data.Data[:dataBytesCount]), 0, nil
	}
}

// ReadUncompHumidity returns uncompensated humidity and CRC.
func (v *Si7021) ReadUncompHumidity(i2c *i2c.I2C) (uint16, byte, error) {
	lg.Debug("Reading uncompensated humidity...")
	rh, crc, err := v.doMeasure(i2c, CMD_REL_HUM, true)
	return rh, crc, err
}

// ReadUncompTemperature returns uncompensated temperature and CRC.
func (v *Si7021) ReadUncompTemprature(i2c *i2c.I2C) (uint16, byte, error) {
	lg.Debug("Reading uncompensated temprature...")
	temp, crc, err := v.doMeasure(i2c, CMD_TEMPRATURE, true)
	return temp, crc, err
}

func (v *Si7021) uncompHumidityToRelativeHumidity(uh uint16) float32 {
	rh := float32(uh)*125/65536 - 6
	rh2 := round32(rh, 2)
	return rh2
}

func (v *Si7021) uncompTemperatureToCelsius(ut uint16) float32 {
	temp := float32(ut)*175.72/65536 - 46.85
	temp2 := round32(temp, 2)
	return temp2
}

// ReadUncompHumidityAndTemperature returns
// uncompensated humidity, temperature and CRC.
func (v *Si7021) ReadUncompHumidityAndTemprature(i2c *i2c.I2C) (uint16, uint16, error) {
	lg.Debug("Reading uncompensated humidity and temperature...")
	rh, _, err := v.doMeasure(i2c, CMD_REL_HUM, true)
	if err != nil {
		return 0, 0, err
	}
	temp, _, err := v.doMeasure(i2c, CMD_TEMP_FROM_PREVIOUS, false)
	if err != nil {
		return 0, 0, err
	}
	return rh, temp, nil
}

// ReadRelativeHumidity return relative humidity.
func (v *Si7021) ReadRelativeHumidity(i2c *i2c.I2C) (float32, error) {
	urh, _, err := v.ReadUncompHumidity(i2c)
	if err != nil {
		return 0, err
	}
	rh := v.uncompHumidityToRelativeHumidity(urh)
	return rh, nil
}

// ReadTemperature return temprature.
func (v *Si7021) ReadTemperature(i2c *i2c.I2C) (float32, error) {
	ut, _, err := v.ReadUncompTemprature(i2c)
	if err != nil {
		return 0, err
	}
	temp := v.uncompTemperatureToCelsius(ut)
	return temp, nil
}

// ReadRelativeHumidityAndTemperature return
// relative humidity and temperature.
func (v *Si7021) ReadRelativeHumidityAndTemperature(i2c *i2c.I2C) (float32, float32, error) {
	urh, ut, err := v.ReadUncompHumidityAndTemprature(i2c)
	if err != nil {
		return 0, 0, err
	}
	rh := v.uncompHumidityToRelativeHumidity(urh)
	temp := v.uncompTemperatureToCelsius(ut)
	return rh, temp, nil
}

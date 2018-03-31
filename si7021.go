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

type FirmwareVersion byte

const (
	FIRMWARE_VER_1_0 FirmwareVersion = 0xFF
	FIRMWARE_VER_2_0 FirmwareVersion = 0x20
)

func (v FirmwareVersion) String() string {
	switch v {
	case FIRMWARE_VER_1_0:
		return "Firmware version 1.0"
	case FIRMWARE_VER_2_0:
		return "Firmware version 2.0"
	default:
		return "<unknown>"
	}
}

type SensorType byte

const (
	SI_ENGINEERING_TYPE1 SensorType = 0x00
	SI_ENGINEERING_TYPE2 SensorType = 0xFF
	SI_7013_TYPE         SensorType = 0x0D
	SI_7020_TYPE         SensorType = 0x14
	SI_7021_TYPE         SensorType = 0x15
)

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

type MeasureResolution byte

const (
	RH_12BIT_TEMP_14BIT MeasureResolution = 0x00
	RH_8BIT_TEMP_12BIT                    = 0x01
	RH_10BIT_TEMP_13BIT                   = 0x80
	RH_11BIT_TEMP_11BIT                   = 0x81
	RH_TEMP_MASK                          = 0x81
)

type HeaterControl byte

const (
	HEATER_DISABLED = 0x00
	HEATER_ENABLED  = 0x40
	HEATER_MASK     = 0x40
)

type HeaterLevel byte

const (
	HEATER_LEVEL_1  HeaterLevel = 0x0
	HEATER_LEVEL_2              = 0x1
	HEATER_LEVEL_3              = 0x2
	HEATER_LEVEL_4              = 0x3
	HEATER_LEVEL_5              = 0x4
	HEATER_LEVEL_6              = 0x5
	HEATER_LEVEL_7              = 0x6
	HEATER_LEVEL_8              = 0x7
	HEATER_LEVEL_9              = 0x8
	HEATER_LEVEL_10             = 0x9
	HEATER_LEVEL_11             = 0xA
	HEATER_LEVEL_12             = 0xB
	HEATER_LEVEL_13             = 0xC
	HEATER_LEVEL_14             = 0xD
	HEATER_LEVEL_15             = 0xE
	HEATER_LEVEL_16             = 0xF
)

type Si7021 struct {
	lastUserReg *byte
}

func NewSi7021() *Si7021 {
	v := &Si7021{}
	return v
}

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
	lg.Debugf("Firmware version = %[0]s(%[0]d)", fv)

	return fv, nil
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

func (v *Si7021) ReadSensoreType(i2c *i2c.I2C) (SensorType, error) {
	buf, err := v.ReadSerialNumber(i2c)
	if err != nil {
		return 0, err
	}
	st := (SensorType)(buf[5])
	lg.Debugf("Sensor type = %[0]s(%[0]d)", st)

	return st, nil
}

func (v *Si7021) SetMeasureResolution(i2c *i2c.I2C, res MeasureResolution) error {
	var ur byte
	if v.lastUserReg != nil {
		ur = *v.lastUserReg
	}
	ur = ur&HEATER_MASK | (byte)(res)
	v.lastUserReg = &ur
	_, err := i2c.WriteBytes(append(CMD_WRITE_USER_REG_1, ur))
	return err
}

func (v *Si7021) GetMeasureResolution(i2c *i2c.I2C) (MeasureResolution, error) {
	_, err := i2c.WriteBytes(CMD_READ_USER_REG_1)
	if err != nil {
		return 0, err
	}
	buf1 := make([]byte, 1)
	_, err = i2c.ReadBytes(buf1)
	if err != nil {
		return 0, err
	}
	ur := buf1[0]
	v.lastUserReg = &ur
	ur = ur & RH_TEMP_MASK
	return (MeasureResolution)(ur), nil
}

func (v *Si7021) SetHeaterControl(i2c *i2c.I2C, heat HeaterControl) error {
	var ur byte
	if v.lastUserReg != nil {
		ur = *v.lastUserReg
	}
	ur = ur&RH_TEMP_MASK | (byte)(heat)
	v.lastUserReg = &ur
	_, err := i2c.WriteBytes(append(CMD_WRITE_USER_REG_1, ur))
	return err
}
func (v *Si7021) GetHeaterControl(i2c *i2c.I2C) (HeaterControl, error) {
	_, err := i2c.WriteBytes(CMD_READ_USER_REG_1)
	if err != nil {
		return 0, err
	}
	buf1 := make([]byte, 1)
	_, err = i2c.ReadBytes(buf1)
	if err != nil {
		return 0, err
	}
	ur := buf1[0]
	v.lastUserReg = &ur
	ur = ur & HEATER_MASK
	return (HeaterControl)(ur), nil
}

func (v *Si7021) SetHeaterLevel(i2c *i2c.I2C, level HeaterLevel) error {
	var hcr byte
	hcr = (byte)(level)
	_, err := i2c.WriteBytes(append(CMD_WRITE_HEATER_REG, hcr))
	return err
}

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
	hcr := buf1[0]
	return (HeaterLevel)(hcr), nil
}

func (v *Si7021) Reset(i2c *i2c.I2C) error {
	_, err := i2c.WriteBytes([]byte{0xFE})
	return err
}

func (v *Si7021) doMeasure(i2c *i2c.I2C, cmd []byte) (uint16, error) {
	_, err := i2c.WriteBytes(CMD_REL_HUM_MASTER_MODE)
	if err != nil {
		return 0, err
	}
	buf1 := make([]byte, 2)
	_, err = i2c.ReadBytes(buf1)
	if err != nil {
		return 0, err
	}
	meas := uint16(buf1[0])<<8 | uint16(buf1[1])
	return meas, nil
}

func (v *Si7021) ReadUncompRelativeHumidityMode1(i2c *i2c.I2C) (uint16, error) {
	rh, err := v.doMeasure(i2c, CMD_REL_HUM_MASTER_MODE)
	return rh, err
}

func (v *Si7021) ReadUncompRelativeHumidityMode2(i2c *i2c.I2C) (uint16, error) {
	rh, err := v.doMeasure(i2c, CMD_REL_HUM_NO_MASTER_MODE)
	return rh, err
}

func (v *Si7021) ReadUncompTempratureMode1(i2c *i2c.I2C) (uint16, error) {
	temp, err := v.doMeasure(i2c, CMD_TEMPRATURE_MASTER_MODE)
	return temp, err
}

func (v *Si7021) ReadUncompTempratureMode2(i2c *i2c.I2C) (uint16, error) {
	temp, err := v.doMeasure(i2c, CMD_TEMPRATURE_NO_MASTER_MODE)
	return temp, err
}

func (v *Si7021) ReadUncompRelativeHumidityAndTempratureMode1(i2c *i2c.I2C) (uint16, uint16, error) {
	rh, err := v.doMeasure(i2c, CMD_REL_HUM_MASTER_MODE)
	if err != nil {
		return 0, 0, err
	}
	temp, err := v.doMeasure(i2c, CMD_TEMP_FROM_PREVIOUS)
	if err != nil {
		return 0, 0, err
	}
	return rh, temp, nil
}

func (v *Si7021) ReadUncompRelativeHumidityAndTempratureMode2(i2c *i2c.I2C) (uint16, uint16, error) {
	rh, err := v.doMeasure(i2c, CMD_REL_HUM_NO_MASTER_MODE)
	if err != nil {
		return 0, 0, err
	}
	temp, err := v.doMeasure(i2c, CMD_TEMP_FROM_PREVIOUS)
	if err != nil {
		return 0, 0, err
	}
	return rh, temp, nil
}

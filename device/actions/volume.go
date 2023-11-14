package actions

import (
	"fmt"
	"strconv"
	"strings"
)

/*
2023/03/02 16:40:13.020914 Setting mute on "zoneOut1" to false
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:13.149921 Set mute
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:13.153742 Setting mute on "zoneOut2" to false
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:13.592596 Set mute
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:18.783588 Setting mute on "zoneOut1" to true
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:18.856657 Set mute
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:18.858791 Setting mute on "zoneOut2" to true
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:19.355018 Set mute
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:21.741883 Setting mute on "zoneOut1" to false
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:21.845413 Set mute
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:21.849429 Setting mute on "zoneOut2" to false
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:22.342521 Set mute
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:24.528289 Setting volume on "zoneOut1" to 35
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:24.599127 Set volume
[ITB-1106-SW1.byu.edu] 2023/03/02 16:40:24.603355 Setting volume on "zoneOut2" to 35

Response format:
{"volume":10}
{"muted":false}  (the variable comes in as either a 1/0 or true/false)


gain 60 - zoneOut1 - VOL x (0-100, sta=status) VOUTMute x (on, off, sta)
62 - zoneOut1, zoneOut2 - VOUTx y (x(output#)=1, 2, 3, 4; y= -90-10, sta=status) VOUTMute x (on, off, sta)
52 - Out1 - VOUTx y (x(output#)=1; y= -80-15, sta=status) VOUTMute x (on, off, sta)
*/

type Volume struct {
	Volume int `json:"volume"`
}

type Mute struct {
	Muted bool `json:"muted"`
}

// gain 60 - zoneOut1 - VOUTMute x - x=(on, off)
// 62 - zoneOut1, zoneOut2 - VOUTMutex y x(output#)=(1, 2, 3, 4)  y=(on, off)
// 52 - Out1 - VOUTMute x - x=(on, off)
func SetMute(address string, output string, status bool, device string) (Mute, error) {
	var state Mute
	port := "23"
	cmd := ""
	parseCmd := ""

	muteCMD := "off"
	if status {
		muteCMD = "on"
	}

	switch device {
	case "AT-UHD-SW-52ED":
		cmd = "VOUTMute1 " + muteCMD + "\r"
		parseCmd = "VOUTMute1"
	case "AT-OME-PS62":
		switch {
		case strings.Contains(output, "1"):
			cmd = "VOUTMute1 " + muteCMD + "\r" + "VOUTMute3 " + muteCMD + "\r"
			parseCmd = "VOUTMute1"
		case strings.Contains(output, "2"):
			cmd = "VOUTMute2 " + muteCMD + "\r" + "VOUTMute4 " + muteCMD + "\r"
			parseCmd = "VOUTMute2"
		default:
			err := fmt.Errorf("invalid output for volume. expected to contain a 1 or 2: %s", output)
			return state, err
		}
	case "AT-GAIN-60":
		cmd = "VOUTMute " + muteCMD + "\r"
		parseCmd = "VOUTMute"
	default:
		err := fmt.Errorf("invalid device. expected AT-UHD-SW-52ED, AT-OME-PS62, or AT-GAIN-60: %s", output)
		return state, err
	}

	resp, err := sendCommand(address, port, []byte(cmd))
	if err != nil {
		return state, err
	}
	respMute, err := parseMuteResponse(resp, output, parseCmd)
	if err != nil {
		return state, err
	}
	state.Muted = respMute

	return state, nil
}

// gain 60 - zoneOut1 - VOUTMute sta
// 62 - zoneOut1, zoneOut2 - VOUTMutex sta - x(output#)=(1, 2, 3, 4)
// 52 - Out1 - VOUTMute sta
func GetMute(address string, output string, device string) (Mute, error) {
	var state Mute
	port := "23"
	cmd := ""
	parseCmd := ""
	switch device {
	case "AT-UHD-SW-52ED":
		cmd = "VOUTMute1 sta\r"
		parseCmd = "VOUTMute1"
	case "AT-OME-PS62":
		switch {
		case strings.Contains(output, "1"):
			cmd = "VOUTMute1 sta\r"
			parseCmd = "VOUTMute1"
		case strings.Contains(output, "2"):
			cmd = "VOUTMute2 sta\r"
			parseCmd = "VOUTMute2"
		default:
			err := fmt.Errorf("invalid output for volume. expected to contain a 1 or 2: %s", output)
			return state, err
		}
	case "AT-GAIN-60":
		cmd = "VOUTMute sta\r"
		parseCmd = "VOUTMute"
	default:
		err := fmt.Errorf("invalid device. expected AT-UHD-SW-52ED, AT-OME-PS62, or AT-GAIN-60: %s", output)
		return state, err
	}

	resp, err := sendCommand(address, port, []byte(cmd))
	if err != nil {
		return state, err
	}
	respMute, err := parseMuteResponse(resp, output, parseCmd)
	if err != nil {
		return state, err
	}
	state.Muted = respMute
	return state, nil
}

// gain 60 - zoneOut1 - VOL x (0-100)
// 62 - zoneOut1, zoneOut2 - VOUTx y - x(output#)=(1, 2, 3, 4; y= -90-10, sta=status)
// 52 - Out1 - VOUT1 y (y= -80-15, sta=status)
func SetVolume(address string, output string, volume string, device string) (Volume, error) {
	var level Volume

	vol := convertVolume(volume, device)
	port := "23"
	cmd := ""
	parseCmd := ""
	switch device {
	case "AT-UHD-SW-52ED":
		cmd = "VOUT1 " + vol + "\r"
		parseCmd = "VOUT1"
	case "AT-OME-PS62":
		switch {
		case strings.Contains(output, "1"):
			cmd = "VOUT1 " + vol + "\r" + "VOUT3 " + vol + "\r"
			parseCmd = "VOUT1"
		case strings.Contains(output, "2"):
			cmd = "VOUT2 " + vol + "\r" + "VOUT4 " + vol + "\r"
			parseCmd = "VOUT2"
		default:
			err := fmt.Errorf("invalid output for volume. expected to contain a 1 or 2: %s", output)
			return level, err
		}
	case "AT-GAIN-60":
		cmd = "VOL " + vol + "\r"
		parseCmd = "VOL"
	default:
		err := fmt.Errorf("invalid device. expected AT-UHD-SW-52ED, AT-OME-PS62, or AT-GAIN-60: %s", output)
		return level, err
	}

	resp, err := sendCommand(address, port, []byte(cmd))
	if err != nil {
		return level, err
	}

	respLevel, err := parseVolumeResponse(resp, output, parseCmd)
	if err != nil {
		return level, err
	}
	stringVol := strconv.Itoa(respLevel)
	tmpVol := convertReceiveVolume(stringVol, device)
	level.Volume, err = strconv.Atoi(tmpVol)
	if err != nil {
		return level, err
	}
	return level, nil
}

// gain 60 - zoneOut1 - VOL sta
// 62 - zoneOut1, zoneOut2 - VOUTx sta (x(output#)=1, 2, 3, 4)
// 52 - Out1 - VOUT1 sta
func GetVolume(address string, output string, device string) (Volume, error) {
	var level Volume
	port := "23"
	cmd := ""
	parseCmd := ""
	switch device {
	case "AT-UHD-SW-52ED":
		cmd = "VOUT1 sta\r"
		parseCmd = "VOUT1"
	case "AT-OME-PS62":
		switch {
		case strings.Contains(output, "1"):
			cmd = "VOUT1 sta\r"
			parseCmd = "VOUT1"
		case strings.Contains(output, "2"):
			cmd = "VOUT2 sta\r"
			parseCmd = "VOUT2"
		default:
			err := fmt.Errorf("invalid output for volume. expected to contain a 1 or 2: %s", output)
			return level, err
		}
	case "AT-GAIN-60":
		cmd = "VOL sta\r"
		parseCmd = "VOL"
	default:
		err := fmt.Errorf("invalid device. expected AT-UHD-SW-52ED, AT-OME-PS62, or AT-GAIN-60: %s", output)
		return level, err
	}

	resp, err := sendCommand(address, port, []byte(cmd))
	if err != nil {
		return level, err
	}

	respLevel, err := parseVolumeResponse(resp, output, parseCmd)
	if err != nil {
		return level, err
	}
	stringVol := strconv.Itoa(respLevel)
	tmpVol := convertReceiveVolume(stringVol, device)
	level.Volume, err = strconv.Atoi(tmpVol)
	if err != nil {
		return level, err
	}

	return level, nil
}

// ****************************************************************************Helper functions
func parseVolumeResponse(resp []byte, output string, parseCmd string) (input int, err error) {
	responses := strings.Split(string(resp), "\r\n")
	responseContainsOut := false
	for _, value := range responses {
		fmt.Println("Slice: ", value)
		if len(value) > 3 {
			responseContainsOut = strings.Contains(string(value), parseCmd)
		} else {
			continue
		}
		if responseContainsOut {
			v := strings.Split(string(value), " ")
			input, err = strconv.Atoi(v[1])
			if err != nil {
				return input, err
			}
			return input, nil
		} else {
			err = fmt.Errorf("invalid volume response: %s", resp)
			continue
		}
	}
	return input, err
}

func parseMuteResponse(resp []byte, output string, parseCmd string) (mute bool, err error) {
	responses := strings.Split(string(resp), "\r\n")
	responseContainsCMD := false

	for _, value := range responses {
		//fmt.Println("Slice: ", value)
		if len(value) > 5 {
			responseContainsCMD = strings.Contains(strings.ToLower(string(value)), strings.ToLower(parseCmd))
		} else {
			continue
		}
		if responseContainsCMD {
			v := strings.Split(string(value), " ")
			state := v[1]
			state = strings.ToLower(state)
			//fmt.Println("state: ", state)
			if err != nil {
				return mute, err
			}

			if state == "on" {
				mute = true
			} else if state == "off" {
				mute = false
			} else {
				err = fmt.Errorf("response not in expected range (\"on\" or \"off\"): %s", string(resp))
				return mute, err
			}
			return mute, nil
		} else {
			err = fmt.Errorf("invalid mute response: %s", resp)
			continue
		}
	}
	return mute, err
}

func convertVolume(volume string, device string) string {
	vtmp, err := strconv.Atoi(volume)
	v := float64(vtmp) //make a float64 for accuracy otherwise 50 returns 49
	if v > 100 {
		v = 100
	}
	if v < 1 {
		v = 0
	}
	outMax := 100.0
	outMin := 0.0
	devHi := 100.0
	devLo := 0.0
	mutedLevel := 0.0

	if err != nil {
		fmt.Println(err)
		return "0"
	}

	switch device {
	case "AT-UHD-SW-52ED": //range -80 - 15
		devHi = 0
		devLo = -50
		mutedLevel = -80

	case "AT-OME-PS62": //range -90 - 10
		devHi = 0
		devLo = -50
		mutedLevel = -90
	case "AT-GAIN-60": //range   0 - 100
		devHi = 100
		devLo = 50
		mutedLevel = 0
	}

	vol := ((devHi-devLo)*(v-outMin))/(outMax-outMin) + devLo
	if v < 1 {
		vol = mutedLevel
	}

	volToSend := int(vol)
	return fmt.Sprint(volToSend)
}

func convertReceiveVolume(volume string, device string) string {
	vtmp, err := strconv.Atoi(volume)
	v := float64(vtmp) //make a float64 for accuracy otherwise 50 returns 49

	outMax := 100.0
	outMin := 0.0
	devHi := 100.0
	devLo := 0.0
	if err != nil {
		return "0"
	}

	switch device {
	case "AT-UHD-SW-52ED": //range -80 - 15
		devHi = 0
		devLo = -50
	case "AT-OME-PS62": //range -90 - 10
		devHi = 0
		devLo = -50
	case "AT-GAIN-60": //range   0 - 100
		devHi = 100
		devLo = 50
	}
	vol := ((outMax-outMin)*(v-devLo))/(devHi-devLo) + outMin
	volToSend := int(vol)
	if volToSend > 100 {
		volToSend = 100
	}
	if volToSend < 1 {
		volToSend = 0
	}
	return fmt.Sprint(volToSend)
}

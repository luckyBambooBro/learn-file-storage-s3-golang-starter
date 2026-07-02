package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"fmt"
	
)

type FFProbeData struct {
	Streams []struct {
		Width              int    `json:"width,omitempty"`
		Height             int    `json:"height,omitempty"`
	} `json:"streams"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	//run ffprobe program and unmarshal data
	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-print_format",
		"json",
		"-show_streams",
		filePath)

	buf := &bytes.Buffer{}
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return "", err
	}
	stdoutBytes := buf.Bytes()

	v := &FFProbeData{}
	if err := json.Unmarshal(stdoutBytes, v); err != nil {
		return "", err
	}

	if len(v.Streams) == 0 {
		return "", fmt.Errorf("no streams found")
	}

	//determine and return asp ratio
	width, height := v.Streams[0].Width, v.Streams[0].Height

	if 1.6 < width/height < 1.8 {
		return "16:9", nil
	} else if width/height <= 1 {
		return "9:16", nil
	} else {
		return "other", nil
	}

}
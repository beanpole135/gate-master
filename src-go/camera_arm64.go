//go:build arm64

package main

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

type Camera struct {
	Frames    <-chan []byte
	CamDevice *device.Device
	err       error
}

type CamConfig struct {
	Device      string `json:"device"`
	PixelFormat string `json:"pixel_format"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

func (cc *CamConfig) PixFmt() v4l2.FourCCType {
	switch strings.ToLower(cc.PixelFormat) {
	case "rgb24":
		return v4l2.PixelFmtRGB24
	case "grey":
		return v4l2.PixelFmtGrey
	case "yuyv":
		return v4l2.PixelFmtYUYV
	case "yyuv":
		return v4l2.PixelFmtYYUV
	case "yvyu":
		return v4l2.PixelFmtYVYU
	case "uyvy":
		return v4l2.PixelFmtUYVY
	case "vyuy":
		return v4l2.PixelFmtVYUY
	case "mjpeg":
		return v4l2.PixelFmtMJPEG
	case "jpeg":
		return v4l2.PixelFmtJPEG
	case "mpeg":
		return v4l2.PixelFmtMPEG
	case "h264":
		return v4l2.PixelFmtH264
	case "mpeg4":
		return v4l2.PixelFmtMPEG4
	}
	return v4l2.PixelFmtH264
}

const (
	devWidth  = 640
	devHeight = 480
)

func NewCamera(cc CamConfig) (*Camera, error) {
	C := Camera{}
	//Now initialize the camera
	var err error
	C.CamDevice, err = device.Open(
		cc.Device,
		device.WithPixFormat(v4l2.PixFormat{PixelFormat: cc.PixFmt(), Width: cc.Width, Height: cc.Height}),
	)
	if err != nil {
		C.err = err
		return &C, fmt.Errorf("Could not open camera device: %w", err)
	}
	err = C.CamDevice.Start(context.TODO())
	if err != nil {
		C.err = err
		return &C, fmt.Errorf("Could not start camera: %w", err)
	}
	C.Frames = C.CamDevice.GetOutput()
	fmt.Println("Initialized Camera")
	return &C, nil
}

func (C *Camera) Close() {
	C.CamDevice.Close()
}

func (C *Camera) ServeImages(w http.ResponseWriter, req *http.Request, p *Page) {
	if C.Frames == nil || C.err != nil {
		http.Error(w, C.err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("Serving images")
	mimeWriter := multipart.NewWriter(w)
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", mimeWriter.Boundary()))
	partHeader := make(textproto.MIMEHeader)
	partHeader.Add("Content-Type", "image/jpeg")

	var frame []byte
	for frame = range C.Frames {
		partWriter, err := mimeWriter.CreatePart(partHeader)
		if err != nil {
			log.Printf("failed to create multi-part writer: %s", err)
			return
		}

		if _, err := partWriter.Write(frame); err != nil {
			log.Printf("failed to write image: %s", err)
		}
	}
}

func (C *Camera) TakePicture() []byte {
	if C.Frames == nil {
		return []byte{}
	}
	// A single picture is just one frame from the current stream
	for frame := range C.Frames {
		return frame
	}
	return nil
}

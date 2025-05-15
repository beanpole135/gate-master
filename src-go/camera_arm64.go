//go:build arm64

package main

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/cleroux/go-rpicamvid"
)

type Camera struct {
	//CamDevice *device.Device
	err    error
	Webcam *rpicamvid.Rpicamvid
}

type CamConfig struct {
	//Device      string `json:"device"`
	//PixelFormat string `json:"pixel_format"`
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
}

func NewCamera(cc CamConfig) (*Camera, error) {
	C := Camera{}
	// Now initialize the camera
	l := log.New(os.StdOut, "", log.LstdFlags)
	C.webcam = rpicamvid.New(l, cc.Width, cc.Height)
	// Start it up and see if it is working (then close it down again)
	stream, err = C.webcam.Start()
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to start camera service:", err))
		C.err = err
	} else {
		fmt.Println("Initialized Camera")
		defer stream.Close()
	}

	return &C, err
}

func (C *Camera) processVideo() {
	for {
		if ok := C.webcam.Read(&C.img); ok {
			fmt.Println("Cannot read video device - stopping")
			return
		}
		if C.img.Empty() {
			continue
		}
		//Perform processing as needed

		//Now send to channel (as JPEG Image in bytes)
		buffer, err := gocv.IMEncodeWithParams(gocv.JPEGFileExt, C.img, nil)
		if err != nil {
			fmt.Println("Could not encode image as JPEG:", err)
			return
		}
		C.Frames <- buffer.GetBytes()
	}
}

func (C *Camera) Close() {
	C.img.Close()
	//C.CamDevice.Close()
}

func (C *Camera) ServeImages(w http.ResponseWriter, req *http.Request, p *Page) {
	if C.Webcam == nil || C.err != nil {
		http.Error(w, C.err.Error(), http.StatusBadRequest)
		return
	}
	C.Webcam.HttpHandler(w, req)
}

func (C *Camera) TakePicture() []byte {
	if C.Webcam == nil || C.err != nil {
		return []byte{}
	}
	// A single picture is just one frame from the current stream
	stream, err := C.Webcam.Start()
	if err != nil {
		fmt.Println("Unable to start camera:", err)
		return []byte{}
	}
	defer stream.Close()
	fr, err2 := stream.GetFrame()
	if err2 != nil {
		fmt.Println("Unable to get camera frame:", err2)
		return []byte{}
	}
	defer fr.Close()
	return fr.GetBytes()
}

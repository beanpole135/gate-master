//go:build arm64

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cleroux/go-rpicamvid"
)

type Camera struct {
	//CamDevice *device.Device
	err    error
	webcam *rpicamvid.Rpicamvid
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
	stream, err := C.webcam.Start()
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to start camera service:", err))
		C.err = err
	} else {
		fmt.Println("Initialized Camera")
		defer stream.Close()
	}

	return &C, err
}

func (C *Camera) Close() {
}

func (C *Camera) ServeImages(w http.ResponseWriter, req *http.Request, p *Page) {
	if C.webcam == nil || C.err != nil {
		http.Error(w, C.err.Error(), http.StatusBadRequest)
		return
	}
	C.webcam.HttpHandler(w, req)
}

func (C *Camera) TakePicture() []byte {
	if C.webcam == nil || C.err != nil {
		return []byte{}
	}
	// A single picture is just one frame from the current stream
	stream, err := C.webcam.Start()
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

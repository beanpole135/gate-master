//go:build arm
package main

import(
	"fmt"
	"net/http"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

type Camera struct {
	Frames <-chan []byte
	CamDevice *device.Device
}

const (
	devName = "/dev/video0"
	devWidth = 640
	devHeight = 480
)


func NewCamera() (*Camera, error) {
	C := Camera{}
	//Now initialize the camera 
	var err error
	C.CamDevice, err = device.Open(
	    devName,
	    device.WithPixFormat(v4l2.PixFormat{PixelFormat: v4l2.PixelFmtMJPEG, Width: devWidth, Height: devHeight}),
	)
	if err != nil {
		return nil, fmt.Errorf("Could not open camera device: %w", err)
	}
	err = C.CamDevice.Start(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("Could not start camera: %w", err)
	}
	C.Frames = C.CamDevice.GetOutput()
	return &C, nil
}

func (C *Camera) Close() {
	C.CamDevice.Close()
}

func (C *Camera) ServeImages(w http.ResponseWriter, req *http.Request) {
	mimeWriter := multipart.NewWriter(w)
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", mimeWriter.Boundary()))
	partHeader := make(textproto.MIMEHeader)
	partHeader.Add("Content-Type", "image/jpeg")

	var frame []byte
	for frame = range frames {
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
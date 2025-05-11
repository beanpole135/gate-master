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

	//"github.com/vladimirvivien/go4vl/device"
	//"github.com/vladimirvivien/go4vl/v4l2"
	"gocv.io/x/gocv"
)

type Camera struct {
	Frames chan []byte
	//CamDevice *device.Device
	err    error
	webcam *gocv.VideoCapture
	img    gocv.Mat
}

type CamConfig struct {
	Device      string `json:"device"`
	PixelFormat string `json:"pixel_format"`
	Width       uint32 `json:"width"`
	Height      uint32 `json:"height"`
}

/*func (cc *CamConfig) PixFmt() v4l2.FourCCType {
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
	fmt.Println("Using MJPEG Camera Format by default")
	return v4l2.PixelFmtMJPEG
}*/

func NewCamera(cc CamConfig) (*Camera, error) {
	C := Camera{}
	//Now initialize the camera
	/*var err error
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
	C.Frames = C.CamDevice.GetOutput()*/

	//Version using gocv (OpenCV)
	dev := strings.TrimPrefix(cc.Device, "video")
	devnum, err := strconv.Atoi(dev)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to parse camera device (%s): using device number 0", cc.Device))
		devnum = 0
	}
	C.webcam, err = gocv.OpenVideoCapture(devnum)
	if err != nil {
		C.err = err
		fmt.Println("Unable to open video capture device:", devnum)
	}
	C.img = gocv.NewMat()
	C.Frames = make(chan []byte)
	go C.processVideo()
	fmt.Println("Initialized Camera")
	return &C, nil
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
	if C.Frames == nil || C.err != nil {
		http.Error(w, C.err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("Serving images")
	mimeWriter := multipart.NewWriter(w)
	defer mimeWriter.Close()
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", mimeWriter.Boundary()))
	partHeader := make(textproto.MIMEHeader)
	partHeader.Add("Content-Type", "image/jpeg")

	var frame []byte
	for frame = range C.Frames {
		if len(frame) == 0 {
			continue //skip empty frame
		}
		//Create the writer
		partWriter, err := mimeWriter.CreatePart(partHeader)
		if err != nil {
			fmt.Printf("failed to create multi-part writer: %s", err)
			return
		}
		// Write the frame (image)
		if _, err := partWriter.Write(frame); err != nil {
			fmt.Printf("failed to write image: %s", err)
			break
		}
		time.Sleep(20 * time.Millisecond) //20ms = 30 frames per second
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

//go:build arm64

package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"syscall"

	"github.com/cleroux/go-rpicamvid"
	"github.com/disintegration/imaging"
)

type Camera struct {
	//CamDevice *device.Device
	err      error
	webcam   *rpicamvid.Rpicamvid
	rotation int `json:"-"` //internal tag for 90 degree rotations
}

type CamConfig struct {
	Rotation int `json:"rotation"`
	Width    int `json:"width"`
	Height   int `json:"height"`
}

func NewCamera(cc CamConfig) (*Camera, error) {
	C := Camera{}
	// Now initialize the camera
	l := log.New(os.Stdout, "", log.LstdFlags)
	var opts []string
	if cc.Rotation >= 180 {
		opts = append(opts, "--rotation", fmt.Sprintf("%d", 180)) //This flag only supports 0 or 180 degrees
		C.rotation = cc.Rotation - 180
	} else {
		C.rotation = cc.Rotation
	}

	C.webcam = rpicamvid.New(l, cc.Width, cc.Height, opts...)
	// Start it up and see if it is working (then close it down again)
	stream, err := C.webcam.Start()
	if err != nil {
		fmt.Printf("Unable to start camera service: %v", err)
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
	C.serveHttp(w, req)
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
	return C.processImage(fr.GetBytes())
}

func (C *Camera) processImage(frame []byte) []byte {
	if C.rotation == 90 {
		img, _, err := image.Decode(bytes.NewReader(frame))
		if err == nil {
			img := imaging.Rotate90(img)
			buf := new(bytes.Buffer)
			err = jpeg.Encode(buf, img, nil)
			if err != nil {
				fmt.Println("Error encoding jpeg image (image not rotated 90 degrees):", err)
				C.rotation = 0 //Skip this logic next time - it will not work
			} else {
				frame = buf.Bytes()
			}
		} else {
			fmt.Println("Error decoding jpeg image (image not rotated 90 degrees):", err)
			C.rotation = 0 //Skip this logic next time - it will not work
		}
	}

	return frame
}

func (C *Camera) serveHttp(w http.ResponseWriter, req *http.Request) {
	// This was a blanket copy of the go-rpicamvid function called "HTTPHandler" (in http.go) - 5/16/2025
	// Just so that we could inject the "processImage()" function above into it for additional 90-degree rotations
	stream, err := C.webcam.Start()
	if err != nil {
		fmt.Printf("Failed to start camera: %v\n", err)
		http.Error(w, "Failed to start camera: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	mimeWriter := multipart.NewWriter(w)
	defer mimeWriter.Close()
	contentType := fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", mimeWriter.Boundary())
	w.Header().Set("Content-Type", contentType)

	partHeader := make(textproto.MIMEHeader, 1)
	partHeader.Add("Content-Type", "image/jpeg")

	ctx := req.Context()

	for {
		if ctx.Err() != nil {
			return
		}

		err := func() error {
			f, err := stream.GetFrame()
			if err != nil {
				fmt.Printf("Failed to get camera frame: %v\n", err)
				return nil // continue for loop
			}
			defer f.Close()

			partWriter, err := mimeWriter.CreatePart(partHeader)
			if err != nil {
				fmt.Printf("Failed to create multi-part section: %v\n", err)
				return err
			}

			if _, err := partWriter.Write(C.processImage(f.GetBytes())); err != nil {
				if errors.Is(err, syscall.EPIPE) {
					// Client went away
					return err
				}

				switch err.Error() {
				case "http2: stream closed", "client disconnected":
					// Client went away
					return err
				}
				fmt.Printf("Failed to write video frame: %v\n", err)
				return err
			}
			return nil
		}()
		if err != nil {
			break
		}
	}
}

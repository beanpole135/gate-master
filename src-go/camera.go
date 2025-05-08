//go:build !arm64

package main

import (
	"net/http"
)

type Camera struct {
}

type CamConfig struct {
	Device      string `json:"device"`
	PixelFormat string `json:"pixel_format"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

func NewCamera(cc CamConfig) (*Camera, error) {
	C := Camera{}
	//Template - for testing build on Windows only
	return &C, nil
}

func (C *Camera) Close() {

}

func (C *Camera) ServeImages(w http.ResponseWriter, req *http.Request, p *Page) {

}

func (C *Camera) TakePicture() []byte {
	return nil
}

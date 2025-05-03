//go:build !arm

package main

import (
	"net/http"
)

type Camera struct {
}

const (
	devWidth  = 640
	devHeight = 480
)

func NewCamera(devname string) (*Camera, error) {
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

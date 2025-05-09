//go:build !arm64

package main

type GateConfig struct {
	GpioPin int `json:"gpio_num"`
}

func (gc *GateConfig) OpenGate() {

}

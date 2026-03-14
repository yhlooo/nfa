package otter

import (
	"embed"
	"fmt"
	"image"

	"github.com/saran13raj/go-pixels"
)

// imagesFS 内嵌的图片文件
//
//go:embed *.png
var imagesFS embed.FS

// GetOtterImage 获取水獭图片
func GetOtterImage(color, bg bool) (image.Image, error) {
	// 确定图片名
	name := "otter-"
	if color {
		name += "color-"
	} else {
		name += "white-"
	}
	if !bg {
		name += "nobg-"
	}
	name += "x4.png"

	// 打开图片
	f, err := imagesFS.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open image %q error: %w", name, err)
	}
	defer func() { _ = f.Close() }()

	// 加载图片
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image %q error: %w", name, err)
	}

	return img, nil
}

// Otter 获取由 unicode 块字符生成的水獭图片
func Otter(color, bg bool, scale int) (string, error) {
	// 获取图片
	img, err := GetOtterImage(color, bg)
	if err != nil {
		return "", err
	}

	if scale < 1 {
		return "", fmt.Errorf("scale must be greater than 0")
	}

	// 确定大小和渲染方式，初始高度 20
	renderType := "halfcell"
	if scale%2 == 0 {
		renderType = "fullcell"
		scale /= 2 // 全格渲染自动变为 2 倍大
	}
	height := 20 * scale

	return gopixels.FromImageStream(img, 0, height, renderType, true)
}

// MustOtter 获取由 unicode 块字符生成的水獭图片
func MustOtter(color, bg bool, scale int) string {
	ret, _ := Otter(color, bg, scale)
	return ret
}

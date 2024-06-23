package screen

import (
	"image"
	"image/color"

	"github.com/google/uuid"
	"golang.org/x/image/draw"
)

type outputProcessor struct {
	outputId uuid.UUID
	nodeId   uuid.UUID
	sinkId   uuid.UUID
	leds     int

	reverse   bool
	scaler    draw.Scaler
	subWidth  int
	subHeight int
	subLeft   int
	subTop    int
	dest      *image.RGBA
}

func newOutputProcessor(cfg OutputConfig, width, height int) *outputProcessor {
	orientation := cfg.Config.Orientation

	var dst *image.RGBA

	switch orientation {
	case LinkSchemaJsonOrientationLandscape, LinkSchemaJsonOrientationLandscapeReverse:
		dst = image.NewRGBA(image.Rect(0, 0, cfg.Leds, 1))
	case LinkSchemaJsonOrientationPortrait, LinkSchemaJsonOrientationPortraitReverse:
		dst = image.NewRGBA(image.Rect(0, 0, 1, cfg.Leds))
	}

	var reverse bool
	switch orientation {
	case LinkSchemaJsonOrientationLandscapeReverse, LinkSchemaJsonOrientationPortraitReverse:
		reverse = true
	}

	return &outputProcessor{
		outputId: cfg.OutputId,
		nodeId:   cfg.NodeId,
		sinkId:   cfg.SinkId,
		leds:     cfg.Leds,

		reverse:   reverse,
		scaler:    draw.BiLinear.NewScaler(width/8, height/8, cfg.Config.Width, cfg.Config.Height),
		subWidth:  cfg.Config.Width,
		subHeight: cfg.Config.Height,
		subLeft:   cfg.Config.Left,
		subTop:    cfg.Config.Top,
		dest:      dst,
	}
}

func (out *outputProcessor) process(thumb *image.NRGBA) []color.NRGBA {
	rect := image.Rect(out.subLeft/8, out.subTop/8, (out.subLeft+out.subWidth)/8, (out.subTop+out.subHeight)/8)

	sub := thumb.SubImage(rect)

	out.scaler.Scale(out.dest, out.dest.Bounds(), sub, sub.Bounds(), draw.Over, nil)

	var colors []color.NRGBA

	for i := 0; i < len(out.dest.Pix); i += 4 {
		clr := color.NRGBA{
			R: out.dest.Pix[i],
			G: out.dest.Pix[i+1],
			B: out.dest.Pix[i+2],
			A: out.dest.Pix[i+3],
		}

		colors = append(colors, clr)
	}

	if out.reverse {
		reverse(colors)
	}

	return colors
}

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

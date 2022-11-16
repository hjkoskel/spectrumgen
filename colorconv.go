package spectrumgen

import (
	"fmt"
	"image/color"
	"math"
)

func colorToIntensity(c color.RGBA) float64 {
	return (0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)) / 255.0
}

const (
	REDWEIGHT   float64 = 0.299
	GREENWEIGHT float64 = 0.587
	BLUEWEIGHT  float64 = 0.114
)

//Intensity back to color (openCV BGR2GRAY compatible)
func intensityToColor_BGR2GRAY(intensity float64) color.RGBA {
	//RGB[A] to Gray:Y←0.299⋅R+0.587⋅G+0.114⋅B
	// Green have most weigh -> overflows first
	//g := math.Floor(math.Max(0, math.Min(255, intensity*255/0.587))) //Green only at first
	g := math.Floor(math.Max(0, math.Min(255, intensity*255/GREENWEIGHT))) //Green only at first

	//iNow := intensity - 0.587*g/255.0 //Remove what was left for green
	iFor_r_b := intensity - GREENWEIGHT*g/255.0 //Remove what was left for green

	//Take as much as red needed
	//r := math.Floor(math.Max(0, math.Min(255, iNow*255/0.299)))
	r := math.Floor(math.Max(0, math.Min(255, iFor_r_b*255/REDWEIGHT)))
	iFor_b := iFor_r_b - REDWEIGHT*r/255.0 //What was left for blue

	b := math.Floor(math.Max(0, math.Min(255, iFor_b*255/BLUEWEIGHT)))

	return color.RGBA{
		R: byte(r),
		G: byte(g),
		B: byte(b),
		A: 255,
	}
}

func ColorToIntensity(c color.RGBA) float64 {
	return float64(c.R)*REDWEIGHT/255 + float64(c.G)*GREENWEIGHT/255 + float64(c.B)*BLUEWEIGHT/255
}

//Trick: pick WavelengthToRGB color and trim so sum is matching?
func intensityToColor_BGR2GRAY_withwavelength(nm float64, intensity float64) color.RGBA {
	result := WavelengthToRGB(nm, intensity)
	errIntensity := colorToIntensity(result) - intensity
	//TODO better ideas? Match sum, intensity and wavelength
	for i := 0; i < 10; i++ {
		errIntensity = colorToIntensity(result) - intensity
		errR := errIntensity * REDWEIGHT
		errG := errIntensity * GREENWEIGHT
		errB := errIntensity * BLUEWEIGHT

		result.R -= uint8(errR * 255)
		result.G -= uint8(errG * 255)
		result.B -= uint8(errB * 255)
	}

	errIntensityAfter := colorToIntensity(result) - intensity

	fmt.Printf("error before %v  after %v\n", errIntensity, errIntensityAfter)

	return result
}

var outOfRangecolor = color.RGBA{R: 0, G: 0, B: 0, A: 255}

//Intensity is 0 to 1.0
func WavelengthToRGB(nm float64, intensity float64) color.RGBA {

	if intensity <= 0 {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}

	//return color.RGBA{R: byte(255 * intensity), G: byte(255 * intensity), B: byte(255 * intensity), A: 255}

	gamma := 0.8
	result := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	r := float64(0)
	g := float64(0)
	b := float64(0)

	if nm < 380 || 780 < nm { //Out of range
		return outOfRangecolor
	}

	switch {
	case nm <= 440:
		r = -(nm - 440) / (440 - 380)
		g = 0.0
		b = 1.0
	case nm < 490:
		r = 0.0
		g = (nm - 440) / (490 - 440)
		b = 1.0
	case nm <= 510:
		r = 0.0
		g = 1.0
		b = -(nm - 510) / (510 - 490)
	case nm < 580:
		r = (nm - 510) / (580 - 510)
		g = 1.0
		b = 0.0
	case nm <= 645:
		r = 1.0
		g = -(nm - 645) / (645 - 580)
		b = 0.0
	default:
		r = 1.0
		g = 0.0
		b = 0.0
	}

	factor := float64(0)
	//calc factor
	switch {
	case nm <= 420:
		factor = 0.3 + 0.7*(nm-380)/(420-380)
	case nm <= 700:
		factor = 1.0
	case nm <= 780:
		factor = 0.3 + 0.7*(780-nm)/(780-700)
	}

	result.R = byte(math.Min(255, intensity*255*math.Pow(r*factor, gamma))) //TODO MAX NOT NEEDED?
	result.G = byte(math.Min(255, intensity*255*math.Pow(g*factor, gamma)))
	result.B = byte(math.Min(255, intensity*255*math.Pow(b*factor, gamma)))

	//display no color as gray

	if int(result.R)+int(result.G)+int(result.B) == 0 {
		fmt.Printf("OUT OF RANGE %v, result=%#v\n", intensity, result)
		return outOfRangecolor
	}

	return result
}

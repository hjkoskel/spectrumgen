/*
frequency data
*/
package spectrumgen

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
)

type FrequencyData struct {
	NmStart   float64
	NmStop    float64
	Intensity []float64
}

func (p *FrequencyData) ToCsv() string {
	var sb strings.Builder

	delta := (p.NmStop - p.NmStart) / float64(len(p.Intensity))
	for index, intensity := range p.Intensity {
		sb.WriteString(fmt.Sprintf("%f\t%f\n", p.NmStart+delta*float64(index), intensity))
	}
	return sb.String()

}

func fakeData() FrequencyData {
	arr := make([]float64, 128)

	for i := range arr {
		x := float64(i)/float64(len(arr)) - 0.5
		x = x * 10
		arr[i] = math.Exp(-x * x)
	}

	fmt.Printf("INTENSITYT %#v\n", arr)

	return FrequencyData{
		NmStart: 340, NmStop: 660,
		Intensity: arr,
	}
}

func (p *FrequencyData) AddPeak(peak PeakModel) error {
	if len(p.Intensity) == 0 {
		return fmt.Errorf("Intensity data not allocated")
	}
	_, newdata, errEval := peak.Eval(p.NmStart, p.NmStop, len(p.Intensity))
	if errEval != nil {
		return errEval
	}

	for i, d := range newdata {
		p.Intensity[i] += d
	}
	return nil
}

func (p *FrequencyData) CreateImage(xreso int, yreso int, rgbmode bool) *image.RGBA {
	result := image.NewRGBA(image.Rect(0, 0, xreso, yreso))
	var col color.RGBA
	for x := 0; x < xreso; x++ {
		ratio := float64(x) / float64(xreso)
		intensity := p.Intensity[int(ratio*float64(len(p.Intensity)))]
		f := p.NmStart + ratio*(p.NmStop-p.NmStart)

		if rgbmode {
			col = WavelengthToRGB(f, intensity)
		} else {
			col = intensityToColor_BGR2GRAY(intensity)
		}
		for y := 0; y < yreso; y++ {
			result.SetRGBA(x, y, col)
		}
	}
	return result
}

type FrequencyDataArr []FrequencyData

func CalcTestTriangle(steps int, startNm float64, endNm float64, points int) (FrequencyDataArr, error) {
	if points < 1 {
		return FrequencyDataArr{}, fmt.Errorf("invalid number of points %v", points)
	}
	if steps < 1 {
		return FrequencyDataArr{}, fmt.Errorf("invalid number of steps %v", steps)
	}
	result := make([]FrequencyData, steps)
	data := FrequencyData{NmStart: startNm, NmStop: endNm, Intensity: make([]float64, points)}
	for i := range data.Intensity {
		data.Intensity[i] = 1 - 2*math.Abs(float64(points/2-i))/float64(points)
	}
	for step := range result {
		result[step] = data
	}
	return result, nil
}

func CalcFrequencyDataArr(model PeakModelFile, steps int, startNm float64, endNm float64, points int) (FrequencyDataArr, error) {
	result := make([]FrequencyData, steps)

	for step := range result {
		progress := float64(0)
		if 1 < steps {
			progress = float64(step) / float64(steps-1)
		}
		peakmodels := model.GetPeakModels(progress)
		data := FrequencyData{NmStart: startNm, NmStop: endNm, Intensity: make([]float64, points)}
		for peakindex, peakmodel := range peakmodels {
			fmt.Printf("PEAK %v: %s\n", peakindex, peakmodel.String())

			errEval := data.AddPeak(peakmodel)
			if errEval != nil {
				return result, fmt.Errorf("error evaluating (progress %f) peak%v err=%s", progress, peakindex, errEval.Error())
			}
		}
		result[step] = data
	}
	return result, nil
}

//Grid, for plotting. (assume same dimension and nm axis)
func (p *FrequencyDataArr) ToGrid() (string, error) {
	if len(*p) == 0 {
		return "", fmt.Errorf("no data")
	}
	first := (*p)[0]

	var sb strings.Builder
	for ndata, fdata := range *p {
		if first.NmStart != fdata.NmStart || first.NmStop != fdata.NmStop {
			return "", fmt.Errorf("All data array items must have same nm range %v does not have", ndata)
		}
		if len(first.Intensity) != len(fdata.Intensity) {
			return "", fmt.Errorf("All data array intensity arrays must have same lengths %v does not have (%v vs %v)", ndata, len(first.Intensity), len(fdata.Intensity))
		}
		sb.WriteString(floatArrToString(fdata.Intensity, NUMBEROFDECIMALS) + "\n")
	}
	return sb.String(), nil
}

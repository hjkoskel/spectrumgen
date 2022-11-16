/*
Peak model

Model peaks to signal

https://github.com/dreading/gospecfunc

voigt?

https://cefrc.princeton.edu/sites/g/files/toruqf1071/files/Files/2013%20Lecture%20Notes/Hanson/pLecture6.pdf


https://en.wikipedia.org/wiki/Full_width_at_half_maximum
*/

package spectrumgen

import (
	"fmt"
	"math"
	"strings"
)

const (
	PEAKMODEL_GAUSS   = 0
	PEAKMODEL_LORENTZ = 1
	PEAKMODEL_HYPSEC  = 2
)

type PeakModelType int

func parsePeakModelType(s string) PeakModelType {
	switch strings.ToUpper(s) {
	case "GAUSS", "GAUSSIAN", "G":
		return PEAKMODEL_GAUSS
	case "LORENTZ", "L":
		return PEAKMODEL_LORENTZ
	case "HYPSEC":
		return PEAKMODEL_HYPSEC
	}
	return -1
}

func (a PeakModelType) String() string {
	result, haz := map[PeakModelType]string{PEAKMODEL_GAUSS: "GAUSS", PEAKMODEL_LORENTZ: "LORENTZ", PEAKMODEL_HYPSEC: "HYPSEC"}[a]
	if !haz {
		return "UNDEFINED"
	}
	return result
}

type PeakModel struct { //Load from config json?
	Type     PeakModelType
	Peak     float64
	Position float64
	Fwhm     float64 //Full width half maximum, puoliarvonleveys in finnish
}

type PeakModelArr []PeakModel

func (a PeakModel) String() string {
	return fmt.Sprintf("%s %.4f@%.2fnm±%.3f", a.Type.String(), a.Peak, a.Position, a.Fwhm/2)
}

func (a PeakModelArr) String() string {
	var sb strings.Builder
	for _, peak := range a {
		sb.WriteString(peak.String() + "\n")
	}
	return sb.String()
}

//TODO optimize calculation. Pre-calc shapes and scale?
func (p *PeakModel) Eval(from float64, to float64, n int) ([]float64, []float64, error) {
	xData := make([]float64, n) //nm axis if needed for plots etc..
	for i := range xData {
		xData[i] = from + (to-from)*float64(i)/float64(n)
	}

	yData := make([]float64, n)

	switch p.Type {
	case PEAKMODEL_GAUSS:
		stdgauss := p.Fwhm / (2 * math.Sqrt(2*math.Log(2)))
		stdgauss2 := stdgauss * stdgauss
		for i, nm := range xData {
			x := nm - p.Position
			yData[i] = p.Peak * math.Exp(-0.5*(x*x)/stdgauss2)
		}
	case PEAKMODEL_LORENTZ:
		for i, nm := range xData {
			x := nm - p.Position
			yData[i] = p.Peak / (1 + (2*x/p.Fwhm)*(2*x/p.Fwhm))
		}
	case PEAKMODEL_HYPSEC:
		hypsecScale := 2 * math.Log(2+math.Sqrt(3)) / p.Fwhm
		for i, nm := range xData {
			x := nm - p.Position
			yData[i] = p.Peak / math.Cosh(x*hypsecScale)
		}
	}
	return xData, yData, nil
}

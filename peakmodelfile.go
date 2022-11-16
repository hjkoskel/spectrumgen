/*
Peak model file
Text formatted file, describes how peakmodels are created
*/
package spectrumgen

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type PeakModelFileRow struct {
	Positions   []float64
	Type        PeakModelType
	Parameters  []float64 //Usually 1, depends on type. can be more at future
	Intencities []float64
}

type PeakModelFile []PeakModelFileRow

//Helper function
func LoadPeakModelFile(filename string) (PeakModelFile, error) {
	byt, readerr := os.ReadFile(filename)
	if readerr != nil {
		return PeakModelFile{}, fmt.Errorf("Error reading %v\n", readerr.Error())
	}

	fmt.Printf("-----%s------\n%s\n----------------\n", filename, byt)

	return ParsePeakModelFile(byt)
}

const COMMENTSYMBOL = "#"

//
func ParsePeakModelFile(content []byte) (PeakModelFile, error) {
	rows := strings.Split(string(content), "\n")
	result := []PeakModelFileRow{}
	for rownumber, rawrow := range rows {
		commentPos := strings.Index(rawrow, COMMENTSYMBOL)
		if 0 <= commentPos {
			rawrow = rawrow[0:commentPos]
		}

		rawrow = strings.TrimSpace(rawrow)

		if 0 == len(rawrow) { //skip empty lines and comment lines
			continue
		}

		parsed, parseErr := ParsePeakModelRow(rawrow)
		if parseErr != nil {
			return PeakModelFile{}, fmt.Errorf("fail at row %v err=%v", rownumber, parseErr.Error())
		}
		result = append(result, parsed)
	}
	return result, nil
}

func (a PeakModelFile) ToOutputFormat() (string, error) {
	var sb strings.Builder
	for rownumber, rowdata := range a {
		s, errOut := rowdata.ToOutputFormat()
		if errOut != nil {
			return "", fmt.Errorf("Invalid peakdata index %v err %v", rownumber, errOut.Error())
		}
		sb.WriteString(s + "\n")
	}
	return sb.String(), nil
}

func floatArrToString(arr []float64, decimals int) string {
	var sb strings.Builder

	fstring := fmt.Sprintf("\t%%.%vf", decimals)
	for _, v := range arr {
		sb.WriteString(fmt.Sprintf(fstring, v))
	}
	return sb.String()
}

const NUMBEROFDECIMALS = 6

func (p *PeakModelFileRow) ToOutputFormat() (string, error) {
	if p.Type != PEAKMODEL_GAUSS && p.Type != PEAKMODEL_HYPSEC && p.Type != PEAKMODEL_LORENTZ {
		return "", fmt.Errorf("Invalid peak type")
	}
	if len(p.Parameters) != 1 { //all require FWHM now
		return "", fmt.Errorf("Invalid number of parameters (%d)", len(p.Parameters))
	}

	if len(p.Intencities) < 1 {
		return "", fmt.Errorf("no intensity data")
	}

	sPositions := floatArrToString(p.Positions, NUMBEROFDECIMALS)
	return sPositions[1:] + "\t" + p.Type.String() + floatArrToString(p.Parameters, NUMBEROFDECIMALS) + floatArrToString(p.Intencities, NUMBEROFDECIMALS), nil
}

func calcFrame(progressRatio float64, n int) (int, float64) {
	f := float64(n+1) * progressRatio
	f = math.Max(0, f)
	f = math.Min(float64(n)-1, f)

	result := int(math.Floor(f))
	return result, f - float64(result) //Use last parameter for interpolation
}

func (p *PeakModelFile) GetPeakModels(progressRatio float64) []PeakModel {
	result := []PeakModel{}
	for _, m := range *p {
		result = append(result, m.GetPeakModel(progressRatio))
	}
	return result
}

func (p *PeakModelFileRow) GetPeakModel(progressRatio float64) PeakModel {
	positionFrame, posInterp := calcFrame(progressRatio, len(p.Positions))
	intensityFrame, intensityInterp := calcFrame(progressRatio, len(p.Intencities))

	nextPositionFrame := positionFrame + 1
	if len(p.Positions) <= nextPositionFrame {
		nextPositionFrame = len(p.Positions) - 1
	}

	nextIntensityFrame := intensityFrame + 1
	if len(p.Intencities) <= nextIntensityFrame {
		nextIntensityFrame = len(p.Intencities) - 1
	}

	result := PeakModel{
		Type:     p.Type,
		Peak:     p.Intencities[intensityFrame]*(1-posInterp) + p.Intencities[nextIntensityFrame]*posInterp,
		Position: p.Positions[positionFrame]*(1-intensityInterp) + p.Positions[nextPositionFrame]*intensityInterp,
		Fwhm:     p.Parameters[0],
	}
	return result
}

func ParsePeakModelRow(row string) (PeakModelFileRow, error) {
	cols := strings.Fields(row)
	/*
		positions
		TYPE string... break here
		parameters (now1, later by type)
		intensities
	*/
	result := PeakModelFileRow{
		Positions:   []float64{},
		Type:        PEAKMODEL_GAUSS,
		Parameters:  []float64{},
		Intencities: []float64{},
	}

	//Get positions and type
	typePos := 0
	for i, s := range cols {
		f, parseErr := strconv.ParseFloat(s, 64)
		if parseErr != nil {
			result.Type = parsePeakModelType(s)
			if result.Type < 0 {
				return result, fmt.Errorf("invalid formatted row %s", row)
			}
			typePos = i
			break
		}
		result.Positions = append(result.Positions, f)
	}
	//Get parameter(s)  (now only 1)
	cols = cols[typePos+1:]
	//fmt.Printf("typePos=%v  cols=%#v\n", typePos, cols)
	if len(cols) < 1 {
		return result, fmt.Errorf("invalid formatted (no parameters or intensity)  row %s", row)
	}
	f, parseErr := strconv.ParseFloat(cols[0], 64)
	if parseErr != nil {
		return result, fmt.Errorf("invalid formatted (invalid par %v) row %s", parseErr.Error(), row)
	}
	result.Parameters = []float64{f}
	cols = cols[1:]
	for _, s := range cols {
		if s == COMMENTSYMBOL {
			break
		}
		f, parseErr := strconv.ParseFloat(s, 64)
		if parseErr != nil {
			return result, fmt.Errorf("invalid formatted (intensities %v) row %s", parseErr.Error(), row)
		}
		result.Intencities = append(result.Intencities, f)
	}
	if len(result.Intencities) == 0 {
		return result, fmt.Errorf("no intensity points row %s", row)
	}
	return result, nil
}

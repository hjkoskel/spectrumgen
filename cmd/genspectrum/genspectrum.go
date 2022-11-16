/*
Generate spectrum data with this command line utility

./genspectrum -? gives help

Create video

Then install  v4l-utils, ffmpeg etc...

- ls /dev/video*, check existing video devices
- Load loopback video: sudo modprobe v4l2loopback

Then
ffmpeg -loop 1 -re -i testpicture.png -f v4l2 -vcodec rawvideo -pix_fmt yuv420p /dev/video2

ffmpeg -stream_loop 10 -re -i koe.mp4 -f v4l2 -vcodec rawvideo -pix_fmt yuv420p /dev/video2
-------------------
*/

package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"math/rand"
	"os"

	vidio "github.com/AlexEidt/Vidio"
	"github.com/hjkoskel/spectrumgen"
)

func imageToPng(filename string, img image.Image) error {
	out, errCreateOut := os.Create(filename)
	if errCreateOut != nil {
		return fmt.Errorf("err creating %v debugfile %v", filename, errCreateOut.Error())
	}
	errEncode := png.Encode(out, img)
	if errEncode != nil {
		return fmt.Errorf("error png-encode debugfile %v err=%v", filename, errEncode.Error())
	}
	closeErr := out.Close()
	if closeErr != nil {
		return fmt.Errorf("closing %v error %v", filename, closeErr.Error())
	}
	return nil
}

func randomizeByte(b byte, amplitude float64) byte {
	return byte(math.Min(255, math.Max(0, float64(b)+rand.NormFloat64()*amplitude)))
}

func addNoiseToImage(img *image.RGBA, amplitude float64) {
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = randomizeByte(img.Pix[i], amplitude)
		img.Pix[i+1] = randomizeByte(img.Pix[i+1], amplitude)
		img.Pix[i+2] = randomizeByte(img.Pix[i+2], amplitude)
	}
}

func main() {
	pNmStart := flag.Float64("a", 380, "shortest wavelength in nm")
	pNmStop := flag.Float64("b", 750, "longest wavelength in nm")

	pDuration := flag.Float64("d", 10, "duration of render in seconds")
	pFps := flag.Float64("fps", 15, "frames per second")
	pFrames := flag.Int("fr", -1, "override duration by giving number of frames required, 0=infinite loop")

	//Filenames
	pPeaksimFile := flag.String("i", "", "peaksim file for input")
	pOutAnimation := flag.String("o", "", "output animation")
	pOutPic := flag.String("pic", "", "filename prefix for image output (png)")
	pOutCsv := flag.String("c", "", "csv file output, one row per frame")

	//Settings
	pXreso := flag.Int("xreso", 800, "x resolution in pixels")
	pYreso := flag.Int("yreso", 600, "y resolution in pixels")

	pRgb := flag.Bool("rgb", false, "set RGB mode and simulate colors instead of BGR2GRAY")

	//Noise generator
	pNoiseAmp := flag.Float64("na", 0, "noise amplitude") //TODO RGB value vs value?

	//Test triangle for testing color<->intensity conversion
	pTestTriangle := flag.Bool("testtriangle", false, "enable test triangle mode instead of peaks")

	flag.Parse()

	steps := int((*pDuration) * (*pFps))
	if 0 < *pFrames {
		steps = *pFrames
	}

	var errCalc error
	var dataArr spectrumgen.FrequencyDataArr

	if *pTestTriangle {
		dataArr, errCalc = spectrumgen.CalcTestTriangle(steps, *pNmStart, *pNmStop, *pXreso)
	} else {
		//Load and initialize sim
		modelfile, errLoadModel := spectrumgen.LoadPeakModelFile(*pPeaksimFile)
		if errLoadModel != nil {
			fmt.Printf("error loading %s\n", errLoadModel)
			return
		}

		modelDump, _ := modelfile.ToOutputFormat()
		fmt.Printf("\n\n%s\n", modelDump)

		dataArr, errCalc = spectrumgen.CalcFrequencyDataArr(modelfile, steps, *pNmStart, *pNmStop, *pXreso)
	}
	if errCalc != nil {
		fmt.Printf("failed calculating %v\n", errCalc.Error())
		return
	}

	if 0 < len(*pOutCsv) {
		gridData, gridDataErr := dataArr.ToGrid()
		if gridDataErr != nil {
			fmt.Printf("creating csv fail %v\n", gridDataErr.Error())
		}

		wErr := os.WriteFile(*pOutCsv, []byte(gridData), 0666)
		if wErr != nil {
			fmt.Printf("csv write err %v\n", wErr.Error())
		}
	}

	if 0 < len(*pOutAnimation) {

		vw, vwErr := vidio.NewVideoWriter(*pOutAnimation, *pXreso, *pYreso, &vidio.Options{
			Loop: 0,
			FPS:  *pFps,
		})
		if vwErr != nil {
			fmt.Printf("video writer error %v\n", vwErr.Error())
			return
		}

		for _, data := range dataArr {
			img := data.CreateImage(*pXreso, *pYreso, *pRgb)
			addNoiseToImage(img, *pNoiseAmp)
			videoWriteError := vw.Write(img.Pix)
			if videoWriteError != nil {
				fmt.Printf("vide writing err %v\n", videoWriteError.Error())
				return
			}
		}
		vw.Close()
	}

	if 0 < len(*pOutPic) {
		for i, data := range dataArr {
			picName := fmt.Sprintf("%s_%v.png", *pOutPic, i)
			imgErr := imageToPng(picName, data.CreateImage(*pXreso, *pYreso, *pRgb))
			if imgErr != nil {
				fmt.Printf("Error saving image %v err=%v\n", picName, imgErr.Error())
				return
			}
		}
	}

}

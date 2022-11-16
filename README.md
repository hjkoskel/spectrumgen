# Spectrum generator (light)
This package is for creating fake test light spectrums for projects like [PySpectrometer2](https://github.com/leswright1977/PySpectrometer2)
or for acting as unit test utility for golang based light spectrometer

# Usage

cmd/genspectrum is example program for generating spectrums.

build with
```sh
go build
```
that folder contains .peaksim files like

- movingPeaks.peaksim
- onenarrow.peaksim, example of one narrow peak
- onewide.peaksim

```sh
./genspectrum -fr 1 -i fluoro.peaksim -o test.mp4 -pic test -rgb
```
produces fake one frame animation and test_0.png picture


## Faking webcam

Install v4l-utils, ffmpeg etc...

check video devices present at the moment. As reference
```sh
ls /dev/video*
```

Load loopback video device
```sh
sudo modprobe v4l2loopback
```
then check is there new video device available
```sh
ls /dev/video*
```

in this example it is */dev/video2*

Playing single picture
```sh
ffmpeg -loop 1 -re -i test_0.png -f v4l2 -vcodec rawvideo -pix_fmt yuv420p /dev/video2
```

# command line options

```
Usage of ./genspectrum:
  -a float
    	shortest wavelength in nm (default 380)
  -b float
    	longest wavelength in nm (default 750)
  -c string
    	csv file output, one row per frame
  -d float
    	duration of render in seconds (default 10)
  -fps float
    	frames per second (default 15)
  -fr int
    	override duration by giving number of frames required, 0=infinite loop (default -1)
  -i string
    	peaksim file for input
  -na float
    	noise amplitude
  -o string
    	output animation
  -pic string
    	filename prefix for image output (png)
  -rgb
    	set RGB mode and simulate colors instead of BGR2GRAY
  -testtriangle
    	enable test triangle mode instead of peaks
  -xreso int
    	x resolution in pixels (default 800)
  -yreso int
    	y resolution in pixels (default 600)
```



# peaksim file format

Simulator generates animated peak patters based on .peaksim file.
Peaksimfile is tab/space separated file with following columns

- 1st to N,  list of positions (nm) durin animation separated by tab or space
- Separator token telling **TYPE** of peak
  - GAUSS, GAUSSIAN or G for gaussian peak
  - LORENTZ L or PEAKMODEL_LORENTZ
  - HYPSEC, for hypsec
- List of parameters depending on **TYPE** of peak
  - Now all have only 1 parameter: "Full width half maximum", ("puoliarvonleveys" in finnish)
- List of intensities during animation

The symbol # acts as line comment starter

![allpeakstyles](/doc/allpeaks.png)

# Issues and future features

This software is updated while experimenting with spectroscopy

## Intensity & wavelength vs RGB values

Mapping intensity vs RGB color is not straightforward
One way to map RGB value to intensity is to use equation where values are weighted
```math
I=0.299*R + 0.587*G + 0.114*B
```

But function WavelengthToRGB( (copied from pyspectrometer) does not produce spectrums where intensity matches to intensity calculated by "R,G,B weighted average"

Better equation that produces "weighted average" results and looks good is now TBD.

Use -rgb command line option to choose mode where spectrum looks like light spectrum, instead of green-yellow thing.  
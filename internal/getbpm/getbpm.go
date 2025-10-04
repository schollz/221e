package getbpm

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-audio/wav"
)

func GetBPM(name string) (beats float64, bpm float64, err error) {
	beats, bpm, err = parseName(name)
	nonSixteenBeats := math.Mod(beats, 16) != 0
	if err != nil || bpm < 100 || bpm > 200 || nonSixteenBeats {
		beats, bpm, err = guessBPM(name)
	}
	return
}

func parseName(name string) (beats float64, bpm float64, err error) {
	_, fname := filepath.Split(name)
	fname = strings.ToLower(fname)
	rBeats, _ := regexp.Compile(`\w+[beats](\d+)`)
	rBPM, _ := regexp.Compile(`\w+[bpm]([0-9]+)`)
	rBPM2 := regexp.MustCompile("[0-9]+")
	foo := rBPM.FindStringSubmatch(fname)
	duration, _, _, err := Length(name)
	if err != nil {
		return
	}

	if len(foo) < 2 {
		err = fmt.Errorf("could not find bpm: %s", name)
		for _, num := range rBPM2.FindAllString(fname, -1) {
			bpm, err = strconv.ParseFloat(num, 64)
			if err == nil && (bpm >= 100 && bpm <= 200 && math.Mod(bpm, 5) == 0) {
				break
			} else {
				err = fmt.Errorf("no bpm detected")
			}
		}
		if err != nil {
			return
		}
	} else {
		bpm, err = strconv.ParseFloat(foo[1], 64)
	}
	if err != nil {
		err = fmt.Errorf("could not parse bpm: %s", name)
		return
	}
	foo = rBeats.FindStringSubmatch(fname)
	if len(foo) > 1 {
		beats, _ = strconv.ParseFloat(foo[1], 64)
	}
	if beats == 0 {
		beats = math.Round(duration / (60 / bpm))
	}
	return
}

func guessBPM(fname string) (beats float64, bpm float64, err error) {
	duration, _, _, err := Length(fname)
	if err != nil {
		return
	}

	multiple := 2.0
	if os.Getenv("MULTIPLE") != "" {
		multiple, _ = strconv.ParseFloat(os.Getenv("MULTIPLE"), 64)
		if multiple == 0 {
			multiple = 2.0
		}
	}
	type guess struct {
		diff, bpm, beats float64
	}
	guesses := make([]guess, 80000)
	i := 0
	for beat := 1.0; beat <= 128; beat++ {
		for bp := 100.0; bp < 200; bp++ {
			guesses[i] = guess{math.Abs(duration - beat*multiple*60.0/bp), bp, beat * multiple}
			i++
		}
	}
	guesses = guesses[:i]

	sort.Slice(guesses, func(i, j int) bool {
		// Primary sort: by diff (smallest first)
		if guesses[i].diff != guesses[j].diff {
			return guesses[i].diff < guesses[j].diff
		}
		// Secondary sort: prefer beats that are powers of 2 (4, 8, 16, 32, 64, etc.)
		// Check if beats is a power of 2: log2(beats) should be an integer
		isPowerOfTwo := func(n float64) bool {
			if n < 1 {
				return false
			}
			log2 := math.Log2(n)
			return math.Abs(log2-math.Round(log2)) < 1e-9
		}
		iPower := isPowerOfTwo(guesses[i].beats)
		jPower := isPowerOfTwo(guesses[j].beats)
		if iPower != jPower {
			return iPower // true (power of 2) comes before false
		}
		// Tertiary sort: if both or neither are powers of 2, prefer smaller beats
		return guesses[i].beats < guesses[j].beats
	})

	beats = guesses[0].beats
	bpm = guesses[0].bpm

	return
}

// Length returns the duration of a WAV file in seconds, along with sample rate and total frames.
// For PCM data, it computes: (bytes / (bytesPerSample * channels)) / sampleRate.
// For non-PCM formats, it falls back to the decoder's Duration(), and returns sample rate and frames as 0.
func Length(filename string) (seconds float64, sampleRate int64, totalFrames int64, err error) {
	f, openErr := os.Open(filename)
	if openErr != nil {
		err = fmt.Errorf("open: %w", openErr)
		return
	}
	defer f.Close()

	d := wav.NewDecoder(f)
	if !d.IsValidFile() {
		err = fmt.Errorf("invalid WAV file")
		return
	}

	// Ensure format header is read.
	d.ReadInfo()

	// Non-PCM (compressed) WAVs: use the library's Duration().
	const wavFormatPCM = 1
	const wavFormatExtensible = 65534
	if int(d.WavAudioFormat) != wavFormatPCM && int(d.WavAudioFormat) != wavFormatExtensible {
		var dur time.Duration
		dur, err = d.Duration()
		if err != nil {
			err = fmt.Errorf("duration (non-PCM): %w", err)
			return
		}
		seconds = dur.Seconds()
		sampleRate = int64(d.SampleRate)
		totalFrames = int64(dur.Seconds() * float64(d.SampleRate))
		return
	}

	// PCM path: compute from data length, bit depth, channels, and sample rate.
	if d.SampleRate == 0 {
		err = fmt.Errorf("invalid sample rate: 0")
		return
	}
	bytesPerSample := int64(d.BitDepth) / 8
	if bytesPerSample <= 0 {
		err = fmt.Errorf("invalid bit depth: %d", d.BitDepth)
		return
	}
	chans := int64(d.NumChans)
	if chans <= 0 {
		err = fmt.Errorf("invalid channel count: %d", d.NumChans)
		return
	}

	// Make sure we're positioned to know the PCM chunk size.
	if !d.WasPCMAccessed() && d.PCMChunk == nil {
		if fwdErr := d.FwdToPCM(); fwdErr != nil {
			err = fmt.Errorf("locate PCM: %w", fwdErr)
			return
		}
	}

	totalBytes := d.PCMLen()
	if totalBytes <= 0 {
		err = fmt.Errorf("no PCM data")
		return
	}

	frameSize := bytesPerSample * chans // bytes per frame (all channels)
	if frameSize == 0 {
		err = fmt.Errorf("invalid frame size")
		return
	}

	totalFrames = totalBytes / frameSize // == total samples per channel
	seconds = float64(totalFrames) / float64(d.SampleRate)
	sampleRate = int64(d.SampleRate)
	return
}

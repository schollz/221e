package sox

import (
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// TempDir is where the temporary intermediate files are held
var TempDir = os.TempDir()

// TempPrefix is a unique indicator of the temporary files
var TempPrefix = "sox"

// TempType is the type of file to be generated (should be "wav")
var TempType = "wav"

var soxbinary = "sox"

var GarbageCollection = false

func Tmpfile() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(TempDir, TempPrefix+hex.EncodeToString(randBytes)+"."+TempType)
}

func Init() (err error) {
	soxbinary, err = getPath()
	if err != nil {
		return
	}
	stdout, _, _ := run(soxbinary, "--help")
	if !strings.Contains(stdout, "SoX") {
		soxbinary = "sox"
		if err != nil {
			return
		}
		stdout, _, _ := run(soxbinary, "--help")
		if !strings.Contains(stdout, "SoX") {
			err = fmt.Errorf("sox not found")
		}
	}
	return
}

func GetBinary() string {
	return soxbinary
}

func run(args ...string) (string, string, error) {
	baseCmd := args[0]
	cmdArgs := args[1:]
	cmd := exec.Command(baseCmd, cmdArgs...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%s%s", outb.String(), errb.String())
	}
	return outb.String(), errb.String(), err
}

func Version() (out string, err error) {
	out, _, err = run(soxbinary, "--version")
	out = strings.TrimSpace(out)
	foo := strings.Fields(out)
	out = foo[len(foo)-1]
	return
}

// MustString returns only the first argument of any function, as a string
func MustString(t ...interface{}) string {
	if len(t) > 0 {
		return t[0].(string)
	}
	return ""
}

// MustFloat returns only the first argument of any function, as a float
func MustFloat(t ...interface{}) float64 {
	if len(t) > 0 {
		return t[0].(float64)
	}
	return 0.0
}

// Clean will remove files created after each function
func Clean() (err error) {
	files, err := filepath.Glob(path.Join(TempDir, TempPrefix+"*."+TempType))
	if err != nil {
		return err
	}
	for _, fname := range files {
		err = os.Remove(fname)
		if err != nil {
			return
		}
	}
	return
}

// Left returns only the left channel
func Left(fname string) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "remix", "1")
	if err != nil {
		return
	}
	return
}

func Convert(fname string, fname2 string) (err error) {
	_, _, err = run(soxbinary, fname, fname2)
	return
}

func ConvertToMatch(fname string, fnameMatch string) (fname2 string, err error) {
	sampleRate, channels, _, err := Info(fnameMatch)
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, "-c", fmt.Sprint(channels), "-r", fmt.Sprint(sampleRate), fname2)
	if err != nil {
		return
	}
	return
}

func Warp(fname string, bpmSet float64, beatDivisionSet float64) (fname2 string, err error) {
	// first determine the BPM / Beats in the file
	beats, bpm, err := GetBPM(fname)
	if err != nil {
		return
	}
	_ = bpm
	_ = beats
	// determine the current duration
	duration, err := Length(fname)
	if err != nil {
		return
	}
	// round the beats to the nearest beat division
	beatsSet := math.Round(beats/beatDivisionSet) * beatDivisionSet
	if beatsSet == 0 {
		beatsSet = beatDivisionSet
	}
	// determine the new duration
	durationSet := float64(beatsSet) * 60 / bpmSet
	// stretch the audio to the new duration
	fname2, err = Stretch(fname, durationSet/duration)
	return
}

// ResampleRate changes the sample rate and precision
func ResampleRate(fname string, sampleRate int, precision int) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, "-r", fmt.Sprint(sampleRate), "-b", fmt.Sprint(precision), "--encoding", "signed-integer", "--endian", "little", fname2)
	if err != nil {
		return
	}
	return

}

func PCM16(fname string) (fname2 string, err error) {
	sr, c, _, err := Info(fname)
	if err != nil {
		return
	}
	rawfile := Tmpfile() + ".raw"
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, "-c", fmt.Sprint(c), "-r", fmt.Sprint(sr), "-b", fmt.Sprint(16), "--encoding", "signed-integer", "--endian", "little", rawfile)
	if err != nil {
		return
	}
	_, _, err = run(soxbinary, "-c", fmt.Sprint(c), "-r", fmt.Sprint(sr), "-b", fmt.Sprint(16), "--encoding", "signed-integer", "--endian", "little", rawfile, fname2)
	if err != nil {
		return
	}
	return
}

// NumSamples returns the number of samples in the file
func NumSamples(fname string) (numSamples int, err error) {
	stdout, _, err := run(soxbinary, "--i", fname)
	if err != nil {
		return
	}

	// Simplified regular expression pattern
	re := regexp.MustCompile(`(\d+)\s+samples`)

	// Find the matches in the input string
	matches := re.FindStringSubmatch(stdout)

	err = fmt.Errorf("could not find number of samples")
	if len(matches) == 2 {
		// Extract the number of samples from the regex match
		samplesStr := matches[1]
		numSamples, err = strconv.Atoi(samplesStr)
		return
	}
	return
}

// Info returns the sample rate and number of channels for file
func Info(fname string) (samplerate int, channels int, precision int, err error) {
	stdout, stderr, err := run(soxbinary, "--i", fname)
	if err != nil {
		return
	}
	stdout += stderr
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, "Channels") && channels == 0 {
			parts := strings.Fields(line)
			channels, err = strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				return
			}
		} else if strings.Contains(line, "Sample Rate") && samplerate == 0 {
			parts := strings.Fields(line)
			samplerate, err = strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				return
			}
		} else if strings.Contains(line, "Precision") && precision == 0 {
			parts := strings.Fields(line)
			precisionPart := parts[len(parts)-1]
			precisionPart2 := strings.Split(precisionPart, "-")
			precision, err = strconv.Atoi(precisionPart2[0])
			if err != nil {
				return
			}
		}
	}
	return
}

// Onsets requires aubioonset
func Onsets(fname string) (onsets []float64, err error) {
	stdout, stderr, err := run("aubioonset", "-i", fname)
	if err != nil {
		err = fmt.Errorf("err: %s: '%s'", err, stderr)
	}
	for _, line := range strings.Split(stdout, "\n") {
		f, err2 := strconv.ParseFloat(line, 64)
		if err2 == nil {
			onsets = append(onsets, f)
		}
	}
	return
}

// FadeIn will fade the audio in using quarter-sine wave
func FadeIn(fname string, duration float64) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "fade", "q", fmt.Sprint(duration))
	if err != nil {
		return
	}
	return
}

// FadeOut will fade the audio out using quarter-sine wave
func FadeOut(fname string, duration float64) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "fade", "q", "0", "-0", fmt.Sprint(duration))
	if err != nil {
		return
	}
	return
}

// Length returns the length of the file in seconds
func Length(fname string) (length float64, err error) {
	stdout, stderr, err := run(soxbinary, fname, "-n", "stat")
	if err != nil {
		return
	}
	stdout += stderr
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, "Length") {
			parts := strings.Fields(line)
			length, err = strconv.ParseFloat(parts[len(parts)-1], 64)
			return
		}
	}
	return
}

// SilenceAppend appends silence to a file
func SilenceAppend(fname string, length float64) (fname2 string, err error) {
	samplerate, channels, _, err := Info(fname)
	if err != nil {
		return
	}
	silencefile := Tmpfile()
	defer os.Remove(silencefile)
	fname2 = Tmpfile()
	// generate silence
	_, _, err = run(soxbinary, "-n", "-r", fmt.Sprint(samplerate), "-c", fmt.Sprint(channels), silencefile, "trim", "0.0", fmt.Sprint(length))
	if err != nil {
		return
	}
	// combine with original file
	_, _, err = run(soxbinary, fname, silencefile, fname2)
	if err != nil {
		return
	}
	if GarbageCollection {
		go func() {
			os.Remove(fname)
		}()
	}
	return
}

// SilencePrepend prepends silence to a file
func SilencePrepend(fname string, length float64) (fname2 string, err error) {
	samplerate, channels, _, err := Info(fname)
	if err != nil {
		return
	}
	silencefile := Tmpfile()
	defer os.Remove(silencefile)
	fname2 = Tmpfile()
	// generate silence
	_, _, err = run(soxbinary, "-n", "-r", fmt.Sprint(samplerate), "-c", fmt.Sprint(channels), silencefile, "trim", "0.0", fmt.Sprint(length))
	if err != nil {
		return
	}
	// combine with original file
	_, _, err = run(soxbinary, silencefile, fname, fname2)
	if err != nil {
		return
	}
	return
}

// FFT
func FFT(fname string) (data string, err error) {
	_, data, err = run(soxbinary, fname, "-n", "stat", "-freq")
	return
}

// Norm normalizes the audio
func Norm(fname string) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "norm")
	return
}

// SilenceTrim trims silence around a file
func SilenceTrim(fname string) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "silence", "1", "0.1", `-50d`, "reverse", "silence", "1", "0.1", `-50d`, "reverse")
	return
}

// SilenceTrimEnd trims silence from end of file
func SilenceTrimEnd(fname string) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "reverse", "silence", "1", "0.1", `-50d`, "reverse")
	return
}

// SilenceTrimFront trims silence from beginning of file
func SilenceTrimFront(fname string) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "silence", "1", "0.1", `-50d`, "reverse")
	return
}

// Trim will trim the audio from the start point (with optional length)
func Trim(fname string, start float64, length ...float64) (fname2 string, err error) {
	fname2 = Tmpfile()
	if len(length) > 0 {
		_, _, err = run(soxbinary, fname, fname2, "trim", fmt.Sprint(start), fmt.Sprint(length[0]))
	} else {
		_, _, err = run(soxbinary, fname, fname2, "trim", fmt.Sprint(start))
	}
	return
}

// Reverse will reverse the audio
func Reverse(fname string) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "reverse")
	return
}

// Pitch repitched the audio
func Pitch(fname string, notes int) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "pitch", fmt.Sprintf("%d", notes*100))
	return
}

// Join will concatonate the files
func Join(fnames ...string) (fname2 string, err error) {
	// match all the files
	for i := 1; i < len(fnames); i++ {
		fnames[i], err = ConvertToMatch(fnames[i], fnames[0])
		if err != nil {
			return
		}
	}
	fname2 = Tmpfile()
	fnames = append(fnames, fname2)
	_, _, err = run(append([]string{soxbinary}, fnames...)...)
	return
}

// Mix will mix the files
func Mix(fnames ...string) (fname2 string, err error) {
	fname2 = Tmpfile()
	fnames = append(fnames, fname2)
	fnames = append(fnames, "norm")
	_, _, err = run(append([]string{soxbinary, "-m"}, fnames...)...)
	return
}

func AddComment(fname string, comment string) (fname2 string, err error) {
	fname2 = Tmpfile() + ".aif"
	_, _, err = run(soxbinary, fname, "--comment", comment, fname2)
	return
}

func GetComment(fname string) (comment string, err error) {
	comment, _, err = run(soxbinary, "--i", "-a", fname)
	comment = strings.TrimSpace(comment)
	return
}

// Repeat will add n repeats to the audio
func Repeat(fname string, n int) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "repeat", fmt.Sprintf("%d", n))
	return
}

// RetempoSpeed will change the tempo of the audio and pitch
func RetempoSpeed(fname string, old_tempo float64, new_tempo float64) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "speed", fmt.Sprint(new_tempo/old_tempo), "rate", "-v", "48k")
	return
}

// RetempoStretch will change the tempo of the audio trying to keep pitch similar
func RetempoStretch(fname string, old_tempo float64, new_tempo float64) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "tempo", "-m", fmt.Sprint(new_tempo/old_tempo))
	return
}

// RetempoStretch will change the tempo of the audio trying to keep pitch similar
func Slowdown(fname string, slowdown float64) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "tempo", "-m", fmt.Sprint(slowdown))
	return
}

func Stereo(fname string) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "channels", "2")
	return
}

func CopyPaste(fname string, startPos float64, endPos float64, pastePos float64, crossfade float64, leeway0 ...float64) (fname2 string, err error) {
	copyLength := endPos - startPos
	if copyLength < 0.05 {
		fname2 = fname
		return
	}
	piece := Tmpfile()
	part1 := Tmpfile()
	part2 := Tmpfile()
	splice1 := Tmpfile()
	defer os.Remove(piece)
	defer os.Remove(part1)
	defer os.Remove(part2)
	defer os.Remove(splice1)
	fname2 = Tmpfile()
	leeway := 0.0
	if len(leeway0) > 0 {
		leeway = leeway0[0]
	}
	// 	os.cmd(string.format("sox %s %s trim %f %f",fname,piece,copy_start-e,copy_length+2*e))
	_, _, err = run(soxbinary, fname, piece, "trim", fmt.Sprint(startPos-crossfade), fmt.Sprint(copyLength+2*crossfade))
	if err != nil {
		fname2 = fname
		return
	}

	// 	os.cmd(string.format("sox %s %s trim 0 %f",fname,part1,paste_start+e))
	_, _, err = run(soxbinary, fname, part1, "trim", "0", fmt.Sprint(pastePos+crossfade))
	if err != nil {
		fname2 = fname
		return
	}

	// 	os.cmd(string.format("sox %s %s trim %f",fname,part2,paste_start+copy_length-e))
	_, _, err = run(soxbinary, fname, part2, "trim", fmt.Sprint(pastePos+copyLength-crossfade))
	if err != nil {
		fname2 = fname
		return
	}

	// 	os.cmd(string.format("sox %s %s %s splice %f,%f,%f",part1,piece,splice1,paste_start+e,e,l))
	_, _, err = run(soxbinary, part1, piece, splice1, "splice", fmt.Sprintf("%f,%f,%f", pastePos+crossfade, crossfade, leeway))
	if err != nil {
		fname2 = fname
		return
	}

	// 	os.cmd(string.format("sox %s %s %s splice %f,%f,%f",splice1,part2,fname2,paste_start+copy_length+e,e,l))
	_, _, err = run(soxbinary, splice1, part2, fname2, "splice", fmt.Sprintf("%f,%f,%f", pastePos+copyLength+crossfade, crossfade, leeway))
	if err != nil {
		fname2 = fname
		return
	}

	return
}

// Paste pastes any piece into a place in the audio, assumes that the piece has "crossfade" length on both sides
// in addition to its current length.
func Paste(fname string, piece string, pasteStart float64, crossfade float64) (fname2 string, err error) {
	copyLength, err := Length(piece)
	if err != nil {
		fname2 = fname
		return
	}
	part1 := Tmpfile()
	part2 := Tmpfile()
	splice1 := Tmpfile()
	defer os.Remove(part1)
	defer os.Remove(part2)
	defer os.Remove(splice1)
	fname2 = Tmpfile()
	leeway := 0.0

	// 	os.cmd(string.format("sox %s %s trim 0 %f",fname,part1,paste_start+e))
	_, _, err = run(soxbinary, fname, part1, "trim", "0", fmt.Sprint(pasteStart+crossfade))
	if err != nil {
		fname2 = fname
		return
	}
	// copy(part1, "1.wav")

	// 	os.cmd(string.format("sox %s %s trim %f",fname,part2,paste_start+copy_length-e*3))
	_, _, err = run(soxbinary, fname, part2, "trim", fmt.Sprint(pasteStart+copyLength-crossfade*3))
	if err != nil {
		fname2 = fname
		return
	}
	// copy(part2, "2.wav")

	// 	os.cmd(string.format("sox %s %s %s splice %f,%f,%f",part1,piece,splice1,paste_start+e,e,l))
	_, _, err = run(soxbinary, part1, piece, splice1, "splice", fmt.Sprintf("%f,%f,%f", pasteStart+crossfade, crossfade, leeway))
	if err != nil {
		fname2 = fname
		return
	}
	// copy(splice1, "3.wav")

	// 	os.cmd(string.format("sox %s %s %s splice %f,%f,%f",splice1,part2,fname2,paste_start+copy_length+e,e,l))
	_, _, err = run(soxbinary, splice1, part2, fname2, "splice", fmt.Sprintf("%f,%f,%f", pasteStart+copyLength+crossfade, crossfade, leeway))
	if err != nil {
		fname2 = fname
		return
	}
	// copy(fname2, "4.wav")

	return
}

// SampleRate changes the sample rate
func SampleRate(fname string, srCh ...int) (fname2 string, err error) {
	sampleRate := int(48000)
	if len(srCh) > 0 {
		sampleRate = srCh[0]
	}
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "rate", fmt.Sprint(sampleRate))
	return
}

// PostProcess
func PostProcess(fname string, gain float64) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "reverse", "silence", "1", "0.1", `0.5%`, "reverse", "gain", fmt.Sprint(gain))
	return
}

// Gain applies gain
func Gain(fname string, gain float64) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "gain", fmt.Sprint(gain))
	return
}

// Stretch does a time stretch
func Stretch(fname string, stretch float64) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "stretch", fmt.Sprint(stretch))
	return
}

func Reverb(fname string) (fname2 string, err error) {
	fname2 = Tmpfile()
	_, _, err = run(soxbinary, fname, fname2, "reverb", "70", "60", "90", "90", "1", "3")
	return
}

// ReverseReverb will reverse a piece, reverb it, reverse it back, and place
// it back in the position that the original note hits so it reverberates up to it.
func ReverseReverb(fname string, beat float64, beats float64) (fname2 string, err error) {
	totalBeats, bpm, err := GetBPM(fname)
	if err != nil {
		return
	}
	if beats >= totalBeats {
		err = fmt.Errorf("not enough beats")
	}
	if beat < beats {
		err = fmt.Errorf("not enough beats")
	}
	fname2, err = Gain(fname, -3)
	if err != nil {
		return
	}
	for _, beat := range []float64{7} {
		var piece string
		piece, err = Trim(fname2, beat*60/bpm, beats*60/bpm)
		if err != nil {
			break
		}
		piece, err = Reverse(piece)
		if err != nil {
			break
		}
		piece, err = SilenceAppend(piece, 60/bpm)
		if err != nil {
			break
		}
		piece, err = Reverb(piece)
		if err != nil {
			break
		}
		piece, err = Reverse(piece)
		if err != nil {
			break
		}
		fname2, err = Paste(fname2, piece, (beat-beats)*60/bpm, 0.02)
		if err != nil {
			break
		}
	}
	return
}

// Stutter does a stutter effect
func Stutter(fname string, stutter_length float64, pos_start float64, count float64, xfadePieceStutterGain ...float64) (fname2 string, err error) {
	crossfade_piece := 0.1
	crossfade_stutter := 0.005
	gain_amt := -2.0
	if count > 8 {
		gain_amt = -1.5
	}
	if len(xfadePieceStutterGain) > 0 {
		crossfade_piece = xfadePieceStutterGain[0]
	}
	if len(xfadePieceStutterGain) > 1 {
		crossfade_stutter = xfadePieceStutterGain[1]
	}
	if len(xfadePieceStutterGain) > 2 {
		gain_amt = xfadePieceStutterGain[2]
	}

	partFirst := Tmpfile()
	partMiddle := Tmpfile()
	partLast := Tmpfile()
	defer os.Remove(partFirst)
	defer os.Remove(partMiddle)
	defer os.Remove(partLast)

	// 	os.cmd(string.format("sox %s %s trim %f %f",fname,partFirst,pos_start-crossfade_piece,stutter_length+crossfade_piece+crossfade_stutter))
	_, _, err = run(soxbinary, fname, partFirst, "trim",
		fmt.Sprint(pos_start-crossfade_piece), fmt.Sprint(stutter_length+crossfade_piece+crossfade_stutter))
	if err != nil {
		fname2 = fname
		return
	}
	// 	os.cmd(string.format("sox %s %s trim %f %f",fname,partMiddle,pos_start-crossfade_stutter,stutter_length+crossfade_stutter+crossfade_stutter))
	_, _, err = run(soxbinary, fname, partMiddle, "trim", fmt.Sprint(pos_start-crossfade_stutter),
		fmt.Sprint(stutter_length+crossfade_stutter+crossfade_stutter))
	if err != nil {
		fname2 = fname
		return
	}
	// 	os.cmd(string.format("sox %s %s trim %f %f",fname,partLast,pos_start-crossfade_stutter,stutter_length+crossfade_piece+crossfade_stutter))
	_, _, err = run(soxbinary, fname, partLast, "trim", fmt.Sprint(pos_start-crossfade_stutter),
		fmt.Sprint(stutter_length+crossfade_piece+crossfade_stutter))
	if err != nil {
		fname2 = fname
		return
	}
	for i := 1.0; i <= count; i++ {
		fnameNext := ""
		if i == 1 {
			fnameNext, err = Gain(partFirst, gain_amt*(count-i))
			if err != nil {
				fname2 = fname
				return
			}
		} else {
			fnameNext = Tmpfile()
			fnameMid := partLast
			if i < count {
				fnameMid = partMiddle
			}
			if gain_amt != 0 {
				var foo string
				foo, err = Gain(fnameMid, gain_amt*(count-i))
				if err != nil {
					fname2 = fname
					return
				}
				fnameMid = foo
			}
			var fname2Length float64
			fname2Length, err = Length(fname2)
			if err != nil {
				fname2 = fname
				return
			}

			// os.cmd(string.format("sox %s %s %s splice %f,%f,0",fname2,fnameMid,fnameNext,audio.length(fname2),crossfade_stutter))
			_, _, err = run(soxbinary, fname2, fnameMid, fnameNext, "splice", fmt.Sprintf("%f,%f,0",
				fname2Length, crossfade_stutter))
			if err != nil {
				fname2 = fname
				return
			}
		}
		fname2 = fnameNext
	}
	return
}

// TrimBeats will take a filename with bpmX_beatsY.wav
// and trim it and make sure its the correct lenght
func TrimBeats(fname string) (fname2 string, tempo float64, beats float64, err error) {
	fname2, err = SilenceTrim(fname)
	if err != nil {
		return
	}
	_, bpm, err := GetBPM(fname)
	if err != nil {
		return
	}
	actualLength := MustFloat(Length(fname2))
	beats = actualLength / (60 / bpm)
	shouldLength := math.Round(beats) * (60 / bpm)
	if actualLength-shouldLength > 1e-4 {
		// trim
		var p1, p2 string

		p1, err = Trim(fname2, 0, shouldLength)
		if err != nil {
			return
		}

		p2, err = Trim(fname2, shouldLength)
		if err != nil {
			return
		}
		p2length := MustFloat(Length(p2))
		p1, err = FadeIn(p1, p2length)
		if err != nil {
			return
		}
		p2, err = FadeOut(p2, p2length)
		if err != nil {
			return
		}
		fname2, err = Mix(p1, p2)
		os.Remove(p1)
		os.Remove(p2)
	} else if shouldLength-actualLength > 1e-4 {
		// pad
		fname2, err = SilenceAppend(fname2, shouldLength-actualLength)
	}
	if err != nil {
		return
	}
	beats = math.Round(beats)
	tempo = bpm

	return
}

// 	for i=1,count do
// 		local fnameNext=""
// 		if i==1 then
// 			fnameNext=audio.gain(partFirst,gain_amt*(count-i))
// 		else
// 			fnameNext=string.random_filename()
//          local fnameMid=i<count and partMiddle or partLast
//          if gain_amt~=0 then
//            fnameMid=audio.gain(fnameMid,gain_amt*(count-i))
//          end
// 			os.cmd(string.format("sox %s %s %s splice %f,%f,0",fname2,fnameMid,fnameNext,audio.length(fname2),crossfade_stutter))
// 		end
// 		fname2=fnameNext
// 	end
// 	return fname2
// end

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// BPM guessing

func GetBPM(name string) (beats float64, bpm float64, err error) {
	beats, bpm, err = parseName(name)
	if err != nil {
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
	duration, err := Length(name)
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
	duration, err := Length(fname)
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
	guesses := make([]guess, 8000)
	i := 0
	for beat := 1.0; beat < 34; beat++ {
		for bp := 100.0; bp < 200; bp++ {
			guesses[i] = guess{math.Abs(duration - beat*multiple*60.0/bp), bp, beat * multiple}
			i++
		}
	}
	guesses = guesses[:i]

	sort.Slice(guesses, func(i, j int) bool {
		return guesses[i].diff < guesses[j].diff
	})

	beats = guesses[0].beats
	bpm = guesses[0].bpm

	return
}

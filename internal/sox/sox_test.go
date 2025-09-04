package sox

import (
	"os"
	"testing"
)

// func TestQuantize(t *testing.T) {
// 	fname := "amen_beats8_bpm172_2.wav"
// 	// fname = "TB_LOOPS_120BPM_2.wav"
// 	fname2, err := Quantize(fname)
// 	fmt.Println(fname2)
// 	os.Rename(fname2, "test.wav")
// 	assert.Nil(t, err)
// }

// func Quantize(fname string) (fname2 string, err error) {
// 	fnameTrimmed, err := SilenceTrim(fname)
// 	if err != nil {
// 		return
// 	}
// 	beats, bpm, err := GetBPM(fname)
// 	actual_length := MustFloat(Length(fnameTrimmed))
// 	fmt.Println("beats,bpm", beats, bpm)
// 	expected_length := beats * 60 / bpm
// 	fmt.Println("exp,act", expected_length, actual_length)
// 	if actual_length < expected_length {
// 		fnameTrimmed, err = SilenceAppend(fnameTrimmed, expected_length-actual_length)
// 		if err != nil {
// 			return
// 		}
// 	}

// 	fname2, err = PCM16(fnameTrimmed)
// 	if err != nil {
// 		return
// 	}
// 	stdout, _, err := run("aubioonset", fname2)
// 	onsets := []float64{}
// 	for _, line := range strings.Fields(stdout) {
// 		num, errnum := strconv.ParseFloat(line, 64)
// 		if errnum == nil {
// 			onsets = append(onsets, num)
// 		}
// 	}
// 	if len(onsets) == 1 {
// 		err = fmt.Errorf("not enough onsets")
// 		return
// 	}
// 	onsets = append(onsets, MustFloat(Length(fnameTrimmed)))
// 	bpm_estimates := make([]float64, len(onsets)-1)
// 	for i, _ := range bpm_estimates {
// 		bpm_estimates[i] = 60 / (onsets[i+1] - onsets[i])
// 		for {
// 			if bpm_estimates[i] >= 100 {
// 				break
// 			}
// 			bpm_estimates[i] = bpm_estimates[i] * 2
// 		}
// 		for {
// 			if bpm_estimates[i] < 200 {
// 				break
// 			}
// 			bpm_estimates[i] = bpm_estimates[i] / 2
// 		}
// 	}
// 	fmt.Println(onsets)
// 	fmt.Println(bpm_estimates)

//		pieces := make([]string, len(bpm_estimates))
//		for i, bpm_estimate := range bpm_estimates {
//			var piece string
//			piece, err = Trim(fname, onsets[i], onsets[i+1]-onsets[i])
//			if err != nil {
//				return
//			}
//			if math.Abs(bpm_estimate-bpm) < 10 {
//				piece, err = RetempoStretch(piece, bpm_estimate, bpm)
//				if err != nil {
//					return
//				}
//			}
//			pieces[i] = piece
//		}
//		fmt.Println(pieces)
//		fname2, err = Join(pieces...)
//		return
//	}
func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	err := Init()
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// func TestVersion(t *testing.T) {
// 	version, err := Version()
// 	assert.Nil(t, err)
// 	assert.Equal(t, "v14.4.2", version)
// }

// func TestComment(t *testing.T) {
// 	fname := "amen_beats8_bpm172.wav"
// 	fname2, err := AddComment(fname, "hello")
// 	assert.Nil(t, err)
// 	assert.Equal(t, "hello", MustString(GetComment(fname2)))
// }

// func TestTrimBeats(t *testing.T) {
// 	os.Remove("test.wav")
// 	fnames, _ := filepath.Glob("*.wav")
// 	for _, fname := range fnames {
// 		fname2, _, _, err := TrimBeats(fname)
// 		assert.Nil(t, err)
// 		os.Rename(fname2, "test.wav")
// 	}
// }

// func TestNumSamples(t *testing.T) {
// 	fname := "amen_beats8_bpm172.wav"
// 	numSamples, err := NumSamples(fname)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 123069, numSamples)
// }

// func TestOnsets(t *testing.T) {
// 	fname := "amen_beats8_bpm172.wav"
// 	onsets, err := Onsets(fname)
// 	_ = onsets
// 	assert.Nil(t, err)
// }

// func TestPCM16(t *testing.T) {
// 	fname := "amen_beats8_bpm172.wav"
// 	fname2, err := PCM16(fname)
// 	os.Rename(fname2, "test.wav")
// 	assert.Nil(t, err)
// }
// func TestPiko(t *testing.T) {
// 	fname := "amen_beats8_bpm172.wav"
// 	beats, bpm, err := GetBPM(fname)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 8.0, beats)
// 	assert.Equal(t, 172.0, bpm)
// }

// func TestReverseReverb(t *testing.T) {
// 	fname := "amen_beats8_bpm172.wav"
// 	fname2, err := ReverseReverb(fname, 7, 3)
// 	assert.Nil(t, err)
// 	if fname2 != fname {
// 		os.Rename(fname2, "test.wav")
// 	}
// 	Clean()
// }
// func TestRun(t *testing.T) {
// 	stdout, stderr, err := run(GetBinary(), "--help")
// 	assert.Nil(t, err)
// 	assert.True(t, strings.Contains(stdout, "SoX"))
// 	assert.Empty(t, stderr)
// }

// func TestLength(t *testing.T) {
// 	length, err := Length("amen_beats8_bpm172.wav")
// 	assert.Nil(t, err)
// 	assert.Equal(t, true, math.Abs(2.790499-length) < 0.1)
// }

// func TestInfo(t *testing.T) {
// 	samplerate, channnels, precision, err := Info("amen_beats8_bpm172.wav")
// 	assert.Nil(t, err)
// 	assert.Equal(t, 44100, samplerate)
// 	assert.Equal(t, 2, channnels)
// 	assert.Equal(t, 24, precision)
// }

// func TestWarp(t *testing.T) {
// 	fname2, err := Warp("amen_beats8_bpm172.wav", 120, 4)
// 	_ = fname2
// 	assert.Nil(t, err)
// }

// func TestSilence(t *testing.T) {
// 	fname2, err := SilenceAppend("amen_beats8_bpm172.wav", 1)
// 	assert.Nil(t, err)
// 	length1, _ := Length("amen_beats8_bpm172.wav")
// 	length2, _ := Length(fname2)
// 	assert.Less(t, math.Abs(length2-length1-1), 0.00001)

// 	fname2, err = SilencePrepend("amen_beats8_bpm172.wav", 1)
// 	assert.Nil(t, err)
// 	length1, _ = Length("amen_beats8_bpm172.wav")
// 	length2, _ = Length(fname2)
// 	assert.Less(t, math.Abs(length2-length1-1), 0.00001)

// 	fname3 := MustString(SilenceTrim(fname2))
// 	length3 := MustFloat(Length(fname3))
// 	assert.Greater(t, length2-length3, 1.0)

// 	os.Rename(fname3, "test.wav")
// }

// func TestTrim(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = Trim("amen_beats8_bpm172.wav", 0.5, 0.5)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 0.5, MustFloat(Length(fname2)))
// 	fname2, err = Trim("amen_beats8_bpm172.wav", 0.5)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 2.29068, MustFloat(Length(fname2)))
// }

// func TestReverse(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = Reverse("amen_beats8_bpm172.wav")
// 	assert.Nil(t, err)
// 	assert.Equal(t, MustFloat(Length("amen_beats8_bpm172.wav")), MustFloat(Length(fname2)))
// }

// func TestPitch(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = Pitch("amen_beats8_bpm172.wav", 3)
// 	assert.Nil(t, err)
// 	assert.Equal(t, MustFloat(Length("amen_beats8_bpm172.wav")), MustFloat(Length(fname2)))
// 	os.Rename(fname2, "test.wav")
// }

// func TestJoin(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = Join("amen_beats8_bpm172.wav", "amen_beats8_bpm172.wav", "amen_beats8_bpm172.wav")
// 	assert.Nil(t, err)
// 	assert.LessOrEqual(t, math.Abs(MustFloat(Length(fname2))-3*MustFloat(Length("amen_beats8_bpm172.wav"))), 0.001)
// 	os.Rename(fname2, "test.wav")
// }

// func TestRepeat(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = Repeat("amen_beats8_bpm172.wav", 2)
// 	assert.Nil(t, err)
// 	assert.LessOrEqual(t, math.Abs(MustFloat(Length(fname2))-3*MustFloat(Length("amen_beats8_bpm172.wav"))), 0.001)
// 	os.Rename(fname2, "test.wav")
// }

// func TestRetempoSpeed(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = RetempoSpeed("amen_beats8_bpm172.wav", 60, 120)
// 	assert.Nil(t, err)
// 	assert.LessOrEqual(t, math.Abs(MustFloat(Length("amen_beats8_bpm172.wav"))/2-MustFloat(Length(fname2))), 0.001)
// }
// func TestRetempoStretch(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = RetempoStretch("amen_beats8_bpm172.wav", 60, 120)
// 	assert.Nil(t, err)
// 	assert.LessOrEqual(t, math.Abs(MustFloat(Length("amen_beats8_bpm172.wav"))/2-MustFloat(Length(fname2))), 0.001)
// }

// func TestCopyPaste(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = CopyPaste("amen_beats8_bpm172.wav", 0.14, 0.27, 0.57, 0.02)
// 	assert.Nil(t, err)

// 	assert.Equal(t, true, math.Abs(MustFloat(Length("amen_beats8_bpm172.wav"))-MustFloat(Length(fname2))) < 0.01)
// 	os.Rename(fname2, "test.wav")
// }

// func TestPaste(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	crossfade := 0.04
// 	piece := MustString(Trim("amen_beats8_bpm172.wav", 0.14-crossfade, 0.27+crossfade))
// 	fname2, err = Paste("amen_beats8_bpm172.wav", piece, 0.57, crossfade)
// 	assert.Nil(t, err)
// 	assert.Equal(t, MustFloat(Length("amen_beats8_bpm172.wav")), MustFloat(Length(fname2)))
// 	os.Rename(fname2, "test.wav")
// }

// func TestGain(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = Gain("amen_beats8_bpm172.wav", 6)
// 	assert.Nil(t, err)
// 	assert.Equal(t, MustFloat(Length("amen_beats8_bpm172.wav")), MustFloat(Length(fname2)))
// 	os.Rename(fname2, "test.wav")
// }

// func TestSampleRate(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = SampleRate("amen_beats8_bpm172.wav", 8000)
// 	assert.Nil(t, err)
// 	assert.Less(t, math.Floor(MustFloat(Length("amen_beats8_bpm172.wav"))-MustFloat(Length(fname2))), 0.001)
// 	os.Rename(fname2, "test.wav")
// }

// func TestStretch(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname := "amen_beats8_bpm172.wav"

// 	fname2, err = Stretch(fname, 2)
// 	assert.Nil(t, err)
// 	assert.Less(t, math.Abs(MustFloat(Length(fname))*2-MustFloat(Length(fname2))), 0.01)
// 	os.Rename(fname2, "test.wav")
// }

// func TestStutter(t *testing.T) {
// 	var fname2 string
// 	var err error
// 	fname2, err = Stutter("amen_beats8_bpm172.wav", 60.0/160/4, 0.5, 4, 0.005)
// 	assert.Nil(t, err)
// 	if fname2 != "amen_beats8_bpm172.wav" {
// 		os.Rename(fname2, "test.wav")
// 	}
// }

// // keep this last
// func TestClean(t *testing.T) {
// 	assert.Nil(t, Clean())
// }

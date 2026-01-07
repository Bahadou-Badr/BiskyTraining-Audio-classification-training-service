package dataset

// "errors"

// "strconv"
// "strings"

func ValidateAudio(path string) error {
	// cmd := exec.Command(
	// 	"ffprobe",
	// 	"-v", "error",
	// 	"-show_entries", "format=duration",
	// 	"-show_entries", "stream=sample_rate",
	// 	"-of", "default=noprint_wrappers=1",
	// 	path,
	// )

	// var out bytes.Buffer
	// cmd.Stdout = &out

	// if err := cmd.Run(); err != nil {
	// 	return errors.New("invalid audio file")
	// }

	// output := out.String()

	// if !strings.Contains(output, "sample_rate=44100") {
	// 	return errors.New("sample rate must be 44100 Hz")
	// }

	// for _, line := range strings.Split(output, "\n") {
	// 	if strings.HasPrefix(line, "duration=") {
	// 		d, _ := strconv.ParseFloat(strings.TrimPrefix(line, "duration="), 64)
	// 		if d < 1.0 {
	// 			return errors.New("audio duration too short")
	// 		}
	// 	}
	// }

	return nil
}

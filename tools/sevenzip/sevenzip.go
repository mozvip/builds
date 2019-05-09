package sevenzip

import (
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	sevenZipExcutable = "c:\\Program Files\\7-Zip\\7z.exe"
	timeLayout = "2006-01-02 15:04:05"
)

type ArchiveEntry struct {
	dateTime time.Time
	Attr string
	Size uint64
	Compressed uint64
	Name string
}

func ReadEntries(archiveFile string) ([]ArchiveEntry, error) {
	combinedOutput, err := exec.Command(sevenZipExcutable, "l", archiveFile).CombinedOutput()
	var entries []ArchiveEntry
	if err == nil {
		lineRegExp := "(\\d[\\d\\-\\:\\s]+)\\s+([\\.\\w]+)\\s+(\\d+)\\s+(\\d+)?\\s+(.*)"
		re := regexp.MustCompile(lineRegExp)

		matches := re.FindAllStringSubmatch(string(combinedOutput), -1)
		for _, match := range matches[:len(matches)-1] {
			dateTime, _ := time.Parse(timeLayout, match[1])
			size, _ := strconv.ParseUint(match[3], 10, 64)
			compressed, _ := strconv.ParseUint(match[4], 10, 64)

			entry := ArchiveEntry{
				dateTime,
				match[2],
				size,
				compressed,
				strings.TrimSpace(match[5]),
			}

			entries = append(entries, entry)
		}
	}
	return entries, err
}

func Uncompress(archiveFile string, destinationFolder string) (err error){
	combinedOutputBytes, err := exec.Command(sevenZipExcutable, "x", archiveFile, "-o"+destinationFolder, "-y").CombinedOutput()
	if err != nil {
		log.Println(string(combinedOutputBytes))
	}
	return err
}
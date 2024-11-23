package bmfont

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

// Parses BMFont text font descriptor file into font
func ParseDescriptor(path string) (parsedFont *Font, err error) {
	// Open file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	parsedFont = &Font{}

	// Initialize scanner that reads line-by-line
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	// Scan font descriptor
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "info"):
			err = parsedFont.parseInfo(line)
		case strings.HasPrefix(line, "common"):
			err = parsedFont.parseCommon(line)
		case strings.HasPrefix(line, "page"):
			err = parsedFont.parsePage(line, filepath.Dir(path))
		case strings.HasPrefix(line, "chars"): // unused
			continue
		case strings.HasPrefix(line, "char"):
			err = parsedFont.parseChar(line)
		case strings.HasPrefix(line, "kernings"): // unused
			continue
		case strings.HasPrefix(line, "kerning"):
			err = parsedFont.parseKerning(line)
		}
		if err != nil {
			return parsedFont, err
		}
	}

	return parsedFont, scanner.Err()
}

// Parses info tag
func (f *Font) parseInfo(line string) error {
	log.Println("parsing info tag")
	faceField := false
	for _, field := range strings.Fields(line) {
		fv := strings.Split(field, "=")
		if len(fv) < 2 {
			if faceField {
				// This field is a continuation of the font name
				f.Name += " " + strings.Trim(field, "\"")
			}
			continue
		}

		switch fv[0] {
		case "face":
			f.Name = strings.Trim(fv[1], "\"")
			faceField = true
		case "size":
			size, err := strconv.Atoi(fv[1])
			if err != nil {
				return err
			}
			f.Size = uint32(size)
		default:
			faceField = false
		}
	}
	f.Name = strcase.ToScreamingSnake(f.Name)
	return nil
}

// Parses common tag
func (f *Font) parseCommon(line string) error {
	log.Println("parsing common tag")
	for _, field := range strings.Fields(line) {
		fv := strings.Split(field, "=")
		if len(fv) < 2 {
			continue
		}
		val, err := strconv.Atoi(fv[1])
		if err != nil {
			log.Printf("failed to parse value for field '%s': %s", fv[0], err)
			continue
		}
		switch fv[0] {
		case "lineHeight":
			f.LineHeight = uint32(val)
		case "base":
			f.Base = uint8(val)
		case "scaleW":
			f.ScaleW = uint16(val)
		case "scaleH":
			f.ScaleH = uint16(val)
		case "packed":
			f.IsPacked = uint8(val)
		case "alphaChnl":
			f.AType = ChannelTypeMapping[val]
		case "redChnl":
			f.RType = ChannelTypeMapping[val]
		case "greenChnl":
			f.GType = ChannelTypeMapping[val]
		case "blueChnl":
			f.BType = ChannelTypeMapping[val]
		}
	}
	return nil
}

// Parses page tag
func (f *Font) parsePage(line string, fileDir string) error {
	log.Println("parsing page tag")
	for _, field := range strings.Fields(line) {
		fv := strings.Split(field, "=")
		if len(fv) < 2 {
			continue
		}
		switch fv[0] {
		case "id":
			log.Printf("found page %s", fv[1])
		case "file":
			fileName := filepath.Join(fileDir, strings.Trim(fv[1], "\""))
			var newPage Page
			log.Printf("opening '%s'", fileName)
			info, err := os.Stat(fileName)
			if err != nil {
				return errors.Join(errors.New("failed to get file stat"), err)
			}
			newPage.PageSize = uint32(info.Size())
			file, err := os.Open(fileName)
			if err != nil {
				log.Fatalf("failed to open file: %s", err)
			}
			newPage.PageData = file
			f.Pages = append(f.Pages, newPage)
		}
	}
	return nil
}

// Parses char tag
func (f *Font) parseChar(line string) error {
	var newChar Char
	for _, field := range strings.Fields(line) {
		fv := strings.Split(field, "=")
		if len(fv) < 2 {
			continue
		}
		val, err := strconv.Atoi(fv[1])
		if err != nil {
			log.Printf("failed to parse value for field '%s': %s", fv[0], err)
			continue
		}
		switch fv[0] {
		case "id":
			newChar.ID = uint32(val)
		case "x":
			newChar.X = uint16(val)
		case "y":
			newChar.Y = uint16(val)
		case "width":
			newChar.Width = uint16(val)
		case "height":
			newChar.Height = uint16(val)
		case "xoffset":
			newChar.XOffset = int16(val)
		case "yoffset":
			newChar.YOffset = int16(val)
		case "xadvance":
			newChar.XAdvance = int16(val)
		case "page":
			newChar.Page = uint8(val)
		case "channels":
			newChar.Channels = uint8(val)
		}
	}
	f.insertChar(newChar)
	return nil
}

// Adds char to a bucket
func (f *Font) insertChar(newChar Char) {
	for i, bucket := range f.Buckets {
		if newChar.ID == bucket.EndChar+1 {
			f.Buckets[i].EndChar = newChar.ID
			f.Buckets[i].Chars = append(bucket.Chars, newChar)
			return
		}
	}
	f.Buckets = append(f.Buckets, Bucket{
		StartChar: newChar.ID,
		EndChar:   newChar.ID,
		Chars:     []Char{newChar},
	})
}

// Parses kerning tags and adds kerning data to the first char in pair
func (f *Font) parseKerning(line string) error {
	var first, second uint32
	var amount int16
	for _, field := range strings.Fields(line) {
		fv := strings.Split(field, "=")
		if len(fv) < 2 {
			continue
		}
		val, err := strconv.Atoi(fv[1])
		if err != nil {
			log.Printf("failed to parse value for field '%s': %s", fv[0], err)
			continue
		}
		switch fv[0] {
		case "first":
			first = uint32(val)
		case "second":
			second = uint32(val)
		case "amount":
			amount = int16(val)
		}
	}
	if amount == 0 {
		return nil
	}

	// Find the first char bucket and insert kerning pair
	for _, bucket := range f.Buckets {
		if first >= bucket.StartChar && first <= bucket.EndChar {
			if bucket.Chars[first-bucket.StartChar].Kernings == nil {
				bucket.Chars[first-bucket.StartChar].Kernings = make(map[uint32]int16)
			}
			bucket.Chars[first-bucket.StartChar].Kernings[second] = amount
		}
	}
	return nil
}

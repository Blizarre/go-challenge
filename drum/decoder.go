package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

/*
	Decode a line from the file
*/
func readLine(p *Pattern, file *os.File) (int64, DrumLine, error) {
	var textSize int8
	var text []byte
	var err error
	var nbReadByte int64
	var line DrumLine

	err = readDataLE(file, &line.id, err)
	err = readDataBE(file, &textSize, err)
	nbReadByte += 4 + 1

	log.Println("id: ", line.id, " textsize: ", textSize, " err: ", err)

	if err != nil {
		return nbReadByte, line, err
	}

	text = make([]byte, textSize)
	err = readDataBE(file, text, err)
	err = readDataBE(file, &line.activate, err)
	nbReadByte += int64(textSize) + 16

	line.name = string(text)
	log.Println("Line name: ", line.name)

	return nbReadByte, line, err
}

/*
	Decode the header of the splice, and return the number of bytes remaining
	in the splice or an error.
*/
func decodeHeader(p *Pattern, file *os.File) (int64, error) {
	magicNumber := []byte{'S', 'P', 'L', 'I', 'C', 'E'}
	var head struct {
		Magic     [6]byte
		TotalSize int64
		Version   [32]byte
	}
	var err error

	err = readDataBE(file, &head, nil)
	err = readDataLE(file, &p.tempo, err)

	if err == nil {

		for i, r := range head.Magic {
			if magicNumber[i] != r {
				return 0, errors.New("Non conformant magic number")
			}
		}

		p.version = string(head.Version[:strings.Index(string(head.Version[:]), "\x00")])
		log.Println("Version:", p.version)
		log.Println("Total data Size:", head.TotalSize)
		log.Println("Tempo:", p.tempo)
	}

	return head.TotalSize - int64(len(head.Version)) - 4, err
}

func readDataLE(f *os.File, data interface{}, err error) error {
	if err != nil {
		return err
	}

	return binary.Read(f, binary.LittleEndian, data)
}

func readDataBE(f *os.File, data interface{}, err error) error {
	if err != nil {
		return err
	}

	return binary.Read(f, binary.BigEndian, data)
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Cannot open file: " + path + "\n error: " + err.Error())
	}
	defer file.Close()

	remainingBytes, err := decodeHeader(p, file)
	if err != nil {
		log.Println("Error during header decoding:", err.Error())
		return nil, err
	}

	if err == nil {
		for err == nil && remainingBytes > 0 {
			var consumedBytes int64
			var line DrumLine
			consumedBytes, line, err = readLine(p, file)
			p.data = append(p.data, line)
			remainingBytes -= consumedBytes
		}
	}

	return p, err
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	data    []DrumLine
}

type DrumLine struct {
	id       int32
	name     string
	activate [16]byte
}

func drawBeats(beatLine [16]byte) (ret string) {
	ret = "|"
	for i, b := range beatLine {
		if i%4 == 0 && i > 0 {
			ret += "|"
		}
		if b == 0 {
			ret += "-"
		} else {
			ret += "x"
		}
	}
	ret += "|"
	return
}

func (p Pattern) String() (repr string) {
	repr = "Saved with HW Version: " + p.version + "\n"
	repr += "Tempo: " + fmt.Sprintf("%g", p.tempo) + "\n"
	for _, d := range p.data {
		repr += fmt.Sprintf("(%d) %s\t", d.id, d.name)
		repr += drawBeats(d.activate) + "\n"
	}
	return
}

package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

// Decode a measure from the file, and return the number of byte read from the file or an error
func decodeMeasure(file *os.File) (nbReadByte int64, m Measure, err error) {
	var textSize int8
	var text []byte
	var steps [16]int8

	err = readDataLE(file, &m.id, err)
	err = readDataLE(file, &textSize, err)
	nbReadByte += 4 + 1

	if err != nil {
		return nbReadByte, m, err
	}

	text = make([]byte, textSize)
	err = readDataLE(file, text, err)
	err = readDataLE(file, &steps, err)

	for i := 0; i < len(steps); i++ {
		m.steps[i] = (steps[i] == 0)
	}

	nbReadByte += int64(textSize) + 16
	m.name = string(text)

	return nbReadByte, m, err
}

// Decode the header of the splice, and return the number of bytes remaining
// in the splice or an error.
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

// Read data from the file in Little Endian. It is a wrapper from binary.read
// If the input err != nil, nothing is read from the file and err is returned.
// otherwise, the error value for binary.read is returned
func readDataLE(f *os.File, data interface{}, err error) error {
	if err != nil {
		return err
	}

	return binary.Read(f, binary.LittleEndian, data)
}

// Read data from the file in Big Endian. It is a wrapper from binary.read
// If the input err != nil, nothing is read from the file and err is returned.
// otherwise, the error value for binary.read is returned
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
			var line Measure
			consumedBytes, line, err = decodeMeasure(file)
			p.measures = append(p.measures, line)
			remainingBytes -= consumedBytes
		}
	}

	return p, err
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version  string
	tempo    float32
	measures []Measure
}

// Measure is the high level representation of a measure.
type Measure struct {
	id    int32
	name  string
	steps [16]bool
}

func (d Measure) String() (repr string) {
	repr = fmt.Sprintf("(%d) %s\t|", d.id, d.name)

	for i, b := range d.steps {
		if i%4 == 0 && i > 0 {
			repr += "|"
		}
		if b {
			repr += "-"
		} else {
			repr += "x"
		}
	}
	repr += "|"
	return
}

func (p Pattern) String() (repr string) {
	repr = "Saved with HW Version: " + p.version + "\n"
	repr += "Tempo: " + fmt.Sprintf("%g", p.tempo) + "\n"
	for _, d := range p.measures {
		repr += fmt.Sprint(d) + "\n"
	}
	return
}

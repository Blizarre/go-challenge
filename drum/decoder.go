package drum

import (
	"encoding/binary"
	"errors"
	"log"
	"os"
)

/*
	Decode a line from the file
*/
func readLine(p *Pattern, file *os.File) (int64, error) {
	var id int32
	var textSize int8
	var err error
	var nbReadByte int64

	err = readData(file, &id, err)
	err = readData(file, &textSize, err)

	if err != nil {
		return 0, err
	}

	text := make([]byte, textSize)
	err = readData(file, text, err)

	return nbReadByte, err
}

/*
	Decode the header of the splice, and return the number of bytes remaining
	in the splice or an error.
*/
func decodeHeader(p *Pattern, file *os.File) (int64, error) {
	const versionLength = 32
	var err error
	var totalSize int64

	version := make([]byte, versionLength)

	spliceMagicNumber := []byte{'S', 'P', 'L', 'I', 'C', 'E'}
	magicNumber := make([]byte, len(spliceMagicNumber))

	err = nil
	err = readData(file, magicNumber, err)
	err = readData(file, &totalSize, err)
	err = readData(file, version, err)

	if err != nil {

		for i, r := range spliceMagicNumber {
			if magicNumber[i] != r {
				return 0, errors.New("Non conformant magic number")
			}
		}

		p.version = string(version)
		log.Println("Version:", p.version)
		log.Println("Total data Size:", totalSize)
	}

	return totalSize - versionLength, err
}

func readData(f *os.File, data interface{}, err error) error {
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

	var unknown int16
	var tempo int8
	var beat byte

	err = binary.Read(file, binary.BigEndian, &unknown)
	err = binary.Read(file, binary.BigEndian, &tempo)

	err = binary.Read(file, binary.BigEndian, &beat)

	if err != nil {

		if beat != 'B' {
			log.Println("Beat doens not have the proper value")
			return nil, errors.New("Invalid Beat")
		}

		log.Println("Unknown value %x", unknown)
		log.Println("Tempo value %x", tempo)
		log.Println("Beat value ", beat)
		p.tempo = tempo >> 2

		for err != nil && remainingBytes > 0 {
			var consumedBytes int64
			consumedBytes, err = readLine(p, file)
			remainingBytes -= consumedBytes
		}
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   int8
	data    []struct {
		time   int
		typeOf string
		data   int
	}
}

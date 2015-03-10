package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version  string
	Tempo    float32
	Measures []Measure
}

// Measure is the high level representation of a measure.
type Measure struct {
	Id    int32
	Name  string
	Steps [16]bool
}

func (d Measure) String() (repr string) {
	repr = fmt.Sprintf("(%d) %s\t|", d.Id, d.Name)

	for i, b := range d.Steps {
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
	repr = "Saved with HW Version: " + p.Version + "\n"
	repr += "Tempo: " + fmt.Sprintf("%g", p.Tempo) + "\n"
	for _, d := range p.Measures {
		repr += fmt.Sprint(d) + "\n"
	}
	return
}

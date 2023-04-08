package XML

import "github.com/beevik/etree"

type Reader struct {
	doc *etree.Document
}

func (x *Reader) ReadFromFile(filename string) error {
	x.doc = etree.NewDocument()
	return x.doc.ReadFromFile(filename)
}

func (x *Reader) FindElement(path string) *etree.Element {
	return x.doc.FindElement(path)
}

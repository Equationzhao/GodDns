package XML

import "github.com/beevik/etree"

type XMLReader struct {
	doc *etree.Document
}

func (x *XMLReader) ReadFromFile(filename string) error {
	x.doc = etree.NewDocument()
	return x.doc.ReadFromFile(filename)
}

func (x *XMLReader) FindElement(path string) *etree.Element {
	return x.doc.FindElement(path)
}

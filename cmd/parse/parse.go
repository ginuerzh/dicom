package main

import (
	"encoding/json"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"strconv"

	"github.com/ginuerzh/dicom"
	"github.com/ginuerzh/dicom/pkg/tag"
)

type Element struct {
	VR     string        `json:"vr"`
	Length uint32        `json:"length"`
	Values []interface{} `json:"values"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ds, err := dicom.ParseFile("cr.dcm", nil)
	if err != nil {
		log.Fatal(err)
	}

	addElement(&ds)

	data, err := encodeDataSet(ds)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
	/*
		pde, err := ds.FindElementByTag(tag.PixelData)
		if err != nil {
			log.Fatal(err)
		}
		extractPixelData(dicom.MustGetPixelDataInfo(pde.Value))
	*/

	for _, e := range ds.Elements {
		if err := verifyValueType(e.Tag, e.Value, e.RawValueRepresentation); err != nil {
			log.Fatal(err)
		}

	}

	f, err := os.Create("cr.export.dcm")
	if err != nil {
		log.Fatal(err)
	}
	if err := dicom.Write(f, ds, dicom.SkipVRVerification()); err != nil {
		log.Println(err)
	}

	f.Close()
}

func encodeDataSet(ds dicom.Dataset) ([]byte, error) {
	m := encodeElement(ds.Elements)
	return json.MarshalIndent(m, "", "  ")
}

func encodeElement(elements []*dicom.Element) map[string]*Element {
	em := make(map[string]*Element)

	for _, e := range elements {
		tag := fmt.Sprintf("%04x,%04x", e.Tag.Group, e.Tag.Element)
		el := &Element{
			VR:     e.RawValueRepresentation,
			Length: e.ValueLength,
		}

		switch e.Value.ValueType() {
		case dicom.Strings:
			for _, v := range dicom.MustGetStrings(e.Value) {
				el.Values = append(el.Values, v)
			}
		case dicom.Bytes:
			el.Values = append(el.Values, fmt.Sprintf("%v", dicom.MustGetBytes(e.Value)))
		case dicom.Ints:
			for _, v := range dicom.MustGetInts(e.Value) {
				el.Values = append(el.Values, strconv.FormatInt(int64(v), 10))
			}
		case dicom.Floats:
			for _, v := range dicom.MustGetFloats(e.Value) {
				el.Values = append(el.Values, strconv.FormatFloat(v, 'f', -1, 64))
			}
		case dicom.PixelData:
			pdi := dicom.MustGetPixelDataInfo(e.Value)
			log.Println("find pixelData:", pdi.IsEncapsulated, len(pdi.Frames))
		case dicom.Sequences:
			for _, item := range e.Value.GetValue().([]*dicom.SequenceItemValue) {
				el.Values = append(el.Values, encodeElement(item.GetValue().([]*dicom.Element)))
			}
		}
		em[tag] = el
	}

	return em
}

func extractPixelData(info dicom.PixelDataInfo) {
	for i, fr := range info.Frames {
		img, _ := fr.GetImage()
		f, _ := os.Create(fmt.Sprintf("image_%d.jpg", i))
		jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
		f.Close()
	}
}

func addElement(ds *dicom.Dataset) {
	t := tag.Tag{Group: 0x0015, Element: 0x0001}
	v, _ := dicom.NewValue([]string{"AppEntity"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "AE",
		ValueRepresentation:    tag.GetVRKind(t, "AE"),
		ValueLength:            10,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0002}
	v, _ = dicom.NewValue([]string{"018M"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "AS",
		ValueRepresentation:    tag.GetVRKind(t, "AS"),
		ValueLength:            4,
		Value:                  v,
	})

	// TODO: parse AT
	t = tag.Tag{Group: 0x0015, Element: 0x0003}
	v, _ = dicom.NewValue([]string{"1234"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "AT",
		ValueRepresentation:    tag.GetVRKind(t, "AT"),
		ValueLength:            4,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0004}
	v, _ = dicom.NewValue([]string{"CS_VR", "TEST"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "CS",
		ValueRepresentation:    tag.GetVRKind(t, "CS"),
		ValueLength:            10,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0005}
	v, _ = dicom.NewValue([]string{"20201112"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "DA",
		ValueRepresentation:    tag.GetVRKind(t, "DA"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0006}
	v, _ = dicom.NewValue([]string{"123.456"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "DS",
		ValueRepresentation:    tag.GetVRKind(t, "DS"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0007}
	v, _ = dicom.NewValue([]string{"20201112142400.999999+0800"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "DT",
		ValueRepresentation:    tag.GetVRKind(t, "DT"),
		ValueLength:            16,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0008}
	v, _ = dicom.NewValue([]float64{123.456})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "FL",
		ValueRepresentation:    tag.GetVRKind(t, "FL"),
		ValueLength:            4,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0009}
	v, _ = dicom.NewValue([]float64{123.456789})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "FD",
		ValueRepresentation:    tag.GetVRKind(t, "FD"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0010}
	v, _ = dicom.NewValue([]string{"-123456"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "IS",
		ValueRepresentation:    tag.GetVRKind(t, "IS"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0011}
	v, _ = dicom.NewValue([]string{"-LO123456"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "LO",
		ValueRepresentation:    tag.GetVRKind(t, "LO"),
		ValueLength:            10,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0012}
	v, _ = dicom.NewValue([]string{"+LT abcdefg 1234567"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "LT",
		ValueRepresentation:    tag.GetVRKind(t, "LT"),
		ValueLength:            20,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0013}
	v, _ = dicom.NewValue([]byte{'O', 'B', 'T', 'E', 'S', 'T', ','})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "OB",
		ValueRepresentation:    tag.GetVRKind(t, "OB"),
		ValueLength:            8,
		Value:                  v,
	})

	// TODO: parse OD
	t = tag.Tag{Group: 0x0015, Element: 0x0014}
	v, _ = dicom.NewValue([]string{"0123456789123456"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "OD",
		ValueRepresentation:    tag.GetVRKind(t, "OD"),
		ValueLength:            16,
		Value:                  v,
	})

	// TODO: parse OF
	t = tag.Tag{Group: 0x0015, Element: 0x0015}
	v, _ = dicom.NewValue([]string{"12345678"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "OF",
		ValueRepresentation:    tag.GetVRKind(t, "OF"),
		ValueLength:            8,
		Value:                  v,
	})

	// TODO: parse OL
	t = tag.Tag{Group: 0x0015, Element: 0x0016}
	v, _ = dicom.NewValue([]string{"ffffffff"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "OL",
		ValueRepresentation:    tag.GetVRKind(t, "OL"),
		ValueLength:            8,
		Value:                  v,
	})

	// TODO: parse OV
	t = tag.Tag{Group: 0x0015, Element: 0x0017}
	v, _ = dicom.NewValue([]string{"ffffffffffffffff"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "OV",
		ValueRepresentation:    tag.GetVRKind(t, "OV"),
		ValueLength:            16,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0018}
	v, _ = dicom.NewValue([]byte{0x12, 0x34, 0x56, 0x78})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "OW",
		ValueRepresentation:    tag.GetVRKind(t, "OW"),
		ValueLength:            4,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0019}
	v, _ = dicom.NewValue([]string{"PN TEST"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "PN",
		ValueRepresentation:    tag.GetVRKind(t, "PN"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0020}
	v, _ = dicom.NewValue([]string{"SH TEST"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "SH",
		ValueRepresentation:    tag.GetVRKind(t, "SH"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0021}
	v, _ = dicom.NewValue([]int{-0x12345678})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "SL",
		ValueRepresentation:    tag.GetVRKind(t, "SL"),
		ValueLength:            4,
		Value:                  v,
	})

	/*
		t = tag.Tag{Group: 0x0015, Element: 0x0022}
		v, _ = dicom.NewValue([]int{-0x12345678})
		ds.Elements = append(ds.Elements, &dicom.Element{
			Tag:                    t,
			RawValueRepresentation: "SQ",
			ValueRepresentation:    tag.GetVRKind(t, "SQ"),
			ValueLength:            4,
			Value:                  v,
		})
	*/

	t = tag.Tag{Group: 0x0015, Element: 0x0023}
	v, _ = dicom.NewValue([]int{-0x1234})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "SS",
		ValueRepresentation:    tag.GetVRKind(t, "SS"),
		ValueLength:            2,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0024}
	v, _ = dicom.NewValue([]string{"ST TEST"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "ST",
		ValueRepresentation:    tag.GetVRKind(t, "ST"),
		ValueLength:            8,
		Value:                  v,
	})

	// TODO: parse SV
	t = tag.Tag{Group: 0x0015, Element: 0x0025}
	v, _ = dicom.NewValue([]string{"12345678"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "SV",
		ValueRepresentation:    tag.GetVRKind(t, "SV"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0026}
	v, _ = dicom.NewValue([]string{"160102.999999"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "TM",
		ValueRepresentation:    tag.GetVRKind(t, "TM"),
		ValueLength:            14,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0027}
	v, _ = dicom.NewValue([]string{"UC TEST"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "UC",
		ValueRepresentation:    tag.GetVRKind(t, "UC"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0028}
	v, _ = dicom.NewValue([]string{"1.2.3.4.5.6.7.8"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "UI",
		ValueRepresentation:    tag.GetVRKind(t, "UI"),
		ValueLength:            16,
		Value:                  v,
	})

	// TODO: parse UL
	t = tag.Tag{Group: 0x0015, Element: 0x0029}
	v, _ = dicom.NewValue([]int{0xffffffff})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "UL",
		ValueRepresentation:    tag.GetVRKind(t, "UL"),
		ValueLength:            4,
		Value:                  v,
	})

	/*
		UN
	*/

	t = tag.Tag{Group: 0x0015, Element: 0x0030}
	v, _ = dicom.NewValue([]string{"https://www.dadax.cn"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "UR",
		ValueRepresentation:    tag.GetVRKind(t, "UR"),
		ValueLength:            20,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0032}
	v, _ = dicom.NewValue([]int{0xffff})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "US",
		ValueRepresentation:    tag.GetVRKind(t, "US"),
		ValueLength:            2,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0033}
	v, _ = dicom.NewValue([]string{"UT TEST"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "UT",
		ValueRepresentation:    tag.GetVRKind(t, "UT"),
		ValueLength:            8,
		Value:                  v,
	})

	// TODO: parse UV
	t = tag.Tag{Group: 0x0015, Element: 0x0034}
	v, _ = dicom.NewValue([]string{"12345678"})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "UV",
		ValueRepresentation:    tag.GetVRKind(t, "UV"),
		ValueLength:            8,
		Value:                  v,
	})
}

func verifyValueType(t tag.Tag, value dicom.Value, vr string) error {
	valueType := value.ValueType()
	var ok bool
	switch vr {
	case "US", "UL", "SL", "SS":
		ok = valueType == dicom.Ints
	case "SQ":
		ok = valueType == dicom.Sequences
	case "NA":
		ok = valueType == dicom.SequenceItem
	case "OW", "OB":
		if t == tag.PixelData {
			ok = valueType == dicom.PixelData
		} else {
			ok = valueType == dicom.Bytes
		}
	case "FL", "FD":
		ok = valueType == dicom.Floats
	case "AT":
		fallthrough
	default:
		ok = valueType == dicom.Strings
	}

	if !ok {
		return fmt.Errorf("%s, %d, ValueType does not match the specified type in the VR", vr, valueType)
	}
	return nil
}

package main

import (
	"encoding/base64"
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
	Name   string        `json:"name"`
	Length uint32        `json:"length"`
	Values []interface{} `json:"values"`
}

type Frame struct {
	Native         bool  `json:"native"`
	Rows           int   `json:"rows"`
	Cols           int   `json:"cols"`
	SamplePerPixel int   `json:"sample_per_pixel"`
	BitsPerSample  int   `json:"bits_per_sample"`
	Size           int64 `json:"size"`
}

var (
	dict = map[tag.Tag]tag.TagInfo{
		{Group: 0x0015, Element: 0x0001}: {VR: "AE", Name: "0015,0001", VM: "1"},
		{Group: 0x0015, Element: 0x0002}: {VR: "AS", Name: "0015,0002", VM: "1"},
		{Group: 0x0015, Element: 0x0003}: {VR: "AT", Name: "0015,0003", VM: "1"},
		{Group: 0x0015, Element: 0x0004}: {VR: "CS", Name: "0015,0004", VM: "1"},
		{Group: 0x0015, Element: 0x0005}: {VR: "DA", Name: "0015,0005", VM: "1"},
		{Group: 0x0015, Element: 0x0006}: {VR: "DS", Name: "0015,0006", VM: "1"},
		{Group: 0x0015, Element: 0x0007}: {VR: "DT", Name: "0015,0007", VM: "1"},
		{Group: 0x0015, Element: 0x0008}: {VR: "FL", Name: "0015,0008", VM: "1"},
		{Group: 0x0015, Element: 0x0009}: {VR: "FD", Name: "0015,0009", VM: "1"},
		{Group: 0x0015, Element: 0x0010}: {VR: "IS", Name: "0015,0010", VM: "1"},
		{Group: 0x0015, Element: 0x0011}: {VR: "LO", Name: "0015,0011", VM: "1"},
		{Group: 0x0015, Element: 0x0012}: {VR: "LT", Name: "0015,0012", VM: "1"},
		{Group: 0x0015, Element: 0x0013}: {VR: "OB", Name: "0015,0013", VM: "1"},
		{Group: 0x0015, Element: 0x0014}: {VR: "OD", Name: "0015,0014", VM: "1"},
		{Group: 0x0015, Element: 0x0015}: {VR: "OF", Name: "0015,0015", VM: "1"},
		{Group: 0x0015, Element: 0x0016}: {VR: "OL", Name: "0015,0016", VM: "1"},
		{Group: 0x0015, Element: 0x0017}: {VR: "OV", Name: "0015,0017", VM: "1"},
		{Group: 0x0015, Element: 0x0018}: {VR: "OW", Name: "0015,0018", VM: "1"},
		{Group: 0x0015, Element: 0x0019}: {VR: "PN", Name: "0015,0019", VM: "1"},
		{Group: 0x0015, Element: 0x0020}: {VR: "SH", Name: "0015,0020", VM: "1"},
		{Group: 0x0015, Element: 0x0021}: {VR: "SL", Name: "0015,0021", VM: "1"},
		{Group: 0x0015, Element: 0x0022}: {VR: "SQ", Name: "0015,0022", VM: "1"},
		{Group: 0x0015, Element: 0x0023}: {VR: "SS", Name: "0015,0023", VM: "1"},
		{Group: 0x0015, Element: 0x0024}: {VR: "ST", Name: "0015,0024", VM: "1"},
		{Group: 0x0015, Element: 0x0025}: {VR: "SV", Name: "0015,0025", VM: "1"},
		{Group: 0x0015, Element: 0x0026}: {VR: "TM", Name: "0015,0026", VM: "1"},
		{Group: 0x0015, Element: 0x0027}: {VR: "UC", Name: "0015,0027", VM: "1"},
		{Group: 0x0015, Element: 0x0028}: {VR: "UI", Name: "0015,0028", VM: "1"},
		{Group: 0x0015, Element: 0x0029}: {VR: "UL", Name: "0015,0029", VM: "1"},
		{Group: 0x0015, Element: 0x0030}: {VR: "UR", Name: "0015,0030", VM: "1"},
		{Group: 0x0015, Element: 0x0031}: {VR: "UN", Name: "0015,0031", VM: "1"},
		{Group: 0x0015, Element: 0x0032}: {VR: "US", Name: "0015,0032", VM: "1"},
		{Group: 0x0015, Element: 0x0033}: {VR: "UT", Name: "0015,0033", VM: "1"},
		{Group: 0x0015, Element: 0x0034}: {VR: "UV", Name: "0015,0034", VM: "1"},
	}
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	tag.SetCustomDict(dict)

	fname := "raw-YBR_FULL_422"
	ds, err := dicom.ParseFile(fname+".dcm", nil)
	if err != nil {
		log.Fatal(err)
	}

	// addElement(&ds)

	data, err := encodeDataSet(ds)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	pde, err := ds.FindElementByTag(tag.PixelData)
	if err != nil {
		log.Fatal(err)
	}
	extractPixelData(dicom.MustGetPixelDataInfo(pde.Value))

	f, err := os.Create(fname + ".export.dcm")
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
		info, err := tag.Find(e.Tag)
		if err != nil {
			log.Println("find Tag:", e.Tag, err)
		}

		el := &Element{
			VR:     e.RawValueRepresentation,
			Name:   info.Name,
			Length: e.ValueLength,
		}
		// log.Println("element:", e.Tag, e.RawValueRepresentation, e.Value)
		switch e.Value.ValueType() {
		case dicom.Strings:
			for _, v := range dicom.MustGetStrings(e.Value) {
				el.Values = append(el.Values, v)
			}
		case dicom.Bytes:
			el.Values = append(el.Values, base64.StdEncoding.EncodeToString(dicom.MustGetBytes(e.Value)))
		case dicom.Ints:
			for _, v := range dicom.MustGetInts(e.Value) {
				el.Values = append(el.Values, strconv.FormatInt(int64(v), 10))
			}
		case dicom.UInts:
			for _, v := range dicom.MustGetUInts(e.Value) {
				el.Values = append(el.Values, strconv.FormatUint(uint64(v), 10))
			}
		case dicom.Floats:
			for _, v := range dicom.MustGetFloats(e.Value) {
				el.Values = append(el.Values, strconv.FormatFloat(v, 'f', -1, 64))
			}
		case dicom.PixelData:
			pdi := dicom.MustGetPixelDataInfo(e.Value)
			for _, frame := range pdi.Frames {
				fr := Frame{}
				if pdi.IsEncapsulated {
					fr.Size = int64(len(frame.EncapsulatedData.Data))
				} else {
					fr.Native = true
					fr.Cols = frame.NativeData.Cols
					fr.Rows = frame.NativeData.Rows
					fr.SamplePerPixel = frame.NativeData.SamplesPerPixel
					fr.BitsPerSample = frame.NativeData.BitsPerSample
					fr.Size = int64(len(frame.NativeData.Data))
				}
				el.Values = append(el.Values, fr)
			}
		case dicom.Sequences:
			for _, item := range e.Value.GetValue().([]*dicom.SequenceItemValue) {
				el.Values = append(el.Values, encodeElement(item.GetValue().([]*dicom.Element)))
			}
		}

		t := fmt.Sprintf("%04x,%04x", e.Tag.Group, e.Tag.Element)
		em[t] = el
	}

	return em
}

func extractPixelData(info dicom.PixelDataInfo) {
	for i, fr := range info.Frames {
		img, err := fr.GetImage()
		if err != nil {
			log.Println(err)
			return
		}
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

	t = tag.Tag{Group: 0x0015, Element: 0x0003}
	v, _ = dicom.NewValue([]uint64{0x12345678})
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
	v, _ = dicom.NewValue([]float64{-123.456})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "FL",
		ValueRepresentation:    tag.GetVRKind(t, "FL"),
		ValueLength:            4,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0009}
	v, _ = dicom.NewValue([]float64{-123.456789})
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

	t = tag.Tag{Group: 0x0015, Element: 0x0014}
	v, _ = dicom.NewValue([]float64{-1234.56789, 1234.56789})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "OD",
		ValueRepresentation:    tag.GetVRKind(t, "OD"),
		ValueLength:            16,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0015}
	v, _ = dicom.NewValue([]float64{-1234.5678, 1234.5678})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "OF",
		ValueRepresentation:    tag.GetVRKind(t, "OF"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0016}
	v, _ = dicom.NewValue([]uint64{0x12345678, 0x87654321, 0xFFFFFFFF})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "OL",
		ValueRepresentation:    tag.GetVRKind(t, "OL"),
		ValueLength:            8,
		Value:                  v,
	})

	t = tag.Tag{Group: 0x0015, Element: 0x0017}
	v, _ = dicom.NewValue([]uint64{0x12345678ABCDEF, 0xFFFFFFFFFFFFFFFF})
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
	v, _ = dicom.NewValue([]int64{-1})
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
	v, _ = dicom.NewValue([]int64{-1})
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

	t = tag.Tag{Group: 0x0015, Element: 0x0025}
	v, _ = dicom.NewValue([]int64{-1})
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

	t = tag.Tag{Group: 0x0015, Element: 0x0029}
	v, _ = dicom.NewValue([]uint64{0xffffffff})
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
	v, _ = dicom.NewValue([]uint64{0xffff})
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

	t = tag.Tag{Group: 0x0015, Element: 0x0034}
	v, _ = dicom.NewValue([]uint64{0xFFFFFFFFFFFFFFFF})
	ds.Elements = append(ds.Elements, &dicom.Element{
		Tag:                    t,
		RawValueRepresentation: "UV",
		ValueRepresentation:    tag.GetVRKind(t, "UV"),
		ValueLength:            8,
		Value:                  v,
	})
}

package cmd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Trimmer struct {
	dec *xml.Decoder
}

func (tr Trimmer) Token() (xml.Token, error) {
	t, err := tr.dec.Token()
	if cd, ok := t.(xml.CharData); ok {
		t = xml.CharData(bytes.TrimSpace(cd))
	}
	return t, err
}

type Aircraft struct {
	Type               string         `xml:"type,attr"`
	Engine             string         `xml:"engine,attr"`
	Wake               string         `xml:"wake,attr"`
	Mass               Mass           `xml:"mass"`
	FlightEnvelope     FlightEnvelope `xml:"flightEnvelope"`
	Aerodynamics       Aerodynamics   `xml:"aerodynamics"`
	EngineThrust       EngineThrust   `xml:"engineThrust"`
	Fuel               Fuel           `xml:"fuel"`
	MMCClimbs          MediumMass     `xml:"mediumMassClimbs"`
	MediumMassClimbs   []FlightLevelProperties
	MMDescents         MediumMass `xml:"mediumMassDescents"`
	MediumMassDescents []FlightLevelProperties

	APF APF `xml:"apf"`
}

func (a *Aircraft) adjust() {
	a.MediumMassClimbs = make([]FlightLevelProperties, 0)
	for _, v := range a.MMCClimbs.FLs {
		v.adjustFL()
		a.MediumMassClimbs = append(a.MediumMassClimbs, v.Field)
	}
	a.MediumMassDescents = make([]FlightLevelProperties, 0)
	for _, v := range a.MMDescents.FLs {
		v.adjustFL()
		a.MediumMassDescents = append(a.MediumMassDescents, v.Field)
	}
}
func (a *Aircraft) toMap() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	m["type"] = a.Type
	m["engine"] = a.Engine
	m["wake"] = a.Wake
	m["mass"] = a.Mass
	m["flightenvelope"] = a.FlightEnvelope
	m["aerodynamics"] = a.Aerodynamics
	m["enginethrust"] = a.EngineThrust
	m["fuel"] = a.Fuel
	m["mediummassclimbs"] = a.MediumMassClimbs
	m["mediummassdescents"] = a.MediumMassDescents
	m["apf"] = a.APF
	return m
}

func (a *Aircraft) persist() error {
	fName := fmt.Sprintf("%s.yaml", a.Type)

	if f, errCr := os.OpenFile(fName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600); errCr == nil {
		defer f.Close()
		enc := yaml.NewEncoder(f)
		m := a.toMap()
		if encErr := enc.Encode(m); encErr != nil {
			Logger.Info("Marshalling error", zap.String("file", fName), zap.Error(encErr))
			return encErr
		}

	} else {
		Logger.Fatal("failed to create file", zap.String("name", fName))
		return errors.New("failed to create file")
	}
	return nil
}

type Mass struct {
	Reference float64 `xml:"reference,attr"`
	Minimum   float64 `xml:"minimum,attr"`
	Maximum   float64 `xml:"maximum,attr"`
	Maxload   float64 `xml:"maxload,attr"`
	Gw        float64 `xml:"Gw,attr"`
}
type FlightEnvelope struct {
	VMO             float64 `xml:"VMO,attr"`
	MMO             float64 `xml:"MMO,attr"`
	MaxAlt          float64 `xml:"maxAlt,attr"`
	HMax            float64 `xml:"h_max,attr"`
	Gt              float64 `xml:"Gt,attr"`
	CrossAltClimb   float64 `xml:"crossAltClimb,attr"`
	CrossAltCruise  float64 `xml:"crossAltCruise,attr"`
	CrossAltDescent float64 `xml:"crossAltDescent,attr"`
}

type EngineThrust struct {
	CTc1        float64 `xml:"CTc1,attr"`
	CTc2        float64 `xml:"CTc2,attr"`
	CTc3        float64 `xml:"CTc3,attr"`
	CTc4        float64 `xml:"CTc4,attr"`
	CTc5        float64 `xml:"CTc5,attr"`
	CTdescLow   float64 `xml:"CTdesc_low,attr"`
	CTdescHigh  float64 `xml:"CTdesc_high,attr"`
	CTdescLevel float64 `xml:"CTdesc_level,attr"`
	CTdescApp   float64 `xml:"CTdesc_app,attr"`
	CTdescLd    float64 `xml:"CTdesc_ld,attr"`
}
type APF struct {
	Vcl1  float32 `xml:"Vcl1,attr"`
	Vcl2  float32 `xml:"Vcl2,attr"`
	Vcr1  float32 `xml:"Vcr1,attr"`
	Vcr2  float32 `xml:"Vcr2,attr"`
	Mcr   float32 `xml:"Mcr,attr"`
	Vdes1 float32 `xml:"Vdes1,attr"`
	Vdes2 float32 `xml:"Vdes2,attr"`
	Mdes  float32 `xml:"Mdes,attr"`
}
type Fuel struct {
	Cf1  float64 `xml:"Cf1,attr"`
	Cf2  float64 `xml:"Cf2,attr"`
	Cf3  float64 `xml:"Cf3,attr"`
	Cf4  float64 `xml:"Cf4,attr"`
	Cfcr float64 `xml:"Cfcr,attr"`
}
type FlightLevelProperties struct {
	FlightLevel  int
	Temperature  int     `xml:"T,attr"`
	Pressure     int     `xml:"P,attr"`
	Rho          float32 `xml:"rho,attr"`
	SpeedOfSound float64 `xml:"a,attr"`
	TAS          float64 `xml:"TAS,attr"`
	CAS          float64 `xml:"CAS,attr"`
	Mach         float32 `xml:"M,attr"`
	Mass         float64 `xml:"mass,attr"`
	Thrust       float64 `xml:"Thrust,attr"`
	Drag         float64 `xml:"Drag,attr"`
	Fuel         float64 `xml:"Fuel,attr"`
	ESF          float32 `xml:"esf,attr"`
	ROC          float32 `xml:"ROC,attr"`
	TDC          int     `xml:"TDC,attr"`
	PWC          float32 `xml:"PWC,attr"`
	GammaTAS     float64 `xml:"gammaTAS,attr"`
}

type FlightLevel struct {
	Value int                   `xml:"value,attr"`
	Field FlightLevelProperties `xml:"field"`
}

//	type FlightLevelDescent struct {
//		Value int                          `xml:"value,attr"`
//		Field FlightLevelPropertiesDescent `xml:"field"`
//	}
type MediumMass struct {
	FLs          []FlightLevel `xml:"FL,omitempty"`
	FlightLevels []FlightLevelProperties
}

type FLAdjuster interface {
	adjustFL()
}

func (c *MediumMass) adjust(p FlightLevelProperties) {
	c.FlightLevels = append(c.FlightLevels, p)
	c.FLs = nil
}

func (f *FlightLevel) adjustFL() {
	f.Field.FlightLevel = f.Value
}

type PerformanceData struct {
	Aircrafts []Aircraft `xml:"aircraft"`
}
type Configuration struct {
	N      int     `xml:"n,attr"`
	Phase  string  `xml:"phase,attr"`
	Name   string  `xml:"name,attr"`
	Vstall float64 `xml:"Vstall,attr"`
	CD0    float64 `xml:"CD0,attr"`
	CD2    float64 `xml:"CD2,attr"`
	Gear   Gear    `xml:"gear"`
}
type Gear struct {
	CD0DeltaDG float64 `xml:"CD0_deltaLDG,attr"`
}
type Aerodynamics struct {
	WingSurf      float64         `xml:"wingSurf,attr"`
	Configuration []Configuration `xml:"configuration"`
}

var (
	xmlFile    string
	xmlContent PerformanceData
	xmlBytes   []byte
	readXmlErr error
)

var formatTransformerCmd = &cobra.Command{
	Use:   "formatTransformer",
	Short: "The utility to read xml file then dump to an xml file",
	Long: `The utility to read xml file then dump to an xml file
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if xmlBytes, readXmlErr = os.ReadFile(xmlFile); readXmlErr != nil {

			Logger.Fatal("error while reading xml", zap.String("path", xmlFile), zap.Error(readXmlErr))
			return readXmlErr
		} else {

			Logger.Info("Xml file is read", zap.String("path", xmlFile), zap.Int("size", len(xmlBytes)))
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		xmlContent = PerformanceData{}
		r := bytes.NewReader(xmlBytes)
		raw := xml.NewDecoder(r)
		dec := xml.NewTokenDecoder(Trimmer{raw})

		if errDecodeErr := dec.Decode(&xmlContent); errDecodeErr != nil {
			Logger.Fatal("error while decoding xml", zap.String("path", xmlFile), zap.Error(errDecodeErr))
			return errDecodeErr
		}
		Logger.Info("unmarshal is completed", zap.Int("total aircraft", len(xmlContent.Aircrafts)))

		return nil
	},
	PostRunE: func(cmd *cobra.Command, args []string) error {
		for _, v := range xmlContent.Aircrafts {
			Logger.Info("content", zap.String("aircraft type", v.Type))
			v.adjust()
			err := v.persist()
			if err != nil {
				return err
			}

		}
		return nil
	},
}

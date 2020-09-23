package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"os"
)

// Command to pretty print XML in linux: xmllint --format as.xml

type NmapRun struct {
	Jid       string    `json:"jid"`
	Args      string    `xml:"args,attr" json:"args"`
	Start     int       `xml:"start,attr" json:"start"`
	StartStr  string    `xml:"startstr,attr" json:"startstr"`
	Version   string    `xml:"version,attr" json:"version"`
	ScanInfo  ScanInfo  `xml:"scaninfo" json:"scaninfo"`
	Verbose   Verbose   `xml:"verbose" json:"verbose"`
	Debugging Debugging `xml:"debugging" json:"debugging"`
	Hosts     []Host    `xml:"host" json:"hosts"`
	RunStats  RunStats  `xml:"runstats" json:"runstats"`
}

type ScanInfo struct {
	Type        string `xml:"type,attr" json:"type"`
	Protocol    string `xml:"protocol,attr" json:"protocol"`
	NumServices int    `xml:"numservices,attr" json:"numservices"`
	Services    string `xml:"services,attr" json:"services"`
}

type Verbose struct {
	Level int `xml:"level,attr" json:"level"`
}

type Debugging struct {
	Level int `xml:"level,attr" json:"level"`
}

type RunStats struct {
	Finished Finished  `xml:"finished" json:"finished"`
	Hosts    HostStats `xml:"hosts" json:"hosts"`
}

type Finished struct {
	Time     int     `xml:"time,attr" json:"time"`
	TimeStr  string  `xml:"timestr,attr" json:"timestr"`
	Elapsed  float32 `xml:"elapsed,attr" json:"elapsed"`
	Summary  string  `xml:"summary,attr" json:"summary"`
	Exit     string  `xml:"exit,attr" json:"exit"`
	ErrorMsg string  `xml:"errormsg,attr" json:"errormsg"`
}

type HostStats struct {
	Up    int `xml:"up,attr" json:"up"`
	Down  int `xml:"down,attr" json:"down"`
	Total int `xml:"total,attr" json:"total"`
}

type Host struct {
	StartTime     int           `xml:"starttime,attr" json:"starttime"`
	EndTime       int           `xml:"endtime,attr" json:"endtime"`
	Status        Status        `xml:"status" json:"status"`
	Addresses     []Address     `xml:"address" json:"addresses"`
	Hostnames     []Hostname    `xml:"hostnames>hostname" json:"hostnames"`
	Ports         []Port        `xml:"ports>port" json:"ports"`
	Os            Os            `xml:"os" json:"os"`
	TcpSequence   TcpSequence   `xml:"tcpsequence" json:"tcpsequence"`
	IpIdSequence  IpIdSequence  `xml:"ipidsequence" json:"ipidsequence"`
	TcpTsSequence TcpTsSequence `xml:"tcptssequence" json:"tcptssequence"`
	Times         Times         `xml:"times" json:"times"`
}

type Status struct {
	State     string  `xml:"state,attr" json:"state"`
	Reason    string  `xml:"reason,attr" json:"reason"`
	ReasonTTL float32 `xml:"reason_ttl,attr" json:"reason_ttl"`
}

type Os struct {
	PortsUsed      []PortUsed      `xml:"portused" json:"portsused"`
	OsMatches      []OsMatch       `xml:"osmatch" json:"osmatches"`
	OsFingerprints []OsFingerprint `xml:"osfingerprint" json:"osfingerprints"`
}

type PortUsed struct {
	State  string `xml:"state,attr" json:"state"`
	Proto  string `xml:"proto,attr" json:"proto"`
	PortId int    `xml:"portid,attr" json:"portid"`
}

type OsMatch struct {
	Name      string    `xml:"name,attr" json:"name"`
	Accuracy  string    `xml:"accuracy,attr" json:"accuracy"`
	Line      string    `xml:"line,attr" json:"line"`
	OsClasses []OsClass `xml:"osclass" json:"osclasses"`
}

type OsClass struct {
	Vendor   string   `xml:"vendor,attr" json:"vendor"`
	OsGen    string   `xml"osgen,attr"`
	Type     string   `xml:"type,attr" json:"type"`
	Accuracy string   `xml:"accurancy,attr" json:"accurancy"`
	OsFamily string   `xml:"osfamily,attr" json:"osfamily"`
	CPEs     []string `xml:"cpe" json:"cpes"`
}

type OsFingerprint struct {
	Fingerprint string `xml:"fingerprint,attr" json:"fingerprint"`
}

type Uptime struct {
	Seconds  int    `xml:"seconds,attr" json:"seconds"`
	Lastboot string `xml:"lastboot,attr" json:"lastboot"`
}

type TcpSequence struct {
	Index      int    `xml:"index,attr" json:"index"`
	Difficulty string `xml:"difficulty,attr" json:"difficulty"`
	Values     string `xml:"vaules,attr" json:"vaules"`
}

type Sequence struct {
	Class  string `xml:"class,attr" json:"class"`
	Values string `xml:"values,attr" json:"values"`
}
type IpIdSequence Sequence
type TcpTsSequence Sequence

type Times struct {
	SRTT string `xml:"srtt,attr" json:"srtt"`
	RTT  string `xml:"rttvar,attr" json:"rttv"`
	To   string `xml:"to,attr" json:"to"`
}

type Address struct {
	Addr     string `xml:"addr,attr" json:"addr"`
	AddrType string `xml:"addrtype,attr" json:"addrtype"`
	Vendor   string `xml:"vendor,attr" json:"vendor"`
}

type Hostname struct {
	Name string `xml:"name,attr" json:"name"`
	Type string `xml:"type,attr" json:"type"`
}

type Port struct {
	Protocol string   `xml:"protocol,attr" json:"protocol"`
	PortId   int      `xml:"portid,attr" json:"id"`
	State    State    `xml:"state" json:"state"`
	Service  Service  `xml:"service" json:"service"`
	Scripts  []Script `xml:"script" json:"scripts"`
}

type State struct {
	State     string  `xml:"state,attr" json:"state"`
	Reason    string  `xml:"reason,attr" json:"reason"`
	ReasonTTL float32 `xml:"reason_ttl,attr" json:"reason_ttl"`
}

type Service struct {
	Name      string   `xml:"name,attr" json:"name"`
	Method    string   `xml:"method,attr" json:"method"`
	Conf      int      `xml:"conf,attr" json:"conf"`
	Version   string   `xml:"version,attr" json:"version"`
	Product   string   `xml:"product,attr" json:"product"`
	ExtraInfo string   `xml:"extrainfo,attr" json:"extrainfo"`
	Tunnel    string   `xml:"tunnel,attr" json:"tunnel"`
	ServiceFp string   `xml:"servicefp,attr" json:"servicefp"`
	CPEs      []string `xml:"cpe" json:"cpes"`
}

type Script struct {
	Id       string    `xml:"id,attr" json:"id"`
	Output   string    `xml:"output,attr" json:"output"`
	Tables   []Table   `xml:"table" json:"tables"`
	Elements []Element `xml:"elem" json:"elements"`
}

type Table struct {
	Key      string    `xml:"key,attr" json:"key"`
	Elements []Element `xml:"elem" json:"elements"`
	Table    []Table   `xml:"table" json:"tables"`
}

type Element struct {
	Key   string `xml:"key,attr" json:"key"`
	Value string `xml:",chardata" json:"value"`
}

var nmapStruct = NmapRun{}

func nmapJSON(path string, jid string) ([]byte, error) {
	byteArr, err := readFile(path)
	if err != nil {
		return nil, err
	}
	toStruct(byteArr, jid)
	result, err := toJSON(nmapStruct)
	return result, err
}

func readFile(path string) ([]byte, error) {
	// Open XML file
	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// Close file after read them to array
	defer xmlFile.Close()
	// Read opened XML file as a byte array
	byteArr, _ := ioutil.ReadAll(xmlFile)
	return byteArr, nil
}

func toStruct(byteArr []byte, jid string) {
	xml.Unmarshal(byteArr, &nmapStruct)
	nmapStruct.Jid = jid
}

func toJSON(nmapStruct NmapRun) ([]byte, error) {
	nmapJSON, err := json.Marshal(nmapStruct)
	if err != nil {
		return nil, err
	}
	return nmapJSON, nil
}

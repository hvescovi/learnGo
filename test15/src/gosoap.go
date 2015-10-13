package main

import (
    "bytes"
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "net/http"
)

// The URL of the SOAP server
const MH_SOAP_URL = "http://sp.mountyhall.com/SP_WebService.php"

// this is just the message I'll send for interrogation, with placeholders
//  for my parameters
const SOAP_VUE_QUERY_FORMAT = "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?><SOAP-ENV:Envelope xmlns:SOAP-ENV=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:tns=\"urn:SP_WebService\" xmlns:soap=\"http://schemas.xmlsoap.org/wsdl/soap/\" xmlns:wsdl=\"http://schemas.xmlsoap.org/wsdl/\" xmlns:SOAP-ENC=\"http://schemas.xmlsoap.org/soap/encoding/\" ><SOAP-ENV:Body><mns:Vue xmlns:mns=\"uri:mhSp\" SOAP-ENV:encodingStyle=\"http://schemas.xmlsoap.org/soap/encoding/\"><numero xsi:type=\"xsd:string\">%d</numero><mdp xsi:type=\"xsd:string\">%s</mdp></mns:Vue></SOAP-ENV:Body></SOAP-ENV:Envelope>"

// Here I define Go structures, almost identical to the structure of the
// XML message we'll fetch
// Note that annotations (the string "return>item") allow to have a slightly
//  different structure or different namings

type SoapItem struct {
    Numero    int
    Nom       string
    Type      string
    PositionX int
    PositionY int
    PositionN int
    Monde     int
}
type SoapVue struct {
    Items []SoapItem "return>item"
}
type SoapFault struct {
    Faultstring string
    Detail      string
}
type SoapBody struct {
    Fault          SoapFault
    ProfilResponse SoapProfil
    VueResponse    SoapVue
}
type SoapEnvelope struct {
    XMLName xml.Name
    Body    SoapBody
}

// Here is the function querying the SOAP server
// It returns the whole answer as a Go structure (a SoapEnvelope)
// You could also return an error in a second returned parameter
func GetSoapEnvelope(query string, numero int, mdp string) (envelope *SoapEnvelope) {
    soapRequestContent := fmt.Sprintf(query, numero, mdp)
    httpClient := new(http.Client)
    resp, err := httpClient.Post(MH_SOAP_URL, "text/xml; charset=utf-8", bytes.NewBufferString(soapRequestContent))
    if err != nil {
        // handle error
    }
    b, e := ioutil.ReadAll(resp.Body) // probably not efficient, done because the stream isn't always a pure XML stream and I have to fix things (not shown here)
    if e != nil {
        // handle error
    }
    in := string(b)
    parser := xml.NewDecoder(bytes.NewBufferString(in))
    envelope = new(SoapEnvelope) // this allocates the structure in which we'll decode the XML
    err = parser.DecodeElement(&envelope, nil)
    if err != nil {
        // handle error
    }
    resp.Body.Close()
    return
}
